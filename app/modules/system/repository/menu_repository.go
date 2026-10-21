// Package repository 存放 system 模块的数据访问实现，只操作本模块的表。
package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
)

const menuTable = model.SystemMenuTable

// MenuRepository 是 system_menu 表的仓储。
type MenuRepository struct{}

func NewMenuRepository() *MenuRepository {
	return &MenuRepository{}
}

// ListAll 查询全部未软删的菜单，按 sort、id 升序，Service 据此在内存组树。
func (r *MenuRepository) ListAll(ctx context.Context) ([]model.SystemMenuModel, error) {
	list := make([]model.SystemMenuModel, 0)

	err := facades.DB().WithContext(ctx).Table(menuTable).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// FindByID 按主键查询未软删的菜单；未命中返回 (nil, nil)。
func (r *MenuRepository) FindByID(ctx context.Context, id types.ID) (*model.SystemMenuModel, error) {
	var m model.SystemMenuModel

	if err := facades.DB().WithContext(ctx).Table(menuTable).
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

// CountChildren 统计直接子节点的数量，用于删除前的父子校验。
func (r *MenuRepository) CountChildren(ctx context.Context, parentID types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(menuTable).
		Where("parent_id", parentID).
		WhereNull("deleted_at").
		Count()
}

// Create 插入菜单。
func (r *MenuRepository) Create(ctx context.Context, m *model.SystemMenuModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(menuTable).Insert(map[string]any{
		"id":             m.ID,
		"name":           m.Name,
		"permission":     m.Permission,
		"type":           m.Type,
		"sort":           m.Sort,
		"parent_id":      m.ParentID,
		"path":           m.Path,
		"icon":           m.Icon,
		"component":      m.Component,
		"component_name": m.ComponentName,
		"status":         m.Status,
		"visible":        m.Visible,
		"keep_alive":     m.KeepAlive,
		"always_show":    m.AlwaysShow,
		"creator":        m.Creator,
		"created_at":     m.CreatedAt,
		"updater":        m.Updater,
		"updated_at":     m.UpdatedAt,
	})

	return err
}

// Update 更新菜单；parent_id 变化由 Service 提前校验。
func (r *MenuRepository) Update(ctx context.Context, m *model.SystemMenuModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(menuTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":           m.Name,
			"permission":     m.Permission,
			"type":           m.Type,
			"sort":           m.Sort,
			"parent_id":      m.ParentID,
			"path":           m.Path,
			"icon":           m.Icon,
			"component":      m.Component,
			"component_name": m.ComponentName,
			"status":         m.Status,
			"visible":        m.Visible,
			"keep_alive":     m.KeepAlive,
			"always_show":    m.AlwaysShow,
			"updater":        m.Updater,
			"updated_at":     m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(systemErrs.SystemMenuNotExists, m.ID)
	}

	return nil
}

// ListPermissions 返回全部非空权限标识，超级管理员直接拥有全部权限（docs 9.3）。
func (r *MenuRepository) ListPermissions(ctx context.Context) ([]string, error) {
	list := make([]model.SystemMenuModel, 0)

	if err := facades.DB().WithContext(ctx).Table(menuTable).
		Select("permission").
		Where("permission <> ?", "").
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return uniquePermissions(list), nil
}

// ListPermissionsByIDs 返回指定菜单集合中的权限标识。
func (r *MenuRepository) ListPermissionsByIDs(ctx context.Context, ids []types.ID) ([]string, error) {
	permissions := make([]string, 0)
	if len(ids) == 0 {
		return permissions, nil
	}

	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}

	list := make([]model.SystemMenuModel, 0)
	if err := facades.DB().WithContext(ctx).Table(menuTable).
		Select("permission").
		WhereIn("id", values).
		Where("permission <> ?", "").
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return uniquePermissions(list), nil
}

// uniquePermissions 去重并过滤空串。
func uniquePermissions(list []model.SystemMenuModel) []string {
	result := make([]string, 0, len(list))
	seen := make(map[string]bool, len(list))

	for _, item := range list {
		if item.Permission == "" || seen[item.Permission] {
			continue
		}
		seen[item.Permission] = true
		result = append(result, item.Permission)
	}

	return result
}

// Delete 软删除；RowsAffected 为 0 说明记录已被删除或不存在。
func (r *MenuRepository) Delete(ctx context.Context, m *model.SystemMenuModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(menuTable).
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
		return baseErrs.New(systemErrs.SystemMenuNotExists, m.ID)
	}

	return nil
}
