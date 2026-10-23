package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// DishStoreSkuTable 是门店可售表名常量。
const DishStoreSkuTable = "dish_store_sku"

// DishStoreSkuModel 对应 dish_store_sku 表，表示某个门店对某个规格的上架配置。
// (store_id, sku_id) 唯一；available / sold_out / online_only 三个开关决定顾客端菜单展示。
type DishStoreSkuModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID，只能指向 type=3 的门店节点
	StoreID types.ID `db:"store_id" json:"storeId"`
	// SkuID 菜品规格 SKU ID
	SkuID types.ID `db:"sku_id" json:"skuId"`
	// Available 门店是否可售：0 下架 / 1 可售
	Available int8 `db:"available" json:"available"`
	// SoldOut 沽清标记：0 正常 / 1 已沽清
	SoldOut int8 `db:"sold_out" json:"soldOut"`
	// OnlineOnly 是否仅线上可售：0 不限 / 1 仅线上
	OnlineOnly int8 `db:"online_only" json:"onlineOnly"`
	// Sort 门店内排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *DishStoreSkuModel) TableName() string {
	return DishStoreSkuTable
}
