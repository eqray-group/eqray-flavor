package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000004SeedPayOrderSyncMenu 给「支付单」页面补「同步支付单」按钮权限点。
// 补单是主动查渠道的动作，单独授权，避免有查询权限的人误触。
type M20261023000004SeedPayOrderSyncMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000004SeedPayOrderSyncMenu) Signature() string {
	return "20261023000004_seed_pay_order_sync_menu"
}

// Up 注册 pay:order:sync 权限点并授权给内置角色。
func (r *M20261023000004SeedPayOrderSyncMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 2).
		Where("path", "/pay/order").
		WhereNull("deleted_at").
		First(&page); err != nil {
		return err
	}
	if page.ID == 0 {
		return nil
	}

	now := time.Now()

	var existing model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("parent_id", page.ID).
		Where("type", 3).
		Where("permission", "pay:order:sync").
		WhereNull("deleted_at").
		First(&existing); err != nil {
		return err
	}

	menuID := existing.ID
	if menuID == 0 {
		menuID = support.NewID()
		if _, err := facades.DB().Table("system_menu").Insert(map[string]any{
			"id":             menuID,
			"name":           "同步支付单",
			"permission":     "pay:order:sync",
			"type":           3,
			"sort":           5,
			"parent_id":      page.ID,
			"path":           "",
			"icon":           "",
			"component":      "",
			"component_name": "",
			"status":         1,
			"visible":        1,
			"keep_alive":     1,
			"always_show":    0,
			"creator":        "system",
			"created_at":     now,
			"updater":        "system",
			"updated_at":     now,
		}); err != nil {
			return err
		}
	}

	if err := r.grantBuiltinRoles(menuID, now); err != nil {
		return err
	}

	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// grantBuiltinRoles 把按钮授权给内置角色，已授权的角色跳过。
func (r *M20261023000004SeedPayOrderSyncMenu) grantBuiltinRoles(menuID types.ID, now time.Time) error {
	if !facades.Schema().HasTable("system_role") || !facades.Schema().HasTable("system_role_menu") {
		return nil
	}

	roles := make([]model.SystemRoleModel, 0)
	if err := facades.DB().Table("system_role").
		Select("id").
		Where("type", 1).
		WhereNull("deleted_at").
		Get(&roles); err != nil {
		return err
	}

	for _, role := range roles {
		var granted model.SystemRoleMenuModel
		if err := facades.DB().Table("system_role_menu").
			Where("role_id", role.ID).
			Where("menu_id", menuID).
			WhereNull("deleted_at").
			First(&granted); err != nil {
			return err
		}
		if granted.ID > 0 {
			continue
		}

		if _, err := facades.DB().Table("system_role_menu").Insert(map[string]any{
			"id":         support.NewID(),
			"role_id":    role.ID,
			"menu_id":    menuID,
			"creator":    "system",
			"created_at": now,
			"updater":    "system",
			"updated_at": now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Down 删除同步支付单按钮及其角色授权。
func (r *M20261023000004SeedPayOrderSyncMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var button model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 3).
		Where("permission", "pay:order:sync").
		WhereNull("deleted_at").
		First(&button); err != nil {
		return err
	}
	if button.ID == 0 {
		return nil
	}

	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").Where("menu_id", button.ID).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", button.ID).Delete()

	return err
}
