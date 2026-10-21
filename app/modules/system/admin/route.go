// Package admin 注册 system 模块的后台管理接口。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/system/admin/controller"
)

// Prefix 是后台管理端接口前缀。
const Prefix = "/api/v1/system"

// Register 注册后台管理接口，由 system.Register() 统一调用（docs 4）。
func Register() {
	authController := controller.NewAuthController()
	userController := controller.NewUserController()
	menuController := controller.NewMenuController()
	roleController := controller.NewRoleController()
	dictController := controller.NewDictController()
	logController := controller.NewLogController()
	dashboardController := controller.NewDashboardController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		r.Get("/auth/captcha", authController.Captcha)
		r.Post("/auth/login", authController.Login)
		// 刷新与登出不挂鉴权：访问令牌过期后仍要能续期，登出也允许用已失效的令牌调用
		r.Post("/auth/refresh", authController.Refresh)
		r.Post("/auth/logout", authController.Logout)

		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 按权限标识分组注册，权限标识与 system_menu.permission 一一对应（docs 7.2、9.3）
		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// operate 给写操作声明留痕语义，中间件注册在 Permission 之内（docs 9.5）
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		auth.Get("/profile", authController.Profile)
		// 个人中心：任何登录者都能维护自己的资料与密码，不挂具体权限点
		auth.Middleware(operate("个人中心", "修改资料")).Put("/profile", authController.UpdateProfile)
		auth.Middleware(operate("个人中心", "修改密码")).Put("/profile/password", authController.UpdatePassword)
		// 动态路由按当前登录者计算，不挂具体权限点，避免没有菜单权限时连导航都拿不到
		auth.Get("/menus/routes", menuController.Routes)

		// 控制台：登录后默认首页，所有登录用户都可查看
		auth.Get("/dashboard/overview", dashboardController.Overview)

		// 用户
		userList := perm("system:user:list")
		userList.Get("/users", userController.Index)
		userList.Get("/users/{id}", userController.Show)

		perm("system:user:create").Middleware(operate("用户管理", "新增用户")).Post("/users", userController.Store)

		userUpdate := perm("system:user:update")
		userUpdate.Middleware(operate("用户管理", "编辑用户")).Put("/users/{id}", userController.Update)
		userUpdate.Middleware(operate("用户管理", "重置密码")).Put("/users/{id}/password", userController.ResetPassword)

		perm("system:user:delete").Middleware(operate("用户管理", "删除用户")).Delete("/users/{id}", userController.Destroy)

		userRole := perm("system:user:assign-role")
		userRole.Get("/users/{id}/roles", userController.Roles)
		userRole.Middleware(operate("用户管理", "分配角色")).Put("/users/{id}/roles", userController.AssignRoles)

		// 菜单：/menus/simple 注册在 /menus/{id} 之前
		menuRead := perm("system:menu:list")
		menuRead.Get("/menus", menuController.Index)
		menuRead.Get("/menus/simple", menuController.Simple)
		menuRead.Get("/menus/{id}", menuController.Show)

		perm("system:menu:create").Middleware(operate("菜单管理", "新增菜单")).Post("/menus", menuController.Store)
		perm("system:menu:update").Middleware(operate("菜单管理", "编辑菜单")).Put("/menus/{id}", menuController.Update)
		perm("system:menu:delete").Middleware(operate("菜单管理", "删除菜单")).Delete("/menus/{id}", menuController.Destroy)

		// 角色：/roles/options 注册在 /roles/{id} 之前
		roleRead := perm("system:role:list")
		roleRead.Get("/roles", roleController.Index)
		roleRead.Get("/roles/options", roleController.Options)
		roleRead.Get("/roles/{id}", roleController.Show)

		perm("system:role:create").Middleware(operate("角色管理", "新增角色")).Post("/roles", roleController.Store)
		perm("system:role:update").Middleware(operate("角色管理", "编辑角色")).Put("/roles/{id}", roleController.Update)
		perm("system:role:delete").Middleware(operate("角色管理", "删除角色")).Delete("/roles/{id}", roleController.Destroy)

		roleMenu := perm("system:role:assign-menu")
		roleMenu.Get("/roles/{id}/menus", roleController.Menus)
		roleMenu.Middleware(operate("角色管理", "分配菜单")).Put("/roles/{id}/menus", roleController.AssignMenus)

		// 字典：/dict-types/options 注册在 /dict-types/{id} 之前
		dictRead := perm("system:dict:list")
		dictRead.Get("/dict-types", dictController.Types)
		dictRead.Get("/dict-types/options", dictController.TypeOptions)
		dictRead.Get("/dict-types/{id}", dictController.TypeShow)

		perm("system:dict:create").Middleware(operate("字典管理", "新增字典类型")).Post("/dict-types", dictController.TypeStore)
		perm("system:dict:update").Middleware(operate("字典管理", "编辑字典类型")).Put("/dict-types/{id}", dictController.TypeUpdate)
		perm("system:dict:delete").Middleware(operate("字典管理", "删除字典类型")).Delete("/dict-types/{id}", dictController.TypeDestroy)

		dictRead.Get("/dict-data", dictController.Data)
		dictRead.Get("/dict-data/{id}", dictController.DataShow)

		perm("system:dict:create").Middleware(operate("字典管理", "新增字典数据")).Post("/dict-data", dictController.DataStore)
		perm("system:dict:update").Middleware(operate("字典管理", "编辑字典数据")).Put("/dict-data/{id}", dictController.DataUpdate)
		perm("system:dict:delete").Middleware(operate("字典管理", "删除字典数据")).Delete("/dict-data/{id}", dictController.DataDestroy)

		// 下拉 / 标签渲染只要求登录态，任何后台页面都可能用到
		auth.Get("/dicts/{type}", dictController.Items)

		// 登录日志：数据来自 base 认证组件，system 只读
		perm("system:login-log:list").Get("/logs/login", logController.Login)

		// 操作日志：由 OperateLog 中间件写入 system_operate_log
		perm("system:operate-log:list").Get("/logs/operate", logController.Operate)
	})
}
