package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

type M20261020000005SeedInfraConfigMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261020000005SeedInfraConfigMenu) Signature() string {
	return "20261020000005_seed_infra_config_menu"
}

// Up 新增「基础设施 / 参数配置」菜单与按钮权限点，并写入内置参数。
func (r *M20261020000005SeedInfraConfigMenu) Up() error {
	now := time.Now()

	if err := r.seedConfigs(now); err != nil {
		return err
	}

	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	infraID, err := r.ensureMenu(0, "基础设施", 1, 20, "/infra", "", "ri:settings-4-line", "", now)
	if err != nil {
		return err
	}

	configID, err := r.ensureMenu(infraID, "参数配置", 2, 1, "config", "/infra/config", "ri:settings-3-line", "", now)
	if err != nil {
		return err
	}

	buttons := []struct {
		name       string
		permission string
		sort       int
	}{
		{"参数查询", "infra:config:list", 1},
		{"参数新增", "infra:config:create", 2},
		{"参数修改", "infra:config:update", 3},
		{"参数删除", "infra:config:delete", 4},
	}
	for _, button := range buttons {
		if _, err := r.ensureMenu(configID, button.name, 3, button.sort, "", "", "", button.permission, now); err != nil {
			return err
		}
	}

	return r.grantBuiltinRoles(now)
}

// seedConfigs 写入内置参数；已存在则跳过，保证重复迁移幂等。
func (r *M20261020000005SeedInfraConfigMenu) seedConfigs(now time.Time) error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	presets := []struct {
		category string
		name     string
		key      string
		value    string
		visible  int
		remark   string
	}{
		{"文件", "上传大小上限（MB）", "infra.upload.max_size", "5", 1, "单个上传文件的体积上限，单位 MB"},
		{"登录", "验证码开关", "system.captcha.enabled", "1", 1, "登录是否校验验证码：1 开启 / 0 关闭（预留）"},
		{"通用", "默认分页大小", "system.page.size", "20", 1, "列表默认每页条数（预留，前端读取后生效）"},
	}

	for _, preset := range presets {
		total, err := facades.DB().Table("infra_config").
			Where("config_key", preset.key).
			Count()
		if err != nil {
			return err
		}
		if total > 0 {
			continue
		}

		if _, err := facades.DB().Table("infra_config").Insert(map[string]any{
			"id":         support.NewID(),
			"category":   preset.category,
			"name":       preset.name,
			"config_key": preset.key,
			"value":      preset.value,
			"type":       1,
			"visible":    preset.visible,
			"remark":     preset.remark,
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

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
func (r *M20261020000005SeedInfraConfigMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261020000005SeedInfraConfigMenu) grantBuiltinRoles(now time.Time) error {
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

// Down 删除「基础设施」菜单及其子节点授权。
func (r *M20261020000005SeedInfraConfigMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	infraMenus := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("path", "/infra").
		Where("type", 1).
		WhereNull("deleted_at").
		Get(&infraMenus); err != nil {
		return err
	}

	for _, root := range infraMenus {
		children := make([]model.SystemMenuModel, 0)
		if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", root.ID).Get(&children); err != nil {
			return err
		}

		for _, child := range children {
			grandchildren := make([]model.SystemMenuModel, 0)
			if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", child.ID).Get(&grandchildren); err != nil {
				return err
			}
			for _, leaf := range grandchildren {
				if facades.Schema().HasTable("system_role_menu") {
					if _, err := facades.DB().Table("system_role_menu").Where("menu_id", leaf.ID).Delete(); err != nil {
						return err
					}
				}
				if _, err := facades.DB().Table("system_menu").Where("id", leaf.ID).Delete(); err != nil {
					return err
				}
			}

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
			if _, err := facades.DB().Table("system_role_menu").Where("menu_id", root.ID).Delete(); err != nil {
				return err
			}
		}
		if _, err := facades.DB().Table("system_menu").Where("id", root.ID).Delete(); err != nil {
			return err
		}
	}

	return nil
}
