package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000030CreateInfraSmsChannelTable 建 infra_sms_channel 短信渠道配置表（对齐 yudao 的 system_sms_channel）。
type M20261020000030CreateInfraSmsChannelTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000030CreateInfraSmsChannelTable) Signature() string {
	return "20261020000030_create_infra_sms_channel_table"
}

// Up 建 infra_sms_channel 表。
// 渠道凭据存库、后台页面维护；每个渠道代码一条配置。
func (r *M20261020000030CreateInfraSmsChannelTable) Up() error {
	if facades.Schema().HasTable("infra_sms_channel") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE infra_sms_channel (
  id         bigint       NOT NULL COMMENT '主键',
  name       varchar(64)  NOT NULL COMMENT '渠道名称',
  code       varchar(32)  NOT NULL COMMENT '渠道代码：DebugLog / AliyunSms / TencentSms',
  api_key    varchar(128) NOT NULL DEFAULT '' COMMENT '阿里云 AccessKeyId / 腾讯云 SecretId',
  api_secret varchar(256) NOT NULL DEFAULT '' COMMENT '阿里云 AccessKeySecret / 腾讯云 SecretKey',
  signature  varchar(64)  NOT NULL DEFAULT '' COMMENT '短信签名',
  extra      varchar(512) NOT NULL DEFAULT '' COMMENT '扩展配置 JSON，如腾讯云 region / sdkAppId',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='短信渠道配置表'`)
}

// Down Reverse the migrations.
func (r *M20261020000030CreateInfraSmsChannelTable) Down() error {
	return facades.Schema().DropIfExists("infra_sms_channel")
}
