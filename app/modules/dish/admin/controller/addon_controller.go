package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/dish/admin/request"
	"eqray-flavor/app/modules/dish/service"
)

// AddonController 提供加料组与加料项的增删改查接口。
type AddonController struct {
	service *service.AddonService
}

// NewAddonController 构造加料控制器。
func NewAddonController() *AddonController {
	return &AddonController{service: service.NewAddonService()}
}

// Index 分页查询加料组，支持名称关键字筛选，并返回加料项数量与菜品引用次数。
// 入参：keyword、pageNo、pageSize（查询串）；返回加料组列表与总数。
func (c *AddonController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), service.AddonGroupListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 返回加料组详情，包含其全部加料项。
// 入参：路径 id；返回加料组与加料项列表。
func (c *AddonController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	detail, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, detail)
}

// Store 新增加料组；名称为品牌级唯一，必选时最少选择数至少为 1。
// 入参：name、required、minSelect、maxSelect、sort、status、remark；返回新加料组 ID。
func (c *AddonController) Store(ctx http.Context) http.Response {
	var req request.AddonGroupSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toAddonGroupSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑加料组。
// 入参：路径 id + 加料组字段；返回 null。
func (c *AddonController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.AddonGroupSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toAddonGroupSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除加料组；已被菜品引用时拒绝，加料项随组一起清理。
// 入参：路径 id；返回 null。
func (c *AddonController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// StoreItem 在指定加料组下新增加料项，同组内名称唯一。
// 入参：路径 groupId + name、price、sort、status、remark；返回新加料项 ID。
func (c *AddonController) StoreItem(ctx http.Context) http.Response {
	groupID, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.AddonItemSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.CreateItem(ctx.Context(), groupID, toAddonItemSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// UpdateItem 编辑加料项；所属加料组不允许修改。
// 入参：路径 id + 加料项字段；返回 null。
func (c *AddonController) UpdateItem(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.AddonItemSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.UpdateItem(ctx.Context(), id, toAddonItemSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// DestroyItem 删除加料项，已下单的历史快照不受影响。
// 入参：路径 id；返回 null。
func (c *AddonController) DestroyItem(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteItem(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toAddonGroupSaveInput 把请求体转换为服务层入参。
func toAddonGroupSaveInput(req request.AddonGroupSaveRequest) service.AddonGroupSaveInput {
	return service.AddonGroupSaveInput{
		Name:      req.Name,
		Required:  req.Required,
		MinSelect: req.MinSelect,
		MaxSelect: req.MaxSelect,
		Sort:      req.Sort,
		Status:    baseStatus(req.Status),
		Remark:    req.Remark,
	}
}

// toAddonItemSaveInput 把请求体转换为服务层入参。
func toAddonItemSaveInput(req request.AddonItemSaveRequest) service.AddonItemSaveInput {
	return service.AddonItemSaveInput{
		Name:   req.Name,
		Price:  req.Price,
		Sort:   req.Sort,
		Status: baseStatus(req.Status),
		Remark: req.Remark,
	}
}
