package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000048SeedPayRefundAuditMenu 给「退款单」页面补「审核退款」按钮权限点。
// 审核与发起拆开授权：能提交退款申请的人不一定能放款，资金动作要二次确认。
type M20261023000048SeedPayRefundAuditMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000048SeedPayRefundAuditMenu) Signature() string {
	return "20261023000048_seed_pay_refund_audit_menu"
}

// Up 注册按钮权限点并授权给内置角色。
func (r *M20261023000048SeedPayRefundAuditMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 2).
		Where("component", "/pay/refund").
		WhereNull("deleted_at").
		First(&page); err != nil {
		return err
	}
	if page.ID == 0 {
		return nil
	}

	now := time.Now()

	menuID, err := r.ensureMenu(page.ID, "审核退款", 3, 3, "pay:refund:audit", now)
	if err != nil {
		return err
	}

	if err := r.grantBuiltinRoles(menuID, now); err != nil {
		return err
	}

	// 新增权限点后前移权限缓存版本，避免管理员等待 30 分钟才拿到新按钮
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找按钮，不存在时插入。
func (r *M20261023000048SeedPayRefundAuditMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, permission string, now time.Time) (types.ID, error) {
	var existing model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("parent_id", parentID).
		Where("type", menuType).
		Where("name", name).
		WhereNull("deleted_at").
		First(&existing); err != nil {
		return 0, err
	}
	if existing.ID > 0 {
		return existing.ID, nil
	}

	id := support.NewID()
	_, err := facades.DB().Table("system_menu").Insert(map[string]any{
		"id":             id,
		"name":           name,
		"permission":     permission,
		"type":           menuType,
		"sort":           sort,
		"parent_id":      parentID,
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
	})

	return id, err
}

// grantBuiltinRoles 把按钮授权给内置角色，已授权的角色跳过。
func (r *M20261023000048SeedPayRefundAuditMenu) grantBuiltinRoles(menuID types.ID, now time.Time) error {
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

// Down 删除按钮及其角色授权。
func (r *M20261023000048SeedPayRefundAuditMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var button model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 3).
		Where("permission", "pay:refund:audit").
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
