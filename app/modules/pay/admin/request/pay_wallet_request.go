package request

import (
	"github.com/goravel/framework/contracts/http"
)

// PayWalletChangeRequest 是后台调整钱包余额的入参，金额单位为分。
// price 正数表示加款、负数表示扣款，0 由服务层拒绝并返回金额非法。
type PayWalletChangeRequest struct {
	// Price 变动金额（分），正数加款 / 负数扣款
	Price int64 `form:"price" json:"price"`
	// Remark 调整原因，会写进钱包流水
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayWalletChangeRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"price":  "required|integer",
		"remark": "max:255",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayWalletChangeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"price.required": "变动金额不能为空",
		"price.integer":  "变动金额必须是整数（单位分）",
		"remark.max":     "调整原因不能超过 255 个字符",
	}
}
