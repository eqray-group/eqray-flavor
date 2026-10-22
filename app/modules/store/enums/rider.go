package enums

// RiderStatusEnum 是骑手档案的接单状态。
// 取值必须与数据字典 store_rider_status 保持一致，改动时两边同步（design/24 §5）。
type RiderStatusEnum int8

const (
	RiderStatusIdle       RiderStatusEnum = 1 // 空闲
	RiderStatusAccepting  RiderStatusEnum = 2 // 接单中：在线且允许抢单，不是配送单的待接单
	RiderStatusDelivering RiderStatusEnum = 3 // 配送中
	RiderStatusResting    RiderStatusEnum = 4 // 休息
	RiderStatusDisabled   RiderStatusEnum = 5 // 停用
)

// String 返回骑手状态的中文描述，便于列表展示与日志定位。
func (s RiderStatusEnum) String() string {
	switch s {
	case RiderStatusIdle:
		return "空闲"
	case RiderStatusAccepting:
		return "接单中"
	case RiderStatusDelivering:
		return "配送中"
	case RiderStatusResting:
		return "休息"
	case RiderStatusDisabled:
		return "停用"
	default:
		return "未知状态"
	}
}

// IsValid 判断骑手状态取值是否合法。
func (s RiderStatusEnum) IsValid() bool {
	switch s {
	case RiderStatusIdle, RiderStatusAccepting, RiderStatusDelivering, RiderStatusResting, RiderStatusDisabled:
		return true
	default:
		return false
	}
}

// CanAccept 判断当前状态是否允许抢单；只有「接单中」是在线可接单态。
func (s RiderStatusEnum) CanAccept() bool {
	return s == RiderStatusAccepting
}

// BlocksOrders 判断该状态是否属于不可接单状态，用于停用 / 休息时的进行中订单校验。
func (s RiderStatusEnum) BlocksOrders() bool {
	return s == RiderStatusResting || s == RiderStatusDisabled
}
