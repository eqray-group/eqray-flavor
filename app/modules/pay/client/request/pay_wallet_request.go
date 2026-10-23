// Package request 定义支付模块 C 端接口的入参。
// C 端入参禁止出现会员 ID，会员身份一律取自令牌。
package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// PayWalletRechargeRequest 是会员发起钱包充值的入参。
type PayWalletRechargeRequest struct {
	// AppID 支付应用编号，雪花 ID 按字符串提交
	AppID types.ID `form:"appId" json:"appId"`
	// ChannelCode 支付渠道编码
	ChannelCode string `form:"channelCode" json:"channelCode"`
	// PackageID 充值套餐编号
	PackageID types.ID `form:"packageId" json:"packageId"`
	// OpenID 微信 JSAPI / 小程序支付必填
	OpenID string `form:"openId" json:"openId"`
	// AuthCode 支付宝条码支付必填
	AuthCode string `form:"authCode" json:"authCode"`
	// UserIP 用户 IP，渠道风控使用
	UserIP string `form:"userIp" json:"userIp"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayWalletRechargeRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"channelCode": "required|max:32",
		"packageId":   "required",
		"openId":      "max:128",
		"authCode":    "max:64",
		"userIp":      "max:64",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayWalletRechargeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"channelCode.required": "渠道编码不能为空",
		"channelCode.max":      "渠道编码不能超过 32 个字符",
		"packageId.required":   "请选择充值套餐",
		"openId.max":           "openid 不能超过 128 个字符",
		"authCode.max":         "付款码不能超过 64 个字符",
		"userIp.max":           "用户 IP 不能超过 64 个字符",
	}
}

// PayWalletPayRequest 是会员用钱包余额支付订单的入参，金额单位为分。
// 付款人身份一律取自令牌，请求体不接受会员 ID。
type PayWalletPayRequest struct {
	// AppID 支付应用编号，雪花 ID 按字符串提交
	AppID types.ID `form:"appId" json:"appId"`
	// MerchantOrderID 业务订单号，用于下单幂等
	MerchantOrderID string `form:"merchantOrderId" json:"merchantOrderId"`
	// Subject 商品标题
	Subject string `form:"subject" json:"subject"`
	// Body 商品描述
	Body string `form:"body" json:"body"`
	// Price 支付金额（分），由业务侧试算后传入
	Price int64 `form:"price" json:"price"`
	// UserIP 用户 IP
	UserIP string `form:"userIp" json:"userIp"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayWalletPayRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"appId":           "required",
		"merchantOrderId": "required|max:64",
		"subject":         "required|max:128",
		"body":            "max:128",
		"price":           "required|integer|min:1",
		"userIp":          "max:50",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayWalletPayRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"appId.required":           "支付应用不能为空",
		"merchantOrderId.required": "商户订单号不能为空",
		"merchantOrderId.max":      "商户订单号不能超过 64 个字符",
		"subject.required":         "商品标题不能为空",
		"subject.max":              "商品标题不能超过 128 个字符",
		"body.max":                 "商品描述不能超过 128 个字符",
		"price.required":           "支付金额不能为空",
		"price.integer":            "支付金额必须是整数（单位分）",
		"price.min":                "支付金额必须大于 0",
		"userIp.max":               "用户 IP 不能超过 50 个字符",
	}
}
