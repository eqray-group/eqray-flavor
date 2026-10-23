package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261022000029SeedDishMenu 写入「菜品管理」目录、三个页面与按钮权限点。
// 前端页面属于 T-FE-04，尚未实现，因此菜单一律以 status=0 落库并暂不授权角色：
// 这样既不会因为组件不存在导致白屏，也不会给角色配上点不进去的菜单。
// T-FE-04 完成后新增迁移：把这些菜单 status 置 1，并补 system_role_menu 授权。
type M20261022000029SeedDishMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261022000029SeedDishMenu) Signature() string {
	return "20261022000029_seed_dish_menu"
}

// Up 注册菜品管理菜单与权限点（停用状态，待前端页面就绪后启用）。
func (r *M20261022000029SeedDishMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	dirID, err := r.ensureMenu(0, "菜品管理", 1, 6, "/dish", "", "", "ri:restaurant-2-line", "", now)
	if err != nil {
		return err
	}

	categoryPageID, err := r.ensureMenu(dirID, "菜品分类", 2, 1, "category", "/dish/category", "DishCategory", "ri:node-tree", "", now)
	if err != nil {
		return err
	}
	spuPageID, err := r.ensureMenu(dirID, "菜品管理", 2, 2, "spu", "/dish/spu", "DishSpu", "ri:dish-line", "", now)
	if err != nil {
		return err
	}
	addonPageID, err := r.ensureMenu(dirID, "加料管理", 2, 3, "addon", "/dish/addon", "DishAddon", "ri:cup-line", "", now)
	if err != nil {
		return err
	}

	categoryButtons := [][2]string{
		{"分类查询", "dish:category:list"},
		{"分类新增", "dish:category:create"},
		{"分类修改", "dish:category:update"},
		{"分类删除", "dish:category:delete"},
	}
	for index, button := range categoryButtons {
		if _, err := r.ensureMenu(categoryPageID, button[0], 3, index+1, "", "", "", "", button[1], now); err != nil {
			return err
		}
	}

	spuButtons := [][2]string{
		{"菜品查询", "dish:spu:list"},
		{"菜品新增", "dish:spu:create"},
		{"菜品修改", "dish:spu:update"},
		{"菜品删除", "dish:spu:delete"},
	}
	for index, button := range spuButtons {
		if _, err := r.ensureMenu(spuPageID, button[0], 3, index+1, "", "", "", "", button[1], now); err != nil {
			return err
		}
	}

	addonButtons := [][2]string{
		{"加料查询", "dish:addon:list"},
		{"加料新增", "dish:addon:create"},
		{"加料修改", "dish:addon:update"},
		{"加料删除", "dish:addon:delete"},
	}
	for index, button := range addonButtons {
		if _, err := r.ensureMenu(addonPageID, button[0], 3, index+1, "", "", "", "", button[1], now); err != nil {
			return err
		}
	}

	// 权限点入库存后前移缓存版本，避免管理员等待 30 分钟
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时以停用状态插入。
// 入参：父级、名称、类型、排序、前端路径、组件、路由名、图标、权限标识、时间；返回菜单 ID。
func (r *M20261022000029SeedDishMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, componentName, icon, permission string, now time.Time) (types.ID, error) {
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
		"status":         0,
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

// Down 删除菜品管理目录及其页面、按钮；按钮权限随菜单一起移除。
func (r *M20261022000029SeedDishMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var dir model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 1).
		Where("path", "/dish").
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
			if facades.Schema().HasTable("system_role_menu") {
				if _, err := facades.DB().Table("system_role_menu").Where("menu_id", button.ID).Delete(); err != nil {
					return err
				}
			}
			if _, err := facades.DB().Table("system_menu").Where("id", button.ID).Delete(); err != nil {
				return err
			}
		}

		if facades.Schema().HasTable("system_role_menu") {
			if _, err := facades.DB().Table("system_role_menu").Where("menu_id", page.ID).Delete(); err != nil {
				return err
			}
		}
		if _, err := facades.DB().Table("system_menu").Where("id", page.ID).Delete(); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").Where("menu_id", dir.ID).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", dir.ID).Delete()

	return err
}
