// Package controller 提供菜品模块的 C 端接口（顾客点餐使用的门店菜单）。
package controller

import (
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/dish/service"
)

// MenuController 提供顾客端门店菜单查询。
type MenuController struct {
	service *service.MenuService
}

// NewMenuController 构造顾客端菜单控制器。
func NewMenuController() *MenuController {
	return &MenuController{service: service.NewMenuService()}
}

// Index 返回指定门店的菜单，价格已按当前会员身份计算。
// 入参：查询串 storeId（必填）；返回分类树、菜品、可售规格与加料组。
func (c *MenuController) Index(ctx http.Context) http.Response {
	storeID := parseStoreID(ctx.Request().Query("storeId"))
	if storeID <= 0 {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "请指定门店"))
	}

	// C 端菜单接口挂会员令牌，会员价档位可直接生效
	isMember := support.MemberID(ctx.Context()) > 0

	menu, err := c.service.Menu(ctx.Context(), storeID, isMember)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, menu)
}

// parseStoreID 解析查询串中的门店 ID，非法值返回 0。
func parseStoreID(raw string) types.ID {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}

	return types.ID(parsed)
}
