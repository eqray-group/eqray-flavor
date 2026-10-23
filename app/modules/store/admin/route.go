// Package admin 注册门店模块的后台管理接口。
// 门店数据是后续所有门店级业务的基础，接口一律挂后台令牌与权限点。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/store/admin/controller"
	storeMiddleware "eqray-flavor/app/modules/store/middleware"
)

// Prefix 是门店后台接口前缀。
const Prefix = "/api/v1/store"

// Register 注册门店后台接口，由 store.Register() 统一调用。
func Register() {
	storeController := controller.NewStoreController()
	employeeController := controller.NewEmployeeController()
	tableAreaController := controller.NewTableAreaController()
	tableController := controller.NewTableController()
	riderController := controller.NewRiderController()
	printerController := controller.NewPrinterController()
	printTemplateController := controller.NewPrintTemplateController()
	printTaskController := controller.NewPrintTaskController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 权限中间件在 base，store 只声明自己的权限点
		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// 写操作留痕，中间件注册在 Permission 之内
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// ---------- 门店 ----------
		// 静态路径必须排在动态路径之前，否则 /stores/tree、/stores/visible 会被 /stores/{id} 吃掉
		perm("store:store:list").Get("/stores/tree", storeController.Tree)
		// 门店切换器是登录后的基础能力，不要求门店列表权限
		auth.Get("/stores/visible", storeController.Visible)
		perm("store:store:list").Get("/stores", storeController.Index)
		perm("store:store:list").Get("/stores/{id}", storeController.Show)
		perm("store:store:create").Middleware(operate("门店管理", "新增门店")).
			Post("/stores", storeController.Store)
		perm("store:store:update").Middleware(operate("门店管理", "编辑门店")).
			Put("/stores/{id}", storeController.Update)
		perm("store:store:delete").Middleware(operate("门店管理", "删除门店")).
			Delete("/stores/{id}", storeController.Destroy)

		// ---------- 门店员工 ----------
		// 员工接口统一挂门店上下文：带 X-Store-Id 时只操作该门店，不带时按可见集合过滤。
		employeeList := perm("store:employee:list").Middleware(&storeMiddleware.StoreContext{})
		employeeList.Get("/employees", employeeController.Index)
		// 账号下拉是静态路径，必须放在任何 /employees/{id} 之前
		perm("store:employee:list").Get("/employees/user-options", employeeController.UserOptions)
		perm("store:employee:create").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "新增员工")).Post("/employees", employeeController.Store)
		perm("store:employee:update").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "编辑员工")).Put("/employees/{id}", employeeController.Update)
		perm("store:employee:delete").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "删除员工")).Delete("/employees/{id}", employeeController.Destroy)

		// ---------- 桌台区域 ----------
		// 区域与桌台都是门店级数据：带 X-Store-Id 时只操作该门店，不带时按可见集合过滤。
		perm("store:table-area:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/table-areas", tableAreaController.Index)
		// 下拉选项是静态路径，必须放在任何 /table-areas/{id} 之前
		perm("store:table-area:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/table-areas/options", tableAreaController.Options)
		perm("store:table-area:create").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "新增桌台区域")).Post("/table-areas", tableAreaController.Store)
		perm("store:table-area:update").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "编辑桌台区域")).Put("/table-areas/{id}", tableAreaController.Update)
		perm("store:table-area:delete").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "删除桌台区域")).Delete("/table-areas/{id}", tableAreaController.Destroy)

		// ---------- 桌台 ----------
		perm("store:table:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/tables", tableController.Index)
		perm("store:table:create").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "新增桌台")).Post("/tables", tableController.Store)
		// 批量生成是静态路径，必须排在 /tables/{id}/... 之前
		perm("store:table:create").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "批量生成桌台")).Post("/tables/batch", tableController.Batch)
		perm("store:table:update").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "编辑桌台")).Put("/tables/{id}", tableController.Update)
		perm("store:table:delete").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "删除桌台")).Delete("/tables/{id}", tableController.Destroy)
		perm("store:table:operate").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "桌台开台")).Post("/tables/{id}/open", tableController.Open)
		perm("store:table:operate").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "桌台清台")).Post("/tables/{id}/clear", tableController.Clear)

		// ---------- 骑手档案 ----------
		// 骑手归属门店员工，新增时未指定账号会自动开通后台账号并返回一次性初始密码。
		perm("store:rider:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/riders", riderController.Index)
		// 账号下拉是静态路径，必须放在任何 /riders/{id} 之前
		perm("store:rider:list").Get("/riders/user-options", riderController.UserOptions)
		perm("store:rider:create").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "新增骑手")).Post("/riders", riderController.Store)
		perm("store:rider:update").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "编辑骑手")).Put("/riders/{id}", riderController.Update)
		perm("store:rider:delete").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "删除骑手")).Delete("/riders/{id}", riderController.Destroy)

		// ---------- 打印设备 ----------
		// 设备按门店隔离；密钥只在写入时接收明文，列表与编辑均只返回脱敏值。
		perm("store:printer:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/printers", printerController.Index)
		perm("store:printer:create").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "新增打印设备")).Post("/printers", printerController.Store)
		perm("store:printer:update").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "编辑打印设备")).Put("/printers/{id}", printerController.Update)
		perm("store:printer:delete").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "删除打印设备")).Delete("/printers/{id}", printerController.Destroy)

		// ---------- 打印模板 ----------
		// 同一门店下设备类型与打印类型唯一；正文变化时后端自动递增版本号。
		perm("store:print-template:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/print-templates", printTemplateController.Index)
		perm("store:print-template:create").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "新增打印模板")).Post("/print-templates", printTemplateController.Store)
		perm("store:print-template:update").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "编辑打印模板")).Put("/print-templates/{id}", printTemplateController.Update)
		perm("store:print-template:delete").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "删除打印模板")).Delete("/print-templates/{id}", printTemplateController.Destroy)

		// ---------- 打印任务 ----------
		// 任务由支付成功时自动生成，后台只做查询与重打；重打会重置为待打印，由设备取走。
		perm("store:print-task:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/print-tasks", printTaskController.Index)
		perm("store:print-task:list").Middleware(&storeMiddleware.StoreContext{}).
			Get("/print-tasks/{id}", printTaskController.Detail)
		perm("store:print-task:reprint").Middleware(&storeMiddleware.StoreContext{}).
			Middleware(operate("门店管理", "重打小票")).Post("/print-tasks/{id}/reprint", printTaskController.Reprint)
	})
}
