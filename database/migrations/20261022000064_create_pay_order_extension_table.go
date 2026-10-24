package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000064CreatePayOrderExtensionTable 创建支付订单拓展表，结构对齐 yudao pay_order_extension。
// 每次调用渠道下单新增一条，保存渠道额外参数与报错信息，便于排查下单失败。
type M20261022000064CreatePayOrderExtensionTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000064CreatePayOrderExtensionTable) Signature() string {
	return "20261022000064_create_pay_order_extension_table"
}

// Up 创建 pay_order_extension 表。
func (r *M20261022000064CreatePayOrderExtensionTable) Up() error {
	if facades.Schema().HasTable("pay_order_extension") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_order_extension (
  id                  bigint        NOT NULL COMMENT '拓展单编号',
  no                  varchar(64)   NOT NULL COMMENT '支付单号',
  order_id            bigint        NOT NULL COMMENT '支付单编号',
  channel_id          bigint        NOT NULL COMMENT '渠道编号',
  channel_code        varchar(32)   NOT NULL COMMENT '渠道编码',
  user_ip             varchar(50)   NOT NULL DEFAULT '' COMMENT '用户 IP',
  status              tinyint       NOT NULL DEFAULT 0 COMMENT '支付状态',
  channel_extras      varchar(1024) NOT NULL DEFAULT '' COMMENT '渠道额外参数 JSON，如二维码链接、调起参数',
  channel_error_code  varchar(128)  NOT NULL DEFAULT '' COMMENT '渠道调用错误码',
  channel_error_msg   varchar(256)  NOT NULL DEFAULT '' COMMENT '渠道调用错误信息',
  channel_notify_data varchar(4096) NOT NULL DEFAULT '' COMMENT '渠道异步通知原文',
  creator             varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at          datetime      NULL COMMENT '创建时间',
  updater             varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at          datetime      NULL COMMENT '更新时间',
  deleted_at          datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_order (order_id),
  KEY idx_no (no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单拓展'`)
}

// Down 删除 pay_order_extension 表。
func (r *M20261022000064CreatePayOrderExtensionTable) Down() error {
	return facades.Schema().DropIfExists("pay_order_extension")
}
