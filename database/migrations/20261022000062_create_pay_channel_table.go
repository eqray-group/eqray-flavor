package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000062CreatePayChannelTable 创建支付渠道表，结构对齐 yudao pay_channel。
// 同一应用下渠道编码唯一；config 存渠道参数 JSON（appId / mchId / 密钥 / 证书内容）。
type M20261022000062CreatePayChannelTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000062CreatePayChannelTable) Signature() string {
	return "20261022000062_create_pay_channel_table"
}

// Up 创建 pay_channel 表。
func (r *M20261022000062CreatePayChannelTable) Up() error {
	if facades.Schema().HasTable("pay_channel") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_channel (
  id         bigint       NOT NULL COMMENT '渠道编号',
  app_id     bigint       NOT NULL COMMENT '应用编号',
  code       varchar(32)  NOT NULL COMMENT '渠道编码：wx_lite / wx_native / alipay_pc / mock 等',
  status     tinyint      NOT NULL DEFAULT 0 COMMENT '开启状态：0 开启 / 1 关闭',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  fee_rate   double       NOT NULL DEFAULT 0 COMMENT '渠道费率，单位：百分比',
  config     varchar(8192) NOT NULL DEFAULT '' COMMENT '支付渠道配置 JSON',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建者',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_app_code (app_id, code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付渠道配置'`)
}

// Down 删除 pay_channel 表。
func (r *M20261022000062CreatePayChannelTable) Down() error {
	return facades.Schema().DropIfExists("pay_channel")
}
