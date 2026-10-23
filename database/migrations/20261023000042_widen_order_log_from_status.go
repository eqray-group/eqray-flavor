package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000042WidenOrderLogFromStatus 放宽订单日志 from_status 列的类型。
// 背景：上一个迁移（20261023000039）只把 cater_order.status 与 cater_order_log.to_status
// 改成了 smallint，漏掉了 from_status。订单日志记的是订单状态的原值，同样是百位段位
// （100 待支付 / … / 800 已退款），tinyint 上限 127 存不下，
// 状态流转到待接单之后写日志会报「Out of range value for column 'from_status'」。
type M20261023000042WidenOrderLogFromStatus struct{}

// Signature The unique signature for the migration.
func (r *M20261023000042WidenOrderLogFromStatus) Signature() string {
	return "20261023000042_widen_order_log_from_status"
}

// Up 把订单日志的 from_status 改成 smallint，与 to_status 保持一致。
func (r *M20261023000042WidenOrderLogFromStatus) Up() error {
	if !facades.Schema().HasTable("cater_order_log") {
		return nil
	}

	// 已是 smallint 时直接跳过，避免重复执行报错
	var currentType string
	if err := facades.DB().Table("information_schema.columns").
		Select("column_type").
		Where("table_schema = DATABASE()").
		Where("table_name", "cater_order_log").
		Where("column_name", "from_status").
		First(&currentType); err != nil {
		return err
	}
	if currentType == "smallint" {
		return nil
	}

	return facades.DB().Statement("ALTER TABLE cater_order_log MODIFY from_status smallint NOT NULL DEFAULT 0 COMMENT '原状态，0 表示初始创建'")
}

// Down 不回退类型收窄：已写入的百位段位状态无法用 tinyint 承载。
func (r *M20261023000042WidenOrderLogFromStatus) Down() error {
	return nil
}
