package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000011AlterPayWalletTransactionBizType 把钱包流水的 biz_type 从字符串改为整型枚举。
// yudao 的 PayWalletBizTypeEnum 用 1 充值 / 2 充值退款 / 3 支付 / 4 支付退款，
// 本项目另加 5 表示后台手工调整。旧实现用 recharge / payment / refund / manual 字符串，
// 且「充值退款」与「支付退款」共用 refund，无法区分资金去向，这里一并拆开。
// 迁移分两步：先把字符串值翻译成数字（幂等，已是数字的原样保留），再把列类型改为 tinyint。
type M20261023000011AlterPayWalletTransactionBizType struct{}

// Signature The unique signature for the migration.
func (r *M20261023000011AlterPayWalletTransactionBizType) Signature() string {
	return "20261023000011_alter_pay_wallet_transaction_biz_type"
}

// Up 翻译历史 biz_type 取值并将列改为 tinyint。
func (r *M20261023000011AlterPayWalletTransactionBizType) Up() error {
	if !facades.Schema().HasTable("pay_wallet_transaction") {
		return nil
	}

	// 先翻译取值：旧 refund 统一视为支付退款（4），充值退款在旧实现里也记 refund，
	// 无法反推，按支付退款处理并在文档中标注历史口径；新数据会正确区分 2 与 4。
	if err := facades.DB().Statement(`
UPDATE pay_wallet_transaction
SET biz_type = CASE biz_type
  WHEN 'recharge' THEN '1'
  WHEN 'payment'  THEN '3'
  WHEN 'refund'   THEN '4'
  WHEN 'manual'   THEN '5'
  ELSE biz_type
END
WHERE biz_type IN ('recharge', 'payment', 'refund', 'manual')`); err != nil {
		return err
	}

	// 再改列类型：重复执行同样成立（MODIFY 幂等）
	return facades.DB().Statement(`
ALTER TABLE pay_wallet_transaction
  MODIFY COLUMN biz_type tinyint NOT NULL DEFAULT 0
  COMMENT '业务类型：1 充值 / 2 充值退款 / 3 支付 / 4 支付退款 / 5 手工调整'`)
}

// Down 把列改回字符串并翻译取值，仅用于回滚到旧实现。
func (r *M20261023000011AlterPayWalletTransactionBizType) Down() error {
	if !facades.Schema().HasTable("pay_wallet_transaction") {
		return nil
	}

	if err := facades.DB().Statement(`
ALTER TABLE pay_wallet_transaction
  MODIFY COLUMN biz_type varchar(32) NOT NULL DEFAULT ''
  COMMENT '业务类型：recharge / payment / refund / transfer'`); err != nil {
		return err
	}

	return facades.DB().Statement(`
UPDATE pay_wallet_transaction
SET biz_type = CASE biz_type
  WHEN '1' THEN 'recharge'
  WHEN '2' THEN 'refund'
  WHEN '3' THEN 'payment'
  WHEN '4' THEN 'refund'
  WHEN '5' THEN 'manual'
  ELSE biz_type
END`)
}
