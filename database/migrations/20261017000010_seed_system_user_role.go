package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/system/model"
)

type M20261017000010SeedSystemUserRole struct{}

// Signature The unique signature for the migration.
func (r *M20261017000010SeedSystemUserRole) Signature() string {
	return "20261017000010_seed_system_user_role"
}

// Up 给初始管理员绑定内置超管角色，保证开箱即用。
// 依赖 20261017000007_seed_system_users 与 20261017000008_seed_system_role 的数据；
// 已存在绑定时跳过。
func (r *M20261017000010SeedSystemUserRole) Up() error {
	if !facades.Schema().HasTable("system_users") ||
		!facades.Schema().HasTable("system_role") ||
		!facades.Schema().HasTable("system_user_role") {
		return nil
	}

	var user model.SystemUserModel
	if err := facades.DB().Table("system_users").
		Select("id").
		Where("username", "admin").
		WhereNull("deleted_at").
		First(&user); err != nil {
		return err
	}

	var role model.SystemRoleModel
	if err := facades.DB().Table("system_role").
		Select("id").
		Where("code", "SUPER_ADMIN").
		WhereNull("deleted_at").
		First(&role); err != nil {
		return err
	}
	if user.ID == 0 || role.ID == 0 {
		return nil
	}

	count, err := facades.DB().Table("system_user_role").
		Where("user_id", user.ID).
		Where("role_id", role.ID).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()

	_, err = facades.DB().Table("system_user_role").Insert(map[string]any{
		"id":         support.NewID(),
		"user_id":    user.ID,
		"role_id":    role.ID,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除初始管理员与超管角色的绑定。
func (r *M20261017000010SeedSystemUserRole) Down() error {
	if !facades.Schema().HasTable("system_users") ||
		!facades.Schema().HasTable("system_role") ||
		!facades.Schema().HasTable("system_user_role") {
		return nil
	}

	var user model.SystemUserModel
	if err := facades.DB().Table("system_users").
		Select("id").
		Where("username", "admin").
		WhereNull("deleted_at").
		First(&user); err != nil {
		return err
	}

	var role model.SystemRoleModel
	if err := facades.DB().Table("system_role").
		Select("id").
		Where("code", "SUPER_ADMIN").
		WhereNull("deleted_at").
		First(&role); err != nil {
		return err
	}
	if user.ID == 0 || role.ID == 0 {
		return nil
	}

	_, err := facades.DB().Table("system_user_role").
		Where("user_id", user.ID).
		Where("role_id", role.ID).
		Delete()

	return err
}
