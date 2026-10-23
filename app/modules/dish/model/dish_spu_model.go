package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// DishSpuTable 是菜品（SPU）表名常量。
const DishSpuTable = "dish_spu"

// DishSpuModel 对应 dish_spu 表，表示品牌级的菜品，一个 SPU 下挂多个 SKU。
// 同一分类下菜品名称唯一；code 允许为空，非空时由业务层保证全局唯一。
type DishSpuModel struct {
	baseModel.BaseModel
	// CategoryID 所属分类 ID
	CategoryID types.ID `db:"category_id" json:"categoryId"`
	// Name 菜品名称，同一分类下唯一
	Name string `db:"name" json:"name"`
	// Code 菜品编码，非空时全局唯一，批量导入按编码匹配
	Code string `db:"code" json:"code"`
	// Unit 销售单位，如份、杯、例
	Unit string `db:"unit" json:"unit"`
	// Description 菜品介绍
	Description string `db:"description" json:"description"`
	// Image 主图地址
	Image string `db:"image" json:"image"`
	// Tags 标签，逗号分隔，如招牌、微辣
	Tags string `db:"tags" json:"tags"`
	// Allergens 过敏原，逗号分隔
	Allergens string `db:"allergens" json:"allergens"`
	// Status 上下架状态：0 下架 / 1 上架
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Sort 分类内排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *DishSpuModel) TableName() string {
	return DishSpuTable
}
