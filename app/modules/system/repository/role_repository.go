// Package repository 存放 system 模块的数据访问实现，只操作本模块的表。
package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
)

const roleTable = model.SystemRoleTable

// RoleFilter 是角色列表的筛选条件，零值表示不限制。
type RoleFilter struct {
	Name   string
	Code   string
	Status *baseEnums.CommonStatusEnum
}

// RoleRepository 是 system_role 表的仓储。
type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}

// Paginate 按条件分页查询未软删的角色。
func (r *RoleRepository) Paginate(ctx context.Context, filter RoleFilter, pageNo, pageSize int) ([]model.SystemRoleModel, int64, error) {
	list := make([]model.SystemRoleModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(roleTable).WhereNull("deleted_at")
	if filter.Name != "" {
		query = query.WhereLike("name", "%"+filter.Name+"%")
	}
	if filter.Code != "" {
		query = query.WhereLike("code", "%"+filter.Code+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderBy("sort").OrderBy("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ListEnabled 返回全部启用角色，供用户分配角色时选择。
func (r *RoleRepository) ListEnabled(ctx context.Context) ([]model.SystemRoleModel, error) {
	list := make([]model.SystemRoleModel, 0)

	err := facades.DB().WithContext(ctx).Table(roleTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// FindByID 按主键查询未软删的角色；未命中返回 (nil, nil)。
func (r *RoleRepository) FindByID(ctx context.Context, id types.ID) (*model.SystemRoleModel, error) {
	var m model.SystemRoleModel

	if err := facades.DB().WithContext(ctx).Table(roleTable).
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

// FindByIDs 批量查询未软删的角色。
func (r *RoleRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.SystemRoleModel, error) {
	list := make([]model.SystemRoleModel, 0)
	if len(ids) == 0 {
		return list, nil
	}

	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}

	if err := facades.DB().WithContext(ctx).Table(roleTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ExistsByCode 判断角色标识是否已被占用；excludeID > 0 时排除自身。
func (r *RoleRepository) ExistsByCode(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	if code == "" {
		return false, nil
	}

	var m model.SystemRoleModel

	// 唯一键建在 code 上，软删除的角色仍占用标识，这里不能过滤 deleted_at
	query := facades.DB().WithContext(ctx).Table(roleTable).
		Where("code", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	if err := query.First(&m); err != nil {
		return false, err
	}

	return m.ID != 0, nil
}

// Create 插入角色；code 唯一键冲突转成业务错误码。
func (r *RoleRepository) Create(ctx context.Context, m *model.SystemRoleModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(roleTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"code":       m.Code,
		"sort":       m.Sort,
		"status":     m.Status,
		"type":       m.Type,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(systemErrs.SystemRoleCodeExists, m.Code)
	}

	return err
}

// Update 更新角色。
func (r *RoleRepository) Update(ctx context.Context, m *model.SystemRoleModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(roleTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"code":       m.Code,
			"sort":       m.Sort,
			"status":     m.Status,
			"type":       m.Type,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(systemErrs.SystemRoleCodeExists, m.Code)
	}

	return err
}

// Delete 软删除；RowsAffected 为 0 说明记录已被删除或不存在。
func (r *RoleRepository) Delete(ctx context.Context, m *model.SystemRoleModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(roleTable).
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
		return baseErrs.New(systemErrs.SystemRoleNotExists, m.ID)
	}

	return nil
}
