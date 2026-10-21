// Package system 聚合系统管理模块的路由。
// 目前只有后台管理接口，C 端无需 system 能力；将来若出现 C 端接口，
// 按 infra / member 的方式新增 client/route.go 并在这里调用 client.Register()。
package system

import (
	"eqray-flavor/app/modules/system/admin"
)

// Register 注册 system 模块的全部路由，由 routes.Modules() 统一调用。
func Register() {
	admin.Register()
}
