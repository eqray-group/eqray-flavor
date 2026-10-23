package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261022000032EnableDishMenu 启用「菜品管理」菜单树并把权限点授权给内置角色。
// 20261022000029 与 20261022000031 为了避开前端页面缺失导致的白屏，把菜单与按钮以
// status=0 落库且未授权；管理端页面就绪后由本迁移统一置为启用并补授权。
type M20261022000032EnableDishMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261022000032EnableDishMenu) Signature() string {
	return "20261022000032_enable_dish_menu"
}

// Up 启用菜品管理目录、三个页面与其下按钮，并授权给内置角色。
func (r *M20261022000032EnableDishMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	ids, err := r.dishMenuIDs()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()

	// 目录、页面、按钮一并启用，避免出现「能点进去但按钮全隐藏」的半开状态
	if _, err := facades.DB().Table("system_menu").
		WhereIn("id", idArgs(ids)).
		Update(map[string]any{
			"status":     1,
			"updated_at": now,
			"updater":    "system",
		}); err != nil {
		return err
	}

	if err := r.grantBuiltinRoles(ids, now); err != nil {
		return err
	}

	// 菜单状态与授权入库存后前移缓存版本，避免管理员等待 30 分钟
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// Down 把菜品管理菜单树恢复为停用状态，并收回内置角色的授权。
func (r *M20261022000032EnableDishMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	ids, err := r.dishMenuIDs()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").
			WhereIn("menu_id", idArgs(ids)).
			Delete(); err != nil {
			return err
		}
	}

	if _, err := facades.DB().Table("system_menu").
		WhereIn("id", idArgs(ids)).
		Update(map[string]any{
			"status":     0,
			"updated_at": time.Now(),
			"updater":    "system",
		}); err != nil {
		return err
	}

	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// dishMenuIDs 收集菜品管理目录、页面与按钮的 ID，供启用与授权复用。
// 返回：按「目录 -> 页面 -> 按钮」顺序排列的菜单 ID 列表，目录不存在时返回空列表。
func (r *M20261022000032EnableDishMenu) dishMenuIDs() ([]types.ID, error) {
	var dir model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("type", 1).
		Where("path", "/dish").
		WhereNull("deleted_at").
		First(&dir); err != nil {
		return nil, err
	}
	if dir.ID == 0 {
		return nil, nil
	}

	pages := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("parent_id", dir.ID).
		Where("type", 2).
		WhereNull("deleted_at").
		Get(&pages); err != nil {
		return nil, err
	}

	ids := make([]types.ID, 0, 1+len(pages)*5)
	ids = append(ids, dir.ID)

	pageIDs := make([]any, 0, len(pages))
	for _, page := range pages {
		ids = append(ids, page.ID)
		pageIDs = append(pageIDs, page.ID)
	}

	if len(pageIDs) == 0 {
		return ids, nil
	}

	buttons := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereIn("parent_id", pageIDs).
		Where("type", 3).
		WhereNull("deleted_at").
		Get(&buttons); err != nil {
		return nil, err
	}

	for _, button := range buttons {
		ids = append(ids, button.ID)
	}

	return ids, nil
}

// grantBuiltinRoles 给内置角色（type=1）补齐指定菜单的授权，已授权的不重复写入。
// 入参：menuIDs 待授权菜单、now 写入时间；返回错误表示写入失败。
func (r *M20261022000032EnableDishMenu) grantBuiltinRoles(menuIDs []types.ID, now time.Time) error {
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
		existing := make([]model.SystemRoleMenuModel, 0)
		if err := facades.DB().Table("system_role_menu").
			Select("menu_id").
			Where("role_id", role.ID).
			WhereNull("deleted_at").
			Get(&existing); err != nil {
			return err
		}

		granted := make(map[types.ID]bool, len(existing))
		for _, row := range existing {
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

// idArgs 把 ID 列表转换成 WhereIn 需要的 []any。
func idArgs(ids []types.ID) []any {
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	return args
}
