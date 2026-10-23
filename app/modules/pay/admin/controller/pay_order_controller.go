package controller

import (
	"time"

	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/pay/admin/request"
	"eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/service"
)

// buildPayOrderQuery 把查询串转成服务层筛选条件，列表与导出共用同一套口径。
func buildPayOrderQuery(ctx http.Context) (service.PayOrderListQuery, error) {
	appID, err := queryID(ctx, "appId")
	if err != nil {
		return service.PayOrderListQuery{}, err
	}

	q := service.PayOrderListQuery{
		Keyword:     ctx.Request().Query("keyword"),
		AppID:       appID,
		ChannelCode: ctx.Request().Query("channelCode"),
	}
	// 状态 0（待支付）是合法值，只有参数缺省才表示不限
	if raw := ctx.Request().Query("status"); raw != "" {
		status := enums.PayOrderStatusEnum(int8(ctx.Request().QueryInt("status", 0)))
		q.Status = &status
	}

	return q, nil
}

// PayOrderController 提供支付单的查询、代下单、重新调起与关闭接口。
type PayOrderController struct {
	service *service.PayOrderService
}

// NewPayOrderController 构造支付单控制器。
func NewPayOrderController() *PayOrderController {
	return &PayOrderController{service: service.NewPayOrderService()}
}

// Index 分页查询支付单，支持单号关键字、应用、渠道与支付状态筛选。
// 入参：keyword、appId、channelCode、status、pageNo、pageSize（查询串）；返回支付单列表与总数。
func (c *PayOrderController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q, err := buildPayOrderQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), q, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Export 按当前筛选条件导出支付单 CSV，供财务对账。
// 入参：与列表一致的筛选条件；返回 CSV 文件流，单次最多导出五千条。
func (c *PayOrderController) Export(ctx http.Context) http.Response {
	q, err := buildPayOrderQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	data, err := c.service.Export(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	filename := "pay-orders-" + time.Now().Format("20060102150405") + ".csv"

	return ctx.Response().
		Header("Content-Disposition", `attachment; filename="`+filename+`"`).
		Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// Show 查询支付单详情，附带每次渠道调用的拓展记录。
// 入参：路径 id；返回支付单详情，不存在返回业务错误码。
func (c *PayOrderController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	detail, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, detail)
}

// Store 创建支付单并立即调用渠道下单，返回前端调起参数。
// 入参：应用、渠道、商户订单号、金额（分）等；返回支付单号与调起参数。
func (c *PayOrderController) Store(ctx http.Context) http.Response {
	var req request.PayOrderCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	submitted, err := c.service.Create(ctx.Context(), service.PayOrderCreateInput{
		AppID:           req.AppID,
		ChannelCode:     req.ChannelCode,
		UserID:          req.UserID,
		UserType:        req.UserType,
		MerchantOrderID: req.MerchantOrderID,
		Subject:         req.Subject,
		Body:            req.Body,
		Price:           req.Price,
		NotifyURL:       req.NotifyURL,
		OpenID:          req.OpenID,
		AuthCode:        req.AuthCode,
		UserIP:          req.UserIP,
		ExpireMinutes:   req.ExpireMinutes,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, submitted)
}

// Submit 对仍待支付且未过期的支付单重新调起渠道。
// 入参：路径 id + openId / authCode；返回新的调起参数。
func (c *PayOrderController) Submit(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PayOrderSubmitRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	submitted, err := c.service.ReSubmit(ctx.Context(), id, req.OpenID, req.AuthCode)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, submitted)
}

// Sync 主动查询渠道订单并补齐本地状态，回调丢失时用于人工补单。
// 入参：路径 id；返回最新支付单（已补成已支付时会带成功时间与渠道订单号）。
func (c *PayOrderController) Sync(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	row, err := c.service.Sync(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}

// Close 关闭待支付的支付单，已支付或已退款的单不允许关闭。
// 入参：路径 id；返回 null。
func (c *PayOrderController) Close(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Close(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
