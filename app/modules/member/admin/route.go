// Package admin 注册会员模块的后台管理接口。
// 管理端访问的是「任意会员」的数据，必须挂后台令牌与权限点。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-flavor/app/facades"
	baseMiddleware "eqray-flavor/app/modules/base/middleware"
	"eqray-flavor/app/modules/member/admin/controller"
)

// Prefix 是后台管理端接口前缀。
const Prefix = "/api/v1/member"

// Register 注册后台管理接口，由 member.Register() 统一调用。
func Register() {
	memberController := controller.NewMemberController()
	levelController := controller.NewMemberLevelController()
	tagController := controller.NewMemberTagController()
	groupController := controller.NewMemberGroupController()
	socialClientController := controller.NewMemberSocialClientController()
	signConfigController := controller.NewMemberSignConfigController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 权限中间件在 base，member 只声明自己的权限点
		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// 写操作留痕，中间件注册在 Permission 之内
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// ---------- 会员 ----------
		perm("member:user:list").Get("/users", memberController.Index)
		perm("member:user:list").Get("/users/{id}", memberController.Show)
		perm("member:user:create").Middleware(operate("会员管理", "新增会员")).
			Post("/users", memberController.Store)
		perm("member:user:update").Middleware(operate("会员管理", "编辑会员")).
			Put("/users/{id}", memberController.Update)
		perm("member:user:update").Middleware(operate("会员管理", "变更会员状态")).
			Put("/users/{id}/status", memberController.UpdateStatus)
		perm("member:user:delete").Middleware(operate("会员管理", "删除会员")).
			Delete("/users/{id}", memberController.Destroy)

		// 积分：列表随会员权限，调整单独授权
		perm("member:user:list").Get("/users/{id}/points", memberController.PointRecords)
		perm("member:user:points").Middleware(operate("会员管理", "调整积分")).
			Post("/users/{id}/point", memberController.AdjustPoint)

		// 等级调整与等级记录
		perm("member:user:list").Get("/users/{id}/level-records", memberController.LevelRecords)
		perm("member:user:level").Middleware(operate("会员管理", "调整等级")).
			Put("/users/{id}/level", memberController.AdjustLevel)

		// 签到记录、收货地址与第三方绑定：后台只读
		perm("member:user:list").Get("/users/{id}/sign-records", memberController.SignRecords)
		perm("member:user:list").Get("/users/{id}/addresses", memberController.Addresses)
		perm("member:user:list").Get("/users/{id}/social-bindings", memberController.SocialBindings)

		// ---------- 等级 ----------
		// 下拉供会员表单使用，因此只要求会员列表权限
		perm("member:user:list").Get("/levels/options", levelController.Options)
		perm("member:level:list").Get("/levels", levelController.Index)
		perm("member:level:list").Get("/levels/{id}", levelController.Show)
		perm("member:level:create").Middleware(operate("会员管理", "新增等级")).
			Post("/levels", levelController.Store)
		perm("member:level:update").Middleware(operate("会员管理", "编辑等级")).
			Put("/levels/{id}", levelController.Update)
		perm("member:level:delete").Middleware(operate("会员管理", "删除等级")).
			Delete("/levels/{id}", levelController.Destroy)

		// ---------- 标签 ----------
		perm("member:user:list").Get("/tags/options", tagController.Options)
		perm("member:tag:list").Get("/tags", tagController.Index)
		perm("member:tag:create").Middleware(operate("会员管理", "新增标签")).
			Post("/tags", tagController.Store)
		perm("member:tag:update").Middleware(operate("会员管理", "编辑标签")).
			Put("/tags/{id}", tagController.Update)
		perm("member:tag:delete").Middleware(operate("会员管理", "删除标签")).
			Delete("/tags/{id}", tagController.Destroy)

		// ---------- 分组 ----------
		perm("member:user:list").Get("/groups/options", groupController.Options)
		perm("member:group:list").Get("/groups", groupController.Index)
		perm("member:group:create").Middleware(operate("会员管理", "新增分组")).
			Post("/groups", groupController.Store)
		perm("member:group:update").Middleware(operate("会员管理", "编辑分组")).
			Put("/groups/{id}", groupController.Update)
		perm("member:group:delete").Middleware(operate("会员管理", "删除分组")).
			Delete("/groups/{id}", groupController.Destroy)

		// ---------- 签到奖励档位 ----------
		// 签到记录在 C 端生成，这里只维护连续第 N 天发多少积分 / 经验
		perm("member:sign-config:list").Get("/sign-configs", signConfigController.Index)
		perm("member:sign-config:create").Middleware(operate("会员管理", "新增签到档位")).
			Post("/sign-configs", signConfigController.Store)
		perm("member:sign-config:update").Middleware(operate("会员管理", "编辑签到档位")).
			Put("/sign-configs/{id}", signConfigController.Update)
		perm("member:sign-config:delete").Middleware(operate("会员管理", "删除签到档位")).
			Delete("/sign-configs/{id}", signConfigController.Destroy)

		// ---------- 第三方登录渠道配置 ----------
		perm("member:social-client:list").Get("/social-clients", socialClientController.Index)
		perm("member:social-client:list").Get("/social-clients/{id}", socialClientController.Show)
		perm("member:social-client:create").Middleware(operate("会员管理", "新增社交客户端")).
			Post("/social-clients", socialClientController.Store)
		perm("member:social-client:update").Middleware(operate("会员管理", "编辑社交客户端")).
			Put("/social-clients/{id}", socialClientController.Update)
		perm("member:social-client:delete").Middleware(operate("会员管理", "删除社交客户端")).
			Delete("/social-clients/{id}", socialClientController.Destroy)
	})
}
