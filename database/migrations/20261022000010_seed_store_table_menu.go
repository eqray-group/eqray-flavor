package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261022000010SeedStoreTableMenu 写入「桌台区域」「桌台管理」页面与按钮权限点。
type M20261022000010SeedStoreTableMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261022000010SeedStoreTableMenu) Signature() string {
	return "20261022000010_seed_store_table_menu"
}

// Up 注册桌台区域与桌台页面菜单并授权给内置角色。
func (r *M20261022000010SeedStoreTableMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	// 门店管理目录通常由历史迁移创建；这里兜底查找 / 创建，保证单独重跑也可用
	storeID, err := r.ensureMenu(0, "门店管理", 1, 5, "/store", "", "", "ri:store-3-line", "", now)
	if err != nil {
		return err
	}

	areaPageID, err := r.ensureMenu(storeID, "桌台区域", 2, 3, "table-area", "/store/table-area", "StoreTableArea", "ri:layout-grid-line", "", now)
	if err != nil {
		return err
	}
	tablePageID, err := r.ensureMenu(storeID, "桌台管理", 2, 4, "table", "/store/table", "StoreTable", "ri:table-line", "", now)
	if err != nil {
		return err
	}

	menuIDs := []types.ID{storeID, areaPageID, tablePageID}
	areaButtons := [][2]string{
		{"区域查询", "store:table-area:list"},
		{"区域新增", "store:table-area:create"},
		{"区域修改", "store:table-area:update"},
		{"区域删除", "store:table-area:delete"},
	}
	for index, button := range areaButtons {
		id, err := r.ensureMenu(areaPageID, button[0], 3, index+1, "", "", "", "", button[1], now)
		if err != nil {
			return err
		}
		menuIDs = append(menuIDs, id)
	}

	tableButtons := [][2]string{
		{"桌台查询", "store:table:list"},
		{"桌台新增", "store:table:create"},
		{"桌台修改", "store:table:update"},
		{"桌台删除", "store:table:delete"},
		{"桌台开台清台", "store:table:operate"},
	}
	for index, button := range tableButtons {
		id, err := r.ensureMenu(tablePageID, button[0], 3, index+1, "", "", "", "", button[1], now)
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

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
// 入参：父级、名称、类型、排序、前端路径、组件、路由名、图标、权限标识、时间；返回菜单 ID。
func (r *M20261022000010SeedStoreTableMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, componentName, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261022000010SeedStoreTableMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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

// Down 删除桌台区域、桌台页面及其按钮与角色授权；保留门店管理目录。
func (r *M20261022000010SeedStoreTableMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	pages := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("component", []any{"/store/table-area", "/store/table"}).
		Where("type", 2).
		WhereNull("deleted_at").
		Get(&pages); err != nil {
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

	return nil
}
