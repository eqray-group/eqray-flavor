// Package enums 定义 order 模块的枚举，取值与字典 cater_order_* 一致。
// 订单状态用百位段位：100 起步，便于后续在状态之间插入子状态而不改历史数据。
package enums

// OrderType 是订单类型。
type OrderType int8

const (
	// OrderTypeDineIn 堂食
	OrderTypeDineIn OrderType = 1
	// OrderTypeTakeout 外卖
	OrderTypeTakeout OrderType = 2
	// OrderTypePickup 自提
	OrderTypePickup OrderType = 3
)

// String 返回订单类型中文描述。
func (t OrderType) String() string {
	switch t {
	case OrderTypeDineIn:
		return "堂食"
	case OrderTypeTakeout:
		return "外卖"
	case OrderTypePickup:
		return "自提"
	default:
		return "未知类型"
	}
}

// OrderChannel 是下单渠道。
type OrderChannel int8

const (
	// OrderChannelMiniProgram 小程序
	OrderChannelMiniProgram OrderChannel = 1
	// OrderChannelPOS POS 收银
	OrderChannelPOS OrderChannel = 2
	// OrderChannelKiosk 自助机
	OrderChannelKiosk OrderChannel = 3
	// OrderChannelDeliveryPlatform 外卖平台
	OrderChannelDeliveryPlatform OrderChannel = 4
)

// String 返回下单渠道中文描述。
func (c OrderChannel) String() string {
	switch c {
	case OrderChannelMiniProgram:
		return "小程序"
	case OrderChannelPOS:
		return "POS"
	case OrderChannelKiosk:
		return "自助机"
	case OrderChannelDeliveryPlatform:
		return "外卖平台"
	default:
		return "未知渠道"
	}
}

// OrderStatus 是订单状态机。流转关系见 CanTransit，与测试设计 SM-00 ~ SM-07 对应：
//
//	100 待支付 → 200 待接单（外卖/自提支付成功）｜300 制作中（堂食支付即制作）｜600 已取消（超时/用户取消）
//	200 待接单 → 300 制作中（接单）｜700 退款中（拒单全额退）
//	300 制作中 → 400 待取餐｜500 已完成（堂食结清）｜700 退款中（未出餐取消）
//	400 待取餐 → 500 已完成
//	500 已完成 → 700 退款中（售后退款）
//	700 退款中 → 800 已退款（退款成功）｜500 已完成（退款驳回）
//	600 已取消、800 已退款 是终态。
type OrderStatus int16

const (
	// OrderStatusPendingPay 待支付
	OrderStatusPendingPay OrderStatus = 100
	// OrderStatusPendingAccept 待接单
	OrderStatusPendingAccept OrderStatus = 200
	// OrderStatusMaking 制作中
	OrderStatusMaking OrderStatus = 300
	// OrderStatusWaitingPickup 待取餐
	OrderStatusWaitingPickup OrderStatus = 400
	// OrderStatusCompleted 已完成
	OrderStatusCompleted OrderStatus = 500
	// OrderStatusCancelled 已取消
	OrderStatusCancelled OrderStatus = 600
	// OrderStatusRefunding 退款中
	OrderStatusRefunding OrderStatus = 700
	// OrderStatusRefunded 已退款
	OrderStatusRefunded OrderStatus = 800
)

// orderTransitions 是合法流转表，key 是原状态，value 是允许的新状态集合。
var orderTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPendingPay:    {OrderStatusPendingAccept, OrderStatusMaking, OrderStatusCancelled},
	OrderStatusPendingAccept: {OrderStatusMaking, OrderStatusRefunding},
	OrderStatusMaking:        {OrderStatusWaitingPickup, OrderStatusCompleted, OrderStatusRefunding},
	OrderStatusWaitingPickup: {OrderStatusCompleted},
	OrderStatusCompleted:     {OrderStatusRefunding},
	OrderStatusCancelled:     {},
	OrderStatusRefunding:     {OrderStatusRefunded, OrderStatusCompleted},
	OrderStatusRefunded:      {},
}

// CanTransit 判断从当前状态流转到目标状态是否合法。
func (s OrderStatus) CanTransit(target OrderStatus) bool {
	for _, next := range orderTransitions[s] {
		if next == target {
			return true
		}
	}

	return false
}

// IsTerminal 判断是否终态（已取消 / 已退款）。
func (s OrderStatus) IsTerminal() bool {
	return s == OrderStatusCancelled || s == OrderStatusRefunded
}

// IsPaid 判断订单是否已进入支付成功后的状态（不含已取消）。
// 用于退款资格、逆向补偿等判断：已支付但可能还没完成。
func (s OrderStatus) IsPaid() bool {
	switch s {
	case OrderStatusPendingAccept, OrderStatusMaking, OrderStatusWaitingPickup,
		OrderStatusCompleted, OrderStatusRefunding, OrderStatusRefunded:
		return true
	default:
		return false
	}
}

// CanAdjust 判断当前状态是否允许改价。
// 只有还未进入支付的早期状态可以改：待支付与待接单；已支付订单金额已被支付单钉住，不能改。
func (s OrderStatus) CanAdjust() bool {
	return s == OrderStatusPendingPay || s == OrderStatusPendingAccept
}

// CanDispatch 判断当前状态是否允许派单。
// 外卖订单在待接单与制作中都可以指派骑手，出餐后不再允许改派。
func (s OrderStatus) CanDispatch() bool {
	return s == OrderStatusPendingAccept || s == OrderStatusMaking
}

// String 返回订单状态中文描述。
func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPendingPay:
		return "待支付"
	case OrderStatusPendingAccept:
		return "待接单"
	case OrderStatusMaking:
		return "制作中"
	case OrderStatusWaitingPickup:
		return "待取餐"
	case OrderStatusCompleted:
		return "已完成"
	case OrderStatusCancelled:
		return "已取消"
	case OrderStatusRefunding:
		return "退款中"
	case OrderStatusRefunded:
		return "已退款"
	default:
		return "未知状态"
	}
}

// PayStatus 是订单支付状态，取值与 pay 模块支付单状态对齐。
type PayStatus int8

const (
	// PayStatusUnpaid 未支付
	PayStatusUnpaid PayStatus = 0
	// PayStatusPaid 已支付
	PayStatusPaid PayStatus = 10
	// PayStatusRefunded 已退款
	PayStatusRefunded PayStatus = 20
)

// String 返回支付状态中文描述。
func (p PayStatus) String() string {
	switch p {
	case PayStatusUnpaid:
		return "未支付"
	case PayStatusPaid:
		return "已支付"
	case PayStatusRefunded:
		return "已退款"
	default:
		return "未知状态"
	}
}

// ItemStatus 是订单明细状态。
type ItemStatus int8

const (
	// ItemStatusNormal 正常
	ItemStatusNormal ItemStatus = 1
	// ItemStatusRefunded 已退
	ItemStatusRefunded ItemStatus = 2
)

// OperatorType 是订单日志的操作方。
type OperatorType int8

const (
	// OperatorTypeMember 顾客
	OperatorTypeMember OperatorType = 1
	// OperatorTypeStaff 店员
	OperatorTypeStaff OperatorType = 2
	// OperatorTypeSystem 系统
	OperatorTypeSystem OperatorType = 3
)
