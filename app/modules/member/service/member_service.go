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
	// MemberPointBizConsume 消费抵扣扣减的积分（退款返还复用同一类型，靠 bizID 关联同一笔订单）
	MemberPointBizConsume int8 = 30
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

// ConsumePoint 原子扣减会员积分并写流水：余额变更与流水写入在同一事务内完成。
// 入参：会员 ID、扣减积分数（正数）、业务标题与描述、关联业务 ID；返回错误表示积分不足或会员不存在。
// 与 AdjustPoint 的区别：这里用带「余额 >= 扣减额」的条件更新，并发下不会扣成负数（docs 07 §2）。
func (s *MemberService) ConsumePoint(ctx context.Context, id types.ID, point int, title, description string, bizID types.ID) error {
	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.ConsumePointTx(ctx, tx, id, point, title, description, bizID)
	})
}

// ConsumePointTx 在调用方事务内扣减会员积分并写流水。
// 入参与 ConsumePoint 相同，外加事务句柄；返回错误表示积分不足或会员不存在。
// 订单结算等需要与业务单据同生共死的场景必须用这个版本，避免业务回滚而积分已扣。
func (s *MemberService) ConsumePointTx(ctx context.Context, tx db.Tx, id types.ID, point int, title, description string, bizID types.ID) error {
	if point <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "扣减积分数量必须大于 0")
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		title = "消费抵扣"
	}

	total, ok, deductErr := s.members.DeductPoint(ctx, tx, m.ID, point, "system")
	if deductErr != nil {
		return deductErr
	}
	if !ok {
		return baseErrs.New(memberErrs.MemberPointNotEnough, m.Point)
	}

	return s.points.Create(ctx, tx, &model.MemberPointRecordModel{
		UserID:      m.ID,
		BizType:     MemberPointBizConsume,
		BizID:       bizID,
		Title:       title,
		Description: strings.TrimSpace(description),
		Point:       -point,
		TotalPoint:  total,
	})
}

// RestorePoint 原子返还会员积分并写流水：整单退款返还积分等场景使用。
// 入参：会员 ID、返还积分数（正数）、业务标题与描述、关联业务 ID；返回错误表示会员不存在。
func (s *MemberService) RestorePoint(ctx context.Context, id types.ID, point int, title, description string, bizID types.ID) error {
	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.RestorePointTx(ctx, tx, id, point, title, description, bizID)
	})
}

// RestorePointTx 在调用方事务内返还会员积分并写流水。
// 入参与 RestorePoint 相同，外加事务句柄；返回错误表示会员不存在。
func (s *MemberService) RestorePointTx(ctx context.Context, tx db.Tx, id types.ID, point int, title, description string, bizID types.ID) error {
	if point <= 0 {
		return nil
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		title = "退款返还"
	}

	total, ok, addErr := s.members.AddPoint(ctx, tx, m.ID, point, "system")
	if addErr != nil {
		return addErr
	}
	if !ok {
		return baseErrs.New(memberErrs.MemberNotExists)
	}

	return s.points.Create(ctx, tx, &model.MemberPointRecordModel{
		UserID:      m.ID,
		BizType:     MemberPointBizConsume,
		BizID:       bizID,
		Title:       title,
		Description: strings.TrimSpace(description),
		Point:       point,
		TotalPoint:  total,
	})
}

// AddExperience 累计会员成长值并在达标时自动升级，等级只升不降。
// 入参：会员 ID、累加成长值、业务标题与描述；返回错误表示会员不存在。
// 等级判定取「启用等级中经验值门槛不超过当前成长值的最高一档」。
func (s *MemberService) AddExperience(ctx context.Context, id types.ID, experience int, title, description string) error {
	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.AddExperienceTx(ctx, tx, id, experience, title, description)
	})
}

// AddExperienceTx 在调用方事务内累计会员成长值并在达标时自动升级。
// 入参与 AddExperience 相同，外加事务句柄；返回错误表示会员不存在。
func (s *MemberService) AddExperienceTx(ctx context.Context, tx db.Tx, id types.ID, experience int, title, description string) error {
	if experience <= 0 {
		return nil
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		title = "消费成长"
	}

	total, ok, addErr := s.members.AddExperience(ctx, tx, m.ID, experience, "system")
	if addErr != nil {
		return addErr
	}
	if !ok {
		return baseErrs.New(memberErrs.MemberNotExists)
	}

	// 重算等级：只升不降，达标才写等级变更记录
	targetLevel, levelErr := s.resolveLevelByExperience(ctx, total)
	if levelErr != nil {
		return levelErr
	}
	if targetLevel == nil || targetLevel.ID == m.LevelID {
		return nil
	}

	// 更新等级：复用已有 model，只改等级与成长值两列
	m.LevelID = targetLevel.ID
	m.Experience = total
	if updateErr := s.members.UpdateLevel(ctx, tx, m); updateErr != nil {
		return updateErr
	}

	return s.members.CreateLevelRecord(ctx, tx, &model.MemberLevelRecordModel{
		UserID:      m.ID,
		LevelID:     targetLevel.ID,
		Experience:  total,
		Title:       title,
		Description: strings.TrimSpace(description),
	})
}

// DeductExperience 回退会员成长值：整单退款扣回已赠送的成长值时使用。
// 入参：会员 ID、回退成长值（正数）；返回错误表示会员不存在。
// 说明：等级只升不降，因此这里只回退成长值，不做降级处理；
// 成长值回退后若低于当前等级门槛，等级保持不变（等级不会因退款而降级）。
func (s *MemberService) DeductExperience(ctx context.Context, id types.ID, experience int) error {
	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.DeductExperienceTx(ctx, tx, id, experience)
	})
}

// DeductExperienceTx 在调用方事务内回退会员成长值。
// 入参与 DeductExperience 相同，外加事务句柄；返回错误表示会员不存在。
// 订单退款等需要与业务单据同生共死的场景必须用这个版本。
func (s *MemberService) DeductExperienceTx(ctx context.Context, tx db.Tx, id types.ID, experience int) error {
	if experience <= 0 {
		return nil
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	_, ok, deductErr := s.members.DeductExperience(ctx, tx, m.ID, experience, "system")
	if deductErr != nil {
		return deductErr
	}
	if !ok {
		return baseErrs.New(memberErrs.MemberNotExists)
	}

	return nil
}

// resolveLevelByExperience 按成长值选出应处的等级：取经验门槛不超过当前成长值的最高一档。
// 入参：当前成长值；返回目标等级，未找到可用等级时返回 (nil, nil)。
func (s *MemberService) resolveLevelByExperience(ctx context.Context, experience int) (*model.MemberLevelModel, error) {
	levels, err := s.levels.Enabled(ctx)
	if err != nil {
		return nil, err
	}

	var matched *model.MemberLevelModel
	for i := range levels {
		level := &levels[i]
		if level.Experience > experience {
			continue
		}
		if matched == nil || level.Experience > matched.Experience {
			matched = level
		}
	}

	return matched, nil
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
