package middleware

import (
	"github.com/goravel/framework/contracts/http"

	baseAuth "eqray-flavor/app/modules/base/auth"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
)

// MemberAuth 校验会员访问令牌，通过后把会员 ID 写入上下文。
// 与后台 TokenAuth 使用不同的令牌作用域，两者不能互相冒充。
// 放在 base 是因为任何模块的 C 端接口都可能用到，业务模块之间不必互相依赖。
type MemberAuth struct{}

// Signature 是中间件的唯一标识。
func (m *MemberAuth) Signature() string {
	return "member_auth"
}

// Handle 校验会员访问令牌，未通过时直接中断请求。
func (m *MemberAuth) Handle(ctx http.Context) {
	token := baseAuth.BearerToken(ctx.Request().Header("Authorization"))
	memberID, err := baseAuth.ResolveScoped(baseAuth.ScopeMember, token)
	if err != nil {
		result.MarkFail(ctx, baseErrs.Unauthorized)
		ctx.Request().Abort(http.StatusUnauthorized)
		return
	}

	ctx.WithContext(support.WithMemberID(ctx.Context(), int64(memberID)))
	ctx.Request().Next()
}
