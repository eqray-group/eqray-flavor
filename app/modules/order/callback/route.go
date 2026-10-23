// Package callback 注册订单模块接收支付模块通知的入口。
// 通知由 pay 模块按应用配置的 order_notify_url 推送，带签名但不带令牌，
// 安全由 pay 侧的签名校验 + 本侧的金额核对保证，因此单独成包避免与后台接口混在一起。
package callback

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/order/callback/controller"
)

// Prefix 是订单回调的接口前缀。
const Prefix = "/api/v1/order"

// Register 注册订单回调接口，由 order.Register() 统一调用。
func Register() {
	orderCallbackController := controller.NewOrderCallbackController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		// 支付结果通知：pay 模块按 merchant_order_id 推送，此处验签后推进订单状态
		r.Post("/callbacks/pay", orderCallbackController.Pay)
	})
}
