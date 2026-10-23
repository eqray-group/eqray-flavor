package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000071CreatePayWalletRechargePackageTable 创建充值套餐表，结构对齐 yudao pay_wallet_recharge_package。
// pay_price 为实付金额、bonus_price 为赠送金额，均按分存储；status：0 开启 / 1 关闭。
type M20261022000071CreatePayWalletRechargePackageTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000071CreatePayWalletRechargePackageTable) Signature() string {
	return "20261022000071_create_pay_wallet_recharge_package_table"
}

// Up 创建 pay_wallet_recharge_package 表。
func (r *M20261022000071CreatePayWalletRechargePackageTable) Up() error {
	if facades.Schema().HasTable("pay_wallet_recharge_package") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_wallet_recharge_package (
  id          bigint       NOT NULL COMMENT '套餐编号',
  name        varchar(64)  NOT NULL COMMENT '套餐名',
  pay_price   bigint       NOT NULL COMMENT '支付金额（分）',
  bonus_price bigint       NOT NULL DEFAULT 0 COMMENT '赠送金额（分）',
  status      tinyint      NOT NULL DEFAULT 0 COMMENT '开启状态：0 开启 / 1 关闭',
  description varchar(128) NOT NULL DEFAULT '' COMMENT '套餐描述',
  creator     varchar(64)  NOT NULL DEFAULT '' COMMENT '创建者',
  created_at  datetime     NULL COMMENT '创建时间',
  updater     varchar(64)  NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at  datetime     NULL COMMENT '更新时间',
  deleted_at  datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值套餐'`)
}

// Down 删除 pay_wallet_recharge_package 表。
func (r *M20261022000071CreatePayWalletRechargePackageTable) Down() error {
	return facades.Schema().DropIfExists("pay_wallet_recharge_package")
}
