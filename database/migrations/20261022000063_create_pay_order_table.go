package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000063CreatePayOrderTable 创建支付单表，结构对齐 yudao pay_order。
// 状态：0 待支付 / 10 已支付 / 20 已退款 / 30 已关闭；no 为对外支付单号。
type M20261022000063CreatePayOrderTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000063CreatePayOrderTable) Signature() string {
	return "20261022000063_create_pay_order_table"
}

// Up 创建 pay_order 表。
func (r *M20261022000063CreatePayOrderTable) Up() error {
	if facades.Schema().HasTable("pay_order") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_order (
  id                 bigint        NOT NULL COMMENT '支付单编号',
  app_id             bigint        NOT NULL COMMENT '应用编号',
  channel_id         bigint        NOT NULL DEFAULT 0 COMMENT '渠道编号',
  channel_code       varchar(32)   NOT NULL DEFAULT '' COMMENT '渠道编码',
  user_id            bigint        NOT NULL DEFAULT 0 COMMENT '付款用户编号',
  user_type          tinyint       NOT NULL DEFAULT 0 COMMENT '用户类型',
  merchant_order_id  varchar(64)   NOT NULL COMMENT '商户订单号',
  subject            varchar(32)   NOT NULL COMMENT '商品标题',
  body               varchar(128)  NOT NULL DEFAULT '' COMMENT '商品描述',
  notify_url         varchar(1024) NOT NULL DEFAULT '' COMMENT '异步通知地址',
  price              bigint        NOT NULL COMMENT '支付金额（分）',
  channel_fee_rate   double        NOT NULL DEFAULT 0 COMMENT '渠道费率，单位：百分比',
  channel_fee_price  bigint        NOT NULL DEFAULT 0 COMMENT '渠道手续费（分）',
  status             tinyint       NOT NULL DEFAULT 0 COMMENT '支付状态：0 待支付 / 10 已支付 / 20 已退款 / 30 已关闭',
  refund_price       bigint        NOT NULL DEFAULT 0 COMMENT '退款总金额（分）',
  user_ip            varchar(50)   NOT NULL DEFAULT '' COMMENT '用户 IP',
  expire_time        datetime      NOT NULL COMMENT '支付单失效时间',
  success_time       datetime      NULL COMMENT '支付成功时间',
  extension_id       bigint        NOT NULL DEFAULT 0 COMMENT '支付成功的拓展单编号',
  no                 varchar(64)   NOT NULL COMMENT '支付单号',
  channel_order_no   varchar(64)   NOT NULL DEFAULT '' COMMENT '渠道订单号',
  channel_user_id    varchar(255)  NOT NULL DEFAULT '' COMMENT '渠道用户标识',
  creator            varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at         datetime      NULL COMMENT '创建时间',
  updater            varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at         datetime      NULL COMMENT '更新时间',
  deleted_at         datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_no (no),
  KEY idx_merchant_order (merchant_order_id),
  KEY idx_channel_order (channel_order_no),
  KEY idx_status_expire (status, expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单'`)
}

// Down 删除 pay_order 表。
func (r *M20261022000063CreatePayOrderTable) Down() error {
	return facades.Schema().DropIfExists("pay_order")
}
