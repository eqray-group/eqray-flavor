package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/pay/admin/request"
	"eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/service"
)

// PayRefundController 提供退款单的查询与发起退款接口。
type PayRefundController struct {
	service *service.PayRefundService
}

// NewPayRefundController 构造退款单控制器。
func NewPayRefundController() *PayRefundController {
	return &PayRefundController{service: service.NewPayRefundService()}
}

// Index 分页查询退款单，支持单号关键字、应用、支付单、渠道与退款状态筛选。
// 入参：keyword、appId、orderId、channelCode、status、pageNo、pageSize（查询串）；返回退款单列表与总数。
func (c *PayRefundController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	appID, err := queryID(ctx, "appId")
	if err != nil {
		return result.Fail(ctx, err)
	}
	orderID, err := queryID(ctx, "orderId")
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.PayRefundListQuery{
		Keyword:     ctx.Request().Query("keyword"),
		AppID:       appID,
		OrderID:     orderID,
		ChannelCode: ctx.Request().Query("channelCode"),
	}
	// 状态 0（退款订单生成）是合法值，只有参数缺省才表示不限
	if raw := ctx.Request().Query("status"); raw != "" {
		status := enums.PayRefundStatusEnum(int8(ctx.Request().QueryInt("status", 0)))
		q.Status = &status
	}

	list, total, err := c.service.Page(ctx.Context(), q, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 查询退款单详情。
// 入参：路径 id；返回退款单详情，不存在返回业务错误码。
func (c *PayRefundController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	row, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}

// Store 发起退款：校验可退金额后调用渠道，渠道同步确认完成的直接置为成功。
// 入参：支付单主键、业务退款单号、退款金额（分）与原因；返回退款单。
func (c *PayRefundController) Store(ctx http.Context) http.Response {
	var req request.PayRefundCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	row, err := c.service.Create(ctx.Context(), service.PayRefundCreateInput{
		OrderID:          req.OrderID,
		MerchantRefundID: req.MerchantRefundID,
		RefundPrice:      req.RefundPrice,
		Reason:           req.Reason,
		UserIP:           req.UserIP,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}
