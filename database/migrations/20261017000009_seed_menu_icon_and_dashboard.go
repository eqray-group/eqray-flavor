package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

type M20261017000009SeedMenuIconAndDashboard struct{}

// Signature The unique signature for the migration.
func (r *M20261017000009SeedMenuIconAndDashboard) Signature() string {
	return "20261017000009_seed_menu_icon_and_dashboard"
}

// Up 补齐系统菜单的图标，并新增「控制台」作为第一个一级菜单。
// 图标统一使用 Remix Icon（ri: 前缀），前端由 ArtSvgIcon 渲染。
func (r *M20261017000009SeedMenuIconAndDashboard) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	icons := []struct {
		name string
		icon string
	}{
		{"系统管理", "ri:settings-3-line"},
		{"用户管理", "ri:user-line"},
		{"角色管理", "ri:shield-user-line"},
		{"菜单管理", "ri:menu-line"},
	}
	for _, item := range icons {
		if _, err := facades.DB().Table("system_menu").
			Where("name", item.name).
			Where("type <> ?", 3).
			WhereNull("deleted_at").
			Update(map[string]any{
				"icon":       item.icon,
				"updater":    "system",
				"updated_at": now,
			}); err != nil {
			return err
		}
	}

	dashboardID, err := r.ensureMenu(0, "控制台", 1, 0, "/dashboard", "", "ri:dashboard-line", now)
	if err != nil {
		return err
	}
	if _, err := r.ensureMenu(dashboardID, "控制台", 2, 1, "console", "/dashboard/console", "ri:pie-chart-line", now); err != nil {
		return err
	}

	return r.grantBuiltinRoles(now)
}

// ensureMenu 按「父级 + 路径 + 类型」查找菜单，不存在时插入，保证迁移可重复执行。
func (r *M20261017000009SeedMenuIconAndDashboard) ensureMenu(parentID types.ID, name string, menuType int, sort int, path, component, icon string, now time.Time) (types.ID, error) {
	var existing model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("parent_id", parentID).
		Where("path", path).
		Where("type", menuType).
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
		"permission":     "",
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

// grantBuiltinRoles 给内置角色补齐新增菜单的授权，避免控制台只有超管代码放行、数据却缺失。
func (r *M20261017000009SeedMenuIconAndDashboard) grantBuiltinRoles(now time.Time) error {
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

// Down 删除控制台菜单并还原图标。
func (r *M20261017000009SeedMenuIconAndDashboard) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var dir model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("path", "/dashboard").
		Where("type", 1).
		First(&dir); err != nil {
		return err
	}
	if dir.ID > 0 {
		if _, err := facades.DB().Table("system_menu").Where("parent_id", dir.ID).Delete(); err != nil {
			return err
		}
		if _, err := facades.DB().Table("system_menu").Where("id", dir.ID).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").
		WhereIn("name", []any{"系统管理", "用户管理", "角色管理", "菜单管理"}).
		Where("type <> ?", 3).
		Update(map[string]any{"icon": ""})

	return err
}
