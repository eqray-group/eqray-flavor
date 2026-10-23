package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/pay/repository"
	"eqray-flavor/app/modules/pay/service"
)

// PayNotifyTaskController 提供业务回调通知任务与通知日志的查询接口。
type PayNotifyTaskController struct {
	service *service.PayNotifyTaskService
}

// NewPayNotifyTaskController 构造通知任务控制器。
func NewPayNotifyTaskController() *PayNotifyTaskController {
	return &PayNotifyTaskController{service: service.NewPayNotifyTaskService()}
}

// Index 分页查询通知任务，支持应用、类型、状态与商户订单号关键字筛选。
// 入参：appId、type、status、keyword、pageNo、pageSize（查询串）；返回任务列表与总数。
func (c *PayNotifyTaskController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	appID, err := queryID(ctx, "appId")
	if err != nil {
		return result.Fail(ctx, err)
	}

	filter := repository.PayNotifyTaskFilter{
		AppID:   appID,
		Keyword: ctx.Request().Query("keyword"),
	}
	// 类型与状态都从 1 起，仍按参数是否出现判断，避免缺省值被当成筛选条件
	if raw := ctx.Request().Query("type"); raw != "" {
		notifyType := int8(ctx.Request().QueryInt("type", 0))
		filter.Type = &notifyType
	}
	if raw := ctx.Request().Query("status"); raw != "" {
		status := int8(ctx.Request().QueryInt("status", 0))
		filter.Status = &status
	}

	list, total, err := c.service.Page(ctx.Context(), filter, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Logs 分页查询某条通知任务的通知日志，用于排查业务方是否收到、响应了什么。
// 入参：路径 id + pageNo、pageSize；返回日志列表与总数。
func (c *PayNotifyTaskController) Logs(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}
	if page.PageSize > payNotifyLogMaxPageSize {
		page.PageSize = payNotifyLogMaxPageSize
	}

	list, total, err := c.service.Logs(ctx.Context(), id, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// payNotifyLogMaxPageSize 单次最多返回的通知日志条数，避免一次拉太多响应报文
const payNotifyLogMaxPageSize = 100
