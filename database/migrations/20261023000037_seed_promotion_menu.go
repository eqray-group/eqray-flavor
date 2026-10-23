package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000037SeedPromotionMenu 写入「营销中心」目录与三个页面及按钮权限点。
// 页面（优惠券模板 / 会员券 / 营销活动）随 promotion 模块前端同步落地，因此以 status=1 落库并授权内置角色。
type M20261023000037SeedPromotionMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000037SeedPromotionMenu) Signature() string {
	return "20261023000037_seed_promotion_menu"
}

// Up 注册营销中心菜单与权限点，并授权给内置角色。
func (r *M20261023000037SeedPromotionMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	dirID, err := r.ensureMenu(0, "营销中心", 1, 6, "/promotion", "", "", "ri:gift-line", "", now)
	if err != nil {
		return err
	}

	templatePageID, err := r.ensureMenu(dirID, "优惠券模板", 2, 1, "coupon-template", "/promotion/coupon-template", "PromotionCouponTemplate", "ri:ticket-line", "", now)
	if err != nil {
		return err
	}
	couponPageID, err := r.ensureMenu(dirID, "会员券", 2, 2, "coupon", "/promotion/coupon", "PromotionCoupon", "ri:coupon-line", "", now)
	if err != nil {
		return err
	}
	activityPageID, err := r.ensureMenu(dirID, "营销活动", 2, 3, "activity", "/promotion/activity", "PromotionActivity", "ri:percent-line", "", now)
	if err != nil {
		return err
	}

	menuIDs := []types.ID{dirID, templatePageID, couponPageID, activityPageID}
	pages := []struct {
		parentID types.ID
		buttons  [][2]string
	}{
		{
			parentID: templatePageID,
			buttons: [][2]string{
				{"券模板查询", "promotion:coupon-template:list"},
				{"券模板新增", "promotion:coupon-template:create"},
				{"券模板修改", "promotion:coupon-template:update"},
				{"券模板删除", "promotion:coupon-template:delete"},
				{"发券", "promotion:coupon-template:issue"},
			},
		},
		{
			parentID: couponPageID,
			buttons: [][2]string{
				{"会员券查询", "promotion:coupon:list"},
				{"会员券作废", "promotion:coupon:revoke"},
			},
		},
		{
			parentID: activityPageID,
			buttons: [][2]string{
				{"活动查询", "promotion:activity:list"},
				{"活动新增", "promotion:activity:create"},
				{"活动修改", "promotion:activity:update"},
				{"活动删除", "promotion:activity:delete"},
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
func (r *M20261023000037SeedPromotionMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, componentName, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261023000037SeedPromotionMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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

// Down 删除营销中心目录、页面、按钮及其角色授权。
func (r *M20261023000037SeedPromotionMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var dir model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 1).
		Where("path", "/promotion").
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
func (r *M20261023000037SeedPromotionMenu) deleteMenu(id types.ID) error {
	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").Where("menu_id", id).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", id).Delete()

	return err
}
