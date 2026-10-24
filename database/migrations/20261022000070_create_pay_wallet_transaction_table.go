package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000070CreatePayWalletTransactionTable 创建钱包流水表，结构对齐 yudao pay_wallet_transaction。
// type：1 收入 / 2 支出；price 恒为正数，方向由 type 决定；balance 记录变动后余额。
type M20261022000070CreatePayWalletTransactionTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000070CreatePayWalletTransactionTable) Signature() string {
	return "20261022000070_create_pay_wallet_transaction_table"
}

// Up 创建 pay_wallet_transaction 表。
func (r *M20261022000070CreatePayWalletTransactionTable) Up() error {
	if facades.Schema().HasTable("pay_wallet_transaction") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_wallet_transaction (
  id         bigint       NOT NULL COMMENT '流水编号',
  wallet_id  bigint       NOT NULL COMMENT '钱包编号',
  biz_type   tinyint      NOT NULL DEFAULT 0 COMMENT '业务类型：1 充值 / 2 充值退款 / 3 支付 / 4 支付退款 / 5 手工调整',
  biz_id     varchar(64)  NOT NULL DEFAULT '' COMMENT '业务编号',
  type       tinyint      NOT NULL COMMENT '收支方向：1 收入 / 2 支出',
  price      bigint       NOT NULL DEFAULT 0 COMMENT '变动金额（分）',
  balance    bigint       NOT NULL DEFAULT 0 COMMENT '变动后余额（分）',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建者',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_wallet (wallet_id),
  KEY idx_biz (biz_type, biz_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包流水'`)
}

// Down 删除 pay_wallet_transaction 表。
func (r *M20261022000070CreatePayWalletTransactionTable) Down() error {
	return facades.Schema().DropIfExists("pay_wallet_transaction")
}
