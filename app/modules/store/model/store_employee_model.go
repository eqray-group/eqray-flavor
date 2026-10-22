package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
)

// StoreEmployeeTable 是门店员工关联表名常量。
const StoreEmployeeTable = "store_employee"

// StoreEmployeeModel 对应 store_employee 表，表示后台账号在门店内的岗位。
// 门店与用户多对多，同一门店内同一用户只允许一条记录；删除采用物理删除。
type StoreEmployeeModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID，只能指向 type=3 的门店节点
	StoreID types.ID `db:"store_id" json:"storeId"`
	// UserID 后台账号 ID，关联 system_users
	UserID types.ID `db:"user_id" json:"userId"`
	// Role 门店角色：1 店长 / 2 收银 / 3 后厨 / 4 服务员 / 5 骑手
	Role enums.StoreEmployeeRoleEnum `db:"role" json:"role"`
	// Status 启停状态：0 禁用 / 1 启用；只有启用记录会计入可见门店
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *StoreEmployeeModel) TableName() string {
	return StoreEmployeeTable
}
