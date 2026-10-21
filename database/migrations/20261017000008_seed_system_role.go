package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

type M20261017000008SeedSystemRole struct{}

// Signature The unique signature for the migration.
func (r *M20261017000008SeedSystemRole) Signature() string {
	return "20261017000008_seed_system_role"
}

// Up 写入内置超管角色 SUPER_ADMIN，权限校验时按 code 直接放行（docs 9.3）。
// 已存在该编码时跳过，保证迁移可重复执行。
func (r *M20261017000008SeedSystemRole) Up() error {
	if !facades.Schema().HasTable("system_role") {
		return nil
	}

	count, err := facades.DB().Table("system_role").
		Where("code", "SUPER_ADMIN").
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()

	_, err = facades.DB().Table("system_role").Insert(map[string]any{
		"id":         support.NewID(),
		"name":       "超级管理员",
		"code":       "SUPER_ADMIN",
		"sort":       0,
		"status":     1,
		"type":       1,
		"remark":     "内置角色，拥有全部权限",
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除内置超管角色。
func (r *M20261017000008SeedSystemRole) Down() error {
	if !facades.Schema().HasTable("system_role") {
		return nil
	}

	_, err := facades.DB().Table("system_role").
		Where("code", "SUPER_ADMIN").
		Delete()

	return err
}
