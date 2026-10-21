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

// MemberGroupSaveInput 是新增 / 编辑会员分组的入参。
type MemberGroupSaveInput struct {
	Name   string
	Status int8
	Remark string
}

// MemberGroupListQuery 是分组列表的查询条件。
type MemberGroupListQuery struct {
	Keyword  string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// MemberGroupService 负责会员分组的增删改查。
type MemberGroupService struct {
	groups *repository.MemberGroupRepository
}

// NewMemberGroupService 构造会员分组服务。
func NewMemberGroupService() *MemberGroupService {
	return &MemberGroupService{groups: repository.NewMemberGroupRepository()}
}

// Page 分页查询分组，关键字做 trim 后按名称模糊匹配。
func (s *MemberGroupService) Page(ctx context.Context, q MemberGroupListQuery) ([]model.MemberGroupModel, int64, error) {
	return s.groups.Paginate(ctx, repository.MemberGroupFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Status:  q.Status,
	}, q.PageNo, q.PageSize)
}

// Options 返回启用中的分组，供会员表单下拉。
func (s *MemberGroupService) Options(ctx context.Context) ([]model.MemberGroupModel, error) {
	return s.groups.Enabled(ctx)
}

// Detail 查询分组详情，不存在时返回业务错误码。
func (s *MemberGroupService) Detail(ctx context.Context, id types.ID) (*model.MemberGroupModel, error) {
	m, err := s.groups.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(memberErrs.MemberGroupNotExists, id)
	}

	return m, nil
}

// Create 新增分组，返回新记录 ID。
func (s *MemberGroupService) Create(ctx context.Context, in MemberGroupSaveInput) (types.ID, error) {
	if err := normalizeGroupInput(&in); err != nil {
		return 0, err
	}

	m := &model.MemberGroupModel{
		Name:   in.Name,
		Status: baseEnums.CommonStatusEnum(in.Status),
		Remark: in.Remark,
	}
	if err := s.groups.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑分组名称、状态与备注。
func (s *MemberGroupService) Update(ctx context.Context, id types.ID, in MemberGroupSaveInput) error {
	if err := normalizeGroupInput(&in); err != nil {
		return err
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	m.Name = in.Name
	m.Status = baseEnums.CommonStatusEnum(in.Status)
	m.Remark = in.Remark

	return s.groups.Update(ctx, m)
}

// Delete 删除分组（软删）。
func (s *MemberGroupService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	return s.groups.Delete(ctx, m)
}

// normalizeGroupInput 清洗入参并校验名称、状态等字段。
func normalizeGroupInput(in *MemberGroupSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "分组名不能为空")
	}
	if len([]rune(in.Name)) > 30 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "分组名不能超过 30 个字符")
	}
	if !baseEnums.CommonStatusEnum(in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "分组状态取值不合法")
	}

	return nil
}
