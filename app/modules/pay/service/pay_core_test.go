package service

import (
	"encoding/json"
	"testing"
	"time"

	payEnums "eqray-flavor/app/modules/pay/enums"
)

// TestPayStatusEnumValues 锁定支付模块三个核心枚举的取值。
// 取值必须与 yudao 一致，改动会直接导致历史数据语义错乱，因此用测试兜住。
func TestPayStatusEnumValues(t *testing.T) {
	orderCases := map[payEnums.PayOrderStatusEnum]int8{
		payEnums.PayOrderStatusWaiting: 0,
		payEnums.PayOrderStatusSuccess: 10,
		payEnums.PayOrderStatusClosed:  20,
	}
	for status, want := range orderCases {
		if int8(status) != want {
			t.Fatalf("支付单状态 %s 取值 %d，期望 %d", status, int8(status), want)
		}
	}

	refundCases := map[payEnums.PayRefundStatusEnum]int8{
		payEnums.PayRefundStatusCreate:  0,
		payEnums.PayRefundStatusSuccess: 1,
		payEnums.PayRefundStatusFail:    2,
		payEnums.PayRefundStatusClose:   99,
	}
	for status, want := range refundCases {
		if int8(status) != want {
			t.Fatalf("退款单状态 %s 取值 %d，期望 %d", status, int8(status), want)
		}
	}

	notifyCases := map[payEnums.PayNotifyStatusEnum]int8{
		payEnums.PayNotifyStatusWaiting:        1,
		payEnums.PayNotifyStatusSuccess:        2,
		payEnums.PayNotifyStatusFail:           3,
		payEnums.PayNotifyStatusRequestSuccess: 4,
		payEnums.PayNotifyStatusRequestFail:    5,
	}
	for status, want := range notifyCases {
		if int8(status) != want {
			t.Fatalf("通知状态 %s 取值 %d，期望 %d", status, int8(status), want)
		}
	}

	if payEnums.PayNotifyTypeOrder != 1 || payEnums.PayNotifyTypeRefund != 2 {
		t.Fatal("通知类型必须是 1 支付单 / 2 退款单")
	}
}

// TestPayOrderCanRefund 验证只有支付成功的单可以发起退款。
func TestPayOrderCanRefund(t *testing.T) {
	if !payEnums.PayOrderStatusSuccess.CanRefund() {
		t.Fatal("支付成功的单应允许退款")
	}
	if payEnums.PayOrderStatusWaiting.CanRefund() {
		t.Fatal("待支付的单不应允许退款")
	}
	if payEnums.PayOrderStatusClosed.CanRefund() {
		t.Fatal("已关闭的单不应允许退款")
	}
}

// TestPayNotifyRetryable 验证通知任务的可重试状态集合。
func TestPayNotifyRetryable(t *testing.T) {
	if !payEnums.PayNotifyStatusWaiting.IsRetryable() ||
		!payEnums.PayNotifyStatusRequestSuccess.IsRetryable() ||
		!payEnums.PayNotifyStatusRequestFail.IsRetryable() {
		t.Fatal("等待 / 请求成功但结果失败 / 请求失败都应继续重试")
	}
	if payEnums.PayNotifyStatusSuccess.IsRetryable() || payEnums.PayNotifyStatusFail.IsRetryable() {
		t.Fatal("成功与最终失败都是终态，不应再重试")
	}
}

// TestNotifyBackoffMatchesYudao 验证重试频率与最大次数和 yudao 一致。
func TestNotifyBackoffMatchesYudao(t *testing.T) {
	want := []time.Duration{
		15 * time.Second, 15 * time.Second, 30 * time.Second, 3 * time.Minute,
		30 * time.Minute, 30 * time.Minute, 30 * time.Minute, time.Hour,
	}
	if len(payNotifyBackoff) != len(want) {
		t.Fatalf("重试档位数 %d，期望 %d", len(payNotifyBackoff), len(want))
	}
	for index, duration := range want {
		if payNotifyBackoff[index] != duration {
			t.Fatalf("第 %d 档重试间隔 %s，期望 %s", index+1, payNotifyBackoff[index], duration)
		}
	}
	if payNotifyMaxTimes != 9 {
		t.Fatalf("最大通知次数 %d，期望 9（8 档 + 首次）", payNotifyMaxTimes)
	}
}

