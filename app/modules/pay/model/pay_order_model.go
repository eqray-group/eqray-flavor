package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayOrderTable 是支付单表名常量，对应 yudao pay_order。
const PayOrderTable = "pay_order"

// PayOrderModel 对应 yudao pay_order 表，表示一笔支付交易。
// 状态取 enums.PayOrderStatusEnum（0 待支付 / 10 已支付 / 20 已退款 / 30 已关闭）。
type PayOrderModel struct {
	baseModel.BaseModel
	// AppID 应用编号
	AppID types.ID `db:"app_id" json:"appId"`
	// ChannelID 渠道编号，下单时选定
	ChannelID types.ID `db:"channel_id" json:"channelId"`
	// ChannelCode 渠道编码
	ChannelCode string `db:"channel_code" json:"channelCode"`
	// UserID 付款用户编号
	UserID types.ID `db:"user_id" json:"userId"`
	// UserType 用户类型，与会员模块对应
	UserType int8 `db:"user_type" json:"userType"`
	// MerchantOrderID 商户订单号，由业务系统生成，用于幂等
	MerchantOrderID string `db:"merchant_order_id" json:"merchantOrderId"`
	// Subject 商品标题
	Subject string `db:"subject" json:"subject"`
	// Body 商品描述
	Body string `db:"body" json:"body"`
	// NotifyURL 异步通知地址
	NotifyURL string `db:"notify_url" json:"notifyUrl"`
	// Price 支付金额（分）
	Price int64 `db:"price" json:"price"`
	// ChannelFeeRate 渠道费率，单位百分比
	ChannelFeeRate float64 `db:"channel_fee_rate" json:"channelFeeRate"`
	// ChannelFeePrice 渠道手续费（分）
	ChannelFeePrice int64 `db:"channel_fee_price" json:"channelFeePrice"`
	// Status 支付状态
	Status enums.PayOrderStatusEnum `db:"status" json:"status"`
	// RefundPrice 退款总金额（分）
	RefundPrice int64 `db:"refund_price" json:"refundPrice"`
	// UserIP 用户 IP
	UserIP string `db:"user_ip" json:"userIp"`
	// ExpireTime 支付单失效时间
	ExpireTime types.DateTime `db:"expire_time" json:"expireTime"`
	// SuccessTime 支付成功时间
	SuccessTime *types.DateTime `db:"success_time" json:"successTime"`
	// ExtensionID 支付成功的拓展单编号
	ExtensionID types.ID `db:"extension_id" json:"extensionId"`
	// No 支付单号，对外展示
	No string `db:"no" json:"no"`
	// ChannelOrderNo 渠道订单号
	ChannelOrderNo string `db:"channel_order_no" json:"channelOrderNo"`
	// ChannelUserID 渠道用户标识，如微信 openid
	ChannelUserID string `db:"channel_user_id" json:"channelUserId"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayOrderModel) TableName() string {
	return PayOrderTable
}
