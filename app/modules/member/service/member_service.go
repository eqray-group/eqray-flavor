package service

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

// 积分业务类型，与 member_point_record.biz_type 对应（参考 yudao）。
const (
	MemberPointBizOther  int8 = 0
	MemberPointBizAdjust int8 = 10
	// MemberPointBizSign 签到获得的积分
	MemberPointBizSign int8 = 20
)

// MemberSaveInput 是新增 / 编辑会员的入参，由 Controller 从请求体转换而来。
type MemberSaveInput struct {
	Nickname   string
	Mobile     string
	Avatar     string
	Name       string
	Sex        int8
	Birthday   *types.DateTime
	AreaID     int
	LevelID    types.ID
	Point      int
	Experience int
	TagIDs     string
	Mark       string
	GroupID    types.ID
	// Status 为空表示调用方没有传状态：新增按启用处理，编辑保持原状态
	Status *int8
}

// MemberListQuery 是会员列表的查询条件。
type MemberListQuery struct {
	Nickname string
	Mobile   string
	LevelID  types.ID
	GroupID  types.ID
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// MemberService 负责会员档案的业务规则。
type MemberService struct {
	members *repository.MemberRepository
	levels  *repository.MemberLevelRepository
	points  *repository.MemberPointRecordRepository
}

// NewMemberService 构造会员档案服务，装配会员、等级与积分流水仓储。
func NewMemberService() *MemberService {
	return &MemberService{
		members: repository.NewMemberRepository(),
		levels:  repository.NewMemberLevelRepository(),
		points:  repository.NewMemberPointRecordRepository(),
	}
}

// Page 分页查询会员。
func (s *MemberService) Page(ctx context.Context, q MemberListQuery) ([]model.MemberUserModel, int64, error) {
	return s.members.Paginate(ctx, repository.MemberFilter{
		Nickname: strings.TrimSpace(q.Nickname),
		Mobile:   strings.TrimSpace(q.Mobile),
		LevelID:  q.LevelID,
		GroupID:  q.GroupID,
		Status:   q.Status,
	}, q.PageNo, q.PageSize)
}

// Detail 查询会员详情，不存在时返回业务错误码。
func (s *MemberService) Detail(ctx context.Context, id types.ID) (*model.MemberUserModel, error) {
	m, err := s.members.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(memberErrs.MemberNotExists, id)
	}

	return m, nil
}

// EnsureActive 校验会员存在且为启用状态，供 C 端只读接口复用。
func (s *MemberService) EnsureActive(ctx context.Context, id types.ID) (*model.MemberUserModel, error) {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.Status != baseEnums.CommonStatusEnabled {
		return nil, baseErrs.New(memberErrs.MemberDisabled)
	}

	return m, nil
}

// Create 新增会员：手机号唯一，等级必须存在且启用；未传状态时默认启用。
func (s *MemberService) Create(ctx context.Context, in MemberSaveInput) (types.ID, error) {
	if err := normalizeMemberInput(&in); err != nil {
		return 0, err
	}
	if err := s.ensureLevel(ctx, in.LevelID); err != nil {
		return 0, err
	}

	exists, err := s.members.ExistsByMobile(ctx, in.Mobile, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(memberErrs.MemberMobileExists, in.Mobile)
	}

	m := &model.MemberUserModel{
		Mobile:     in.Mobile,
		Nickname:   in.Nickname,
		Avatar:     in.Avatar,
		Name:       in.Name,
		Sex:        in.Sex,
		Birthday:   in.Birthday,
		AreaID:     in.AreaID,
		LevelID:    in.LevelID,
		Point:      in.Point,
		Experience: in.Experience,
		TagIDs:     in.TagIDs,
		Mark:       in.Mark,
		GroupID:    in.GroupID,
		Status:     memberStatusOrDefault(in.Status),
	}
	if err := s.members.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 更新会员档案；手机号变更时排除自身做唯一校验。
func (s *MemberService) Update(ctx context.Context, id types.ID, in MemberSaveInput) error {
	if err := normalizeMemberInput(&in); err != nil {
		return err
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	// 等级未变更时不校验启用状态：等级被停用后，存量会员仍可正常编辑其它字段
	if in.LevelID != m.LevelID {
		if err := s.ensureLevel(ctx, in.LevelID); err != nil {
			return err
		}
	}

	exists, err := s.members.ExistsByMobile(ctx, in.Mobile, id)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(memberErrs.MemberMobileExists, in.Mobile)
	}

	m.Nickname = in.Nickname
	m.Mobile = in.Mobile
	m.Avatar = in.Avatar
	m.Name = in.Name
	m.Sex = in.Sex
	m.Birthday = in.Birthday
	m.AreaID = in.AreaID
	m.LevelID = in.LevelID
	m.Point = in.Point
	m.Experience = in.Experience
	m.TagIDs = in.TagIDs
	m.Mark = in.Mark
	m.GroupID = in.GroupID
	// 未传状态时保留原状态，避免编辑漏传把会员静默改成禁用
	if in.Status != nil {
		m.Status = baseEnums.CommonStatusEnum(*in.Status)
	}

	return s.members.Update(ctx, m)
}

// UpdateStatus 启用 / 禁用会员，供列表开关直接调用。
func (s *MemberService) UpdateStatus(ctx context.Context, id types.ID, status int8) error {
	if !baseEnums.CommonStatusEnum(status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "会员状态取值不合法")
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	m.Status = baseEnums.CommonStatusEnum(status)

	return s.members.UpdateStatus(ctx, m)
}

// Delete 软删除会员。
func (s *MemberService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	return s.members.Delete(ctx, m)
}

// AdjustPoint 调整会员积分：余额变更与积分记录写入在同一事务内完成。
// point 用正负表示增减，与 yudao 的 member_point_record 设计一致。
func (s *MemberService) AdjustPoint(ctx context.Context, id types.ID, point int, title, description string) error {
	if point == 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "积分变动数量不能为 0")
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	total := m.Point + point
	if total < 0 {
		return baseErrs.New(memberErrs.MemberPointNotEnough, m.Point)
	}

	title = strings.TrimSpace(title)
	if title == "" {
		title = "后台调整"
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		m.Point = total
		if err := s.members.UpdatePoint(ctx, tx, m); err != nil {
			return err
		}

		return s.points.Create(ctx, tx, &model.MemberPointRecordModel{
			UserID:      m.ID,
			BizType:     MemberPointBizAdjust,
			Title:       title,
			Description: strings.TrimSpace(description),
			Point:       point,
			TotalPoint:  total,
		})
	})
}

