package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

type M20261020000022SeedMemberTagGroupMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261020000022SeedMemberTagGroupMenu) Signature() string {
	return "20261020000022_seed_member_tag_group_menu"
}

// Up 补「会员标签」「会员分组」页面与会员列表的「等级调整」权限点。
func (r *M20261020000022SeedMemberTagGroupMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	memberID, err := r.ensureMenu(0, "会员管理", 1, 30, "/member", "", "ri:user-heart-line", "", now)
	if err != nil {
		return err
	}

	userPageID, err := r.ensureMenu(memberID, "会员列表", 2, 1, "user", "/member/user", "ri:user-3-line", "", now)
	if err != nil {
		return err
	}

	// 会员列表下补等级调整按钮，与编辑解耦
	if _, err := r.ensureMenu(userPageID, "等级调整", 3, 4, "", "", "", "member:user:level", now); err != nil {
		return err
	}

	pages := []struct {
		name      string
		path      string
		component string
		icon      string
		sort      int
		prefix    string
	}{
		{"会员标签", "tag", "/member/tag", "ri:price-tag-3-line", 3, "member:tag"},
		{"会员分组", "group", "/member/group", "ri:group-line", 4, "member:group"},
	}

	actions := [][2]string{
		{"查询", "list"},
		{"新增", "create"},
		{"修改", "update"},
		{"删除", "delete"},
	}

	for _, page := range pages {
		pageID, err := r.ensureMenu(memberID, page.name, 2, page.sort, page.path, page.component, page.icon, "", now)
		if err != nil {
			return err
		}

		for index, action := range actions {
			if _, err := r.ensureMenu(pageID, page.name+action[0], 3, index+1, "", "", "", page.prefix+":"+action[1], now); err != nil {
				return err
			}
		}
	}

	if err := r.grantBuiltinRoles(now); err != nil {
		return err
	}

	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
func (r *M20261020000022SeedMemberTagGroupMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261020000022SeedMemberTagGroupMenu) grantBuiltinRoles(now time.Time) error {
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

// Down 删除会员标签、会员分组页面及等级调整按钮。
func (r *M20261020000022SeedMemberTagGroupMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	names := []any{"会员标签", "会员分组", "等级调整"}
	menus := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("name", names).
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
