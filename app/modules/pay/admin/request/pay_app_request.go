// Package request 定义支付模块后台接口的入参与校验规则。
package request

import (
	"github.com/goravel/framework/contracts/http"
)

// PayAppSaveRequest 是新增 / 编辑支付应用的入参。
// appSecret 留空表示由系统生成（新增）或保持原值（编辑），接口不会回显明文。
type PayAppSaveRequest struct {
	// Name 应用名
	Name string `form:"name" json:"name"`
	// Status 开启状态：0 开启 / 1 关闭（yudao 口径）
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
	// OrderNotifyURL 支付结果回调地址
	OrderNotifyURL string `form:"orderNotifyUrl" json:"orderNotifyUrl"`
	// RefundNotifyURL 退款结果回调地址
	RefundNotifyURL string `form:"refundNotifyUrl" json:"refundNotifyUrl"`
	// TransferNotifyURL 转账结果回调地址
	TransferNotifyURL string `form:"transferNotifyUrl" json:"transferNotifyUrl"`
	// AppSecret 应用密钥，留空由系统生成
	AppSecret string `form:"appSecret" json:"appSecret"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayAppSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":              "required|max:64",
		"status":            "required|integer|in:0,1",
		"remark":            "max:255",
		"orderNotifyUrl":    "required|max:255",
		"refundNotifyUrl":   "required|max:255",
		"transferNotifyUrl": "max:255",
		"appSecret":         "max:64",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayAppSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":            "应用名不能为空",
		"name.max":                 "应用名不能超过 64 个字符",
		"status.required":          "应用状态不能为空",
		"status.integer":           "应用状态必须是整数",
		"status.in":                "应用状态只能是 0（开启）或 1（关闭）",
		"remark.max":               "备注不能超过 255 个字符",
		"orderNotifyUrl.required":  "支付回调地址不能为空",
		"orderNotifyUrl.max":       "支付回调地址不能超过 255 个字符",
		"refundNotifyUrl.required": "退款回调地址不能为空",
		"refundNotifyUrl.max":      "退款回调地址不能超过 255 个字符",
		"transferNotifyUrl.max":    "转账回调地址不能超过 255 个字符",
		"appSecret.max":            "应用密钥不能超过 64 个字符",
	}
}
