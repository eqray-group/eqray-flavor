package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/pay/enums"
)

// PayWalletRechargePackageTable 是充值套餐表名常量，对应 yudao pay_wallet_recharge_package。
const PayWalletRechargePackageTable = "pay_wallet_recharge_package"

// PayWalletRechargePackageModel 对应 yudao pay_wallet_recharge_package 表，表示一个可选充值档位。
// pay_price 为实付金额、bonus_price 为赠送金额，均按分存储。
type PayWalletRechargePackageModel struct {
	baseModel.BaseModel
	// Name 套餐名
	Name string `db:"name" json:"name"`
	// PayPrice 支付金额（分）
	PayPrice int64 `db:"pay_price" json:"payPrice"`
	// BonusPrice 赠送金额（分）
	BonusPrice int64 `db:"bonus_price" json:"bonusPrice"`
	// Status 开启状态：0 开启 / 1 关闭（yudao 口径）
	Status enums.PayEnableStatusEnum `db:"status" json:"status"`
	// Description 套餐描述
	Description string `db:"description" json:"description"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayWalletRechargePackageModel) TableName() string {
	return PayWalletRechargePackageTable
}
