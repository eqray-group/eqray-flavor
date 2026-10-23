package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261022000030SeedDishPriceFloatConfig 写入门店价浮动区间参数。
// 值留空表示不限制；填 20 表示门店价允许偏离品牌标准价 ±20%，超出时后端直接拒绝。
type M20261022000030SeedDishPriceFloatConfig struct{}

// Signature The unique signature for the migration.
func (r *M20261022000030SeedDishPriceFloatConfig) Signature() string {
	return "20261022000030_seed_dish_price_float_config"
}

// Up 按参数键幂等插入门店价浮动区间参数。
func (r *M20261022000030SeedDishPriceFloatConfig) Up() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	const configKey = "dish_store_price_float_rate"

	total, err := facades.DB().Table("infra_config").
		Where("config_key", configKey).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	now := time.Now()

	_, err = facades.DB().Table("infra_config").Insert(map[string]any{
		"id":         support.NewID(),
		"category":   "菜品",
		"name":       "门店价浮动区间(%)",
		"config_key": configKey,
		"value":      "",
		"type":       2,
		"visible":    0,
		"remark":     "门店价允许偏离品牌标准价的百分比，留空表示不限制，例如填 20 表示上下浮动不超过 20%",
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除门店价浮动区间参数。
func (r *M20261022000030SeedDishPriceFloatConfig) Down() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	_, err := facades.DB().Table("infra_config").
		Where("config_key", "dish_store_price_float_rate").
		Delete()

	return err
}
