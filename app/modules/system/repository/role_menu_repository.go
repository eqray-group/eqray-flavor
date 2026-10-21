// Package repository 存放 system 模块的数据访问实现，只操作本模块的表。
package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

const roleMenuTable = model.SystemRoleMenuTable

// RoleMenuRepository 是 system_role_menu 关联表的仓储。
type RoleMenuRepository struct{}

func NewRoleMenuRepository() *RoleMenuRepository {
	return &RoleMenuRepository{}
}

// ListMenuIDsByRole 返回角色已授权的菜单 ID（含按钮）。
func (r *RoleMenuRepository) ListMenuIDsByRole(ctx context.Context, roleID types.ID) ([]types.ID, error) {
	rows := make([]model.SystemRoleMenuModel, 0)

	if err := facades.DB().WithContext(ctx).Table(roleMenuTable).
		Where("role_id", roleID).
		WhereNull("deleted_at").
		Get(&rows); err != nil {
		return nil, err
	}

	ids := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.MenuID)
	}

	return ids, nil
}

// ListMenuIDsByRoles 返回多个角色的菜单 ID 并集，权限计算时使用。
func (r *RoleMenuRepository) ListMenuIDsByRoles(ctx context.Context, roleIDs []types.ID) ([]types.ID, error) {
	ids := make([]types.ID, 0)
	if len(roleIDs) == 0 {
		return ids, nil
	}

	values := make([]any, 0, len(roleIDs))
	for _, id := range roleIDs {
		values = append(values, id)
	}

	rows := make([]model.SystemRoleMenuModel, 0)
	if err := facades.DB().WithContext(ctx).Table(roleMenuTable).
		WhereIn("role_id", values).
		WhereNull("deleted_at").
		Get(&rows); err != nil {
		return nil, err
	}

	seen := make(map[types.ID]bool, len(rows))
	for _, row := range rows {
		if seen[row.MenuID] {
			continue
		}
		seen[row.MenuID] = true
		ids = append(ids, row.MenuID)
	}

	return ids, nil
}

// Sync 把角色的菜单集合对齐到 menuIDs：多余的软删除，缺少的补插入。
func (r *RoleMenuRepository) Sync(ctx context.Context, tx db.Tx, roleID types.ID, menuIDs []types.ID) error {
	rows := make([]model.SystemRoleMenuModel, 0)
	if err := tx.Table(roleMenuTable).
		Select("id", "menu_id").
		Where("role_id", roleID).
		WhereNull("deleted_at").
		Get(&rows); err != nil {
		return err
	}

	wanted := toIDSet(menuIDs)
	alive := make(map[types.ID]bool, len(rows))

	for _, row := range rows {
		alive[row.MenuID] = true
		if wanted[row.MenuID] {
			continue
		}
		// 关联表不再软删除，直接物理删除，避免占用 (role_id, menu_id) 唯一键
		if _, err := tx.Table(roleMenuTable).
			Where("id", row.ID).
			Delete(); err != nil {
			return err
		}
	}

	for _, menuID := range menuIDs {
		if menuID <= 0 || alive[menuID] {
			continue
		}

		m := &model.SystemRoleMenuModel{RoleID: roleID, MenuID: menuID}
		m.FillCreate(ctx)

		if _, err := tx.Table(roleMenuTable).Insert(map[string]any{
			"id":         m.ID,
			"role_id":    m.RoleID,
			"menu_id":    m.MenuID,
			"creator":    m.Creator,
			"created_at": m.CreatedAt,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		}); err != nil {
			return err
		}
	}

	return nil
}

// DeleteByRole 物理删除角色的全部授权，删除角色时一并清理。
func (r *RoleMenuRepository) DeleteByRole(ctx context.Context, tx db.Tx, roleID types.ID) error {
	_, err := tx.Table(roleMenuTable).
		Where("role_id", roleID).
		Delete()

	return err
}
