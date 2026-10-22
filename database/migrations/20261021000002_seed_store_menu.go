package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261021000002SeedStoreMenu 写入「门店管理」目录、门店列表页面与按钮权限点。
type M20261021000002SeedStoreMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261021000002SeedStoreMenu) Signature() string {
	return "20261021000002_seed_store_menu"
}

// Up 注册菜单并授权给内置角色。
func (r *M20261021000002SeedStoreMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	storeID, err := r.ensureMenu(0, "门店管理", 1, 5, "/store", "", "", "ri:store-3-line", "", now)
	if err != nil {
		return err
	}

	// 组件路径决定前端文件位置：/store/store -> src/views/store/store/index.vue。
	// 路由名必须显式给 StoreStore：页面路径 store 推导出的名字是 Store，会与上级目录路由重名。
	pageID, err := r.ensureMenu(storeID, "门店列表", 2, 1, "store", "/store/store", "StoreStore", "ri:store-3-line", "", now)
	if err != nil {
		return err
	}

	buttons := [][2]string{
		{"门店查询", "store:store:list"},
		{"门店新增", "store:store:create"},
		{"门店修改", "store:store:update"},
		{"门店删除", "store:store:delete"},
	}
	for index, button := range buttons {
		if _, err := r.ensureMenu(pageID, button[0], 3, index+1, "", "", "", "", button[1], now); err != nil {
			return err
		}
	}

	if err := r.grantBuiltinRoles(now); err != nil {
		return err
	}

	// 新增权限点后前移权限缓存版本，避免管理员等待 30 分钟才拿到新按钮
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
// componentName 传空表示按前端路径推导路由名；显式传入用于规避父子同名。
func (r *M20261021000002SeedStoreMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, componentName, icon, permission string, now time.Time) (types.ID, error) {
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
		// 已存在但缺组件名时补齐，保证重复执行迁移能修正历史数据
		if componentName != "" && existing.ComponentName == "" {
			if _, err := facades.DB().Table("system_menu").
				Where("id", existing.ID).
				Update(map[string]any{"component_name": componentName, "updated_at": now}); err != nil {
				return 0, err
			}
		}

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

// grantBuiltinRoles 给内置角色补齐全部菜单授权。
func (r *M20261021000002SeedStoreMenu) grantBuiltinRoles(now time.Time) error {
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

	menus := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereNull("deleted_at").
		Get(&menus); err != nil {
		return err
	}

	for _, role := range roles {
		granted := make(map[types.ID]bool, len(menus))

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

		for _, menu := range menus {
			if granted[menu.ID] {
				continue
			}

			if _, err := facades.DB().Table("system_role_menu").Insert(map[string]any{
				"id":         support.NewID(),
				"role_id":    role.ID,
				"menu_id":    menu.ID,
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

// Down 删除门店管理目录、门店列表页面及其按钮与授权。
func (r *M20261021000002SeedStoreMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	menus := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("name", []any{"门店管理"}).
		Where("type", 1).
		WhereNull("deleted_at").
		Get(&menus); err != nil {
		return err
	}

	for _, menu := range menus {
		pages := make([]model.SystemMenuModel, 0)
		if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", menu.ID).Get(&pages); err != nil {
			return err
		}

		for _, page := range pages {
			buttons := make([]model.SystemMenuModel, 0)
			if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", page.ID).Get(&buttons); err != nil {
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

			if facades.Schema().HasTable("system_role_menu") {
				if _, err := facades.DB().Table("system_role_menu").Where("menu_id", page.ID).Delete(); err != nil {
					return err
				}
			}
			if _, err := facades.DB().Table("system_menu").Where("id", page.ID).Delete(); err != nil {
				return err
			}
		}

		if facades.Schema().HasTable("system_role_menu") {
			if _, err := facades.DB().Table("system_role_menu").Where("menu_id", menu.ID).Delete(); err != nil {
				return err
			}
		}
		if _, err := facades.DB().Table("system_menu").Where("id", menu.ID).Delete(); err != nil {
			return err
		}
	}

	return nil
}
