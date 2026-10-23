// Package client 注册支付模块的 C 端接口（小程序 / H5 / App）。
// C 端只处理会员本人的钱包数据，只做登录态校验，不进入后台 RBAC 权限体系。
package client

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/pay/client/controller"
)

// Prefix 是 C 端支付接口前缀。
const Prefix = "/api/v1/client/pay"

// Register 注册 C 端接口，由 pay.Register() 统一调用。
// 入参：无；返回：无。
func Register() {
	walletController := controller.NewPayWalletController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		// 会员身份统一取令牌里的 MemberID，请求体不接受会员 ID
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})

		memberAuth.Get("/wallet", walletController.Wallet)
		memberAuth.Get("/wallet/transactions", walletController.Transactions)
		memberAuth.Get("/wallet/packages", walletController.Packages)
		memberAuth.Post("/wallet/recharges", walletController.CreateRecharge)
		memberAuth.Post("/wallet/pay", walletController.PayByBalance)
	})
}
