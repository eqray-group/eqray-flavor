// Package store 聚合门店模块的路由。
// 当前只有管理端；C 端的扫码解析后续在 client/ 下注册。
package store

import (
	"eqray-flavor/app/modules/store/admin"
)

// Register 注册门店模块的全部路由，由 routes.Modules() 统一调用。
func Register() {
	admin.Register()
}
