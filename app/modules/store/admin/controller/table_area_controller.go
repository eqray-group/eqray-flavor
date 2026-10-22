package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/store/admin/request"
	"eqray-flavor/app/modules/store/service"
)

// TableAreaController 提供桌台区域的增删改查与下拉选项接口。
type TableAreaController struct {
	service *service.StoreTableAreaService
}

// NewTableAreaController 构造桌台区域控制器。
func NewTableAreaController() *TableAreaController {
	return &TableAreaController{service: service.NewStoreTableAreaService()}
}

// Index 分页查询桌台区域，支持区域名称关键字筛选。
// 门店范围由请求头门店上下文或当前用户可见门店集合决定。
func (c *TableAreaController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), service.TableAreaListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Options 返回当前门店上下文或可见范围内的全部区域，供桌台筛选与弹窗选择。
func (c *TableAreaController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Store 新增桌台区域；区域名称在同一门店内唯一。
// 入参：storeId、name、sort、remark；返回新区域 ID。
func (c *TableAreaController) Store(ctx http.Context) http.Response {
	var req request.TableAreaSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toTableAreaSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑桌台区域；区域不能跨门店移动。
// 入参：路径 id + 区域字段；返回 null。
func (c *TableAreaController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.TableAreaSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toTableAreaSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除桌台区域；区域下仍有桌台时拒绝。
// 入参：路径 id；返回 null。
func (c *TableAreaController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toTableAreaSaveInput 把请求体转换为服务层入参。
func toTableAreaSaveInput(req request.TableAreaSaveRequest) service.TableAreaSaveInput {
	return service.TableAreaSaveInput{
		StoreID: req.StoreID,
		Name:    req.Name,
		Sort:    req.Sort,
		Remark:  req.Remark,
	}
}
