package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// InfraSmsTemplateTable 是短信模板表名，对齐 yudao 的 system_sms_template。
const InfraSmsTemplateTable = "infra_sms_template"

const (
	// InfraSmsTemplateTypeCode 验证码短信
	InfraSmsTemplateTypeCode int8 = 1
	// InfraSmsTemplateTypeNotify 通知短信
	InfraSmsTemplateTypeNotify int8 = 2
)

// InfraSmsTemplateModel 对应 infra_sms_template 表：模板绑定渠道，发送时按 code 查找。
type InfraSmsTemplateModel struct {
	baseModel.BaseModel
	// ChannelID 关联 infra_sms_channel.id
	ChannelID types.ID `db:"channel_id" json:"channelId"`
	// Code 模板标识，如 member_sms_login
	Code string `db:"code" json:"code"`
	// Name 模板名称，仅后台展示
	Name string `db:"name" json:"name"`
	// Type 类型：1 验证码 / 2 通知
	Type int8 `db:"type" json:"type"`
	// APITemplateID 第三方平台上的模板 ID（阿里云模板 Code / 腾讯云模板 ID）
	APITemplateID string `db:"api_template_id" json:"apiTemplateId"`
	// Content 模板内容，变量用 {name} 占位，如「您的验证码是 {code}」
	Content string `db:"content" json:"content"`
	// Params 模板变量名，逗号分隔，如 code
	Params string                     `db:"params" json:"params"`
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Remark string                     `db:"remark" json:"remark"`
}

// TableName 返回表名。
func (m *InfraSmsTemplateModel) TableName() string {
	return InfraSmsTemplateTable
}
