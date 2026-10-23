package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000041AddOrderPointsUsed 给订单主表补一个 points_used 列。
// 背景：订单原本只存 points_amount（积分抵扣金额，分），
// 结算扣减与退款返还都按「当前兑换比例」现算比例再折算，
// 一旦运营调整了兑换比例，同一笔订单就会出现「扣 2000 分、退 1500 分」的错账。
// 这里在下单时把实际使用的积分数落库，支付按此扣减、退款按此返还，口径与订单一一对应。
type M20261023000041AddOrderPointsUsed struct{}

// Signature The unique signature for the migration.
func (r *M20261023000041AddOrderPointsUsed) Signature() string {
	return "20261023000041_add_order_points_used"
}

// Up 给 cater_order 增加 points_used 列。
func (r *M20261023000041AddOrderPointsUsed) Up() error {
	if !facades.Schema().HasTable("cater_order") {
		return nil
	}
	if facades.Schema().HasColumn("cater_order", "points_used") {
		return nil
	}

	return facades.DB().Statement("ALTER TABLE cater_order ADD COLUMN points_used int NOT NULL DEFAULT 0 COMMENT '实际使用的积分数：下单时按当时兑换比例折算并落库'")
}

// Down 删除 points_used 列。
func (r *M20261023000041AddOrderPointsUsed) Down() error {
	if !facades.Schema().HasTable("cater_order") {
		return nil
	}
	if !facades.Schema().HasColumn("cater_order", "points_used") {
		return nil
	}

	return facades.DB().Statement("ALTER TABLE cater_order DROP COLUMN points_used")
}
