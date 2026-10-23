package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
)

// DishAddonGroupTable 是加料组表名常量。
const DishAddonGroupTable = "dish_addon_group"

// DishAddonGroupModel 对应 dish_addon_group 表，表示可复用的加料组（加料、甜度、温度等）。
// 名称为品牌级唯一；通过 dish_spu_addon 挂到具体菜品。
type DishAddonGroupModel struct {
	baseModel.BaseModel
	// Name 加料组名称，全局唯一
	Name string `db:"name" json:"name"`
	// Required 是否必选：0 可选 / 1 必选
	Required int8 `db:"required" json:"required"`
	// MinSelect 最少选择数量，必选时至少为 1
	MinSelect int `db:"min_select" json:"minSelect"`
	// MaxSelect 最多选择数量，0 表示不限
	MaxSelect int `db:"max_select" json:"maxSelect"`
	// Sort 排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Status 启停状态：0 禁用 / 1 启用
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *DishAddonGroupModel) TableName() string {
	return DishAddonGroupTable
}
