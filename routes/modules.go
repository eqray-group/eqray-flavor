package routes

import (
	"eqray-flavor/app/modules/dish"
	"eqray-flavor/app/modules/infra"
	"eqray-flavor/app/modules/member"
	"eqray-flavor/app/modules/pay"
	"eqray-flavor/app/modules/store"
	"eqray-flavor/app/modules/system"
)

// Modules 聚合各业务模块的路由，在 bootstrap/app.go 的 WithRouting 中调用。
// 顺序按依赖方向排列：门店是其他业务模块的前置条件，菜品依赖门店做可售配置，
// 支付依赖会员（钱包）与门店（门店维度对账），因此排在最后。
func Modules() {
	system.Register()
	infra.Register()
	member.Register()
	store.Register()
	dish.Register()
	pay.Register()
}
