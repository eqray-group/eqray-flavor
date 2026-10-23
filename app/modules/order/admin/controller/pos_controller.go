package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/order/admin/request"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
	"eqray-flavor/app/modules/order/service"
)

// PosController 提供收银端接口：代客试算、代客下单与现金结账。
// 门店一律取请求头 X-Store-Id，由 StoreContext 中间件校验可见范围后写入上下文。
type PosController struct {
	service *service.OrderService
}

// NewPosController 构造 POS 收银控制器。
func NewPosController() *PosController {
	return &PosController{service: service.NewOrderService()}
}

// Estimate 代客试算：与 C 端共用计价引擎，唯一差别是渠道固定 POS。
// 入参：memberId（可选）、items、couponId、usePoints；返回完整的优惠与应付明细。
func (c *PosController) Estimate(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())
	if storeID == 0 {
		return result.Fail(ctx, errs.New(orderErrs.OrderPosStoreRequired))
	}

	var req request.PosEstimateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	out, err := c.service.Estimate(ctx.Context(), service.EstimateInput{
		StoreID:   storeID,
		MemberID:  req.MemberID,
		IsMember:  req.MemberID > 0,
		Items:     toPosEstimateItems(req.Items),
		CouponID:  req.CouponID,
		UsePoints: req.UsePoints,
		Channel:   int8(orderEnums.OrderChannelPOS),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// Store 代客下单：散客可直接下单，带 memberId 则按会员价与积分结算。
// 入参：订单类型、桌台 / 收货信息与明细；返回订单与是否幂等命中。
func (c *PosController) Store(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())
	if storeID == 0 {
		return result.Fail(ctx, errs.New(orderErrs.OrderPosStoreRequired))
	}

	var req request.PosCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	out, err := c.service.CreateByStaff(ctx.Context(), service.PosCreateInput{
		StoreID:         storeID,
		MemberID:        req.MemberID,
		Type:            orderEnums.OrderType(req.Type),
		TableID:         req.TableID,
		RequestNo:       req.RequestNo,
		Remark:          req.Remark,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		Items:           toPosCreateItems(req.Items),
		CouponID:        req.CouponID,
		UsePoints:       req.UsePoints,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// SettleCash 现金结账：不生成支付单，直接置订单为已支付并推进到制作 / 待接单。
// 入参：路径 id；返回空数据表示收款完成。
func (c *PosController) SettleCash(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())
	if storeID == 0 {
		return result.Fail(ctx, errs.New(orderErrs.OrderPosStoreRequired))
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.SettleCash(ctx.Context(), service.PosSettleInput{
		OrderID: id,
		StoreID: storeID,
		UserID:  types.ID(support.UserID(ctx.Context())),
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Report 返回当前门店最近 N 天的经营报表。
// 入参：days（默认 7，限 1~31）、top（默认 10，限 1~50）；返回趋势、收款构成与菜品排行。
func (c *PosController) Report(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())

	out, err := c.service.PosReport(
		ctx.Context(),
		storeID,
		types.ID(support.UserID(ctx.Context())),
		ctx.Request().QueryInt("days", 7),
		ctx.Request().QueryInt("top", 10),
	)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// Append 给未支付的堂食订单加菜：整单重新计价后返回新的应付金额。
// 入参：路径 id + items；返回加菜后的整单应付。
func (c *PosController) Append(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())
	if storeID == 0 {
		return result.Fail(ctx, errs.New(orderErrs.OrderPosStoreRequired))
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PosAppendRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	out, err := c.service.AppendOrderItems(ctx.Context(), service.AppendOrderItemsInput{
		OrderID: id,
		StoreID: storeID,
		UserID:  types.ID(support.UserID(ctx.Context())),
		Items:   toPosCreateItems(req.Items),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// Merge 并台：把源单明细并入目标单，源单随后取消。
// 入参：路径 id（保留单）+ sourceOrderId；返回合并后的应付与源台（供清台）。
func (c *PosController) Merge(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())
	if storeID == 0 {
		return result.Fail(ctx, errs.New(orderErrs.OrderPosStoreRequired))
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PosMergeRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	out, err := c.service.MergeOrders(ctx.Context(), service.MergeOrdersInput{
		TargetOrderID: id,
		SourceOrderID: req.SourceOrderID,
		StoreID:       storeID,
		UserID:        types.ID(support.UserID(ctx.Context())),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// Overview 返回当前门店今日的经营概览，供门店端首页展示。
// 入参：无（门店取请求头）；返回营收、客单价、退款额与各状态单量。
func (c *PosController) Overview(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())

	out, err := c.service.PosOverview(ctx.Context(), storeID, types.ID(support.UserID(ctx.Context())))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// Pay 扫码收款：微信返回二维码等顾客扫，支付宝条码同步确认结果。
// 入参：路径 id + payType、authCode；返回支付单号、调起参数与是否已收款。
func (c *PosController) Pay(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())
	if storeID == 0 {
		return result.Fail(ctx, errs.New(orderErrs.OrderPosStoreRequired))
	}

	var req request.PosPayRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	out, err := c.service.PosPay(ctx.Context(), service.PosPayInput{
		OrderID:  id,
		StoreID:  storeID,
		UserID:   types.ID(support.UserID(ctx.Context())),
		PayType:  service.PosPayType(req.PayType),
		AuthCode: req.AuthCode,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// PayStatus 查询收款状态：收银端展示二维码后轮询，成功即返回已收款。
// 入参：路径 id；返回是否已收款与是否仍在等待扫码。
func (c *PosController) PayStatus(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())
	if storeID == 0 {
		return result.Fail(ctx, errs.New(orderErrs.OrderPosStoreRequired))
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	out, err := c.service.PosPayStatus(ctx.Context(), service.PosPayStatusInput{
		OrderID: id,
		StoreID: storeID,
		UserID:  types.ID(support.UserID(ctx.Context())),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, out)
}

// toPosEstimateItems 把请求明细转成试算入参。
func toPosEstimateItems(lines []request.PosLine) []service.EstimateItemInput {
	items := make([]service.EstimateItemInput, 0, len(lines))
	for _, line := range lines {
		items = append(items, service.EstimateItemInput{
			SkuID:    line.SkuID,
			Quantity: line.Quantity,
			Addons:   toPosAddons(line.Addons),
		})
	}

	return items
}

// toPosCreateItems 把请求明细转成下单入参。
func toPosCreateItems(lines []request.PosLine) []service.CreateOrderItemInput {
	items := make([]service.CreateOrderItemInput, 0, len(lines))
	for _, line := range lines {
		items = append(items, service.CreateOrderItemInput{
			SkuID:    line.SkuID,
			Quantity: line.Quantity,
			Addons:   toPosAddons(line.Addons),
			Orderer:  line.Orderer,
			Remark:   line.Remark,
		})
	}

	return items
}

// toPosAddons 把请求里的加料转成服务层入参。
func toPosAddons(addons []request.PosAddonItem) []service.EstimateAddonInput {
	items := make([]service.EstimateAddonInput, 0, len(addons))
	for _, addon := range addons {
		items = append(items, service.EstimateAddonInput{
			AddonItemID: addon.AddonItemID,
			Quantity:    addon.Quantity,
		})
	}

	return items
}
