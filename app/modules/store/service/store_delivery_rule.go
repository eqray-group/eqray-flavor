package service

import (
	"encoding/json"
	"strings"

	baseErrs "eqray-flavor/app/modules/base/errs"
)

// DeliveryRule 是 store.delivery_rule 的 JSON 结构，字段单位统一：
// 距离用米、金额用分、时长用分钟，避免前端传元导致精度问题。
// 外卖开关在门店上，本结构只描述「配送到哪、怎么收费」，不承载开关语义。
type DeliveryRule struct {
	// Radius 配送半径（米），0 表示不限范围
	Radius int `json:"radius"`
	// MinAmount 起送价（分），0 表示无起送门槛
	MinAmount int `json:"minAmount"`
	// DeliveryFee 配送费（分）
	DeliveryFee int `json:"deliveryFee"`
	// PackingFee 每份打包费（分）
	PackingFee int `json:"packingFee"`
	// EtaMinutes 预计送达时长（分钟），0 表示未设置
	EtaMinutes int `json:"etaMinutes"`
}

// ParseDeliveryRule 把库中的 JSON 解析成结构体，空串视为未配置并返回零值。
// 供下单与试算复用，解析失败返回业务错误码而不是裸 error。
func ParseDeliveryRule(raw string) (DeliveryRule, error) {
	rule := DeliveryRule{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return rule, nil
	}
	if err := json.Unmarshal([]byte(raw), &rule); err != nil {
		return DeliveryRule{}, baseErrs.WithMsg(baseErrs.BadRequest, "配送规则不是合法的 JSON")
	}

	return rule, nil
}

// normalizeDeliveryRule 校验配送规则并返回统一序列化结果。
// 空串表示未配置，原样返回；金额与时长不允许为负，半径 0 表示不限范围。
func normalizeDeliveryRule(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	rule, err := ParseDeliveryRule(raw)
	if err != nil {
		return "", err
	}
	if rule.Radius < 0 {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "配送半径不能为负数，不限范围请填 0")
	}
	if rule.MinAmount < 0 {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "起送价不能为负数")
	}
	if rule.DeliveryFee < 0 {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "配送费不能为负数")
	}
	if rule.PackingFee < 0 {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "打包费不能为负数")
	}
	if rule.EtaMinutes < 0 {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "预计送达时长不能为负数")
	}

	data, err := json.Marshal(rule)
	if err != nil {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "配送规则序列化失败")
	}

	return string(data), nil
}
