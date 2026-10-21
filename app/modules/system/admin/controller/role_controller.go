package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/system/admin/request"
	"eqray-flavor/app/modules/system/service"
)

// RoleController 提供 system_role 的 CRUD 与角色授权接口。
type RoleController struct {
	service *service.RoleService
}

func NewRoleController() *RoleController {
	return &RoleController{service: service.NewRoleService()}
}

// Index 分页查询角色，支持按名称、编码与状态筛选。
// 入参：name、code、status、pageNo、pageSize（查询串）；返回角色列表与总数。
func (c *RoleController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.RoleListQuery{
		Name:     ctx.Request().Query("name"),
		Code:     ctx.Request().Query("code"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := baseEnums.CommonStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Options 查询启用中的角色，供用户分配角色时下拉选择。
// 无入参；返回角色列表。
func (c *RoleController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show 查询单个角色详情。
// 入参：路径 id；返回角色详情，不存在返回业务错误码。
func (c *RoleController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	role, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, role)
}

// Store 新增角色；角色编码唯一，重复时返回业务错误码。
// 入参：name、code、sort、status、remark；返回新角色 ID。
func (c *RoleController) Store(ctx http.Context) http.Response {
	var req request.RoleCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.RoleSaveInput{
		Name:   req.Name,
		Code:   req.Code,
		Sort:   req.Sort,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑角色名称、编码、排序、状态与备注，并刷新权限缓存版本。
// 入参：路径 id + 角色字段；返回 null。
func (c *RoleController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.RoleUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.RoleSaveInput{
		Name:   req.Name,
		Code:   req.Code,
		Sort:   req.Sort,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除角色（软删）；内置角色或仍被用户使用时不建议删除。
// 入参：路径 id；返回 null。
func (c *RoleController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Menus 查询角色已授权的菜单 ID 列表，供权限树回显。
// 入参：路径 id；返回菜单 ID 数组。
func (c *RoleController) Menus(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	menuIDs, err := c.service.MenuIDs(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, menuIDs)
}

// AssignMenus 覆盖式设置角色菜单权限，并前移 rbac 缓存版本让新权限立即生效。
// 入参：路径 id + menuIds；返回 null。
func (c *RoleController) AssignMenus(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.RoleAssignMenuRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.AssignMenus(ctx.Context(), id, req.MenuIDs); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
