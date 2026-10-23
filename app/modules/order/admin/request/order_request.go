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
