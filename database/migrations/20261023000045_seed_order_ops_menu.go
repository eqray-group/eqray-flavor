package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000045SeedOrderOpsMenu 给「订单管理」补备注、改价、派单三个按钮权限点。
// 三个动作分别涉及文案、金额与配送，权限拆开授权，避免只读角色直接改钱或改派单。
type M20261023000045SeedOrderOpsMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000045SeedOrderOpsMenu) Signature() string {
	return "20261023000045_seed_order_ops_menu"
}

// Up 注册三个按钮权限点并授权给内置角色。
func (r *M20261023000045SeedOrderOpsMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	// 订单页面的 path 是 index、component 才是 /order，按 component 定位才不会漏
	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 2).
		Where("component", "/order").
		WhereNull("deleted_at").
		First(&page); err != nil {
		return err
	}
	if page.ID == 0 {
		return nil
	}

	now := time.Now()

	buttons := [][2]string{
		{"修改订单备注", "order:order:remark"},
		{"订单改价", "order:order:adjust"},
		{"订单派单", "order:order:dispatch"},
	}

	menuIDs := make([]types.ID, 0, len(buttons))
	for index, button := range buttons {
		id, err := r.ensureMenu(page.ID, button[0], 3, 4+index, button[1], now)
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

// ensureMenu 按「父级 + 类型 + 名称」查找按钮，不存在时插入。
// 入参：父级、名称、类型、排序、权限标识与时间；返回菜单 ID。
func (r *M20261023000045SeedOrderOpsMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, permission string, now time.Time) (types.ID, error) {
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
		"path":           "",
		"icon":           "",
		"component":      "",
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

// grantBuiltinRoles 把本次新增的按钮授权给内置角色，已授权的组合跳过。
func (r *M20261023000045SeedOrderOpsMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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
		rows := make([]model.SystemRoleMenuModel, 0)
		if err := facades.DB().Table("system_role_menu").
			Select("menu_id").
			Where("role_id", role.ID).
			WhereNull("deleted_at").
			Get(&rows); err != nil {
			return err
		}

		granted := make(map[types.ID]bool, len(rows))
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

// Down 删除本次新增的三个按钮及其角色授权。
func (r *M20261023000045SeedOrderOpsMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	buttons := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("permission", []any{"order:order:remark", "order:order:adjust", "order:order:dispatch"}).
		WhereNull("deleted_at").
		Get(&buttons); err != nil {
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

	return nil
}
