package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// PayOrderCreateRequest 是后台代下单（创建支付单并调起渠道）的入参。
// 金额单位分，只允许传业务侧试算结果，前端不得自行调整。
type PayOrderCreateRequest struct {
	// AppID 应用编号，雪花 ID 按字符串提交
	AppID types.ID `form:"appId" json:"appId"`
	// ChannelCode 渠道编码，如 wx_lite / alipay_qr
	ChannelCode string `form:"channelCode" json:"channelCode"`
	// UserID 付款用户编号，雪花 ID 按字符串提交
	UserID types.ID `form:"userId" json:"userId"`
	// UserType 用户类型，与会员模块对应
	UserType int8 `form:"userType" json:"userType"`
	// MerchantOrderID 业务订单号，用于下单幂等
	MerchantOrderID string `form:"merchantOrderId" json:"merchantOrderId"`
	// Subject 商品标题
	Subject string `form:"subject" json:"subject"`
	// Body 商品描述
	Body string `form:"body" json:"body"`
	// Price 支付金额（分）
	Price int64 `form:"price" json:"price"`
	// NotifyURL 业务回调地址，留空取应用配置的支付回调地址
	NotifyURL string `form:"notifyUrl" json:"notifyUrl"`
	// OpenID 微信 JSAPI / 小程序支付必填
	OpenID string `form:"openId" json:"openId"`
	// AuthCode 支付宝条码支付必填
	AuthCode string `form:"authCode" json:"authCode"`
	// UserIP 用户 IP，渠道风控使用
	UserIP string `form:"userIp" json:"userIp"`
	// ExpireMinutes 支付单有效期（分钟），0 表示默认 30 分钟
	ExpireMinutes int `form:"expireMinutes" json:"expireMinutes"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayOrderCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"channelCode":     "required|max:32",
		"merchantOrderId": "required|max:64",
		"subject":         "required|max:128",
		"body":            "max:255",
		"price":           "required|integer|min:1",
		"notifyUrl":       "max:1024",
		"openId":          "max:128",
		"authCode":        "max:64",
		"userIp":          "max:64",
		"userType":        "integer|min:0",
		"expireMinutes":   "integer|min:0|max:1440",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayOrderCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"channelCode.required":     "渠道编码不能为空",
		"channelCode.max":          "渠道编码不能超过 32 个字符",
		"merchantOrderId.required": "商户订单号不能为空",
		"merchantOrderId.max":      "商户订单号不能超过 64 个字符",
		"subject.required":         "商品标题不能为空",
		"subject.max":              "商品标题不能超过 128 个字符",
		"body.max":                 "商品描述不能超过 255 个字符",
		"price.required":           "支付金额不能为空",
		"price.integer":            "支付金额必须是整数（单位分）",
		"price.min":                "支付金额必须大于 0",
		"notifyUrl.max":            "回调地址不能超过 1024 个字符",
		"openId.max":               "openid 不能超过 128 个字符",
		"authCode.max":             "付款码不能超过 64 个字符",
		"userIp.max":               "用户 IP 不能超过 64 个字符",
		"userType.integer":         "用户类型必须是整数",
		"userType.min":             "用户类型不能为负数",
		"expireMinutes.integer":    "有效期必须是整数",
		"expireMinutes.min":        "有效期不能为负数",
		"expireMinutes.max":        "有效期不能超过 1440 分钟",
	}
}

// PayOrderSubmitRequest 是待支付单重新调起渠道的入参，只需补渠道需要的用户标识。
type PayOrderSubmitRequest struct {
	// OpenID 微信 JSAPI / 小程序支付必填
	OpenID string `form:"openId" json:"openId"`
	// AuthCode 支付宝条码支付必填
	AuthCode string `form:"authCode" json:"authCode"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayOrderSubmitRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"openId":   "max:128",
		"authCode": "max:64",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayOrderSubmitRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"openId.max":   "openid 不能超过 128 个字符",
		"authCode.max": "付款码不能超过 64 个字符",
	}
}
