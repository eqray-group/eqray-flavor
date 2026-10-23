package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000069CreatePayWalletTable 创建会员钱包表，结构对齐 yudao pay_wallet。
// 一个用户一个钱包，余额变动必须同步写 pay_wallet_transaction。
type M20261022000069CreatePayWalletTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000069CreatePayWalletTable) Signature() string {
	return "20261022000069_create_pay_wallet_table"
}

// Up 创建 pay_wallet 表。
func (r *M20261022000069CreatePayWalletTable) Up() error {
	if facades.Schema().HasTable("pay_wallet") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_wallet (
  id             bigint      NOT NULL COMMENT '钱包编号',
  user_id        bigint      NOT NULL COMMENT '用户编号',
  user_type      tinyint     NOT NULL DEFAULT 0 COMMENT '用户类型',
  balance        bigint      NOT NULL DEFAULT 0 COMMENT '余额（分）',
  freeze_price   bigint      NOT NULL DEFAULT 0 COMMENT '冻结金额（分）',
  total_expense  bigint      NOT NULL DEFAULT 0 COMMENT '累计支出（分）',
  total_recharge bigint      NOT NULL DEFAULT 0 COMMENT '累计充值（分）',
  creator        varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
  created_at     datetime    NULL COMMENT '创建时间',
  updater        varchar(64) NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at     datetime    NULL COMMENT '更新时间',
  deleted_at     datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user (user_id, user_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员钱包'`)
}

// Down 删除 pay_wallet 表。
func (r *M20261022000069CreatePayWalletTable) Down() error {
	return facades.Schema().DropIfExists("pay_wallet")
}
