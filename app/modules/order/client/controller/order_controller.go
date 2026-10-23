// Package controller 提供订单模块的 C 端接口（顾客侧）。
// 身份只从会员令牌取，请求体不收会员 ID；门店由参数显式传入并校验营业状态。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/order/client/request"
	orderEnums "eqray-flavor/app/modules/order/enums"
	"eqray-flavor/app/modules/order/service"
)

// OrderController 提供试算、下单、我的订单、取消与发起支付接口。
type OrderController struct {
	service *service.OrderService
}

// NewOrderController 构造 C 端订单控制器。
func NewOrderController() *OrderController {
	return &OrderController{service: service.NewOrderService()}
}

// Estimate 价格试算：按门店上下文与会员身份返回逐行计价与合计。
// 入参：门店、明细行（含加料）；返回金额拆分与计价步骤。
func (c *OrderController) Estimate(ctx http.Context) http.Response {
	var req request.EstimateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	memberID := support.MemberID(ctx.Context())
	out, err := c.service.Estimate(ctx.Context(), service.EstimateInput{
		StoreID:  req.StoreID,
		IsMember: memberID > 0,
		Items:    toEstimateItems(req.Items),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// Create 下单：幂等靠请求号，重复提交返回同一张订单。
// 入参：门店、类型、渠道、桌台 / 收货信息与明细行；返回订单与是否幂等命中。
func (c *OrderController) Create(ctx http.Context) http.Response {
	var req request.CreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	memberID := support.MemberID(ctx.Context())
	out, err := c.service.Create(ctx.Context(), service.CreateOrderInput{
		StoreID:         req.StoreID,
		MemberID:        memberID,
		Type:            orderEnums.OrderType(req.Type),
		Channel:         orderEnums.OrderChannel(req.Channel),
		TableID:         req.TableID,
		RequestNo:       req.RequestNo,
		Remark:          req.Remark,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		Items:           toCreateItems(req.Items),
		IsMember:        memberID > 0,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// Index 我的订单：分页返回当前会员的订单列表。
// 入参：storeId（可选）、status（可选）、pageNo、pageSize；返回订单列表与总数。
func (c *OrderController) Index(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	q := service.PageQuery{MemberID: &memberID}
	if storeID := param.QueryID(ctx, "storeId"); storeID > 0 {
		q.StoreID = &storeID
	}
	if ctx.Request().Query("status") != "" {
		status := orderEnums.OrderStatus(ctx.Request().QueryInt("status", 0))
		q.Status = &status
	}

	pageNo := ctx.Request().QueryInt("pageNo", 1)
	pageSize := ctx.Request().QueryInt("pageSize", 10)

	rows, total, err := c.service.Page(ctx.Context(), q, pageNo, pageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, rows, total)
}

// Show 查询自己的订单详情。
// 入参：路径 id；非本人订单按不存在处理，避免枚举探测。
func (c *OrderController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	memberID := support.MemberID(ctx.Context())
	detail, err := c.service.MemberDetail(ctx.Context(), id, memberID)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, detail)
}

// Cancel 取消自己的待支付订单。
// 入参：路径 id + 取消原因；已支付订单走退款，不在本接口范围。
func (c *OrderController) Cancel(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.CancelRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Cancel(ctx.Context(), id, support.MemberID(ctx.Context()), req.Reason)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Pay 为待支付订单发起支付，返回渠道调起参数。
// 入参：路径 id + 支付应用 / 渠道 / openId；重复调用返回既有支付单参数（幂等）。
func (c *OrderController) Pay(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PayRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	out, err := c.service.CreatePayment(ctx.Context(), service.CreatePaymentInput{
		OrderID:     id,
		AppID:       req.AppID,
		ChannelCode: req.ChannelCode,
		OpenID:      req.OpenID,
		MemberID:    support.MemberID(ctx.Context()),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// toEstimateItems 把请求行转成试算输入。
func toEstimateItems(lines []request.ClientLine) []service.EstimateItemInput {
	items := make([]service.EstimateItemInput, 0, len(lines))
	for _, line := range lines {
		addons := make([]service.EstimateAddonInput, 0, len(line.Addons))
		for _, addon := range line.Addons {
			addons = append(addons, service.EstimateAddonInput{
				AddonItemID: addon.AddonItemID,
				Quantity:    addon.Quantity,
			})
		}
		items = append(items, service.EstimateItemInput{
			SkuID: line.SkuID, Quantity: line.Quantity, Addons: addons,
		})
	}

	return items
}

// toCreateItems 把请求行转成下单输入。
func toCreateItems(lines []request.ClientLine) []service.CreateOrderItemInput {
	items := make([]service.CreateOrderItemInput, 0, len(lines))
	for _, line := range lines {
		addons := make([]service.EstimateAddonInput, 0, len(line.Addons))
		for _, addon := range line.Addons {
			addons = append(addons, service.EstimateAddonInput{
				AddonItemID: addon.AddonItemID,
				Quantity:    addon.Quantity,
			})
		}
		items = append(items, service.CreateOrderItemInput{
			SkuID: line.SkuID, Quantity: line.Quantity,
			Addons: addons, Orderer: line.Orderer, Remark: line.Remark,
		})
	}

	return items
}
