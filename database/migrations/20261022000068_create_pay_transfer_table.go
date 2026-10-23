package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000068CreatePayTransferTable 创建转账单表，结构对齐 yudao pay_transfer。
// 状态：0 等待转账 / 10 转账成功 / 20 转账失败。
type M20261022000068CreatePayTransferTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000068CreatePayTransferTable) Signature() string {
	return "20261022000068_create_pay_transfer_table"
}

// Up 创建 pay_transfer 表。
func (r *M20261022000068CreatePayTransferTable) Up() error {
	if facades.Schema().HasTable("pay_transfer") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_transfer (
  id                    bigint        NOT NULL COMMENT '转账单编号',
  no                    varchar(64)   NOT NULL COMMENT '转账单号',
  app_id                bigint        NOT NULL COMMENT '应用编号',
  channel_id            bigint        NOT NULL COMMENT '渠道编号',
  channel_code          varchar(32)   NOT NULL COMMENT '渠道编码',
  user_id               bigint        NOT NULL DEFAULT 0 COMMENT '收款用户编号',
  user_type             tinyint       NOT NULL DEFAULT 0 COMMENT '用户类型',
  merchant_transfer_id  varchar(64)   NOT NULL COMMENT '商户转账单号',
  subject               varchar(32)   NOT NULL COMMENT '转账标题',
  price                 bigint        NOT NULL COMMENT '转账金额（分）',
  channel_fee_rate      double        NOT NULL DEFAULT 0 COMMENT '渠道费率，单位：百分比',
  channel_fee_price     bigint        NOT NULL DEFAULT 0 COMMENT '渠道手续费（分）',
  status                tinyint       NOT NULL DEFAULT 0 COMMENT '转账状态：0 等待转账 / 10 转账成功 / 20 转账失败',
  success_time          datetime      NULL COMMENT '转账成功时间',
  channel_extras        varchar(256)  NOT NULL DEFAULT '' COMMENT '渠道额外参数',
  channel_error_code    varchar(128)  NOT NULL DEFAULT '' COMMENT '渠道调用错误码',
  channel_error_msg     varchar(256)  NOT NULL DEFAULT '' COMMENT '渠道调用错误信息',
  channel_notify_data   varchar(4096) NOT NULL DEFAULT '' COMMENT '渠道转账通知原文',
  creator               varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at            datetime      NULL COMMENT '创建时间',
  updater               varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at            datetime      NULL COMMENT '更新时间',
  deleted_at            datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_no (no),
  KEY idx_merchant_transfer (merchant_transfer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='转账订单'`)
}

// Down 删除 pay_transfer 表。
func (r *M20261022000068CreatePayTransferTable) Down() error {
	return facades.Schema().DropIfExists("pay_transfer")
}
