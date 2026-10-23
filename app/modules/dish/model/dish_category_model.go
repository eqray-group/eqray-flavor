package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// DishCategoryTable 是菜品分类表名常量。
const DishCategoryTable = "dish_category"

// DishCategoryModel 对应 dish_category 表，表示品牌级的菜品分类树节点。
// 同一上级下名称唯一；删除采用物理删除，删除前需校验下级分类与菜品引用。
type DishCategoryModel struct {
	baseModel.BaseModel
	// ParentID 上级分类 ID，0 表示根分类
	ParentID types.ID `db:"parent_id" json:"parentId"`
	// Name 分类名称，同一上级下唯一
	Name string `db:"name" json:"name"`
	// Image 分类图标地址
	Image string `db:"image" json:"image"`
	// Sort 排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Status 启停状态：0 禁用 / 1 启用
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *DishCategoryModel) TableName() string {
	return DishCategoryTable
}
