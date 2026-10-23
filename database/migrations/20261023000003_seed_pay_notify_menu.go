package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000003SeedPayNotifyMenu 在「支付管理」目录下补「回调通知」页面与查询按钮权限点。
// 通知任务由支付 / 退款成功时自动登记，派发靠定时任务，页面只读。
type M20261023000003SeedPayNotifyMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000003SeedPayNotifyMenu) Signature() string {
	return "20261023000003_seed_pay_notify_menu"
}

// Up 注册回调通知页面与权限点，并授权给内置角色。
func (r *M20261023000003SeedPayNotifyMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	var dir model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 1).
		Where("path", "/pay").
		WhereNull("deleted_at").
		First(&dir); err != nil {
		return err
	}
	if dir.ID == 0 {
		return nil
	}

	pageID, err := r.ensureMenu(dir.ID, "回调通知", 2, 5, "/pay/notify", "/pay/notify", "PayNotify", "ri:notification-3-line", "", now)
	if err != nil {
		return err
	}

	menuIDs := []types.ID{pageID}
	buttonID, err := r.ensureMenu(pageID, "通知任务查询", 3, 1, "", "", "", "", "pay:notify:list", now)
	if err != nil {
		return err
	}
	menuIDs = append(menuIDs, buttonID)

	if err := r.grantBuiltinRoles(menuIDs, now); err != nil {
		return err
	}

	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
// 入参：父级、名称、类型、排序、路径、组件、路由名、图标、权限标识、时间；返回菜单 ID。
func (r *M20261023000003SeedPayNotifyMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, componentName, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261023000003SeedPayNotifyMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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

// Down 删除回调通知页面、按钮及其角色授权。
func (r *M20261023000003SeedPayNotifyMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 2).
		Where("path", "/pay/notify").
		WhereNull("deleted_at").
		First(&page); err != nil {
		return err
	}
	if page.ID == 0 {
		return nil
	}

	buttons := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", page.ID).Get(&buttons); err != nil {
		return err
	}
	for _, button := range buttons {
		if err := r.deleteMenu(button.ID); err != nil {
			return err
		}
	}

	return r.deleteMenu(page.ID)
}

// deleteMenu 删除单个菜单及其角色授权。
func (r *M20261023000003SeedPayNotifyMenu) deleteMenu(id types.ID) error {
	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").Where("menu_id", id).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", id).Delete()

	return err
}
