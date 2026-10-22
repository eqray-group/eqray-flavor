package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
)

// StoreRiderTable 是骑手档案表名常量。
const StoreRiderTable = "store_rider"

// StoreRiderModel 对应 store_rider 表，表示门店员工中的骑手接单档案。
// 与 store_employee 一对一：employee_id 唯一；停用只改状态，不删除历史档案。
type StoreRiderModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID，必须与关联员工的 store_id 一致
	StoreID types.ID `db:"store_id" json:"storeId"`
	// EmployeeID 门店员工关联 ID，一名员工只能有一份骑手档案
	EmployeeID types.ID `db:"employee_id" json:"employeeId"`
	// Status 骑手状态：1 空闲 / 2 接单中 / 3 配送中 / 4 休息 / 5 停用
	Status enums.RiderStatusEnum `db:"status" json:"status"`
	// MaxOrders 接单并发上限，骑手手上进行中的配送单不能超过该值
	MaxOrders int `db:"max_orders" json:"maxOrders"`
	// CurrentOrders 当前进行中的配送单数量，接单时条件更新递增，送达后释放
	CurrentOrders int `db:"current_orders" json:"currentOrders"`
	// TodayOrders 今日累计接单量；stat_date 跨日后由接单条件更新重置
	TodayOrders int `db:"today_orders" json:"todayOrders"`
	// StatDate 今日单量对应的自然日，跨日接单时自动切换
	StatDate *types.DateTime `db:"stat_date" json:"statDate"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *StoreRiderModel) TableName() string {
	return StoreRiderTable
}
