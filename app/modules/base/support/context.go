package support

import (
	"context"

	"eqray-flavor/app/modules/base/types"
)

type userIDKey struct{}

// WithUserID 把当前登录者写入上下文，供 Model 基类自动填充 creator / updater。
func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// UserID 从上下文取当前登录者，未登录或未注入时返回 0。
func UserID(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if v, ok := ctx.Value(userIDKey{}).(int64); ok {
		return v
	}
	return 0
}

type memberIDKey struct{}

// WithMemberID 把 C 端会员写入上下文，与后台登录者（UserID）分开存放，
// 避免会员令牌被后台权限体系误判。
func WithMemberID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, memberIDKey{}, id)
}

// MemberID 从上下文取当前登录会员，未登录或未注入时返回 0。
func MemberID(ctx context.Context) types.ID {
	if ctx == nil {
		return 0
	}
	if v, ok := ctx.Value(memberIDKey{}).(int64); ok {
		return types.ID(v)
	}

	return 0
}
