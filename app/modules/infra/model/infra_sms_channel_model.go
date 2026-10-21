package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
)

// InfraSmsChannelTable 是短信渠道配置表名，对齐 yudao 的 system_sms_channel。
const InfraSmsChannelTable = "infra_sms_channel"

// 短信渠道代码，与后台「短信渠道」页面的 code 对应。
const (
	InfraSmsChannelDebugLog = "DebugLog"
	InfraSmsChannelAliyun   = "AliyunSms"
	InfraSmsChannelTencent  = "TencentSms"
)

// InfraSmsChannelModel 对应 infra_sms_channel 表：一个渠道一条配置。
type InfraSmsChannelModel struct {
	baseModel.BaseModel
	// Name 渠道名称，仅后台展示
	Name string `db:"name" json:"name"`
	// Code 渠道代码：DebugLog / AliyunSms / TencentSms
	Code string `db:"code" json:"code"`
	// APIKey 阿里云 AccessKeyId / 腾讯云 SecretId
	APIKey string `db:"api_key" json:"apiKey"`
	// APISecret 阿里云 AccessKeySecret / 腾讯云 SecretKey
	APISecret string `db:"api_secret" json:"-"`
	// Signature 短信签名
	Signature string `db:"signature" json:"signature"`
	// Extra 渠道扩展配置 JSON：阿里云 {"region"}，腾讯云 {"region","sdkAppId"}
	Extra  string                     `db:"extra" json:"extra"`
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Remark string                     `db:"remark" json:"remark"`
}

// TableName 返回表名。
func (m *InfraSmsChannelModel) TableName() string {
	return InfraSmsChannelTable
}
