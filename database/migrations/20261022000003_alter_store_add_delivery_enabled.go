package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000003AlterStoreAddDeliveryEnabled 为 store 表补充外卖开关。
// 默认关闭，与 design/24 §1「外卖默认关，需单独开启」一致；重复执行不会重复加列。
type M20261022000003AlterStoreAddDeliveryEnabled struct{}

// Signature The unique signature for the migration.
func (r *M20261022000003AlterStoreAddDeliveryEnabled) Signature() string {
	return "20261022000003_alter_store_add_delivery_enabled"
}

// Up 在 online_enabled 之后新增 delivery_enabled 字段。
func (r *M20261022000003AlterStoreAddDeliveryEnabled) Up() error {
	if !facades.Schema().HasTable("store") {
		return nil
	}
	if facades.Schema().HasColumn("store", "delivery_enabled") {
		return nil
	}

	return facades.DB().Statement(`
ALTER TABLE store
  ADD COLUMN delivery_enabled tinyint NOT NULL DEFAULT 0
  COMMENT '是否开启外卖：1 是 / 0 否；默认关闭，关闭时顾客端不展示外卖入口'
  AFTER online_enabled`)
}

// Down 删除 delivery_enabled 字段。
func (r *M20261022000003AlterStoreAddDeliveryEnabled) Down() error {
	if !facades.Schema().HasTable("store") {
		return nil
	}
	if !facades.Schema().HasColumn("store", "delivery_enabled") {
		return nil
	}

	return facades.DB().Statement("ALTER TABLE store DROP COLUMN delivery_enabled")
}
