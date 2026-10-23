package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayOrderExtensionTable 是支付订单拓展表名常量，对应 yudao pay_order_extension。
const PayOrderExtensionTable = "pay_order_extension"

// PayOrderExtensionModel 对应 yudao pay_order_extension 表，保存每次下单的渠道交互明细。
// 一个支付单每次换渠道 / 重新下单都会新增一条拓展记录，用于排查渠道报错。
type PayOrderExtensionModel struct {
	baseModel.BaseModel
	// No 支付单号，与 pay_order.no 对应
	No string `db:"no" json:"no"`
	// OrderID 支付单编号
	OrderID types.ID `db:"order_id" json:"orderId"`
	// ChannelID 渠道编号
	ChannelID types.ID `db:"channel_id" json:"channelId"`
	// ChannelCode 渠道编码
	ChannelCode string `db:"channel_code" json:"channelCode"`
	// UserIP 用户 IP
	UserIP string `db:"user_ip" json:"userIp"`
	// Status 支付状态
	Status enums.PayOrderStatusEnum `db:"status" json:"status"`
	// ChannelExtras 渠道额外参数，如二维码链接
	ChannelExtras string `db:"channel_extras" json:"channelExtras"`
	// ChannelErrorCode 渠道调用报错错误码
	ChannelErrorCode string `db:"channel_error_code" json:"channelErrorCode"`
	// ChannelErrorMsg 渠道调用报错信息
	ChannelErrorMsg string `db:"channel_error_msg" json:"channelErrorMsg"`
	// ChannelNotifyData 渠道异步通知原文
	ChannelNotifyData string `db:"channel_notify_data" json:"channelNotifyData"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayOrderExtensionModel) TableName() string {
	return PayOrderExtensionTable
}
