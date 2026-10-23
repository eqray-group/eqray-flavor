package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// PayRefundCreateRequest 是发起退款的入参。
// 退款金额单位为分，后端会校验不超过支付单可退金额；商户退款单号同一应用下不可重复。
type PayRefundCreateRequest struct {
	// OrderID 支付单主键，雪花 ID 按字符串提交
	OrderID types.ID `form:"orderId" json:"orderId"`
	// MerchantRefundID 业务退款单号，用于退款幂等
	MerchantRefundID string `form:"merchantRefundId" json:"merchantRefundId"`
	// RefundPrice 本次退款金额（分）
	RefundPrice int64 `form:"refundPrice" json:"refundPrice"`
	// Reason 退款原因
	Reason string `form:"reason" json:"reason"`
	// UserIP 操作人 IP，渠道风控使用
	UserIP string `form:"userIp" json:"userIp"`
	// NeedAudit 是否需要审核：true 时只生成退款单，审核通过后才真正退钱
	NeedAudit bool `form:"needAudit" json:"needAudit"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayRefundCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"merchantRefundId": "required|max:64",
		"refundPrice":      "required|integer|min:1",
		"reason":           "max:255",
		"userIp":           "max:64",
	}
}

// PayRefundAuditRequest 是退款审核的入参。
// 通过时备注可选，驳回时必须填写原因，便于后续追查。
type PayRefundAuditRequest struct {
	// Remark 审核意见 / 驳回原因
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *PayRefundAuditRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"remark": "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *PayRefundAuditRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"remark.max": "审核意见不能超过 255 个字符",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayRefundCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"merchantRefundId.required": "商户退款单号不能为空",
		"merchantRefundId.max":      "商户退款单号不能超过 64 个字符",
		"refundPrice.required":      "退款金额不能为空",
		"refundPrice.integer":       "退款金额必须是整数（单位分）",
		"refundPrice.min":           "退款金额必须大于 0",
		"reason.max":                "退款原因不能超过 255 个字符",
		"userIp.max":                "操作人 IP 不能超过 64 个字符",
	}
}
