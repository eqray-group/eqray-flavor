package migrations

import (
	"strconv"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

// M20261023000046SeedMemberAddressMenu 给「会员列表」补「维护收货地址」按钮权限点。
// 背景：收货地址此前只有查询接口，电话改址、代客下单都需要后台能增删改。
type M20261023000046SeedMemberAddressMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261023000046SeedMemberAddressMenu) Signature() string {
	return "20261023000046_seed_member_address_menu"
}

// Up 注册按钮权限点并授权给内置角色。
func (r *M20261023000046SeedMemberAddressMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	// 会员列表页面的 path 是 user、component 才是 /member/user，按 component 定位
	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 2).
		Where("component", "/member/user").
		WhereNull("deleted_at").
		First(&page); err != nil {
		return err
	}
	if page.ID == 0 {
		return nil
	}

	now := time.Now()

	menuID, err := r.ensureMenu(page.ID, "维护收货地址", 3, 10, "member:user:address", now)
	if err != nil {
		return err
	}

	if err := r.grantBuiltinRoles(menuID, now); err != nil {
		return err
	}

	// 新增权限点后前移权限缓存版本，避免管理员等待 30 分钟才拿到新按钮
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找按钮，不存在时插入。
func (r *M20261023000046SeedMemberAddressMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, permission string, now time.Time) (types.ID, error) {
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

// grantBuiltinRoles 把按钮授权给内置角色，已授权的角色跳过。
func (r *M20261023000046SeedMemberAddressMenu) grantBuiltinRoles(menuID types.ID, now time.Time) error {
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
		var granted model.SystemRoleMenuModel
		if err := facades.DB().Table("system_role_menu").
			Where("role_id", role.ID).
			Where("menu_id", menuID).
			WhereNull("deleted_at").
			First(&granted); err != nil {
			return err
		}
		if granted.ID > 0 {
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

	return nil
}

// Down 删除按钮及其角色授权。
func (r *M20261023000046SeedMemberAddressMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var button model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Where("type", 3).
		Where("permission", "member:user:address").
		WhereNull("deleted_at").
		First(&button); err != nil {
		return err
	}
	if button.ID == 0 {
		return nil
	}

	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").Where("menu_id", button.ID).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", button.ID).Delete()

	return err
}
