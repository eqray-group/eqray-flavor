package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000060DropPayTablesBeforeYudao 清空旧版支付模块的全部表。
// 支付模块整体切换为 yudao（ruoyi-vue-pro）结构：旧的 pay_channel / pay_order / pay_refund /
// pay_wallet / pay_wallet_flow / pay_recharge_plan 数据结构与 yudao 不兼容，直接删除重建。
// 当前无生产数据，不迁移历史数据。
type M20261022000060DropPayTablesBeforeYudao struct{}

// Signature The unique signature for the migration.
func (r *M20261022000060DropPayTablesBeforeYudao) Signature() string {
	return "20261022000060_drop_pay_tables_before_yudao"
}

// Up 依次删除旧版支付模块的六张表。
func (r *M20261022000060DropPayTablesBeforeYudao) Up() error {
	tables := []string{
		"pay_wallet_flow",
		"pay_wallet",
		"pay_recharge_plan",
		"pay_refund",
		"pay_order",
		"pay_channel",
	}

	for _, table := range tables {
		if err := facades.Schema().DropIfExists(table); err != nil {
			return err
		}
	}

	return nil
}

// Down 不恢复旧表结构，数据已不可用。
func (r *M20261022000060DropPayTablesBeforeYudao) Down() error {
	return nil
}
