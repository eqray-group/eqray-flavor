// order_core_test.go 锁定订单模块的核心口径：状态机边界、枚举取值与通知报文解析。
// 这些是「错了就出事故」的地方，用单测兜住，避免后续改动悄悄破坏。
package service

import (
	"encoding/json"
	"testing"

	orderEnums "eqray-flavor/app/modules/order/enums"
)

// TestOrderStatusTerminal 验证终态判定：已取消与已退款是终态，其余不是。
func TestOrderStatusTerminal(t *testing.T) {
	terminals := map[orderEnums.OrderStatus]bool{
		orderEnums.OrderStatusCancelled: true,
		orderEnums.OrderStatusRefunded:  true,
	}
	statuses := []orderEnums.OrderStatus{
		orderEnums.OrderStatusPendingPay,
		orderEnums.OrderStatusPendingAccept,
		orderEnums.OrderStatusMaking,
		orderEnums.OrderStatusWaitingPickup,
		orderEnums.OrderStatusCompleted,
		orderEnums.OrderStatusRefunding,
	}
	for _, status := range statuses {
		if status.IsTerminal() {
			t.Fatalf("状态 %s 不应是终态", status)
		}
	}
	if !orderEnums.OrderStatusCancelled.IsTerminal() || !orderEnums.OrderStatusRefunded.IsTerminal() {
		t.Fatal("已取消与已退款都应是终态")
	}
	for status, want := range terminals {
		if status.IsTerminal() != want {
			t.Fatalf("状态 %s 终态判定应为 %v", status, want)
		}
	}
}

// TestOrderStatusIsPaid 验证「已收款」判定：待支付不算，已取消不算，其余支付后状态都算。
func TestOrderStatusIsPaid(t *testing.T) {
	if orderEnums.OrderStatusPendingPay.IsPaid() {
		t.Fatal("待支付不应算作已收款")
	}
	if orderEnums.OrderStatusCancelled.IsPaid() {
		t.Fatal("已取消不应算作已收款")
	}
	paid := []orderEnums.OrderStatus{
		orderEnums.OrderStatusPendingAccept,
		orderEnums.OrderStatusMaking,
		orderEnums.OrderStatusWaitingPickup,
		orderEnums.OrderStatusCompleted,
		orderEnums.OrderStatusRefunding,
		orderEnums.OrderStatusRefunded,
	}
	for _, status := range paid {
		if !status.IsPaid() {
			t.Fatalf("状态 %s 应算作已收款", status)
		}
	}
}

// TestOrderStatusEnumValues 锁定订单状态取值：百位段位，改动会让历史数据语义错乱。
func TestOrderStatusEnumValues(t *testing.T) {
	cases := map[orderEnums.OrderStatus]int16{
		orderEnums.OrderStatusPendingPay:    100,
		orderEnums.OrderStatusPendingAccept: 200,
		orderEnums.OrderStatusMaking:        300,
		orderEnums.OrderStatusWaitingPickup: 400,
		orderEnums.OrderStatusCompleted:     500,
		orderEnums.OrderStatusCancelled:     600,
		orderEnums.OrderStatusRefunding:     700,
		orderEnums.OrderStatusRefunded:      800,
	}
	for status, want := range cases {
		if int16(status) != want {
			t.Fatalf("订单状态 %s 取值 %d，期望 %d", status, int16(status), want)
		}
	}
}

// TestOrderTypeAndChannelValues 锁定订单类型与下单渠道取值。
func TestOrderTypeAndChannelValues(t *testing.T) {
	typeCases := map[orderEnums.OrderType]int8{
		orderEnums.OrderTypeDineIn:  1,
		orderEnums.OrderTypeTakeout: 2,
		orderEnums.OrderTypePickup:  3,
	}
	for typ, want := range typeCases {
		if int8(typ) != want {
			t.Fatalf("订单类型 %s 取值 %d，期望 %d", typ, int8(typ), want)
		}
	}

	channelCases := map[orderEnums.OrderChannel]int8{
		orderEnums.OrderChannelMiniProgram:      1,
		orderEnums.OrderChannelPOS:              2,
		orderEnums.OrderChannelKiosk:            3,
		orderEnums.OrderChannelDeliveryPlatform: 4,
	}
	for channel, want := range channelCases {
		if int8(channel) != want {
			t.Fatalf("下单渠道 %s 取值 %d，期望 %d", channel, int8(channel), want)
		}
	}
}

