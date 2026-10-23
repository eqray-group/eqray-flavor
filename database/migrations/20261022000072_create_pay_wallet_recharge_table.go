package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000072CreatePayWalletRechargeTable 创建钱包充值记录表，结构对齐 yudao pay_wallet_recharge。
// 状态：0 未支付 / 10 已支付 / 20 已退款；支付与退款分别关联支付单、退款单。
type M20261022000072CreatePayWalletRechargeTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000072CreatePayWalletRechargeTable) Signature() string {
	return "20261022000072_create_pay_wallet_recharge_table"
}

// Up 创建 pay_wallet_recharge 表。
func (r *M20261022000072CreatePayWalletRechargeTable) Up() error {
	if facades.Schema().HasTable("pay_wallet_recharge") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_wallet_recharge (
  id               bigint       NOT NULL COMMENT '充值记录编号',
  wallet_id        bigint       NOT NULL COMMENT '钱包编号',
  total_price      bigint       NOT NULL DEFAULT 0 COMMENT '充值总金额（分）',
  pay_price        bigint       NOT NULL DEFAULT 0 COMMENT '实付金额（分）',
  bonus_price      bigint       NOT NULL DEFAULT 0 COMMENT '赠送金额（分）',
  package_id       bigint       NOT NULL DEFAULT 0 COMMENT '充值套餐编号，0 表示自定义金额',
  status           tinyint      NOT NULL DEFAULT 0 COMMENT '充值状态：0 未支付 / 10 已支付 / 20 已退款',
  pay_order_id     bigint       NOT NULL DEFAULT 0 COMMENT '支付单编号',
  pay_channel_code varchar(32)  NOT NULL DEFAULT '' COMMENT '支付渠道编码',
  pay_time         datetime     NULL COMMENT '支付成功时间',
  refund_order_id  bigint       NOT NULL DEFAULT 0 COMMENT '退款单编号',
  refund_price     bigint       NOT NULL DEFAULT 0 COMMENT '退款金额（分）',
  refund_time      datetime     NULL COMMENT '退款时间',
  creator          varchar(64)  NOT NULL DEFAULT '' COMMENT '创建者',
  created_at       datetime     NULL COMMENT '创建时间',
  updater          varchar(64)  NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at       datetime     NULL COMMENT '更新时间',
  deleted_at       datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_wallet (wallet_id),
  KEY idx_pay_order (pay_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包充值记录'`)
}

// Down 删除 pay_wallet_recharge 表。
func (r *M20261022000072CreatePayWalletRechargeTable) Down() error {
	return facades.Schema().DropIfExists("pay_wallet_recharge")
}
