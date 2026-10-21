package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
)

const groupTable = model.MemberGroupTable

// MemberGroupFilter 是分组列表的筛选条件。
type MemberGroupFilter struct {
	Keyword string
	Status  *baseEnums.CommonStatusEnum
}

// MemberGroupRepository 是 member_group 表的仓储。
type MemberGroupRepository struct{}

func NewMemberGroupRepository() *MemberGroupRepository {
	return &MemberGroupRepository{}
}

// Create 新增会员分组。
func (r *MemberGroupRepository) Create(ctx context.Context, m *model.MemberGroupModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(groupTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// Update 更新分组名称、状态与备注。
func (r *MemberGroupRepository) Update(ctx context.Context, m *model.MemberGroupModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(groupTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// Delete 软删除分组；记录不存在时返回业务错误码。
func (r *MemberGroupRepository) Delete(ctx context.Context, m *model.MemberGroupModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(groupTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(memberErrs.MemberGroupNotExists, m.ID)
	}

	return nil
}

// FindByID 按主键查询分组，不存在返回 nil。
func (r *MemberGroupRepository) FindByID(ctx context.Context, id types.ID) (*model.MemberGroupModel, error) {
	var m model.MemberGroupModel

	if err := facades.DB().WithContext(ctx).Table(groupTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// Enabled 返回启用中的分组，供会员表单下拉使用。
func (r *MemberGroupRepository) Enabled(ctx context.Context) ([]model.MemberGroupModel, error) {
	list := make([]model.MemberGroupModel, 0)

	if err := facades.DB().WithContext(ctx).Table(groupTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 分页查询分组，支持名称模糊匹配与状态筛选。
func (r *MemberGroupRepository) Paginate(ctx context.Context, filter MemberGroupFilter, pageNo, pageSize int) ([]model.MemberGroupModel, int64, error) {
	list := make([]model.MemberGroupModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(groupTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
