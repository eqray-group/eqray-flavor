// Package order 聚合订单模块的路由。
// 管理端用于接单与履约，C 端用于顾客下单与查看订单，回调入口接收支付模块的通知。
package order

import (
	"eqray-flavor/app/modules/order/admin"
	"eqray-flavor/app/modules/order/callback"
	"eqray-flavor/app/modules/order/client"
)

// Register 注册订单模块的全部路由，由 routes.Modules() 统一调用。
func Register() {
	admin.Register()
	callback.Register()
	client.Register()
}
