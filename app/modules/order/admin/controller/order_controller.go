// Package controller 提供订单模块的后台管理接口。
package controller

import (
	"time"

	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/order/admin/request"
	orderEnums "eqray-flavor/app/modules/order/enums"
	"eqray-flavor/app/modules/order/service"
)

// buildPageQuery 把查询串转成服务层查询条件，列表与导出共用同一套筛选口径。
func buildPageQuery(ctx http.Context, req request.OrderPageRequest) service.PageQuery {
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

	return q
}

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

	q := buildPageQuery(ctx, req)

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

// UpdateRemark 修改订单备注：顾客备注与商家备注分开更新。
// 入参：路径 id + remark、merchantRemark；返回空数据表示修改成功。
func (c *OrderController) UpdateRemark(ctx http.Context) http.Response {
	var req request.OrderRemarkRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.UpdateRemark(ctx.Context(), service.UpdateRemarkInput{
		OrderID:        id,
		Remark:         req.Remark,
		MerchantRemark: req.MerchantRemark,
		UserID:         types.ID(support.UserID(ctx.Context())),
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// AdjustPrice 对未支付订单改价（加收或减免）。
// 入参：路径 id + 差额 delta（分）；返回空数据表示改价成功。
func (c *OrderController) AdjustPrice(ctx http.Context) http.Response {
	var req request.OrderAdjustPriceRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.AdjustPrice(ctx.Context(), service.AdjustPriceInput{
		OrderID: id,
		Delta:   req.Delta,
		UserID:  types.ID(support.UserID(ctx.Context())),
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// AssignRider 给外卖订单派单。
// 入参：路径 id + riderId；返回空数据表示派单成功。
func (c *OrderController) AssignRider(ctx http.Context) http.Response {
	var req request.OrderAssignRiderRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.AssignRider(ctx.Context(), service.AssignRiderInput{
		OrderID: id,
		RiderID: req.RiderID,
		UserID:  types.ID(support.UserID(ctx.Context())),
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Riders 返回该订单门店下可接单的骑手，供派单弹窗选择。
// 入参：路径 id；返回骑手列表（含当前负载与是否还有额度）。
func (c *OrderController) Riders(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	options, err := c.service.DispatchableRiders(ctx.Context(), id, types.ID(support.UserID(ctx.Context())))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, options)
}

// CancelDispatch 取消订单派单并释放骑手额度。
// 入参：路径 id；返回空数据表示取消成功。
func (c *OrderController) CancelDispatch(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.CancelDispatch(ctx.Context(), id, types.ID(support.UserID(ctx.Context()))); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Ticket 返回打印小票所需数据，用于前台补打 / 重打。
// 入参：路径 id；返回门店抬头、明细（含加料）与金额拆分。
func (c *OrderController) Ticket(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	ticket, err := c.service.Ticket(ctx.Context(), id, types.ID(support.UserID(ctx.Context())))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, ticket)
}

// Export 按当前筛选条件导出订单 CSV（UTF-8 BOM，Excel 可直接打开）。
// 入参：与列表一致的筛选条件；返回 CSV 文件流，单次最多导出一万条。
func (c *OrderController) Export(ctx http.Context) http.Response {
	var req request.OrderPageRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	data, err := c.service.Export(ctx.Context(), buildPageQuery(ctx, req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	filename := "orders-" + time.Now().Format("20060102150405") + ".csv"

	return ctx.Response().
		Header("Content-Disposition", `attachment; filename="`+filename+`"`).
		Data(http.StatusOK, "text/csv; charset=utf-8", data)
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
