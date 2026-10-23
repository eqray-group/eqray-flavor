// Package controller 提供支付模块的后台接口，只做取参、调用服务与统一响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/pay/admin/request"
	"eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/service"
)

// PayAppController 提供支付应用的增删改查与下拉选项接口。
type PayAppController struct {
	service *service.PayAppService
}

// NewPayAppController 构造支付应用控制器。
func NewPayAppController() *PayAppController {
	return &PayAppController{service: service.NewPayAppService()}
}

// Index 分页查询支付应用，支持应用名 / appKey 关键字与开启状态筛选。
// 入参：keyword、status、pageNo、pageSize（查询串）；返回应用列表与总数。
func (c *PayAppController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.PayAppListQuery{Keyword: ctx.Request().Query("keyword")}
	// status 的合法值包含 0（开启），因此不能按「取不到就跳过」判断，必须看参数是否出现
	if raw := ctx.Request().Query("status"); raw != "" {
		status := int8(ctx.Request().QueryInt("status", 0))
		q.Status = &status
	}

	list, total, err := c.service.Page(ctx.Context(), q, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Options 返回开启状态的支付应用，供渠道配置与代下单页面的应用下拉使用。
// 无入参；返回 id / name 列表。
func (c *PayAppController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show 查询单个支付应用详情，密钥字段只返回脱敏值。
// 入参：路径 id；返回应用详情，不存在返回业务错误码。
func (c *PayAppController) Show(ctx http.Context) http.Response {
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

// Store 新增支付应用，appKey 与 appSecret 由系统生成，明文密钥只在本次响应里返回一次。
// 入参：应用字段；返回应用详情（含一次性 appSecret）。
func (c *PayAppController) Store(ctx http.Context) http.Response {
	var req request.PayAppSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	row, err := c.service.Create(ctx.Context(), toPayAppSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}

// Update 编辑支付应用，appKey 不可修改，appSecret 留空保持原值。
// 入参：路径 id + 应用字段；返回 null。
func (c *PayAppController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PayAppSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toPayAppSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除支付应用，应用下仍有渠道配置时拒绝。
// 入参：路径 id；返回 null。
func (c *PayAppController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toPayAppSaveInput 把请求体转换为服务层入参，避免 Controller 直接拼装模型。
// 入参：已通过校验的请求体；返回服务层可直接使用的应用字段。
func toPayAppSaveInput(req request.PayAppSaveRequest) service.PayAppSaveInput {
	return service.PayAppSaveInput{
		Name:              req.Name,
		Status:            enums.PayEnableStatusEnum(req.Status),
		Remark:            req.Remark,
		OrderNotifyURL:    req.OrderNotifyURL,
		RefundNotifyURL:   req.RefundNotifyURL,
		TransferNotifyURL: req.TransferNotifyURL,
		AppSecret:         req.AppSecret,
	}
}
