package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/infra/admin/request"
	"eqray-flavor/app/modules/infra/service"
)

// InfraDistrictController 提供行政区划的增删改查与树查询接口。
type InfraDistrictController struct {
	service *service.InfraDistrictService
}

// NewInfraDistrictController 构造行政区划控制器。
func NewInfraDistrictController() *InfraDistrictController {
	return &InfraDistrictController{service: service.NewInfraDistrictService()}
}

// Index 分页查询行政区划，支持名称 / 简称 / 编码关键字与级别、上级筛选。
// 入参：keyword、level、pid、pageNo、pageSize（查询串）；返回区划列表与总数。
func (c *InfraDistrictController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.DistrictListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		Level:    ctx.Request().QueryInt("level", 0),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	// pid 传 0 查顶级，不传则不限
	if pid := ctx.Request().Query("pid"); pid != "" {
		value := types.ID(ctx.Request().QueryInt("pid", 0))
		q.Pid = &value
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Tree 返回完整行政区划树，供管理页树形展示与级联选择。
// 无入参；返回以顶级区划为根的嵌套节点数组。
func (c *InfraDistrictController) Tree(ctx http.Context) http.Response {
	tree, err := c.service.Tree(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tree)
}

// Options 返回指定上级的直接下级，pid 缺省为 0（顶级）。
// 入参：pid（查询串）；返回区划数组，用于级联选择逐级加载。
func (c *InfraDistrictController) Options(ctx http.Context) http.Response {
	list, err := c.service.Children(ctx.Context(), types.ID(ctx.Request().QueryInt("pid", 0)))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show 查询单个区划详情。
// 入参：路径 id；返回区划详情。
func (c *InfraDistrictController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// Store 新增行政区划；层级与父级编码由上级自动推导。
// 入参：pid、areaName、shortName、areaCode、sort；返回新区划 ID。
func (c *InfraDistrictController) Store(ctx http.Context) http.Response {
	var req request.InfraDistrictSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.DistrictSaveInput{
		Pid:       types.ID(req.Pid),
		AreaName:  req.AreaName,
		ShortName: req.ShortName,
		AreaCode:  req.AreaCode,
		Sort:      req.Sort,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑行政区划；上级与行政编码不可修改，入参中的对应字段会被忽略。
// 入参：路径 id + areaName、shortName、sort；返回 null。
func (c *InfraDistrictController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.InfraDistrictSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.DistrictSaveInput{
		Pid:       types.ID(req.Pid),
		AreaName:  req.AreaName,
		ShortName: req.ShortName,
		AreaCode:  req.AreaCode,
		Sort:      req.Sort,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除行政区划；仍有下级时拒绝。
// 入参：路径 id；返回 null。
func (c *InfraDistrictController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
