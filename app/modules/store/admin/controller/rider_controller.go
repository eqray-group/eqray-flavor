package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/store/admin/request"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/service"
)

// RiderController 提供骑手档案的增删改查与已有账号选择接口。
type RiderController struct {
	service *service.StoreRiderService
}

// NewRiderController 构造骑手档案控制器。
func NewRiderController() *RiderController {
	return &RiderController{service: service.NewStoreRiderService()}
}

// Index 分页查询骑手档案，支持账号关键字与状态筛选。
// 门店范围由请求头门店上下文或当前用户可见门店集合决定。
func (c *RiderController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.RiderListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", 0); status > 0 {
		value := enums.RiderStatusEnum(status)
		q.Status = &value
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// UserOptions 搜索可绑定为骑手的启用后台账号，供新增弹窗选择已有账号。
func (c *RiderController) UserOptions(ctx http.Context) http.Response {
	list, err := c.service.UserOptions(ctx.Context(), ctx.Request().Query("keyword"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Store 新增骑手档案；未指定账号时自动开通后台账号并返回一次性初始密码。
func (c *RiderController) Store(ctx http.Context) http.Response {
	var req request.RiderCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	data, err := c.service.Create(ctx.Context(), service.RiderCreateInput{
		StoreID:   req.StoreID,
		UserID:    req.UserID,
		Name:      req.Name,
		Mobile:    req.Mobile,
		MaxOrders: req.MaxOrders,
		Status:    enums.RiderStatusEnum(req.Status),
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, data)
}

// Update 编辑骑手档案；门店、员工与账号不可变更。
func (c *RiderController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.RiderUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, service.RiderUpdateInput{
		MaxOrders: req.MaxOrders,
		Status:    enums.RiderStatusEnum(req.Status),
		Remark:    req.Remark,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除骑手档案；仍有进行中配送单时拒绝。
func (c *RiderController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
