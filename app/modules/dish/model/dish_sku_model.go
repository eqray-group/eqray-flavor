package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// DishSkuTable 是菜品规格（SKU）表名常量。
const DishSkuTable = "dish_sku"

// DishSkuModel 对应 dish_sku 表，表示菜品的一个可售卖规格。
// 同一 SPU 下规格名唯一；三档价格一律按分存储，标准价必须大于 0。
type DishSkuModel struct {
	baseModel.BaseModel
	// SpuID 所属菜品 SPU ID
	SpuID types.ID `db:"spu_id" json:"spuId"`
	// Name 规格名称，如大份、中杯；同一 SPU 内唯一
	Name string `db:"name" json:"name"`
	// Code 规格编码，非空时全局唯一，批量导入按编码匹配
	Code string `db:"code" json:"code"`
	// StandardPrice 品牌标准价（分）
	StandardPrice int64 `db:"standard_price" json:"standardPrice"`
	// MemberPrice 品牌会员价（分），0 表示未设置
	MemberPrice int64 `db:"member_price" json:"memberPrice"`
	// CostPrice 成本价（分），允许为 0
	CostPrice int64 `db:"cost_price" json:"costPrice"`
	// Status 上下架状态：0 下架 / 1 上架
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Sort SPU 内规格排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *DishSkuModel) TableName() string {
	return DishSkuTable
}
