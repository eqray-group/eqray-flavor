// Package dish 聚合菜品模块的路由。
// 管理端维护菜品与门店配置，C 端提供顾客点餐使用的门店菜单。
package dish

import (
	"eqray-flavor/app/modules/dish/admin"
	"eqray-flavor/app/modules/dish/client"
)

// Register 注册菜品模块的全部路由，由 routes.Modules() 统一调用。
func Register() {
	admin.Register()
	client.Register()
}
