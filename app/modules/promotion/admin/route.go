// Package admin 注册 promotion 模块的后台管理接口。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/promotion/admin/controller"
)

// Prefix 是营销后台接口前缀。
const Prefix = "/api/v1/promotion"

// Register 注册营销后台接口，由 promotion.Register() 统一调用。
func Register() {
	promotionController := controller.NewPromotionController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// ---------- 优惠券模板 ----------
		perm("promotion:coupon-template:list").Get("/coupon-templates", promotionController.Index)
		perm("promotion:coupon-template:list").Get("/coupon-templates/{id}", promotionController.Show)
		perm("promotion:coupon-template:create").Post("/coupon-templates", promotionController.Store)
		perm("promotion:coupon-template:update").Put("/coupon-templates/{id}", promotionController.Store)
		perm("promotion:coupon-template:delete").Delete("/coupon-templates/{id}", promotionController.Destroy)
		perm("promotion:coupon-template:issue").Post("/coupon-templates/{id}/issue", promotionController.Issue)

		// ---------- 会员券 ----------
		perm("promotion:coupon:list").Get("/coupons", promotionController.CouponIndex)
		perm("promotion:coupon:revoke").Post("/coupons/{id}/revoke", promotionController.Revoke)

		// ---------- 营销活动 ----------
		perm("promotion:activity:list").Get("/activities", promotionController.ActivityIndex)
		perm("promotion:activity:list").Get("/activities/{id}", promotionController.ActivityShow)
		perm("promotion:activity:create").Post("/activities", promotionController.ActivityStore)
		perm("promotion:activity:update").Put("/activities/{id}", promotionController.ActivityStore)
		perm("promotion:activity:delete").Delete("/activities/{id}", promotionController.ActivityDestroy)

		// ---------- 积分规则 ----------
		perm("promotion:activity:list").Get("/point-rule", promotionController.PointRule)
		perm("promotion:activity:update").Put("/point-rule", promotionController.PointRuleStore)
	})
}
