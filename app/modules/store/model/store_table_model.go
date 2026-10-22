package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
)

// StoreTableTable 是桌台表名常量。
const StoreTableTable = "store_table"

// StoreTableModel 对应 store_table 表，表示门店内的一个座位单元。
// 桌码全局唯一，用于扫码识别门店与桌号；桌号在同一门店内唯一。
type StoreTableModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID，只能指向 type=3 的门店节点
	StoreID types.ID `db:"store_id" json:"storeId"`
	// AreaID 所属区域 ID，必须与 StoreID 属于同一门店
	AreaID types.ID `db:"area_id" json:"areaId"`
	// Name 桌号，如 A01、包间 1
	Name string `db:"name" json:"name"`
	// Code 桌码，全局唯一；默认按「门店编码-桌号」生成，创建后不可修改
	Code string `db:"code" json:"code"`
	// Seats 座位数
	Seats int `db:"seats" json:"seats"`
	// Sort 区域内排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Status 桌台状态：1 空闲 / 2 使用中 / 3 待清台
	Status enums.TableStatusEnum `db:"status" json:"status"`
	// CurrentOrderID 当前台位单 ID，0 表示没有进行中的台位单
	CurrentOrderID types.ID `db:"current_order_id" json:"currentOrderId"`
	// OpenedAt 开台时间，清台后置空
	OpenedAt *types.DateTime `db:"opened_at" json:"openedAt"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *StoreTableModel) TableName() string {
	return StoreTableTable
}
