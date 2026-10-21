// Package admin 注册 infra 模块的后台管理接口。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/infra/admin/controller"
)

// Prefix 是后台管理端接口前缀。
const Prefix = "/api/v1/infra"

// Register 注册后台管理接口，由 infra.Register() 统一调用（docs 4）。
func Register() {
	fileController := controller.NewFileController()
	fileConfigController := controller.NewFileConfigController()
	configController := controller.NewConfigController()
	apiLogController := controller.NewApiLogController()
	smsChannelController := controller.NewInfraSmsChannelController()
	smsTemplateController := controller.NewInfraSmsTemplateController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 权限中间件已下沉到 base，infra 按自身权限点校验，不再依赖 system
		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// operate 给写操作声明留痕语义，中间件注册在 Permission 之内（docs 9.5）
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// 上传只要求登录态：头像等自助场景不应被具体权限点挡住，但上传动作需要留痕
		auth.Middleware(operate("文件管理", "上传文件")).Post("/files", fileController.Upload)

		// 文件列表与删除；C 端上传见 client/route.go
		perm("infra:file:list").Get("/files", fileController.Index)
		perm("infra:file:delete").Middleware(operate("文件管理", "删除文件")).
			Delete("/files/{id}", fileController.Destroy)

		perm("infra:file-config:list").Get("/file-configs", fileConfigController.Index)
		perm("infra:file-config:create").Middleware(operate("文件配置", "新增配置")).
			Post("/file-configs", fileConfigController.Store)
		perm("infra:file-config:update").Middleware(operate("文件配置", "编辑配置")).
			Put("/file-configs/{id}", fileConfigController.Update)
		perm("infra:file-config:update").Middleware(operate("文件配置", "设为主配置")).
			Put("/file-configs/{id}/master", fileConfigController.SetMaster)
		perm("infra:file-config:delete").Middleware(operate("文件配置", "删除配置")).
			Delete("/file-configs/{id}", fileConfigController.Destroy)

		perm("infra:config:list").Get("/configs", configController.Index)
		perm("infra:config:create").Middleware(operate("参数配置", "新增参数")).
			Post("/configs", configController.Store)
		perm("infra:config:update").Middleware(operate("参数配置", "编辑参数")).
			Put("/configs/{id}", configController.Update)
		perm("infra:config:delete").Middleware(operate("参数配置", "删除参数")).
			Delete("/configs/{id}", configController.Destroy)

		perm("infra:api-log:list").Get("/api-logs/access", apiLogController.Access)
		perm("infra:api-log:list").Get("/api-logs/error", apiLogController.Error)
		perm("infra:api-log:update").Middleware(operate("API 日志", "标记处理")).
			Put("/api-logs/error/{id}/process", apiLogController.Process)

		// 短信渠道与模板：配置在 infra，member 等模块通过 InfraSmsService 复用发送能力
		perm("infra:sms-channel:list").Get("/sms-channels", smsChannelController.Index)
		perm("infra:sms-template:list").Get("/sms-channels/options", smsChannelController.Options)
		perm("infra:sms-channel:list").Get("/sms-channels/{id}", smsChannelController.Show)
		perm("infra:sms-channel:create").Middleware(operate("短信管理", "新增渠道")).
			Post("/sms-channels", smsChannelController.Store)
		perm("infra:sms-channel:update").Middleware(operate("短信管理", "编辑渠道")).
			Put("/sms-channels/{id}", smsChannelController.Update)
		perm("infra:sms-channel:delete").Middleware(operate("短信管理", "删除渠道")).
			Delete("/sms-channels/{id}", smsChannelController.Destroy)

		perm("infra:sms-template:list").Get("/sms-templates", smsTemplateController.Index)
		perm("infra:sms-template:list").Get("/sms-templates/{id}", smsTemplateController.Show)
		perm("infra:sms-template:create").Middleware(operate("短信管理", "新增模板")).
			Post("/sms-templates", smsTemplateController.Store)
		perm("infra:sms-template:update").Middleware(operate("短信管理", "编辑模板")).
			Put("/sms-templates/{id}", smsTemplateController.Update)
		perm("infra:sms-template:delete").Middleware(operate("短信管理", "删除模板")).
			Delete("/sms-templates/{id}", smsTemplateController.Destroy)
	})
}
