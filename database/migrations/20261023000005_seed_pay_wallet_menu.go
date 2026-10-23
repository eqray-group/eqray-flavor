package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000005SeedPayWalletMenu 在「支付管理」目录下补钱包、充值套餐、充值记录三个页面与按钮权限点。
type M20261023000005SeedPayWalletMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000005SeedPayWalletMenu) Signature() string {
	return "20261023000005_seed_pay_wallet_menu"
}

// Up 注册钱包相关页面与权限点，并授权给内置角色。
func (r *M20261023000005SeedPayWalletMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var dir model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 1).
		Where("path", "/pay").
		WhereNull("deleted_at").
		First(&dir); err != nil {
		return err
	}
	if dir.ID == 0 {
		return nil
	}

	now := time.Now()
	menuIDs := make([]types.ID, 0, 12)

	walletPageID, err := r.ensureMenu(dir.ID, "会员钱包", 2, 6, "/pay/wallet", "/pay/wallet", "PayWallet", "ri:wallet-3-line", "", now)
	if err != nil {
		return err
	}
	menuIDs = append(menuIDs, walletPageID)
	for index, button := range [][2]string{
		{"钱包查询", "pay:wallet:list"},
		{"钱包调账", "pay:wallet:update"},
	} {
		id, err := r.ensureMenu(walletPageID, button[0], 3, index+1, "", "", "", "", button[1], now)
		if err != nil {
			return err
		}
		menuIDs = append(menuIDs, id)
	}

	packagePageID, err := r.ensureMenu(dir.ID, "充值套餐", 2, 7, "/pay/wallet-package", "/pay/wallet-package", "PayWalletPackage", "ri:coupon-3-line", "", now)
	if err != nil {
		return err
	}
	menuIDs = append(menuIDs, packagePageID)
	for index, button := range [][2]string{
		{"充值套餐查询", "pay:wallet-package:list"},
		{"充值套餐新增", "pay:wallet-package:create"},
		{"充值套餐修改", "pay:wallet-package:update"},
		{"充值套餐删除", "pay:wallet-package:delete"},
	} {
		id, err := r.ensureMenu(packagePageID, button[0], 3, index+1, "", "", "", "", button[1], now)
		if err != nil {
			return err
		}
		menuIDs = append(menuIDs, id)
	}

	rechargePageID, err := r.ensureMenu(dir.ID, "充值记录", 2, 8, "/pay/wallet-recharge", "/pay/wallet-recharge", "PayWalletRecharge", "ri:file-list-2-line", "", now)
	if err != nil {
		return err
	}
	menuIDs = append(menuIDs, rechargePageID)
	rechargeButtonID, err := r.ensureMenu(rechargePageID, "充值记录查询", 3, 1, "", "", "", "", "pay:wallet-recharge:list", now)
	if err != nil {
		return err
	}
	menuIDs = append(menuIDs, rechargeButtonID)

	if err := r.grantBuiltinRoles(menuIDs, now); err != nil {
		return err
	}

	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
// 入参：父级、名称、类型、排序、路径、组件、路由名、图标、权限标识、时间；返回菜单 ID。
func (r *M20261023000005SeedPayWalletMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, componentName, icon, permission string, now time.Time) (types.ID, error) {
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
		"path":           path,
		"icon":           icon,
		"component":      component,
		"component_name": componentName,
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

// grantBuiltinRoles 把本次新增的菜单授权给内置角色，已授权的组合跳过。
func (r *M20261023000005SeedPayWalletMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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
		rows := make([]model.SystemRoleMenuModel, 0)
		if err := facades.DB().Table("system_role_menu").
			Select("menu_id").
			Where("role_id", role.ID).
			WhereNull("deleted_at").
			Get(&rows); err != nil {
			return err
		}

		granted := make(map[types.ID]bool, len(rows))
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

// Down 删除钱包相关页面、按钮及其角色授权。
func (r *M20261023000005SeedPayWalletMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	paths := []string{"/pay/wallet", "/pay/wallet-package", "/pay/wallet-recharge"}
	for _, path := range paths {
		var page model.SystemMenuModel
		if err := facades.DB().Table("system_menu").
			Where("type", 2).
			Where("path", path).
			WhereNull("deleted_at").
			First(&page); err != nil {
			return err
		}
		if page.ID == 0 {
			continue
		}

		buttons := make([]model.SystemMenuModel, 0)
		if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", page.ID).Get(&buttons); err != nil {
			return err
		}
		for _, button := range buttons {
			if err := r.deleteMenu(button.ID); err != nil {
				return err
			}
		}
		if err := r.deleteMenu(page.ID); err != nil {
			return err
		}
	}

	return nil
}

// deleteMenu 删除单个菜单及其角色授权。
func (r *M20261023000005SeedPayWalletMenu) deleteMenu(id types.ID) error {
	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").Where("menu_id", id).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", id).Delete()

	return err
}
