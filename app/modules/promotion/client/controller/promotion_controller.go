// Package controller 提供 promotion 模块的 C 端接口（会员侧查看与领取优惠券）。
// 身份只从会员令牌取，请求体不收会员 ID。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/promotion/enums"
	"eqray-flavor/app/modules/promotion/service"
)

// PromotionController 提供我的券、可领券与领券接口。
type PromotionController struct {
	service *service.PromotionService
}

// NewPromotionController 构造营销 C 端控制器。
func NewPromotionController() *PromotionController {
	return &PromotionController{service: service.NewPromotionService()}
}

// MyCoupons 查询我持有的优惠券。
// 入参：status（可选，1 未使用 / 2 锁定 / 3 已使用 / 4 已过期）；返回券列表。
func (c *PromotionController) MyCoupons(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())

	status := enums.CouponStatus(ctx.Request().QueryInt("status", 0))

	rows, err := c.service.MyCoupons(ctx.Context(), memberID, status)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, rows)
}

// AvailableTemplates 查询当前可主动领取的优惠券模板。
// 入参：无；返回模板列表，含剩余可领张数与本人已领张数。
func (c *PromotionController) AvailableTemplates(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())

	rows, err := c.service.AvailableTemplates(ctx.Context(), memberID)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, rows)
}

// Take 领取一张优惠券。
// 入参：路径 id（券模板 ID）；返回领取张数。
func (c *PromotionController) Take(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	memberID := support.MemberID(ctx.Context())

	count, err := c.service.Take(ctx.Context(), id, memberID)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, count)
}
