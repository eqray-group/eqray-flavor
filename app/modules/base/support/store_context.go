package support

import (
	"context"

	"eqray-flavor/app/modules/base/types"
)

// storeIDKey 用于把当前请求选定的门店写入上下文。
// 0 表示未指定门店，服务层按当前用户可见门店集合查询（docs design/23 §1）。
type storeIDKey struct{}

// WithStoreID 把当前请求的门店上下文写入 ctx，供门店级业务统一读取。
// 入参：ctx 请求上下文、id 门店 ID；返回带门店上下文的新 ctx。
func WithStoreID(ctx context.Context, id types.ID) context.Context {
	return context.WithValue(ctx, storeIDKey{}, id)
}

// StoreID 读取当前请求的门店上下文，未指定时返回 0。
func StoreID(ctx context.Context) types.ID {
	if ctx == nil {
		return 0
	}
	if v, ok := ctx.Value(storeIDKey{}).(types.ID); ok {
		return v
	}

	return 0
}
