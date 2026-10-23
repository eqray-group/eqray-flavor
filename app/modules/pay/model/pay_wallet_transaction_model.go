package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayWalletTransactionTable 是钱包流水表名常量，对应 yudao pay_wallet_transaction。
const PayWalletTransactionTable = "pay_wallet_transaction"

// PayWalletTransactionModel 对应 yudao pay_wallet_transaction 表，记录每一次余额变动。
// type 表示收支方向；biz_type + biz_id 指向来源业务（充值记录 / 支付单 / 退款单）。
type PayWalletTransactionModel struct {
	baseModel.BaseModel
	// WalletID 钱包编号
	WalletID types.ID `db:"wallet_id" json:"walletId"`
	// BizType 业务类型：1 充值 / 2 充值退款 / 3 支付 / 4 支付退款 / 5 手工调整
	BizType enums.PayWalletBizTypeEnum `db:"biz_type" json:"bizType"`
	// BizID 业务编号
	BizID string `db:"biz_id" json:"bizId"`
	// Type 收支方向：1 收入 / 2 支出
	Type enums.PayWalletTransactionTypeEnum `db:"type" json:"type"`
	// Price 变动金额（分），恒为正数，方向由 type 决定
	Price int64 `db:"price" json:"price"`
	// Balance 变动后余额（分）
	Balance int64 `db:"balance" json:"balance"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayWalletTransactionModel) TableName() string {
	return PayWalletTransactionTable
}
