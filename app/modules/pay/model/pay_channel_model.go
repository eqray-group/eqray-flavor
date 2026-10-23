package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayChannelTable 是支付渠道表名常量，对应 yudao pay_channel。
const PayChannelTable = "pay_channel"

// PayChannelModel 对应 yudao pay_channel 表，表示某个应用下的一条渠道配置。
// code 取值见 enums.PayChannelItems；config 是渠道参数 JSON（appId / mchId / 密钥 / 证书内容）。
type PayChannelModel struct {
	baseModel.BaseModel
	// AppID 所属应用
	AppID types.ID `db:"app_id" json:"appId"`
	// Code 渠道编码：wx_lite / wx_native / alipay_pc / mock 等
	Code string `db:"code" json:"code"`
	// Status 开启状态：0 开启 / 1 关闭（yudao 口径）
	Status enums.PayEnableStatusEnum `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
	// FeeRate 渠道费率，单位百分比
	FeeRate float64 `db:"fee_rate" json:"feeRate"`
	// Config 渠道配置 JSON，字段随渠道不同
	Config string `db:"config" json:"config"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayChannelModel) TableName() string {
	return PayChannelTable
}
