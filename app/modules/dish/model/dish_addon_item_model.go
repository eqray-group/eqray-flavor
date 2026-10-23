package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// DishAddonItemTable 是加料项表名常量。
const DishAddonItemTable = "dish_addon_item"

// DishAddonItemModel 对应 dish_addon_item 表，表示加料组内的一个选项。
// 组内名称唯一；加价按分存储，允许为 0（免费加料）。
type DishAddonItemModel struct {
	baseModel.BaseModel
	// GroupID 所属加料组 ID
	GroupID types.ID `db:"group_id" json:"groupId"`
	// Name 加料项名称，组内唯一，如加珍珠、少冰
	Name string `db:"name" json:"name"`
	// Price 加价（分），0 表示免费
	Price int64 `db:"price" json:"price"`
	// Sort 组内排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Status 启停状态：0 禁用 / 1 启用
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *DishAddonItemModel) TableName() string {
	return DishAddonItemTable
}
