package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// PayWalletTable 是会员钱包表名常量，对应 yudao pay_wallet。
const PayWalletTable = "pay_wallet"

// PayWalletModel 对应 yudao pay_wallet 表，表示一个用户的储值账户。
// balance 为可用余额（分），freeze_price 为冻结金额（分）；余额变动必须同步写流水。
type PayWalletModel struct {
	baseModel.BaseModel
	// UserID 用户编号
	UserID types.ID `db:"user_id" json:"userId"`
	// UserType 用户类型
	UserType int8 `db:"user_type" json:"userType"`
	// Balance 余额（分）
	Balance int64 `db:"balance" json:"balance"`
	// FreezePrice 冻结金额（分）
	FreezePrice int64 `db:"freeze_price" json:"freezePrice"`
	// TotalExpense 累计支出（分）
	TotalExpense int64 `db:"total_expense" json:"totalExpense"`
	// TotalRecharge 累计充值（分）
	TotalRecharge int64 `db:"total_recharge" json:"totalRecharge"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayWalletModel) TableName() string {
	return PayWalletTable
}
