package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000065CreatePayRefundTable 创建退款单表，结构对齐 yudao pay_refund。
// 状态：0 等待退款 / 10 退款成功 / 20 退款失败；同一支付单可多次部分退款。
type M20261022000065CreatePayRefundTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000065CreatePayRefundTable) Signature() string {
	return "20261022000065_create_pay_refund_table"
}

// Up 创建 pay_refund 表。
func (r *M20261022000065CreatePayRefundTable) Up() error {
	if facades.Schema().HasTable("pay_refund") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_refund (
  id                  bigint        NOT NULL COMMENT '退款单编号',
  no                  varchar(64)   NOT NULL COMMENT '退款单号',
  app_id              bigint        NOT NULL COMMENT '应用编号',
  channel_id          bigint        NOT NULL COMMENT '渠道编号',
  channel_code        varchar(32)   NOT NULL COMMENT '渠道编码',
  order_id            bigint        NOT NULL COMMENT '支付单编号',
  order_no            varchar(64)   NOT NULL COMMENT '支付单号',
  merchant_order_id   varchar(64)   NOT NULL COMMENT '商户订单号',
  merchant_refund_id  varchar(64)   NOT NULL COMMENT '商户退款单号',
  notify_url          varchar(1024) NOT NULL DEFAULT '' COMMENT '退款结果通知地址',
  status              tinyint       NOT NULL DEFAULT 0 COMMENT '退款状态：0 等待退款 / 10 退款成功 / 20 退款失败',
  pay_price           bigint        NOT NULL COMMENT '原支付金额（分）',
  refund_price        bigint        NOT NULL COMMENT '退款金额（分）',
  reason              varchar(256)  NOT NULL DEFAULT '' COMMENT '退款原因',
  user_ip             varchar(50)   NOT NULL DEFAULT '' COMMENT '用户 IP',
  channel_order_no    varchar(64)   NOT NULL DEFAULT '' COMMENT '渠道订单号',
  channel_refund_no   varchar(64)   NOT NULL DEFAULT '' COMMENT '渠道退款单号',
  success_time        datetime      NULL COMMENT '退款成功时间',
  channel_error_code  varchar(128)  NOT NULL DEFAULT '' COMMENT '渠道调用错误码',
  channel_error_msg   varchar(256)  NOT NULL DEFAULT '' COMMENT '渠道调用错误信息',
  channel_notify_data varchar(4096) NOT NULL DEFAULT '' COMMENT '渠道退款通知原文',
  audit_status        tinyint       NOT NULL DEFAULT 0 COMMENT '审核状态：0 无需审核 / 1 待审核 / 2 已通过 / 3 已驳回',
  audit_time          datetime      NULL COMMENT '审核时间',
  audit_by            bigint        NOT NULL DEFAULT 0 COMMENT '审核人后台用户 ID',
  audit_remark        varchar(255)  NOT NULL DEFAULT '' COMMENT '审核意见 / 驳回原因',
  creator             varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at          datetime      NULL COMMENT '创建时间',
  updater             varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at          datetime      NULL COMMENT '更新时间',
  deleted_at          datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_no (no),
  KEY idx_order (order_id),
  KEY idx_merchant_refund (merchant_refund_id),
  KEY idx_app_merchant_refund (app_id, merchant_refund_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款订单'`)
}

// Down 删除 pay_refund 表。
func (r *M20261022000065CreatePayRefundTable) Down() error {
	return facades.Schema().DropIfExists("pay_refund")
}
