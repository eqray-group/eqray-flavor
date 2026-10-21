package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/infra/admin/request"
	"eqray-flavor/app/modules/infra/service"
)

// ConfigController 提供参数配置的增删改查接口。
type ConfigController struct {
	service *service.ConfigService
}

func NewConfigController() *ConfigController {
	return &ConfigController{service: service.NewConfigService()}
}

// Index 分页查询参数配置，支持按分类与关键字筛选。
// 入参：category、keyword、pageNo、pageSize（查询串）；返回配置列表与总数。
func (c *ConfigController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), service.ConfigListQuery{
		Category: ctx.Request().Query("category"),
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Store 新增参数配置；配置键唯一，重复时返回业务错误码。
// 入参：category、name、configKey、value、type、visible、remark；返回 null。
func (c *ConfigController) Store(ctx http.Context) http.Response {
	var req request.ConfigCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err := c.service.Create(ctx.Context(), service.ConfigCreateInput{
		Category:  req.Category,
		Name:      req.Name,
		ConfigKey: req.ConfigKey,
		Value:     req.Value,
		Type:      req.Type,
		Visible:   req.Visible,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Update 编辑参数配置；配置键不可修改。
// 入参：路径 id + category、name、value、visible、remark；返回 null。
func (c *ConfigController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.ConfigUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.ConfigUpdateInput{
		Category: req.Category,
		Name:     req.Name,
		Value:    req.Value,
		Visible:  req.Visible,
		Remark:   req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除参数配置（软删），删除后业务读取该键会回落默认值。
// 入参：路径 id；返回 null。
func (c *ConfigController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