// TestBackoffDuration 验证按已通知次数取间隔，超出档位沿用最后一档。
func TestBackoffDuration(t *testing.T) {
	if got := backoffDuration(1); got != 15*time.Second {
		t.Fatalf("首次重试应等 15s，实际 %s", got)
	}
	if got := backoffDuration(8); got != time.Hour {
		t.Fatalf("第 8 次重试应等 1h，实际 %s", got)
	}
	if got := backoffDuration(99); got != time.Hour {
		t.Fatalf("超出档位应沿用最后一档，实际 %s", got)
	}
	if got := backoffDuration(0); got != 15*time.Second {
		t.Fatalf("非法次数应按第一档处理，实际 %s", got)
	}
}

// TestCalcChannelFeePrice 验证渠道手续费按百分比计算并四舍五入到分。
func TestCalcChannelFeePrice(t *testing.T) {
	cases := []struct {
		price   int64
		feeRate float64
		want    int64
	}{
		{10000, 0.6, 60},
		{10000, 0, 0},
		{0, 0.6, 0},
		{999, 0.6, 6},   // 5.994 → 6
		{100, 0.005, 0}, // 0.005 → 0
		{100, 0.5, 1},   // 0.5 → 1（四舍五入）
		{-100, 0.6, 0},  // 负数直接返回 0
		{10000, 100, 10000},
	}
	for _, item := range cases {
		if got := calcChannelFeePrice(item.price, item.feeRate); got != item.want {
			t.Fatalf("金额 %d 费率 %v 的手续费为 %d，期望 %d", item.price, item.feeRate, got, item.want)
		}
	}
}

// TestNormalizePayOrderInput 验证下单入参清洗与校验。
func TestNormalizePayOrderInput(t *testing.T) {
	// 合法入参：去除首尾空格，有效期缺省 30 分钟
	valid := PayOrderCreateInput{
		MerchantOrderID: " M001 ",
		Subject:         " 堂食订单 ",
		Price:           100,
	}
	if err := normalizePayOrderInput(&valid); err != nil {
		t.Fatalf("合法入参被拒：%v", err)
	}
	if valid.MerchantOrderID != "M001" || valid.Subject != "堂食订单" {
		t.Fatal("首尾空格未被清理")
	}
	if valid.ExpireMinutes != payOrderExpireMinutesDefault {
		t.Fatalf("默认有效期 %d，期望 %d", valid.ExpireMinutes, payOrderExpireMinutesDefault)
	}

	// 有效期超上限时被截断，避免长期占用商户订单号
	tooLong := PayOrderCreateInput{MerchantOrderID: "M002", Subject: "订单", Price: 1, ExpireMinutes: 99999}
	if err := normalizePayOrderInput(&tooLong); err != nil {
		t.Fatalf("合法入参被拒：%v", err)
	}
	if tooLong.ExpireMinutes != payOrderExpireMinutesMax {
		t.Fatalf("有效期应被截断到 %d，实际 %d", payOrderExpireMinutesMax, tooLong.ExpireMinutes)
	}

	// 非法入参：订单号为空、标题为空、标题超长、金额非正
	invalid := []PayOrderCreateInput{
		{MerchantOrderID: "", Subject: "订单", Price: 100},
		{MerchantOrderID: "M003", Subject: "", Price: 100},
		{MerchantOrderID: "M004", Subject: string(make([]rune, payOrderSubjectMaxLength+1)), Price: 100},
		{MerchantOrderID: "M005", Subject: "订单", Price: 0},
		{MerchantOrderID: "M006", Subject: "订单", Price: -1},
	}
	for index, item := range invalid {
		if err := normalizePayOrderInput(&item); err == nil {
			t.Fatalf("第 %d 组非法入参未被拒绝", index+1)
		}
	}
}

