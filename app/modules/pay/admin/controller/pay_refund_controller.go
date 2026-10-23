package controller

import (
	"time"

	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/admin/request"
	"eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/service"
)

// buildPayRefundQuery 把查询串转成服务层筛选条件，列表与导出共用同一套口径。
func buildPayRefundQuery(ctx http.Context) (service.PayRefundListQuery, error) {
	appID, err := queryID(ctx, "appId")
	if err != nil {
		return service.PayRefundListQuery{}, err
	}
	orderID, err := queryID(ctx, "orderId")
	if err != nil {
		return service.PayRefundListQuery{}, err
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
	// 审核状态 0（无需审核）也是合法值，只有参数缺省才表示不限
	if raw := ctx.Request().Query("auditStatus"); raw != "" {
		auditStatus := enums.PayRefundAuditStatusEnum(int8(ctx.Request().QueryInt("auditStatus", 0)))
		q.AuditStatus = &auditStatus
	}

	return q, nil
}

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

	q, err := buildPayRefundQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), q, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Export 按当前筛选条件导出售后退款单 CSV，含审核状态便于对账溯源。
// 入参：与列表一致的筛选条件；返回 CSV 文件流。
func (c *PayRefundController) Export(ctx http.Context) http.Response {
	q, err := buildPayRefundQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	data, err := c.service.Export(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	filename := "pay-refunds-" + time.Now().Format("20060102150405") + ".csv"

	return ctx.Response().
		Header("Content-Disposition", `attachment; filename="`+filename+`"`).
		Data(http.StatusOK, "text/csv; charset=utf-8", data)
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
		NeedAudit:        req.NeedAudit,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}

// Approve 审核通过退款申请：钱包单回冲余额，渠道单提交退款。
// 入参：路径 id + remark（审核意见，选填）；返回更新后的退款单。
func (c *PayRefundController) Approve(ctx http.Context) http.Response {
	var req request.PayRefundAuditRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	row, err := c.service.ApproveRefund(ctx.Context(), id, types.ID(support.UserID(ctx.Context())), req.Remark)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}

// Reject 审核驳回退款申请：退款单关闭，不动钱。
// 入参：路径 id + remark（驳回原因，必填）；返回空数据表示驳回成功。
func (c *PayRefundController) Reject(ctx http.Context) http.Response {
	var req request.PayRefundAuditRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.RejectRefund(ctx.Context(), id, types.ID(support.UserID(ctx.Context())), req.Remark); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
