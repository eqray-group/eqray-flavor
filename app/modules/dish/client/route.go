// Package client 注册菜品模块的 C 端接口（顾客点餐）。
// C 端只校验会员令牌，不进入后台 RBAC 权限体系，门店由查询串显式指定。
package client

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/dish/client/controller"
)

// Prefix 是菜品 C 端接口前缀。
const Prefix = "/api/v1/client/dish"

// Register 注册菜品 C 端接口，由 dish.Register() 统一调用。
func Register() {
	menuController := controller.NewMenuController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})
		// 门店菜单带会员价，必须登录后按会员身份取价
		memberAuth.Get("/menus", menuController.Index)
	})
}
