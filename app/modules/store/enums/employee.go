package enums

// StoreEmployeeRoleEnum 是门店员工在门店内承担的角色。
// 取值必须与数据字典 store_role_type 保持一致，改动时两边同步。
type StoreEmployeeRoleEnum int8

const (
	StoreEmployeeRoleManager StoreEmployeeRoleEnum = 1 // 店长
	StoreEmployeeRoleCashier StoreEmployeeRoleEnum = 2 // 收银
	StoreEmployeeRoleKitchen StoreEmployeeRoleEnum = 3 // 后厨
	StoreEmployeeRoleWaiter  StoreEmployeeRoleEnum = 4 // 服务员
	StoreEmployeeRoleRider   StoreEmployeeRoleEnum = 5 // 骑手
)

// String 返回门店角色的中文描述，便于列表展示与日志定位。
func (r StoreEmployeeRoleEnum) String() string {
	switch r {
	case StoreEmployeeRoleManager:
		return "店长"
	case StoreEmployeeRoleCashier:
		return "收银"
	case StoreEmployeeRoleKitchen:
		return "后厨"
	case StoreEmployeeRoleWaiter:
		return "服务员"
	case StoreEmployeeRoleRider:
		return "骑手"
	default:
		return "未知角色"
	}
}

// IsValid 判断门店角色取值是否合法。
func (r StoreEmployeeRoleEnum) IsValid() bool {
	switch r {
	case StoreEmployeeRoleManager, StoreEmployeeRoleCashier, StoreEmployeeRoleKitchen,
		StoreEmployeeRoleWaiter, StoreEmployeeRoleRider:
		return true
	default:
		return false
	}
}
