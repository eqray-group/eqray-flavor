// Package admin 注册订单模块的后台管理接口。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/order/admin/controller"
)

// Prefix 是订单后台接口前缀。
const Prefix = "/api/v1/order"

// Register 注册订单后台接口，由 order.Register() 统一调用。
func Register() {
	orderController := controller.NewOrderController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// ---------- 订单管理 ----------
		perm("order:order:list").Get("/orders", orderController.Index)
		perm("order:order:list").Get("/orders/{id}", orderController.Show)
		perm("order:order:status").Post("/orders/{id}/transit", orderController.Transit)
		perm("order:order:cancel").Post("/orders/{id}/cancel", orderController.Cancel)
	})
}
