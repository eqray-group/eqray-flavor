package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayWalletRechargeTable 是钱包充值记录表名常量，对应 yudao pay_wallet_recharge。
const PayWalletRechargeTable = "pay_wallet_recharge"

// PayWalletRechargeModel 对应 yudao pay_wallet_recharge 表，表示一次钱包充值的完整记录。
// 先创建记录与支付单，支付成功后入账；退款时回写 refund_price 与状态。
type PayWalletRechargeModel struct {
	baseModel.BaseModel
	// WalletID 钱包编号
	WalletID types.ID `db:"wallet_id" json:"walletId"`
	// TotalPrice 充值总金额（分）= 实付 + 赠送
	TotalPrice int64 `db:"total_price" json:"totalPrice"`
	// PayPrice 实付金额（分）
	PayPrice int64 `db:"pay_price" json:"payPrice"`
	// BonusPrice 赠送金额（分）
	BonusPrice int64 `db:"bonus_price" json:"bonusPrice"`
	// PackageID 充值套餐编号，0 表示自定义金额
	PackageID types.ID `db:"package_id" json:"packageId"`
	// Status 充值状态：0 未支付 / 10 已支付 / 20 已退款
	Status enums.PayWalletRechargeStatusEnum `db:"status" json:"status"`
	// PayOrderID 支付单编号
	PayOrderID types.ID `db:"pay_order_id" json:"payOrderId"`
	// PayChannelCode 支付渠道编码
	PayChannelCode string `db:"pay_channel_code" json:"payChannelCode"`
	// PayTime 支付成功时间
	PayTime *types.DateTime `db:"pay_time" json:"payTime"`
	// RefundOrderID 退款单编号
	RefundOrderID types.ID `db:"refund_order_id" json:"refundOrderId"`
	// RefundPrice 退款金额（分）
	RefundPrice int64 `db:"refund_price" json:"refundPrice"`
	// RefundTime 退款时间
	RefundTime *types.DateTime `db:"refund_time" json:"refundTime"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayWalletRechargeModel) TableName() string {
	return PayWalletRechargeTable
}
