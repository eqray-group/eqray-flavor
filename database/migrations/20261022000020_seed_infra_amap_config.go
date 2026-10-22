package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261022000020SeedInfraAmapConfig 写入高德地图前端所需的两个公开参数。
// 值留空，由管理员在「基础设施 → 参数配置」里填写，避免把 AK 硬编码进代码或前端环境变量。
type M20261022000020SeedInfraAmapConfig struct{}

// Signature The unique signature for the migration.
func (r *M20261022000020SeedInfraAmapConfig) Signature() string {
	return "20261022000020_seed_infra_amap_config"
}

// Up 按参数键幂等插入高德地图 Key 与安全密钥。
func (r *M20261022000020SeedInfraAmapConfig) Up() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	now := time.Now()

	items := []struct {
		name    string
		key     string
		remark  string
		visible int8
	}{
		{
			name:    "高德地图 Key",
			key:     "amap_js_key",
			remark:  "前端地图选点使用，需在高德开放平台申请 Web 端（JS API）Key",
			visible: 1,
		},
		{
			name:    "高德地图安全密钥",
			key:     "amap_security_code",
			remark:  "JS API 2.0 的安全密钥；未开启安全密钥时留空",
			visible: 1,
		},
	}

	for _, item := range items {
		exists, err := r.exists(item.key)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		if _, err := facades.DB().Table("infra_config").Insert(map[string]any{
			"id":         support.NewID(),
			"category":   "通用",
			"name":       item.name,
			"config_key": item.key,
			"value":      "",
			"type":       2,
			"visible":    item.visible,
			"remark":     item.remark,
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

// exists 判断参数键是否已存在，保证迁移可重复执行。
func (r *M20261022000020SeedInfraAmapConfig) exists(key string) (bool, error) {
	total, err := facades.DB().Table("infra_config").
		Where("config_key", key).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return false, err
	}

	return total > 0, nil
}

// Down 删除本次写入的两个参数；管理员改过的值一并清理。
func (r *M20261022000020SeedInfraAmapConfig) Down() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	_, err := facades.DB().Table("infra_config").
		WhereIn("config_key", []any{"amap_js_key", "amap_security_code"}).
		Delete()

	return err
}
