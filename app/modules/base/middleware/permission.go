package middleware

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
)

// PermissionChecker 由持有权限数据的模块注入（当前为 system），
// 避免 base 反向依赖业务模块，同时让 infra / member 也能复用权限中间件（docs 3、4）。
type PermissionChecker func(ctx context.Context, userID types.ID, code string) (bool, error)

var permissionChecker PermissionChecker

// RegisterPermissionChecker 在应用启动时注入权限实现，未注入时所有鉴权请求会被拒绝。
func RegisterPermissionChecker(checker PermissionChecker) {
	permissionChecker = checker
}

// Permission 按路由声明的权限标识做接口级 RBAC 校验（docs 9.3）。
// 超级管理员的放行规则由注入的 PermissionChecker 负责。
type Permission struct {
	Code string
}

// Signature 让每个权限标识对应独立的中间件签名。
func (m *Permission) Signature() string {
	return "permission_" + strings.NewReplacer(":", "_", "-", "_").Replace(m.Code)
}

// Handle 校验当前用户是否拥有路由要求的权限。
func (m *Permission) Handle(ctx http.Context) {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		result.MarkFail(ctx, baseErrs.Unauthorized)
		ctx.Request().Abort(http.StatusUnauthorized)
		return
	}

	// 未声明权限标识的路由只要求登录态
	if m.Code == "" {
		ctx.Request().Next()
		return
	}

	// 未注入实现说明启动装配缺失，直接拒绝而不是放行
	if permissionChecker == nil {
		result.MarkFail(ctx, baseErrs.InternalServerError)
		ctx.Request().Abort(http.StatusInternalServerError)
		return
	}

	allowed, err := permissionChecker(ctx.Context(), types.ID(userID), m.Code)
	if err != nil {
		result.MarkFail(ctx, baseErrs.InternalServerError)
		ctx.Request().Abort(http.StatusInternalServerError)
		return
	}
	if !allowed {
		result.MarkFail(ctx, baseErrs.Forbidden)
		ctx.Request().Abort(http.StatusForbidden)
		return
	}

	ctx.Request().Next()
}
