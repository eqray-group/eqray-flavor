package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/dish/service"
)

// MenuController 提供门店端菜单查询。
// 与 C 端菜单的区别：走后台令牌鉴权，门店取请求头 X-Store-Id，供 POS 点单使用。
type MenuController struct {
	service *service.MenuService
}

// NewMenuController 构造菜单控制器。
func NewMenuController() *MenuController {
	return &MenuController{service: service.NewMenuService()}
}

// Index 返回当前门店的可售菜单：分类树 + 菜品 + 规格 + 加料组。
// 入参：memberId（可选，传了按会员价算）；返回菜单快照。
// 说明：菜单里已过滤不可售规格，沽清规格保留并带 soldOut 标记，由前端置灰。
func (c *MenuController) Index(ctx http.Context) http.Response {
	storeID := support.StoreID(ctx.Context())
	if storeID == 0 {
		return result.Fail(ctx, errs.WithMsg(errs.BadRequest, "请先选择门店"))
	}

	// 收银员给会员点单时传 memberId，价格按会员价展示，与结算口径一致
	memberID := param.QueryID(ctx, "memberId")

	snapshot, err := c.service.Menu(ctx.Context(), storeID, memberID > 0)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, snapshot)
}
