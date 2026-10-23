// Package pay 聚合支付模块的路由，结构对齐 yudao pay 模块。
// 管理端接口在 admin 包，渠道异步回调在 callback 包；C 端（会员支付）接口随订单模块接单场景补齐。
package pay

import (
	"eqray-flavor/app/modules/pay/admin"
	"eqray-flavor/app/modules/pay/callback"
	"eqray-flavor/app/modules/pay/client"
)

// Register 注册支付模块的全部路由，由 routes.Modules() 统一调用。
// 入参：无；返回：无。
func Register() {
	admin.Register()
	callback.Register()
	client.Register()
}
