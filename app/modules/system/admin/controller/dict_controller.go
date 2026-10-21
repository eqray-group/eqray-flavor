package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/system/admin/request"
	"eqray-flavor/app/modules/system/service"
)

// DictController 提供字典类型与字典数据的管理接口。
type DictController struct {
	service *service.DictService
}

func NewDictController() *DictController {
	return &DictController{service: service.NewDictService()}
}

// Types 分页查询字典类型，支持按名称、类型编码与状态筛选。
// 入参：name、type、status、pageNo、pageSize（查询串）；返回类型列表与总数。
func (c *DictController) Types(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.DictTypeListQuery{
		Name:     ctx.Request().Query("name"),
		Type:     ctx.Request().Query("type"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := baseEnums.CommonStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.PageTypes(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// TypeOptions 查询启用中的字典类型，供字典数据弹窗选择所属类型。
// 无入参；返回类型列表。
func (c *DictController) TypeOptions(ctx http.Context) http.Response {
	list, err := c.service.TypeOptions(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// TypeShow 查询单个字典类型详情。
// 入参：路径 id；返回类型详情，不存在返回业务错误码。
func (c *DictController) TypeShow(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.TypeDetail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// TypeStore 新增字典类型；类型编码唯一，重复时返回业务错误码。
// 入参：name、type、status、remark；返回新类型 ID。
func (c *DictController) TypeStore(ctx http.Context) http.Response {
	var req request.DictTypeCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.CreateType(ctx.Context(), service.DictTypeSaveInput{
		Name:   req.Name,
		Type:   req.Type,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// TypeUpdate 编辑字典类型名称、编码、状态与备注。
// 入参：路径 id + name、type、status、remark；返回 null。
func (c *DictController) TypeUpdate(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.DictTypeUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.UpdateType(ctx.Context(), id, service.DictTypeSaveInput{
		Name:   req.Name,
		Type:   req.Type,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// TypeDestroy 删除字典类型（软删）；其下仍有字典数据时由服务层拒绝。
// 入参：路径 id；返回 null。
func (c *DictController) TypeDestroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteType(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Data 分页查询字典数据，支持按类型编码、标签与状态筛选。
// 入参：dictType、label、status、pageNo、pageSize（查询串）；返回数据列表与总数。
func (c *DictController) Data(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.DictDataListQuery{
		DictType: ctx.Request().Query("dictType"),
		Label:    ctx.Request().Query("label"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := baseEnums.CommonStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.PageData(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// DataShow 查询单条字典数据详情。
// 入参：路径 id；返回数据详情，不存在返回业务错误码。
func (c *DictController) DataShow(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.DataDetail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// DataStore 新增字典数据；同一类型下数据值唯一。
// 入参：dictType、label、value、sort、status、colorType、cssClass、remark；返回新数据 ID。
func (c *DictController) DataStore(ctx http.Context) http.Response {
	var req request.DictDataCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.CreateData(ctx.Context(), service.DictDataSaveInput{
		Sort:      req.Sort,
		Label:     req.Label,
		Value:     req.Value,
		DictType:  req.DictType,
		Status:    req.Status,
		ColorType: req.ColorType,
		CSSClass:  req.CSSClass,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// DataUpdate 编辑字典数据的标签、值、排序、状态与样式。
// 入参：路径 id + 数据字段；返回 null。
func (c *DictController) DataUpdate(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.DictDataUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.UpdateData(ctx.Context(), id, service.DictDataSaveInput{
		Sort:      req.Sort,
		Label:     req.Label,
		Value:     req.Value,
		DictType:  req.DictType,
		Status:    req.Status,
		ColorType: req.ColorType,
		CSSClass:  req.CSSClass,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// DataDestroy 删除单条字典数据（软删）。
// 入参：路径 id；返回 null。
func (c *DictController) DataDestroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteData(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Items 按类型编码查询启用中的字典数据，前端下拉与标签渲染统一走这里，只要求登录态。
// 入参：路径 type（类型编码）；返回字典项列表。
func (c *DictController) Items(ctx http.Context) http.Response {
	items, err := c.service.Items(ctx.Context(), ctx.Request().Route("type"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, items)
}
