// Package member 聚合会员模块的路由。
// 管理端与 C 端分别在自己的目录里注册：admin/route.go、client/route.go。
package member

import (
	"eqray-flavor/app/modules/member/admin"
	"eqray-flavor/app/modules/member/client"
)

// Register 注册会员模块的全部路由，由 routes.Modules() 统一调用。
// 两端前缀不同、中间件不同，互不影响。
func Register() {
	admin.Register()
	client.Register()
}
