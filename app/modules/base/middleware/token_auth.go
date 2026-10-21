// Package middleware 提供框架级中间件，见 docs 4。
package middleware

import (
	"github.com/goravel/framework/contracts/http"

	baseAuth "eqray-flavor/app/modules/base/auth"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
)

// TokenAuth 校验 Authorization: Bearer <accessToken>，通过后把用户 ID 写入上下文，
// 供 Model 基类填充 creator / updater（docs 9.5）。
type TokenAuth struct{}

// Signature 是中间件的唯一标识。
func (m *TokenAuth) Signature() string {
	return "token_auth"
}

// Handle 校验访问令牌，未通过时直接中断请求。
func (m *TokenAuth) Handle(ctx http.Context) {
	token := baseAuth.BearerToken(ctx.Request().Header("Authorization"))
	userID, err := baseAuth.Resolve(token)
	if err != nil {
		result.MarkFail(ctx, baseErrs.Unauthorized)
		ctx.Request().Abort(http.StatusUnauthorized)
		return
	}

	ctx.WithContext(support.WithUserID(ctx.Context(), int64(userID)))
	ctx.Request().Next()
}
