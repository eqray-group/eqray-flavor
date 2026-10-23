package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000013AddPayCompositeIndexes 为支付单与退款单补「应用 + 商户单号」复合索引。
// 现有索引只有 merchant_order_id 单列，而幂等查询与退款校验都带 app_id 前缀，
// 复合索引能直接命中。这里保持 yudao 的做法用普通索引而不是唯一索引：
// 两张表都是软删除，唯一索引会让已删除的单永久占用商户单号，删除后无法重建。
type M20261023000013AddPayCompositeIndexes struct{}

// Signature The unique signature for the migration.
func (r *M20261023000013AddPayCompositeIndexes) Signature() string {
	return "20261023000013_add_pay_composite_indexes"
}

// Up 补两个复合索引，已存在则跳过。
func (r *M20261023000013AddPayCompositeIndexes) Up() error {
	if facades.Schema().HasTable("pay_order") && !facades.Schema().HasIndex("pay_order", "idx_app_merchant_order") {
		if err := facades.DB().Statement(
			"ALTER TABLE pay_order ADD KEY idx_app_merchant_order (app_id, merchant_order_id)",
		); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("pay_refund") && !facades.Schema().HasIndex("pay_refund", "idx_app_merchant_refund") {
		if err := facades.DB().Statement(
			"ALTER TABLE pay_refund ADD KEY idx_app_merchant_refund (app_id, merchant_refund_id)",
		); err != nil {
			return err
		}
	}

	return nil
}

// Down 删除两个复合索引。
func (r *M20261023000013AddPayCompositeIndexes) Down() error {
	if facades.Schema().HasTable("pay_order") && facades.Schema().HasIndex("pay_order", "idx_app_merchant_order") {
		if err := facades.DB().Statement("ALTER TABLE pay_order DROP INDEX idx_app_merchant_order"); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("pay_refund") && facades.Schema().HasIndex("pay_refund", "idx_app_merchant_refund") {
		return facades.DB().Statement("ALTER TABLE pay_refund DROP INDEX idx_app_merchant_refund")
	}

	return nil
}
