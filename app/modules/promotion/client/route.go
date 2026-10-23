// Package client 注册 promotion 模块的 C 端接口（会员查看与领取优惠券）。
// C 端只校验会员令牌，不进入后台 RBAC 权限体系；会员身份只从令牌取，请求体不接受会员 ID。
package client

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/promotion/client/controller"
)

// Prefix 是营销 C 端接口前缀。
const Prefix = "/api/v1/client/promotion"

// Register 注册营销 C 端接口，由 promotion.Register() 统一调用。
func Register() {
	promotionController := controller.NewPromotionController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})

		memberAuth.Get("/coupons", promotionController.MyCoupons)
		memberAuth.Get("/coupon-templates", promotionController.AvailableTemplates)
		memberAuth.Post("/coupon-templates/{id}/take", promotionController.Take)
	})
}
