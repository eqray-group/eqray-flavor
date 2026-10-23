package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000051SeedOrderPosMenu 注册 POS 代客下单与现金结账两个按钮权限点。
// 门店端（merchant-manage-ui）走静态路由，不依赖菜单树，但权限点必须存在才能授权。
type M20261023000051SeedOrderPosMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000051SeedOrderPosMenu) Signature() string {
	return "20261023000051_seed_order_pos_menu"
}

// Up 在订单管理页面下挂两个按钮并授权给内置角色。
func (r *M20261023000051SeedOrderPosMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	// 订单页面的 path 是 index、component 才是 /order，按 component 定位才不会漏
	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 2).
		Where("component", "/order").
		WhereNull("deleted_at").
		First(&page); err != nil {
		return err
	}
	if page.ID == 0 {
		return nil
	}

	buttons := [][2]string{
		{"POS 代客下单", "order:pos:create"},
		{"POS 现金结账", "order:pos:settle"},
	}

	menuIDs := make([]types.ID, 0, len(buttons))
	for index, button := range buttons {
		id, err := r.ensureMenu(page.ID, button[0], 3, index+10, button[1], now)
		if err != nil {
			return err
		}
		menuIDs = append(menuIDs, id)
	}

	if err := r.grantBuiltinRoles(menuIDs, now); err != nil {
		return err
	}

	// 新增权限点后前移权限缓存版本，避免管理员等待 30 分钟才拿到新按钮
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 名称」查找按钮，不存在时插入。
func (r *M20261023000051SeedOrderPosMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, permission string, now time.Time) (types.ID, error) {
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
		"id":          id,
		"name":        name,
		"permission":  permission,
		"type":        menuType,
		"sort":        sort,
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

// grantBuiltinRoles 把本次新增的按钮授权给内置角色，已授权的组合跳过。
func (r *M20261023000051SeedOrderPosMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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

// Down 删除两个按钮及其角色授权。
func (r *M20261023000051SeedOrderPosMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	buttons := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("permission", []any{"order:pos:create", "order:pos:settle"}).
		WhereNull("deleted_at").
		Get(&buttons); err != nil {
		return err
	}

	for _, button := range buttons {
		if facades.Schema().HasTable("system_role_menu") {
			if _, err := facades.DB().Table("system_role_menu").Where("menu_id", button.ID).Delete(); err != nil {
				return err
			}
		}
		if _, err := facades.DB().Table("system_menu").Where("id", button.ID).Delete(); err != nil {
			return err
		}
	}

	return nil
}
