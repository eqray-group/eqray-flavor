package model

import (
	"eqray-flavor/app/modules/pay/enums"

	baseModel "eqray-flavor/app/modules/base/model"
)

// PayAppTable 是支付应用表名常量，对应 yudao pay_app。
const PayAppTable = "pay_app"

// PayAppModel 对应 yudao pay_app 表，表示接入支付能力的一个商户应用。
// 一个应用可以配置多个渠道；订单 / 退款 / 转账回调地址按应用维度配置。
type PayAppModel struct {
	baseModel.BaseModel
	// Name 应用名
	Name string `db:"name" json:"name"`
	// Status 开启状态：0 开启 / 1 关闭（yudao 口径）
	Status enums.PayEnableStatusEnum `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
	// OrderNotifyURL 支付结果回调地址
	OrderNotifyURL string `db:"order_notify_url" json:"orderNotifyUrl"`
	// RefundNotifyURL 退款结果回调地址
	RefundNotifyURL string `db:"refund_notify_url" json:"refundNotifyUrl"`
	// TransferNotifyURL 转账结果回调地址
	TransferNotifyURL string `db:"transfer_notify_url" json:"transferNotifyUrl"`
	// AppKey 应用标识，商户接入时校验
	AppKey string `db:"app_key" json:"appKey"`
	// AppSecret 应用密钥，参与回调签名
	AppSecret string `db:"app_secret" json:"-"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayAppModel) TableName() string {
	return PayAppTable
}
