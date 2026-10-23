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

		// 写操作留痕，中间件注册在 Permission 之内
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// ---------- 订单管理 ----------
		// 静态路径必须排在 /orders/{id} 之前，否则 /orders/export 会被当成详情
		perm("order:order:list").Get("/orders/export", orderController.Export)
		perm("order:order:list").Get("/orders", orderController.Index)
		perm("order:order:list").Get("/orders/{id}", orderController.Show)
		// 小票重打是只读动作，跟查询同权限
		perm("order:order:list").Get("/orders/{id}/ticket", orderController.Ticket)
		perm("order:order:status").Middleware(operate("订单管理", "订单状态流转")).
			Post("/orders/{id}/transit", orderController.Transit)
		perm("order:order:cancel").Middleware(operate("订单管理", "取消订单")).
			Post("/orders/{id}/cancel", orderController.Cancel)

		// ---------- 订单经营操作 ----------
		// 备注可随时改；改价与派单各自独立授权，避免只读角色误操作金额与配送。
		perm("order:order:remark").Middleware(operate("订单管理", "修改订单备注")).
			Put("/orders/{id}/remark", orderController.UpdateRemark)
		perm("order:order:adjust").Middleware(operate("订单管理", "订单改价")).
			Put("/orders/{id}/price", orderController.AdjustPrice)
		perm("order:order:dispatch").Get("/orders/{id}/riders", orderController.Riders)
		perm("order:order:dispatch").Middleware(operate("订单管理", "订单派单")).
			Post("/orders/{id}/rider", orderController.AssignRider)
		perm("order:order:dispatch").Middleware(operate("订单管理", "取消派单")).
			Delete("/orders/{id}/rider", orderController.CancelDispatch)
	})
}
