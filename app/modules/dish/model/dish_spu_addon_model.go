package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// DishSpuAddonTable 是菜品加料关联表名常量。
const DishSpuAddonTable = "dish_spu_addon"

// DishSpuAddonModel 对应 dish_spu_addon 表，表示某个菜品挂载了哪个加料组。
// (spu_id, addon_group_id) 唯一，避免同一菜品重复挂载同一加料组。
type DishSpuAddonModel struct {
	baseModel.BaseModel
	// SpuID 菜品 SPU ID
	SpuID types.ID `db:"spu_id" json:"spuId"`
	// AddonGroupID 加料组 ID
	AddonGroupID types.ID `db:"addon_group_id" json:"addonGroupId"`
	// Sort 菜品内加料组排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *DishSpuAddonModel) TableName() string {
	return DishSpuAddonTable
}
