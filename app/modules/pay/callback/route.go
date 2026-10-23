// Package callback 注册支付模块的渠道异步回调路由。
// 回调由渠道服务器直接访问，不挂令牌与权限中间件，单独成包避免与后台接口混在一起。
package callback

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/pay/callback/controller"
)

// Prefix 是支付回调的接口前缀，与 admin.Prefix 一致但不共享中间件。
const Prefix = "/api/v1/pay"

// Register 注册支付回调接口，由 pay.Register() 统一调用。
// 入参：无；返回：无。
func Register() {
	callbackController := controller.NewPayCallbackController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		// 路径里的渠道主键决定用哪套渠道配置验签与解密
		r.Post("/callbacks/order/{channelId}", callbackController.Order)
		// 退款回调单独一个地址，避免和支付回调互相覆盖状态
		r.Post("/callbacks/refund/{channelId}", callbackController.Refund)
	})
}
