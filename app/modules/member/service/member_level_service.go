package service

import (
	"context"
	"strings"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

// MemberLevelSaveInput 是新增 / 编辑会员等级的入参。
type MemberLevelSaveInput struct {
	Level           int
	Name            string
	Icon            string
	Experience      int
	DiscountPercent int
	Description     string
	Status          int8
	Sort            int
}

// MemberLevelListQuery 是会员等级列表的查询条件。
type MemberLevelListQuery struct {
	Keyword  string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// MemberLevelService 负责会员等级配置的业务规则。
type MemberLevelService struct {
	levels *repository.MemberLevelRepository
}

// NewMemberLevelService 构造会员等级服务。
func NewMemberLevelService() *MemberLevelService {
	return &MemberLevelService{levels: repository.NewMemberLevelRepository()}
}

// Page 分页查询会员等级。
func (s *MemberLevelService) Page(ctx context.Context, q MemberLevelListQuery) ([]model.MemberLevelModel, int64, error) {
	return s.levels.Paginate(ctx, repository.MemberLevelFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Status:  q.Status,
	}, q.PageNo, q.PageSize)
}

// Options 返回启用中的等级，供会员表单下拉。
func (s *MemberLevelService) Options(ctx context.Context) ([]model.MemberLevelModel, error) {
	return s.levels.Enabled(ctx)
}

// Detail 查询等级详情，不存在时返回业务错误码。
func (s *MemberLevelService) Detail(ctx context.Context, id types.ID) (*model.MemberLevelModel, error) {
	m, err := s.levels.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(memberErrs.MemberLevelNotExists, id)
	}

	return m, nil
}

// Create 新增会员等级；等级序号唯一。
func (s *MemberLevelService) Create(ctx context.Context, in MemberLevelSaveInput) (types.ID, error) {
	if err := normalizeLevelInput(&in); err != nil {
		return 0, err
	}

	exists, err := s.levels.ExistsByLevel(ctx, in.Level, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(memberErrs.MemberLevelExists, in.Level)
	}

	m := &model.MemberLevelModel{
		Level:           in.Level,
		Name:            in.Name,
		Icon:            in.Icon,
		Experience:      in.Experience,
		DiscountPercent: in.DiscountPercent,
		Description:     in.Description,
		Status:          baseEnums.CommonStatusEnum(in.Status),
		Sort:            in.Sort,
	}
	if err := s.levels.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 更新等级配置；等级序号创建后不可修改，避免会员等级错位。
func (s *MemberLevelService) Update(ctx context.Context, id types.ID, in MemberLevelSaveInput) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	// 入参里的 level 一律忽略，始终沿用库中序号
	in.Level = m.Level
	if err := normalizeLevelInput(&in); err != nil {
		return err
	}

	m.Name = in.Name
	m.Icon = in.Icon
	m.Experience = in.Experience
	m.DiscountPercent = in.DiscountPercent
	m.Description = in.Description
	m.Status = baseEnums.CommonStatusEnum(in.Status)
	m.Sort = in.Sort

	return s.levels.Update(ctx, m)
}

// Delete 删除等级；仍有会员在该等级时拒绝，避免会员等级悬空。
func (s *MemberLevelService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	count, err := s.levels.CountMembersByLevelID(ctx, m.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(memberErrs.MemberLevelInUse, m.Name)
	}

	return s.levels.Delete(ctx, m)
}

// normalizeLevelInput 清洗入参并校验取值范围。
func normalizeLevelInput(in *MemberLevelSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Icon = strings.TrimSpace(in.Icon)
	in.Description = strings.TrimSpace(in.Description)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "等级名称不能为空")
	}
	if in.Level < 1 || in.Level > 99 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "等级序号需在 1 ~ 99 之间")
	}
	if !baseEnums.CommonStatusEnum(in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "等级状态取值不合法")
	}
	if in.Experience < 0 {
		in.Experience = 0
	}
	if in.DiscountPercent <= 0 || in.DiscountPercent > 100 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "折扣百分比需在 1 ~ 100 之间")
	}
	if in.Sort < 0 {
		in.Sort = 0
	}

	return nil
}
