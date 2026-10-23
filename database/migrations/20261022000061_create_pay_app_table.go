package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000061CreatePayAppTable 创建支付应用表，结构对齐 yudao pay_app。
// 一个应用可配置多个渠道；支付 / 退款 / 转账回调地址按应用维度配置。
type M20261022000061CreatePayAppTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000061CreatePayAppTable) Signature() string {
	return "20261022000061_create_pay_app_table"
}

// Up 创建 pay_app 表。
func (r *M20261022000061CreatePayAppTable) Up() error {
	if facades.Schema().HasTable("pay_app") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_app (
  id                  bigint        NOT NULL COMMENT '应用编号',
  name                varchar(64)   NOT NULL COMMENT '应用名',
  status              tinyint       NOT NULL DEFAULT 0 COMMENT '开启状态：0 开启 / 1 关闭',
  remark              varchar(255)  NOT NULL DEFAULT '' COMMENT '备注',
  order_notify_url    varchar(1024) NOT NULL DEFAULT '' COMMENT '支付结果的回调地址',
  refund_notify_url   varchar(1024) NOT NULL DEFAULT '' COMMENT '退款结果的回调地址',
  transfer_notify_url varchar(1024) NOT NULL DEFAULT '' COMMENT '转账结果的回调地址',
  app_key             varchar(64)   NOT NULL COMMENT '应用标识',
  app_secret          varchar(64)   NOT NULL COMMENT '应用密钥',
  creator             varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at          datetime      NULL COMMENT '创建时间',
  updater             varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at          datetime      NULL COMMENT '更新时间',
  deleted_at          datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_app_key (app_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付应用信息'`)
}

// Down 删除 pay_app 表。
func (r *M20261022000061CreatePayAppTable) Down() error {
	return facades.Schema().DropIfExists("pay_app")
}