// TestMaskAndMergeSensitiveConfig 验证渠道配置的密钥脱敏与掩码回填。
func TestMaskAndMergeSensitiveConfig(t *testing.T) {
	raw := `{"appId":"wx123","mchId":"1600001","apiV3Key":"abcdef123456","privateKey":"MIIEvQ..."}`

	masked := maskConfigJSON(raw)
	var maskedMap map[string]any
	if err := json.Unmarshal([]byte(masked), &maskedMap); err != nil {
		t.Fatalf("脱敏结果不是合法 JSON：%v", err)
	}
	if maskedMap["apiV3Key"] == "abcdef123456" || maskedMap["privateKey"] == "MIIEvQ..." {
		t.Fatal("密钥字段未被脱敏")
	}
	if maskedMap["appId"] != "wx123" {
		t.Fatal("非敏感字段不应被改动")
	}

	// 编辑时回传掩码 → 沿用旧值；回传新值 → 用新值
	oldConfig := `{"apiV3Key":"old-key-1234","appId":"wx123"}`
	merged := mergeSensitiveConfig(oldConfig, `{"apiV3Key":"****1234","appId":"wx456"}`)
	var mergedMap map[string]any
	if err := json.Unmarshal([]byte(merged), &mergedMap); err != nil {
		t.Fatalf("合并结果不是合法 JSON：%v", err)
	}
	if mergedMap["apiV3Key"] != "old-key-1234" {
		t.Fatalf("掩码值应沿用旧密钥，实际 %v", mergedMap["apiV3Key"])
	}
	if mergedMap["appId"] != "wx456" {
		t.Fatal("普通字段应使用新值")
	}

	if got := maskConfigJSON(""); got != "" {
		t.Fatal("空配置应原样返回")
	}
	if got := maskConfigJSON("not-json"); got != "not-json" {
		t.Fatal("非法 JSON 应原样返回")
	}
}

// TestSignPayload 验证回调签名：同样入参结果稳定，任一要素变化结果不同。
func TestSignPayload(t *testing.T) {
	base := signPayload("secret", "app-key", 1700000000, "nonce-1", `{"a":1}`)
	if base == "" {
		t.Fatal("签名不应为空")
	}
	if base != signPayload("secret", "app-key", 1700000000, "nonce-1", `{"a":1}`) {
		t.Fatal("同样入参的签名必须一致")
	}

	changed := []string{
		signPayload("secret2", "app-key", 1700000000, "nonce-1", `{"a":1}`),
		signPayload("secret", "app-key-2", 1700000000, "nonce-1", `{"a":1}`),
		signPayload("secret", "app-key", 1700000001, "nonce-1", `{"a":1}`),
		signPayload("secret", "app-key", 1700000000, "nonce-2", `{"a":1}`),
		signPayload("secret", "app-key", 1700000000, "nonce-1", `{"a":2}`),
	}
	for index, signature := range changed {
		if signature == base {
			t.Fatalf("第 %d 个要素变化后签名未改变", index+1)
		}
	}
}

// TestNotifyTypeText 验证通知类型到业务方可读类型名的映射。
func TestNotifyTypeText(t *testing.T) {
	if got := notifyTypeText(payEnums.PayNotifyTypeOrder); got != "order" {
		t.Fatalf("支付通知类型名 %s，期望 order", got)
	}
	if got := notifyTypeText(payEnums.PayNotifyTypeRefund); got != "refund" {
		t.Fatalf("退款通知类型名 %s，期望 refund", got)
	}
	if got := notifyTypeText(payEnums.PayNotifyTypeEnum(99)); got != "unknown" {
		t.Fatalf("未知类型的类型名 %s，期望 unknown", got)
	}
}
