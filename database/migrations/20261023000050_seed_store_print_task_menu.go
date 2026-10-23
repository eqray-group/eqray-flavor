package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000050SeedStorePrintTaskMenu 写入「打印任务」页面与按钮权限点。
// 打印任务由支付成功时自动生成，后台只做查询与重打。
type M20261023000050SeedStorePrintTaskMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000050SeedStorePrintTaskMenu) Signature() string {
	return "20261023000050_seed_store_print_task_menu"
}

// Up 注册打印任务菜单并授权给内置角色。
func (r *M20261023000050SeedStorePrintTaskMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	// 目录通常由 20261021000002 创建；这里兜底查找/创建，保证迁移可单独重跑
	storeID, err := r.ensureMenu(0, "门店管理", 1, 5, "/store", "", "", "ri:store-3-line", "", now)
	if err != nil {
		return err
	}

	pageID, err := r.ensureMenu(storeID, "打印任务", 2, 8, "print-task", "/store/print-task", "StorePrintTask", "ri:file-paper-2-line", "", now)
	if err != nil {
		return err
	}

	menuIDs := []types.ID{pageID}
	buttons := [][2]string{
		{"打印任务查询", "store:print-task:list"},
		{"小票重打", "store:print-task:reprint"},
	}
	for index, button := range buttons {
		id, err := r.ensureMenu(pageID, button[0], 3, index+1, "", "", "", "", button[1], now)
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
func (r *M20261023000050SeedStorePrintTaskMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, componentName, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261023000050SeedStorePrintTaskMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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

// Down 删除打印任务页面、按钮及其角色授权；保留门店管理目录。
func (r *M20261023000050SeedStorePrintTaskMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	pages := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("component", "/store/print-task").
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
