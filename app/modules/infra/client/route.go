// Package client 注册 infra 模块的 C 端接口。
// 这里只放会员本人会直接调用的通用能力（如文件上传），不含任何管理配置类接口。
package client

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/infra/client/controller"
)

// Prefix 是 C 端接口前缀。
const Prefix = "/api/v1/client/infra"

// Register 注册 C 端接口，由 infra.Register() 统一调用。
func Register() {
	fileController := controller.NewFileController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		// 上传需要会员登录态；不挂操作日志，C 端行为不进后台操作审计
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})
		memberAuth.Post("/files", fileController.Upload)
	})
}
