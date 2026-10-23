package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// DishStorePriceTable 是门店价格表名常量。
const DishStorePriceTable = "dish_store_price"

// DishStorePriceModel 对应 dish_store_price 表，表示门店对某个规格的价格覆盖。
// (store_id, sku_id) 唯一；start_at / end_at 为空表示长期有效，价格一律按分存储。
type DishStorePriceModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID，只能指向 type=3 的门店节点
	StoreID types.ID `db:"store_id" json:"storeId"`
	// SkuID 菜品规格 SKU ID
	SkuID types.ID `db:"sku_id" json:"skuId"`
	// StorePrice 门店价（分），0 表示不覆盖品牌标准价
	StorePrice int64 `db:"store_price" json:"storePrice"`
	// StoreMemberPrice 门店会员价（分），0 表示不覆盖品牌会员价
	StoreMemberPrice int64 `db:"store_member_price" json:"storeMemberPrice"`
	// StartAt 生效开始时间，空表示立即生效
	StartAt *types.DateTime `db:"start_at" json:"startAt"`
	// EndAt 生效结束时间，空表示长期有效
	EndAt *types.DateTime `db:"end_at" json:"endAt"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *DishStorePriceModel) TableName() string {
	return DishStorePriceTable
}
