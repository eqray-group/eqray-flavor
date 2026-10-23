package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayRefundTable 是退款单表名常量，对应 yudao pay_refund。
const PayRefundTable = "pay_refund"

// PayRefundModel 对应 yudao pay_refund 表，表示一笔退款交易。
// 状态取 enums.PayRefundStatusEnum（0 等待退款 / 10 退款成功 / 20 退款失败）。
type PayRefundModel struct {
	baseModel.BaseModel
	// No 退款单号
	No string `db:"no" json:"no"`
	// AppID 应用编号
	AppID types.ID `db:"app_id" json:"appId"`
	// ChannelID 渠道编号
	ChannelID types.ID `db:"channel_id" json:"channelId"`
	// ChannelCode 渠道编码
	ChannelCode string `db:"channel_code" json:"channelCode"`
	// OrderID 支付单编号
	OrderID types.ID `db:"order_id" json:"orderId"`
	// OrderNo 支付单号
	OrderNo string `db:"order_no" json:"orderNo"`
	// MerchantOrderID 商户订单号
	MerchantOrderID string `db:"merchant_order_id" json:"merchantOrderId"`
	// MerchantRefundID 商户退款单号
	MerchantRefundID string `db:"merchant_refund_id" json:"merchantRefundId"`
	// NotifyURL 退款结果通知地址
	NotifyURL string `db:"notify_url" json:"notifyUrl"`
	// Status 退款状态
	Status enums.PayRefundStatusEnum `db:"status" json:"status"`
	// PayPrice 原支付金额（分）
	PayPrice int64 `db:"pay_price" json:"payPrice"`
	// RefundPrice 退款金额（分）
	RefundPrice int64 `db:"refund_price" json:"refundPrice"`
	// Reason 退款原因
	Reason string `db:"reason" json:"reason"`
	// UserIP 用户 IP
	UserIP string `db:"user_ip" json:"userIp"`
	// ChannelOrderNo 渠道订单号
	ChannelOrderNo string `db:"channel_order_no" json:"channelOrderNo"`
	// ChannelRefundNo 渠道退款单号
	ChannelRefundNo string `db:"channel_refund_no" json:"channelRefundNo"`
	// SuccessTime 退款成功时间
	SuccessTime *types.DateTime `db:"success_time" json:"successTime"`
	// ChannelErrorCode 渠道调用报错错误码
	ChannelErrorCode string `db:"channel_error_code" json:"channelErrorCode"`
	// ChannelErrorMsg 渠道调用报错信息
	ChannelErrorMsg string `db:"channel_error_msg" json:"channelErrorMsg"`
	// ChannelNotifyData 渠道退款通知原文
	ChannelNotifyData string `db:"channel_notify_data" json:"channelNotifyData"`
	// AuditStatus 审核状态：0 无需审核 / 1 待审核 / 2 已通过 / 3 已驳回
	// 与 status 解耦：审核通过的退款单才提交渠道，状态机仍按 yudao 的 0/1/2/99 走
	AuditStatus enums.PayRefundAuditStatusEnum `db:"audit_status" json:"auditStatus"`
	// AuditTime 审核时间
	AuditTime *types.DateTime `db:"audit_time" json:"auditTime"`
	// AuditBy 审核人后台用户 ID
	AuditBy types.ID `db:"audit_by" json:"auditBy"`
	// AuditRemark 审核意见 / 驳回原因
	AuditRemark string `db:"audit_remark" json:"auditRemark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayRefundModel) TableName() string {
	return PayRefundTable
}
