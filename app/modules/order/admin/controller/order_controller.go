// Package controller 提供订单模块的后台管理接口。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/order/admin/request"
	orderEnums "eqray-flavor/app/modules/order/enums"
	"eqray-flavor/app/modules/order/service"
)

// OrderController 提供订单列表、详情与状态流转接口。
type OrderController struct {
	service *service.OrderService
}

// NewOrderController 构造订单控制器。
func NewOrderController() *OrderController {
	return &OrderController{service: service.NewOrderService()}
}

// Index 分页查询订单列表。门店、状态、支付状态与类型可筛选，关键字匹配订单号 / 取餐码 / 收货电话。
// 入参：storeId、status、payStatus、type、keyword、pageNo、pageSize（查询串）；返回订单列表与总数。
func (c *OrderController) Index(ctx http.Context) http.Response {
	var req request.OrderPageRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}
	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 15
	}

	q := service.PageQuery{Keyword: req.Keyword, UserID: types.ID(support.UserID(ctx.Context()))}
	if req.StoreID > 0 {
		storeID := req.StoreID
		q.StoreID = &storeID
	}
	// 状态 0（未支付）等是合法值，只有参数缺省才表示不限
	if ctx.Request().Query("status") != "" {
		status := orderEnums.OrderStatus(req.Status)
		q.Status = &status
	}
	if ctx.Request().Query("payStatus") != "" {
		payStatus := orderEnums.PayStatus(req.PayStatus)
		q.PayStatus = &payStatus
	}
	if req.Type > 0 {
		typ := orderEnums.OrderType(req.Type)
		q.Type = &typ
	}

	rows, total, err := c.service.Page(ctx.Context(), q, pageNo, pageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, rows, total)
}

// Show 查询订单详情：金额拆分、明细（含加料快照）与状态日志。
// 入参：路径 id；返回订单详情，不存在返回业务错误码。
func (c *OrderController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	detail, err := c.service.DetailForUser(ctx.Context(), id, types.ID(support.UserID(ctx.Context())))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, detail)
}

// Transit 执行一次订单状态流转（接单 / 出餐 / 完成 / 拒单等）。
// 入参：路径 id + 目标状态 to 与原因 reason；合法流转由状态机校验，并发由条件更新兜底，全部留痕。
func (c *OrderController) Transit(ctx http.Context) http.Response {
	var req request.OrderTransitRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Transit(ctx.Context(), service.TransitInput{
		OrderID:      id,
		To:           orderEnums.OrderStatus(req.To),
		Reason:       req.Reason,
		OperatorType: int8(orderEnums.OperatorTypeStaff),
		OperatorID:   types.ID(support.UserID(ctx.Context())),
		OperatorName: "店员",
		UserID:       types.ID(support.UserID(ctx.Context())),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Cancel 取消订单（仅待支付状态）；已支付订单走退款流程。
// 入参：路径 id + 取消原因 reason。
func (c *OrderController) Cancel(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Transit(ctx.Context(), service.TransitInput{
		OrderID:      id,
		To:           orderEnums.OrderStatusCancelled,
		Reason:       ctx.Request().Input("reason", ""),
		OperatorType: int8(orderEnums.OperatorTypeStaff),
		OperatorID:   types.ID(support.UserID(ctx.Context())),
		OperatorName: "店员",
		UserID:       types.ID(support.UserID(ctx.Context())),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
