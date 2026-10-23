package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000039WidenOrderStatusColumns 放宽订单状态列的类型。
// 背景：订单状态用百位段位（100 待支付 / 200 待接单 / … / 800 已退款），
// 建表迁移误用 tinyint（上限 127），写入 200 及以上会报
// 「Out of range value for column 'status'」，状态流转在待接单之后必然失败。
// 模型侧 enums.OrderStatus 是 int16，这里同步把列改成 smallint。
type M20261023000039WidenOrderStatusColumns struct{}

// Signature The unique signature for the migration.
func (r *M20261023000039WidenOrderStatusColumns) Signature() string {
	return "20261023000039_widen_order_status_columns"
}

// Up 把订单主表与订单日志的状态列改成 smallint。
func (r *M20261023000039WidenOrderStatusColumns) Up() error {
	if !facades.Schema().HasTable("cater_order") {
		return nil
	}

	// 已是 smallint 时直接跳过，避免重复执行报错
	var currentType string
	if err := facades.DB().Table("information_schema.columns").
		Select("column_type").
		Where("table_schema = DATABASE()").
		Where("table_name", "cater_order").
		Where("column_name", "status").
		First(&currentType); err != nil {
		return err
	}
	if currentType == "smallint" {
		return nil
	}

	if err := facades.DB().Statement("ALTER TABLE cater_order MODIFY status smallint NOT NULL DEFAULT 100 COMMENT '订单状态：100 待支付 / 200 待接单 / 300 制作中 / 400 待取餐 / 500 已完成 / 600 已取消 / 700 退款中 / 800 已退款'"); err != nil {
		return err
	}

	if !facades.Schema().HasTable("cater_order_log") {
		return nil
	}

	var logType string
	if err := facades.DB().Table("information_schema.columns").
		Select("column_type").
		Where("table_schema = DATABASE()").
		Where("table_name", "cater_order_log").
		Where("column_name", "to_status").
		First(&logType); err != nil {
		return err
	}
	if logType == "smallint" {
		return nil
	}

	return facades.DB().Statement("ALTER TABLE cater_order_log MODIFY to_status smallint NOT NULL COMMENT '新状态'")
}

// Down 不回退类型收窄：已写入的百位段位状态无法用 tinyint 承载。
func (r *M20261023000039WidenOrderStatusColumns) Down() error {
	return nil
}
