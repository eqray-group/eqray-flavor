package migrations

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

type M20261017000007SeedSystemUsers struct{}

// Signature The unique signature for the migration.
func (r *M20261017000007SeedSystemUsers) Signature() string {
	return "20261017000007_seed_system_users"
}

// Up 写入初始管理员 admin / admin123，上线前必须修改密码。
// 已存在同名账号时跳过，保证迁移可重复执行。
func (r *M20261017000007SeedSystemUsers) Up() error {
	if !facades.Schema().HasTable("system_users") {
		return nil
	}

	count, err := facades.DB().Table("system_users").
		Where("username", "admin").
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()

	_, err = facades.DB().Table("system_users").Insert(map[string]any{
		"id":         support.NewID(),
		"username":   "admin",
		"password":   string(hash),
		"nickname":   "超级管理员",
		"status":     1,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除初始管理员。
func (r *M20261017000007SeedSystemUsers) Down() error {
	if !facades.Schema().HasTable("system_users") {
		return nil
	}

	_, err := facades.DB().Table("system_users").
		Where("username", "admin").
		Delete()

	return err
}
