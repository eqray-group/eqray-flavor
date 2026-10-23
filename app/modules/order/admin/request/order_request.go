// Package request 定义订单管理端接口的请求结构。
package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// OrderTransitRequest 是店员对订单做状态流转的入参（接单 / 出餐 / 完成 / 拒单等）。
type OrderTransitRequest struct {
	// To 目标状态值，取值见字典 cater_order_status
	To int16 `form:"to" json:"to"`
	// Reason 变更原因，拒单时必填给顾客看
	Reason string `form:"reason" json:"reason"`
}

// Rules 返回校验规则。
func (r *OrderTransitRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"to":     "required|integer|min:100|max:800",
		"reason": "max:255",
	}
}

// Messages 返回中文提示。
func (r *OrderTransitRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"to.required": "目标状态不能为空",
		"to.integer":  "目标状态必须是整数",
	}
}

// OrderRemarkRequest 是修改订单备注的入参。
// 顾客备注与商家备注分列存放，店员修改商家备注不会覆盖顾客原始诉求。
type OrderRemarkRequest struct {
	// Remark 顾客备注，空串表示清空
	Remark string `form:"remark" json:"remark"`
	// MerchantRemark 商家备注，空串表示清空
	MerchantRemark string `form:"merchantRemark" json:"merchantRemark"`
}

// Rules 返回校验规则。
func (r *OrderRemarkRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"remark":         "max:255",
		"merchantRemark": "max:255",
	}
}

// Messages 返回中文提示。
func (r *OrderRemarkRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"remark.max":         "顾客备注不能超过 255 个字符",
		"merchantRemark.max": "商家备注不能超过 255 个字符",
	}
}

// OrderAdjustPriceRequest 是订单改价入参，金额单位为分。
type OrderAdjustPriceRequest struct {
	// Delta 改价差额（分）：正数加价、负数减免
	Delta int64 `form:"delta" json:"delta"`
}

// Rules 返回校验规则。
func (r *OrderAdjustPriceRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"delta": "required|integer",
	}
}

// Messages 返回中文提示。
func (r *OrderAdjustPriceRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"delta.required": "改价金额不能为空",
		"delta.integer":  "改价金额必须是整数（单位分）",
	}
}

// OrderAssignRiderRequest 是订单派单入参。
type OrderAssignRiderRequest struct {
	// RiderID 骑手 ID
	RiderID types.ID `form:"riderId" json:"riderId"`
}

// Rules 返回校验规则。
func (r *OrderAssignRiderRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"riderId": "required",
	}
}

// Messages 返回中文提示。
func (r *OrderAssignRiderRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"riderId.required": "请选择配送骑手",
	}
}

// OrderPageRequest 是订单列表查询入参。
type OrderPageRequest struct {
	// StoreID 门店，缺省按当前用户可见门店集合
	StoreID types.ID `form:"storeId" json:"storeId"`
	// Status 订单状态，缺省不限
	Status int16 `form:"status" json:"status"`
	// PayStatus 支付状态，缺省不限
	PayStatus int8 `form:"payStatus" json:"payStatus"`
	// Type 订单类型，缺省不限
	Type int8 `form:"type" json:"type"`
	// Keyword 订单号 / 取餐码 / 收货人电话
	Keyword string `form:"keyword" json:"keyword"`
	// PageNo 页码
	PageNo int `form:"pageNo" json:"pageNo"`
	// PageSize 页大小
	PageSize int `form:"pageSize" json:"pageSize"`
}

// Rules 返回校验规则。
func (r *OrderPageRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"keyword": "max:64",
	}
}

// Messages 返回中文提示。
func (r *OrderPageRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}
