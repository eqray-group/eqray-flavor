package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/infra/repository"
)

// ApiLogController 提供 API 访问 / 错误日志的查询与处理接口。
type ApiLogController struct {
	repo *repository.ApiLogRepository
}

func NewApiLogController() *ApiLogController {
	return &ApiLogController{repo: repository.NewApiLogRepository()}
}

// Access 分页查询 API 访问日志，支持按请求地址与响应码筛选。
// 入参：requestUrl、resultCode、pageNo、pageSize（查询串）；返回日志列表与总数。
func (c *ApiLogController) Access(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	filter := repository.ApiLogFilter{
		RequestURL: ctx.Request().Query("requestUrl"),
	}
	if code := ctx.Request().QueryInt("resultCode", -1); code >= 0 {
		filter.ResultCode = &code
	}

	list, total, err := c.repo.PaginateAccess(ctx.Context(), filter, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Error 分页查询 API 错误日志，支持按请求地址与关键字（错误信息）筛选。
// 入参：requestUrl、keyword、pageNo、pageSize（查询串）；返回日志列表与总数。
func (c *ApiLogController) Error(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.repo.PaginateError(ctx.Context(), repository.ApiLogFilter{
		RequestURL: ctx.Request().Query("requestUrl"),
		Keyword:    ctx.Request().Query("keyword"),
	}, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Process 把错误日志标记为已处理，用于区分待排查与已解决。
// 入参：路径 id；返回 null。
func (c *ApiLogController) Process(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.repo.MarkErrorProcessed(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
