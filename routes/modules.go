package routes

import (
	"eqray-flavor/app/modules/infra"
	"eqray-flavor/app/modules/member"
	"eqray-flavor/app/modules/system"
)

// Modules 聚合各业务模块的路由，在 bootstrap/app.go 的 WithRouting 中调用。
func Modules() {
	system.Register()
	infra.Register()
	member.Register()
}
