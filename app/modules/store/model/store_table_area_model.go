package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// StoreTableAreaTable 是桌台区域表名常量。
const StoreTableAreaTable = "store_table_area"

// StoreTableAreaModel 对应 store_table_area 表，表示门店内的用餐分区（大厅、包间等）。
// 区域按门店归属，同一门店内名称唯一；删除采用物理删除，避免软删除记录占用唯一键。
type StoreTableAreaModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID，只能指向 type=3 的门店节点
	StoreID types.ID `db:"store_id" json:"storeId"`
	// Name 区域名称，如大厅、包间、露台
	Name string `db:"name" json:"name"`
	// Sort 区域内桌台的展示顺序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *StoreTableAreaModel) TableName() string {
	return StoreTableAreaTable
}
