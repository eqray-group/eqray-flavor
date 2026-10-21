package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000031CreateInfraSmsTemplateTable 建 infra_sms_template 短信模板表（对齐 yudao 的 system_sms_template）。
type M20261020000031CreateInfraSmsTemplateTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000031CreateInfraSmsTemplateTable) Signature() string {
	return "20261020000031_create_infra_sms_template_table"
}

// Up 建 infra_sms_template 表。
// 模板绑定渠道，发送时按 code 查找；content 用 {name} 占位，params 声明变量名。
func (r *M20261020000031CreateInfraSmsTemplateTable) Up() error {
	if facades.Schema().HasTable("infra_sms_template") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE infra_sms_template (
  id              bigint       NOT NULL COMMENT '主键',
  channel_id      bigint       NOT NULL COMMENT '短信渠道 ID',
  code            varchar(64)  NOT NULL COMMENT '模板代码，如 member_sms_login',
  name            varchar(64)  NOT NULL COMMENT '模板名称',
  type            tinyint      NOT NULL DEFAULT 1 COMMENT '类型：1 验证码 / 2 通知',
  api_template_id varchar(128) NOT NULL DEFAULT '' COMMENT '第三方平台模板 ID',
  content         varchar(512) NOT NULL COMMENT '模板内容，变量用 {name} 占位',
  params          varchar(128) NOT NULL DEFAULT '' COMMENT '模板变量名，逗号分隔',
  status          tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  remark          varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator         varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at      datetime     NULL COMMENT '创建时间',
  updater         varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at      datetime     NULL COMMENT '更新时间',
  deleted_at      datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code),
  KEY idx_channel (channel_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='短信模板表'`)
}

// Down Reverse the migrations.
func (r *M20261020000031CreateInfraSmsTemplateTable) Down() error {
	return facades.Schema().DropIfExists("infra_sms_template")
}
