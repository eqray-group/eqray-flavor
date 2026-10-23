package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261022000031SeedDishStoreSkuMenu 在「菜品管理」页面下补门店可售与门店价的按钮权限点。
// 与 20261022000029 一致，按钮以 status=0 落库并暂不授权：前端页面属 T-FE-04，
// 页面就绪后再统一启用并授权，避免出现点不进去的菜单与按钮。
type M20261022000031SeedDishStoreSkuMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261022000031SeedDishStoreSkuMenu) Signature() string {
	return "20261022000031_seed_dish_store_sku_menu"
}

// Up 幂等写入「门店菜品查询」「门店菜品配置」两个按钮权限点。
func (r *M20261022000031SeedDishStoreSkuMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 2).
		Where("component", "/dish/spu").
		WhereNull("deleted_at").
		First(&page); err != nil {
		return err
	}
	if page.ID == 0 {
		return nil
	}

	now := time.Now()

	buttons := [][2]string{
		{"门店菜品查询", "dish:store-sku:list"},
		{"门店菜品配置", "dish:store-sku:update"},
	}
	for index, button := range buttons {
		if err := r.ensureButton(page.ID, button[0], button[1], index+1, now); err != nil {
			return err
		}
	}

	return nil
}

// ensureButton 按「父级 + 类型 + 名称」查找按钮，不存在时以停用状态插入。
// 入参：父级菜单、名称、权限标识、排序、时间；返回错误表示写入失败。
func (r *M20261022000031SeedDishStoreSkuMenu) ensureButton(parentID types.ID, name, permission string, sort int, now time.Time) error {
	var existing model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("parent_id", parentID).
		Where("type", 3).
		Where("name", name).
		WhereNull("deleted_at").
		First(&existing); err != nil {
		return err
	}
	if existing.ID > 0 {
		return nil
	}

	_, err := facades.DB().Table("system_menu").Insert(map[string]any{
		"id":             support.NewID(),
		"name":           name,
		"permission":     permission,
		"type":           3,
		"sort":           sort,
		"parent_id":      parentID,
		"path":           "",
		"icon":           "",
		"component":      "",
		"component_name": "",
		"status":         0,
		"visible":        1,
		"keep_alive":     1,
		"always_show":    0,
		"creator":        "system",
		"created_at":     now,
		"updater":        "system",
		"updated_at":     now,
	})

	return err
}

// Down 删除本迁移写入的两个按钮及其角色授权。
func (r *M20261022000031SeedDishStoreSkuMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	rows := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("permission", []any{"dish:store-sku:list", "dish:store-sku:update"}).
		Where("type", 3).
		Get(&rows); err != nil {
		return err
	}

	for _, row := range rows {
		if facades.Schema().HasTable("system_role_menu") {
			if _, err := facades.DB().Table("system_role_menu").Where("menu_id", row.ID).Delete(); err != nil {
				return err
			}
		}
		if _, err := facades.DB().Table("system_menu").Where("id", row.ID).Delete(); err != nil {
			return err
		}
	}

	return nil
}
