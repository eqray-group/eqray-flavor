package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/service"
)

// PrintTaskController 提供打印任务的查询与重打接口。
type PrintTaskController struct {
	service *service.StorePrintTaskService
}

// NewPrintTaskController 构造打印任务控制器。
func NewPrintTaskController() *PrintTaskController {
	return &PrintTaskController{service: service.NewStorePrintTaskService()}
}

// Index 分页查询打印任务，支持打印类型、状态、设备与订单号筛选。
// 门店范围由请求头门店上下文或当前用户可见门店集合决定。
func (c *PrintTaskController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.PrintTaskListQuery{
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
		OrderNo:  ctx.Request().Query("orderNo"),
	}
	if value := ctx.Request().QueryInt("printType", 0); value > 0 {
		item := enums.PrintTypeEnum(value)
		if !item.IsValid() {
			return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "打印类型取值不合法"))
		}
		q.PrintType = &item
	}
	// 状态 0（待打印）是合法筛选值，只有缺省才表示不限
	if ctx.Request().Query("status") != "" {
		item := enums.PrintTaskStatusEnum(int8(ctx.Request().QueryInt("status", 0)))
		if !item.IsValid() {
			return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "任务状态取值不合法"))
		}
		q.Status = &item
	}

	q.PrinterID = param.QueryID(ctx, "printerId")

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Reprint 重打一张单据：任务重置为待打印，正文沿用生成时的快照。
// 入参：路径 id；返回空数据表示重打已入队。
func (c *PrintTaskController) Reprint(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Reprint(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Detail 查询打印任务详情：返回渲染后的正文，便于排查打印内容是否正确。
// 入参：路径 id；返回任务详情。
func (c *PrintTaskController) Detail(ctx http.Context) http.Response {
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
