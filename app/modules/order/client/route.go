// Package client 注册订单模块的 C 端接口（顾客点单）。
// C 端只校验会员令牌，不进入后台 RBAC 权限体系；会员身份只从令牌取，请求体不接受会员 ID。
package client

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/order/client/controller"
)

// Prefix 是订单 C 端接口前缀。
const Prefix = "/api/v1/client/order"

// Register 注册订单 C 端接口，由 order.Register() 统一调用。
func Register() {
	orderController := controller.NewOrderController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})

		// 试算不落库，登录即可用（登录后才能按会员价试算）
		memberAuth.Post("/estimates", orderController.Estimate)
		memberAuth.Post("/orders", orderController.Create)
		memberAuth.Get("/orders", orderController.Index)
		memberAuth.Get("/orders/{id}", orderController.Show)
		memberAuth.Post("/orders/{id}/cancel", orderController.Cancel)
		memberAuth.Post("/orders/{id}/pay", orderController.Pay)
	})
}
