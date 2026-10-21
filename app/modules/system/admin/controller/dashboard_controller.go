package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/system/service"
)

// DashboardController 提供控制台首页的聚合统计。
type DashboardController struct {
	service *service.DashboardService
}

func NewDashboardController() *DashboardController {
	return &DashboardController{service: service.NewDashboardService()}
}

// Overview 查询控制台概览数据（用户数、登录趋势等）。
// 无入参，仅要求登录态；返回概览统计。
func (c *DashboardController) Overview(ctx http.Context) http.Response {
	data, err := c.service.Overview(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, data)
}
