package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000053SeedPayWalletRechargeMenu 注册「门店充值」按钮权限点。
// 门店端 POS 用现金给会员钱包充值，与后台「钱包调账」是两件事：
// 调账是纠错，充值是收钱，必须能单独授权给收银员。
type M20261023000053SeedPayWalletRechargeMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000053SeedPayWalletRechargeMenu) Signature() string {
	return "20261023000053_seed_pay_wallet_recharge_menu"
}

// Up 在「会员钱包」页面下挂按钮并授权给内置角色。
func (r *M20261023000053SeedPayWalletRechargeMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 2).
		Where("component", "/pay/wallet").
		WhereNull("deleted_at").
		First(&page); err != nil {
		return err
	}
	if page.ID == 0 {
		return nil
	}

	menuID, err := r.ensureMenu(page.ID, "门店充值", "pay:wallet:recharge", now)
	if err != nil {
		return err
	}
	if err := r.grantBuiltinRoles([]types.ID{menuID}, now); err != nil {
		return err
	}

	// 新权限点要立刻生效，前移权限缓存版本
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 权限码」查找按钮，不存在时插入。
func (r *M20261023000053SeedPayWalletRechargeMenu) ensureMenu(parentID types.ID, name, permission string, now time.Time) (types.ID, error) {
	var existing model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("parent_id", parentID).
		Where("type", 3).
		Where("permission", permission).
		WhereNull("deleted_at").
		First(&existing); err != nil {
		return 0, err
	}
	if existing.ID > 0 {
		return existing.ID, nil
	}

	id := support.NewID()
	_, err := facades.DB().Table("system_menu").Insert(map[string]any{
		"id":          id,
		"name":        name,
		"permission":  permission,
		"type":        3,
		"sort":        20,
		"parent_id":   parentID,
		"status":      1,
		"visible":     1,
		"creator":     "system",
		"created_at":  now,
		"updater":     "system",
		"updated_at":  now,
		"keep_alive":  1,
		"always_show": 0,
	})

	return id, err
}

// grantBuiltinRoles 把按钮授权给内置角色，已授权的跳过。
func (r *M20261023000053SeedPayWalletRechargeMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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
		granted := make(map[types.ID]bool, len(menuIDs))
		rows := make([]model.SystemRoleMenuModel, 0)
		if err := facades.DB().Table("system_role_menu").
			Select("menu_id").
			Where("role_id", role.ID).
			WhereNull("deleted_at").
			Get(&rows); err != nil {
			return err
		}
		for _, row := range rows {
			granted[row.MenuID] = true
		}

		for _, menuID := range menuIDs {
			if granted[menuID] {
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
	}

	return nil
}

// Down 删除按钮与授权。
func (r *M20261023000053SeedPayWalletRechargeMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var button model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("permission", "pay:wallet:recharge").
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
