package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

type M20261019000005SeedDictAndLoginLogMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261019000005SeedDictAndLoginLogMenu) Signature() string {
	return "20261019000005_seed_dict_and_login_log_menu"
}

// Up 在系统管理下新增「字典管理」「登录日志」菜单及按钮权限点，并给内置角色授权。
func (r *M20261019000005SeedDictAndLoginLogMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var system model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("path", "/system").
		Where("type", 1).
		WhereNull("deleted_at").
		First(&system); err != nil {
		return err
	}
	if system.ID == 0 {
		return nil
	}

	now := time.Now()

	dictID, err := r.ensureMenu(system.ID, "字典管理", 2, 40, "dict", "/system/dict", "ri:book-2-line", "", now)
	if err != nil {
		return err
	}
	dictButtons := []struct {
		name       string
		permission string
		sort       int
	}{
		{"字典查询", "system:dict:list", 1},
		{"字典新增", "system:dict:create", 2},
		{"字典修改", "system:dict:update", 3},
		{"字典删除", "system:dict:delete", 4},
	}
	for _, button := range dictButtons {
		if _, err := r.ensureMenu(dictID, button.name, 3, button.sort, "", "", "", button.permission, now); err != nil {
			return err
		}
	}

	logID, err := r.ensureMenu(system.ID, "登录日志", 2, 50, "login-log", "/system/login-log", "ri:history-line", "", now)
	if err != nil {
		return err
	}
	if _, err := r.ensureMenu(logID, "日志查询", 3, 1, "", "", "", "system:login-log:list", now); err != nil {
		return err
	}

	return r.grantBuiltinRoles(now)
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
func (r *M20261019000005SeedDictAndLoginLogMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
	var existing model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
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

// grantBuiltinRoles 给内置角色补齐全部菜单授权。
func (r *M20261019000005SeedDictAndLoginLogMenu) grantBuiltinRoles(now time.Time) error {
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

// Down 删除本迁移新增的菜单与授权。
func (r *M20261019000005SeedDictAndLoginLogMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	menus := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("name", []any{"字典管理", "登录日志"}).
		WhereIn("type", []any{2}).
		WhereNull("deleted_at").
		Get(&menus); err != nil {
		return err
	}

	for _, menu := range menus {
		if facades.Schema().HasTable("system_role_menu") {
			if _, err := facades.DB().Table("system_role_menu").Where("menu_id", menu.ID).Delete(); err != nil {
				return err
			}
		}

		children := make([]model.SystemMenuModel, 0)
		if err := facades.DB().Table("system_menu").
			Select("id").
			Where("parent_id", menu.ID).
			Get(&children); err != nil {
			return err
		}
		for _, child := range children {
			if facades.Schema().HasTable("system_role_menu") {
				if _, err := facades.DB().Table("system_role_menu").Where("menu_id", child.ID).Delete(); err != nil {
					return err
				}
			}
			if _, err := facades.DB().Table("system_menu").Where("id", child.ID).Delete(); err != nil {
				return err
			}
		}

		if _, err := facades.DB().Table("system_menu").Where("id", menu.ID).Delete(); err != nil {
			return err
		}
	}

	return nil
}
