package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261022000019SeedInfraDistrictMenu 补「行政区划」后台页面与按钮权限点。
type M20261022000019SeedInfraDistrictMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261022000019SeedInfraDistrictMenu) Signature() string {
	return "20261022000019_seed_infra_district_menu"
}

// Up 写入行政区划菜单并授权给内置角色。
func (r *M20261022000019SeedInfraDistrictMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	// 挂在既有「基础设施」目录下，与文件、短信等页面同级
	infraID, err := r.ensureMenu(0, "基础设施", 1, 20, "/infra", "", "ri:settings-4-line", "", now)
	if err != nil {
		return err
	}

	pageID, err := r.ensureMenu(infraID, "行政区划", 2, 8, "district", "/infra/district", "ri:map-pin-2-line", "", now)
	if err != nil {
		return err
	}

	actions := [][2]string{
		{"查询", "list"},
		{"新增", "create"},
		{"修改", "update"},
		{"删除", "delete"},
	}
	for index, action := range actions {
		if _, err := r.ensureMenu(pageID, "行政区划"+action[0], 3, index+1, "", "", "", "infra:district:"+action[1], now); err != nil {
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
func (r *M20261022000019SeedInfraDistrictMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261022000019SeedInfraDistrictMenu) grantBuiltinRoles(now time.Time) error {
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

// Down 删除行政区划页面、按钮与授权。
func (r *M20261022000019SeedInfraDistrictMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	menus := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("name", "行政区划").
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
