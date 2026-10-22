// Package middleware 提供门店级接口共用的请求上下文校验。
package middleware

import (
	"net/http"
	"strconv"
	"strings"

	contractsHttp "github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/service"
)

// StoreContextHeader 是前端注入当前门店的请求头名称。
// 服务端同时兼容查询串 storeId，便于调试与不支持自定义请求头的场景。
const StoreContextHeader = "X-Store-Id"

// StoreContext 读取请求门店标识，校验当前用户可见范围后写入上下文。
// 未携带门店标识时不拦截，服务层按当前用户可见门店集合查询（design/23 §1）。
type StoreContext struct{}

// Signature 是中间件的唯一标识。
func (m *StoreContext) Signature() string {
	return "store_context"
}

// Handle 校验门店上下文：
// 1. 请求头 X-Store-Id 优先，其次查询串 storeId；
// 2. 门店必须存在且未删除；
// 3. 门店必须在当前用户可见范围内；
// 4. 通过后写入 support.StoreID，越权时返回固定业务错误码并中断。
func (m *StoreContext) Handle(ctx contractsHttp.Context) {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		result.MarkFail(ctx, baseErrs.Unauthorized)
		ctx.Request().Abort(http.StatusUnauthorized)
		return
	}

	raw := strings.TrimSpace(ctx.Request().Header(StoreContextHeader))
	if raw == "" {
		raw = strings.TrimSpace(ctx.Request().Query("storeId"))
	}
	if raw == "" {
		ctx.Request().Next()
		return
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		result.AbortFail(ctx, baseErrs.New(storeErrs.StoreContextInvalid))
		return
	}

	storeService := service.NewStoreService()
	storeID := types.ID(value)
	store, err := storeService.FindByID(ctx.Context(), storeID)
	if err != nil {
		result.AbortFail(ctx, err)
		return
	}
	if store == nil {
		result.AbortFail(ctx, baseErrs.New(storeErrs.StoreContextNotExists))
		return
	}
	if store.Type != enums.StoreTypeShop {
		result.AbortFail(ctx, baseErrs.New(storeErrs.StoreContextTypeInvalid))
		return
	}

	if err := storeService.EnsureStoreVisible(ctx.Context(), types.ID(userID), storeID); err != nil {
		result.AbortFail(ctx, err)
		return
	}

	ctx.WithContext(support.WithStoreID(ctx.Context(), storeID))
	ctx.Request().Next()
}