// TestPayStatusValues 锁定订单支付状态取值，与 pay 模块保持一致：0 未支付 / 10 已支付 / 20 已退款。
func TestPayStatusValues(t *testing.T) {
	cases := map[orderEnums.PayStatus]int8{
		orderEnums.PayStatusUnpaid:   0,
		orderEnums.PayStatusPaid:     10,
		orderEnums.PayStatusRefunded: 20,
	}
	for status, want := range cases {
		if int8(status) != want {
			t.Fatalf("支付状态 %s 取值 %d，期望 %d", status, int8(status), want)
		}
	}
}

// TestOrderNoFormat 验证订单号格式：O 前缀 + 时间戳 + 雪花尾 6 位，全局唯一且可读。
func TestOrderNoFormat(t *testing.T) {
	first := newOrderNo()
	second := newOrderNo()

	if len(first) != 1+14+6 {
		t.Fatalf("订单号长度 %d，期望 21", len(first))
	}
	if first[0] != 'O' {
		t.Fatalf("订单号应以 O 开头，实际 %s", first)
	}
	if first == second {
		t.Fatal("连续生成的订单号不应相同")
	}
}

// TestPickupCodeFormat 验证取餐码为 4 位数字，落在 1000 ~ 9999。
func TestPickupCodeFormat(t *testing.T) {
	for i := 0; i < 200; i++ {
		code := newPickupCode()
		if len(code) != 4 {
			t.Fatalf("取餐码长度 %d，期望 4：%s", len(code), code)
		}
		for _, char := range code {
			if char < '0' || char > '9' {
				t.Fatalf("取餐码应全为数字：%s", code)
			}
		}
	}
}

// TestPayNotifyPayloadParse 验证支付通知报文能解析出关键字段，字段名与 pay 模块推送的一致。
func TestPayNotifyPayloadParse(t *testing.T) {
	body := []byte(`{"type":"order","appKey":"ak_test","merchantOrderId":"O20260101120000abcdef","orderNo":"P20260101120000abcdef","channelCode":"wechat_mini","channelOrderNo":"wx123","price":8800,"successTime":"2026-01-01 12:00:00"}`)

	var payload PayNotifyPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("通知报文解析失败：%v", err)
	}
	if payload.Type != "order" {
		t.Fatalf("通知类型 %q，期望 order", payload.Type)
	}
	if payload.MerchantOrderID != "O20260101120000abcdef" {
		t.Fatalf("商户订单号 %q 不匹配", payload.MerchantOrderID)
	}
	if payload.OrderNo != "P20260101120000abcdef" {
		t.Fatalf("支付单号 %q 不匹配", payload.OrderNo)
	}
	if payload.Price != 8800 {
		t.Fatalf("支付金额 %d，期望 8800", payload.Price)
	}
}

// TestMarshalSnapshot 验证优惠快照能序列化出金额与步骤，且是合法 JSON。
func TestMarshalSnapshot(t *testing.T) {
	pricing := &PricingResult{
		ItemAmount:         10_000,
		AddonAmount:        500,
		LineDiscountAmount: 200,
		ActivityAmount:     1_000,
		CouponAmount:       300,
		PointsAmount:       100,
		PayAmount:          9_100,
		Steps:              []PricingStep{{Step: 1, Name: "商品原价", Amount: 10_000}},
	}

	snapshot := marshalSnapshot(pricing)
	if snapshot == "" {
		t.Fatal("优惠快照不应为空")
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(snapshot), &decoded); err != nil {
		t.Fatalf("优惠快照不是合法 JSON：%v", err)
	}
	if decoded["payAmount"] != float64(9_100) {
		t.Fatalf("快照应付金额 %v，期望 9100", decoded["payAmount"])
	}
	if decoded["itemAmount"] != float64(10_000) {
		t.Fatalf("快照商品小计 %v，期望 10000", decoded["itemAmount"])
	}
}
