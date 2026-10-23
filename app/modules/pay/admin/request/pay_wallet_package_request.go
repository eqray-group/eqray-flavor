package request

import (
	"github.com/goravel/framework/contracts/http"
)

// PayWalletPackageSaveRequest 是新增 / 编辑充值套餐的入参，金额单位为分。
// payPrice 是用户实付金额，bonusPrice 是额外赠送金额，到账金额为两者之和。
type PayWalletPackageSaveRequest struct {
	// Name 套餐名
	Name string `form:"name" json:"name"`
	// PayPrice 实付金额（分）
	PayPrice int64 `form:"payPrice" json:"payPrice"`
	// BonusPrice 赠送金额（分）
	BonusPrice int64 `form:"bonusPrice" json:"bonusPrice"`
	// Status 开启状态：0 开启 / 1 关闭
	Status int8 `form:"status" json:"status"`
	// Description 套餐描述
	Description string `form:"description" json:"description"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayWalletPackageSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":        "required|max:64",
		"payPrice":    "required|integer|min:1",
		"bonusPrice":  "integer|min:0",
		"status":      "required|integer|in:0,1",
		"description": "max:255",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayWalletPackageSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":      "套餐名不能为空",
		"name.max":           "套餐名不能超过 64 个字符",
		"payPrice.required":  "套餐支付金额不能为空",
		"payPrice.integer":   "套餐支付金额必须是整数（单位分）",
		"payPrice.min":       "套餐支付金额必须大于 0",
		"bonusPrice.integer": "赠送金额必须是整数（单位分）",
		"bonusPrice.min":     "赠送金额不能为负数",
		"status.required":    "套餐状态不能为空",
		"status.integer":     "套餐状态必须是整数",
		"status.in":          "套餐状态只能是 0（开启）或 1（关闭）",
		"description.max":    "套餐描述不能超过 255 个字符",
	}
}
