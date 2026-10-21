// Package infra 聚合基础设施模块的路由。
// 管理端与 C 端分别在自己的目录里注册：admin/route.go、client/route.go。
package infra

import (
	"eqray-flavor/app/modules/infra/admin"
	"eqray-flavor/app/modules/infra/client"
)

// Register 注册 infra 模块的全部路由，由 routes.Modules() 统一调用。
func Register() {
	admin.Register()
	client.Register()
}
