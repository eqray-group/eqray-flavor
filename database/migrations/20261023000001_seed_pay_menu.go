package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000001SeedPayMenu 写入「支付管理」目录、三个页面与按钮权限点。
// 页面（支付应用 / 支付渠道 / 支付单）前端已同步实现，因此菜单以 status=1 落库并授权内置角色；
// 退款、通知、转账、钱包等页面随对应后端完成后再补种子。
type M20261023000001SeedPayMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000001SeedPayMenu) Signature() string {
	return "20261023000001_seed_pay_menu"
}

// Up 注册支付管理菜单与权限点，并授权给内置角色。
func (r *M20261023000001SeedPayMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	// 目录已在 20261022000073 中清空，这里重建；ensureMenu 幂等，可单独重跑
	dirID, err := r.ensureMenu(0, "支付管理", 1, 7, "/pay", "", "", "ri:money-cny-circle-line", "", now)
	if err != nil {
		return err
	}

	appPageID, err := r.ensureMenu(dirID, "支付应用", 2, 1, "app", "/pay/app", "PayApp", "ri:apps-2-line", "", now)
	if err != nil {
		return err
	}
	channelPageID, err := r.ensureMenu(dirID, "支付渠道", 2, 2, "channel", "/pay/channel", "PayChannel", "ri:links-line", "", now)
	if err != nil {
		return err
	}
	orderPageID, err := r.ensureMenu(dirID, "支付单", 2, 3, "order", "/pay/order", "PayOrder", "ri:file-list-3-line", "", now)
	if err != nil {
		return err
	}

	menuIDs := []types.ID{dirID, appPageID, channelPageID, orderPageID}
	pages := []struct {
		parentID types.ID
		buttons  [][2]string
	}{
		{
			parentID: appPageID,
			buttons: [][2]string{
				{"支付应用查询", "pay:app:list"},
				{"支付应用新增", "pay:app:create"},
				{"支付应用修改", "pay:app:update"},
				{"支付应用删除", "pay:app:delete"},
			},
		},
		{
			parentID: channelPageID,
			buttons: [][2]string{
				{"支付渠道查询", "pay:channel:list"},
				{"支付渠道新增", "pay:channel:create"},
				{"支付渠道修改", "pay:channel:update"},
				{"支付渠道删除", "pay:channel:delete"},
			},
		},
		{
			parentID: orderPageID,
			buttons: [][2]string{
				{"支付单查询", "pay:order:list"},
				{"支付单创建", "pay:order:create"},
				{"重新调起支付", "pay:order:submit"},
				{"关闭支付单", "pay:order:close"},
			},
		},
	}
	for _, page := range pages {
		for index, button := range page.buttons {
			id, err := r.ensureMenu(page.parentID, button[0], 3, index+1, "", "", "", "", button[1], now)
			if err != nil {
				return err
			}
			menuIDs = append(menuIDs, id)
		}
	}

	if err := r.grantBuiltinRoles(menuIDs, now); err != nil {
		return err
	}

	// 新增权限点后前移权限缓存版本，避免管理员等待 30 分钟才拿到新按钮
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
// 入参：父级、名称、类型、排序、前端路径、组件、路由名、图标、权限标识、时间；返回菜单 ID。
func (r *M20261023000001SeedPayMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, componentName, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261023000001SeedPayMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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

// Down 删除支付管理目录、页面、按钮及其角色授权。
func (r *M20261023000001SeedPayMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

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

	pages := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("parent_id", dir.ID).
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
			if err := r.deleteMenu(button.ID); err != nil {
				return err
			}
		}
		if err := r.deleteMenu(page.ID); err != nil {
			return err
		}
	}

	return r.deleteMenu(dir.ID)
}

// deleteMenu 删除单个菜单及其角色授权。
func (r *M20261023000001SeedPayMenu) deleteMenu(id types.ID) error {
	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").Where("menu_id", id).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", id).Delete()

	return err
}
