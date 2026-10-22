// Package enums 定义门店域的业务枚举。
// 取值必须与数据字典 store_type / store_status 保持一致，改动时两边同步。
package enums

// StoreTypeEnum 是门店组织节点的类型，用于区分总部、大区与门店。
// 同一张表靠该字段区分层级，父子自引用成树。
type StoreTypeEnum int8

const (
	StoreTypeHeadquarter StoreTypeEnum = 1 // 总部
	StoreTypeRegion      StoreTypeEnum = 2 // 大区
	StoreTypeShop        StoreTypeEnum = 3 // 门店
)

// String 返回节点类型的中文描述，便于错误提示与日志定位。
func (t StoreTypeEnum) String() string {
	switch t {
	case StoreTypeHeadquarter:
		return "总部"
	case StoreTypeRegion:
		return "大区"
	case StoreTypeShop:
		return "门店"
	default:
		return "未知类型"
	}
}

// IsValid 判断节点类型取值是否合法。
func (t StoreTypeEnum) IsValid() bool {
	return t == StoreTypeHeadquarter || t == StoreTypeRegion || t == StoreTypeShop
}

// IsBranch 判断是否为可挂子节点的中间层级（总部、大区）。
// 门店是最末级，不允许再挂下级。
func (t StoreTypeEnum) IsBranch() bool {
	return t == StoreTypeHeadquarter || t == StoreTypeRegion
}

// StoreStatusEnum 是门店的经营状态，决定能否接单。
// 营业可接单；休息与装修不可下单；停业保留历史、禁止新单。
type StoreStatusEnum int8

const (
	StoreStatusOpen       StoreStatusEnum = 1 // 营业
	StoreStatusRest       StoreStatusEnum = 2 // 休息
	StoreStatusDecorating StoreStatusEnum = 3 // 装修
	StoreStatusClosed     StoreStatusEnum = 4 // 停业
)

// String 返回经营状态的中文描述。
func (s StoreStatusEnum) String() string {
	switch s {
	case StoreStatusOpen:
		return "营业"
	case StoreStatusRest:
		return "休息"
	case StoreStatusDecorating:
		return "装修"
	case StoreStatusClosed:
		return "停业"
	default:
		return "未知状态"
	}
}

// IsValid 判断经营状态取值是否合法。
func (s StoreStatusEnum) IsValid() bool {
	switch s {
	case StoreStatusOpen, StoreStatusRest, StoreStatusDecorating, StoreStatusClosed:
		return true
	default:
		return false
	}
}