// AdjustLevel 调整会员等级与经验值，并写入等级变更记录。
func (s *MemberService) AdjustLevel(ctx context.Context, id types.ID, levelID types.ID, experience int, title, description string) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	if levelID > 0 && levelID != m.LevelID {
		if err := s.ensureLevel(ctx, levelID); err != nil {
			return err
		}
	}
	if levelID == 0 {
		levelID = m.LevelID
	}
	if experience < 0 {
		experience = 0
	}

	title = strings.TrimSpace(title)
	if title == "" {
		title = "后台调整等级"
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		m.LevelID = levelID
		m.Experience = experience
		if err := s.members.UpdateLevel(ctx, tx, m); err != nil {
			return err
		}

		return s.members.CreateLevelRecord(ctx, tx, &model.MemberLevelRecordModel{
			UserID:      m.ID,
			LevelID:     levelID,
			Experience:  experience,
			Title:       title,
			Description: strings.TrimSpace(description),
		})
	})
}

// PointRecords 分页查询会员积分记录。
func (s *MemberService) PointRecords(ctx context.Context, id types.ID, pageNo, pageSize int) ([]model.MemberPointRecordModel, int64, error) {
	if _, err := s.Detail(ctx, id); err != nil {
		return nil, 0, err
	}

	return s.points.Paginate(ctx, id, pageNo, pageSize)
}

// LevelRecords 分页查询会员等级记录。
func (s *MemberService) LevelRecords(ctx context.Context, id types.ID, pageNo, pageSize int) ([]model.MemberLevelRecordModel, int64, error) {
	if _, err := s.Detail(ctx, id); err != nil {
		return nil, 0, err
	}

	return s.members.LevelRecords(ctx, id, pageNo, pageSize)
}

// SignRecords 分页查询会员签到记录。
func (s *MemberService) SignRecords(ctx context.Context, id types.ID, pageNo, pageSize int) ([]model.MemberSignRecordModel, int64, error) {
	if _, err := s.Detail(ctx, id); err != nil {
		return nil, 0, err
	}

	return s.members.SignRecords(ctx, id, pageNo, pageSize)
}

// Addresses 查询会员收货地址。
func (s *MemberService) Addresses(ctx context.Context, id types.ID) ([]model.MemberAddressModel, error) {
	if _, err := s.Detail(ctx, id); err != nil {
		return nil, err
	}

	return s.members.Addresses(ctx, id)
}

// ensureLevel 校验会员引用的等级存在且启用，避免写入悬空等级。
func (s *MemberService) ensureLevel(ctx context.Context, levelID types.ID) error {
	if levelID == 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "请选择会员等级")
	}

	m, err := s.levels.FindByID(ctx, levelID)
	if err != nil {
		return err
	}
	if m == nil || m.Status != baseEnums.CommonStatusEnabled {
		return baseErrs.New(memberErrs.MemberLevelUnavailable, levelID)
	}

	return nil
}

// memberStatusOrDefault 把可选的会员状态转成落库值：调用方没传时默认启用。
// 参数 status：请求里的状态指针，nil 表示未传；返回值：可直接写入 model 的状态枚举。
func memberStatusOrDefault(status *int8) baseEnums.CommonStatusEnum {
	if status == nil {
		return baseEnums.CommonStatusEnabled
	}

	return baseEnums.CommonStatusEnum(*status)
}

// normalizeMemberInput 统一清洗入参并补齐默认值，供新增与编辑复用。
func normalizeMemberInput(in *MemberSaveInput) error {
	in.Nickname = strings.TrimSpace(in.Nickname)
	in.Mobile = strings.TrimSpace(in.Mobile)
	in.Avatar = strings.TrimSpace(in.Avatar)
	in.Name = strings.TrimSpace(in.Name)
	in.Mark = strings.TrimSpace(in.Mark)
	in.TagIDs = strings.Trim(strings.TrimSpace(in.TagIDs), ",")

	if in.Nickname == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "昵称不能为空")
	}
	if in.Mobile == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "手机号不能为空")
	}
	// 状态可以不传，传了才校验取值范围
	if in.Status != nil && !baseEnums.CommonStatusEnum(*in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "会员状态取值不合法")
	}
	if in.Sex < 0 || in.Sex > 2 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "性别取值不合法")
	}
	if in.AreaID < 0 {
		in.AreaID = 0
	}
	if in.Point < 0 {
		in.Point = 0
	}
	if in.Experience < 0 {
		in.Experience = 0
	}

	return nil
}
