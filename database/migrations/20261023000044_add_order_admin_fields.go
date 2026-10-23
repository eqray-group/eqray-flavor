package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000044AddOrderAdminFields 给订单主表补四个后台经营字段。
// 背景：订单管理此前只有查询与状态流转，缺「商家备注 / 改价 / 派单」的落库位置。
//   - merchant_remark 与顾客备注分开存，避免店员改动顾客原始诉求；
//   - adjust_amount 记改价差额（正加价 / 负减免），pay_amount 同步调整，保留差额可追溯；
//   - rider_id + dispatch_time 记录外卖派单结果，取消或完成时释放骑手。
type M20261023000044AddOrderAdminFields struct{}

// Signature The unique signature for the migration.
func (r *M20261023000044AddOrderAdminFields) Signature() string {
	return "20261023000044_add_order_admin_fields"
}

// Up 给 cater_order 增加商家备注、改价差额、骑手与派单时间四个列。
func (r *M20261023000044AddOrderAdminFields) Up() error {
	if !facades.Schema().HasTable("cater_order") {
		return nil
	}

	columns := []struct {
		name      string
		statement string
	}{
		{
			name:      "merchant_remark",
			statement: "ALTER TABLE cater_order ADD COLUMN merchant_remark varchar(255) NOT NULL DEFAULT '' COMMENT '商家备注，与顾客备注分列存放'",
		},
		{
			name:      "adjust_amount",
			statement: "ALTER TABLE cater_order ADD COLUMN adjust_amount bigint NOT NULL DEFAULT 0 COMMENT '改价差额（分）：正数加价、负数减免'",
		},
		{
			name:      "rider_id",
			statement: "ALTER TABLE cater_order ADD COLUMN rider_id bigint NOT NULL DEFAULT 0 COMMENT '配送骑手 ID，0 表示未派单'",
		},
		{
			name:      "dispatch_time",
			statement: "ALTER TABLE cater_order ADD COLUMN dispatch_time datetime NULL COMMENT '派单时间'",
		},
	}

	for _, column := range columns {
		if facades.Schema().HasColumn("cater_order", column.name) {
			continue
		}
		if err := facades.DB().Statement(column.statement); err != nil {
			return err
		}
	}

	return nil
}

// Down 删除本次新增的四个列。
func (r *M20261023000044AddOrderAdminFields) Down() error {
	if !facades.Schema().HasTable("cater_order") {
		return nil
	}

	for _, name := range []string{"merchant_remark", "adjust_amount", "rider_id", "dispatch_time"} {
		if !facades.Schema().HasColumn("cater_order", name) {
			continue
		}
		if err := facades.DB().Statement("ALTER TABLE cater_order DROP COLUMN " + name); err != nil {
			return err
		}
	}

	return nil
}
