package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

type M20261020000012SeedInfraOpsMenus struct{}

// Signature The unique signature for the migration.
func (r *M20261020000012SeedInfraOpsMenus) Signature() string {
	return "20261020000012_seed_infra_ops_menus"
}

// Up 在「基础设施」下补齐文件配置 / 访问日志 / 错误日志 / 定时任务菜单与权限点。
func (r *M20261020000012SeedInfraOpsMenus) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	infraID, err := r.ensureMenu(0, "基础设施", 1, 20, "/infra", "", "ri:settings-4-line", "", now)
	if err != nil {
		return err
	}

	pages := []struct {
		name      string
		path      string
		component string
		icon      string
		sort      int
		buttons   [][2]string
	}{
		{
			"文件配置", "file-config", "/infra/file-config", "ri:database-2-line", 3,
			[][2]string{
				{"配置查询", "infra:file-config:list"},
				{"配置新增", "infra:file-config:create"},
				{"配置修改", "infra:file-config:update"},
				{"配置删除", "infra:file-config:delete"},
			},
		},
		{
			"访问日志", "api-access-log", "/infra/api-access-log", "ri:route-line", 4,
			[][2]string{
				{"日志查询", "infra:api-log:list"},
				{"日志处理", "infra:api-log:update"},
			},
		},
		{
			"错误日志", "api-error-log", "/infra/api-error-log", "ri:error-warning-line", 5,
			[][2]string{
				{"日志查询", "infra:api-log:list"},
				{"日志处理", "infra:api-log:update"},
			},
		},
	}

	for _, page := range pages {
		menuID, err := r.ensureMenu(infraID, page.name, 2, page.sort, page.path, page.component, page.icon, "", now)
		if err != nil {
			return err
		}

		for index, button := range page.buttons {
			if _, err := r.ensureMenu(menuID, button[0], 3, index+1, "", "", "", button[1], now); err != nil {
				return err
			}
		}
	}

	if err := r.grantBuiltinRoles(now); err != nil {
		return err
	}

	// 新增权限点后前移权限缓存版本，避免管理员要等 30 分钟才能拿到新按钮权限
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
func (r *M20261020000012SeedInfraOpsMenus) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261020000012SeedInfraOpsMenus) grantBuiltinRoles(now time.Time) error {
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

// Down 删除本次新增的四个页面菜单及其按钮与授权。
func (r *M20261020000012SeedInfraOpsMenus) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	names := []any{"文件配置", "访问日志", "错误日志", "定时任务"}
	menus := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("name", names).
		Where("type", 2).
		WhereNull("deleted_at").
		Get(&menus); err != nil {
		return err
	}

	for _, menu := range menus {
		children := make([]model.SystemMenuModel, 0)
		if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", menu.ID).Get(&children); err != nil {
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
