package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayTransferTable 是转账单表名常量，对应 yudao pay_transfer。
const PayTransferTable = "pay_transfer"

// PayTransferModel 对应 yudao pay_transfer 表，表示一笔渠道转账（提现 / 打款）。
// 状态取 enums.PayTransferStatusEnum（0 等待转账 / 10 转账成功 / 20 转账失败）。
type PayTransferModel struct {
	baseModel.BaseModel
	// No 转账单号
	No string `db:"no" json:"no"`
	// AppID 应用编号
	AppID types.ID `db:"app_id" json:"appId"`
	// ChannelID 渠道编号
	ChannelID types.ID `db:"channel_id" json:"channelId"`
	// ChannelCode 渠道编码
	ChannelCode string `db:"channel_code" json:"channelCode"`
	// UserID 收款用户编号
	UserID types.ID `db:"user_id" json:"userId"`
	// UserType 用户类型
	UserType int8 `db:"user_type" json:"userType"`
	// MerchantTransferID 商户转账单号
	MerchantTransferID string `db:"merchant_transfer_id" json:"merchantTransferId"`
	// Subject 转账标题
	Subject string `db:"subject" json:"subject"`
	// Price 转账金额（分）
	Price int64 `db:"price" json:"price"`
	// ChannelFeeRate 渠道费率，单位百分比
	ChannelFeeRate float64 `db:"channel_fee_rate" json:"channelFeeRate"`
	// ChannelFeePrice 渠道手续费（分）
	ChannelFeePrice int64 `db:"channel_fee_price" json:"channelFeePrice"`
	// Status 转账状态
	Status enums.PayTransferStatusEnum `db:"status" json:"status"`
	// SuccessTime 转账成功时间
	SuccessTime *types.DateTime `db:"success_time" json:"successTime"`
	// ChannelExtras 渠道额外参数
	ChannelExtras string `db:"channel_extras" json:"channelExtras"`
	// ChannelErrorCode 渠道调用报错错误码
	ChannelErrorCode string `db:"channel_error_code" json:"channelErrorCode"`
	// ChannelErrorMsg 渠道调用报错信息
	ChannelErrorMsg string `db:"channel_error_msg" json:"channelErrorMsg"`
	// ChannelNotifyData 渠道转账通知原文
	ChannelNotifyData string `db:"channel_notify_data" json:"channelNotifyData"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayTransferModel) TableName() string {
	return PayTransferTable
}
