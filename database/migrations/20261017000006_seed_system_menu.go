package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
)

type M20261017000006SeedSystemMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261017000006SeedSystemMenu) Signature() string {
	return "20261017000006_seed_system_menu"
}

// Up 写入后台基础菜单与按钮权限点，保证角色授权页与权限校验开箱可用。
// 已有菜单数据时直接跳过，避免覆盖使用者自己维护的菜单。
func (r *M20261017000006SeedSystemMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	count, err := facades.DB().Table("system_menu").Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()

	insert := func(parentID types.ID, name string, menuType int, sort int, path, component, permission string) (types.ID, error) {
		id := support.NewID()

		_, err := facades.DB().Table("system_menu").Insert(map[string]any{
			"id":             id,
			"name":           name,
			"permission":     permission,
			"type":           menuType,
			"sort":           sort,
			"parent_id":      parentID,
			"path":           path,
			"icon":           "",
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

	systemID, err := insert(0, "系统管理", 1, 1, "/system", "", "")
	if err != nil {
		return err
	}

	userID, err := insert(systemID, "用户管理", 2, 1, "user", "/system/user", "")
	if err != nil {
		return err
	}
	roleID, err := insert(systemID, "角色管理", 2, 2, "role", "/system/role", "")
	if err != nil {
		return err
	}
	menuID, err := insert(systemID, "菜单管理", 2, 3, "menu", "/system/menu", "")
	if err != nil {
		return err
	}

	buttons := []struct {
		parent     types.ID
		name       string
		permission string
	}{
		{userID, "查询用户", "system:user:list"},
		{userID, "新增用户", "system:user:create"},
		{userID, "编辑用户", "system:user:update"},
		{userID, "删除用户", "system:user:delete"},
		{userID, "重置密码", "system:user:reset-password"},
		{userID, "分配角色", "system:user:assign-role"},
		{roleID, "查询角色", "system:role:list"},
		{roleID, "新增角色", "system:role:create"},
		{roleID, "编辑角色", "system:role:update"},
		{roleID, "删除角色", "system:role:delete"},
		{roleID, "角色授权", "system:role:assign-menu"},
		{menuID, "查询菜单", "system:menu:list"},
		{menuID, "新增菜单", "system:menu:create"},
		{menuID, "编辑菜单", "system:menu:update"},
		{menuID, "删除菜单", "system:menu:delete"},
	}

	for i, button := range buttons {
		if _, err := insert(button.parent, button.name, 3, i+1, "", "", button.permission); err != nil {
			return err
		}
	}

	return nil
}

// Down 仅删除本迁移写入的菜单（初始化数据），表结构由 0002 负责。
func (r *M20261017000006SeedSystemMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	_, err := facades.DB().Table("system_menu").Where("creator", "system").Delete()

	return err
}
