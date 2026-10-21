package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

type M20261020000013SeedInfraFileConfig struct{}

// Signature The unique signature for the migration.
func (r *M20261020000013SeedInfraFileConfig) Signature() string {
	return "20261020000013_seed_infra_file_config"
}

// Up 写入默认的本地磁盘文件配置，并标记为主配置，保证首次部署即可上传文件。
// 已有任意配置时直接跳过，避免覆盖用户维护的数据。
func (r *M20261020000013SeedInfraFileConfig) Up() error {
	if !facades.Schema().HasTable("infra_file_config") {
		return nil
	}

	total, err := facades.DB().Table("infra_file_config").
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	now := time.Now()

	_, err = facades.DB().Table("infra_file_config").Insert(map[string]any{
		"id":         support.NewID(),
		"name":       "本地存储",
		"storage":    "local",
		"master":     1,
		"config":     `{"dir":"uploads"}`,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除本迁移写入的默认配置。
func (r *M20261020000013SeedInfraFileConfig) Down() error {
	if !facades.Schema().HasTable("infra_file_config") {
		return nil
	}

	_, err := facades.DB().Table("infra_file_config").
		Where("name", "本地存储").
		Where("storage", "local").
		Where("master", 1).
		Delete()

	return err
}
