package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/admin/request"
	"eqray-flavor/app/modules/system/service"
)

// MenuController 提供 system_menu 的树形管理接口。
type MenuController struct {
	service *service.MenuService
}

func NewMenuController() *MenuController {
	return &MenuController{service: service.NewMenuService()}
}

// Index 查询完整菜单树，用于菜单管理页面的表格展示。
// 无入参；返回菜单树（含按钮节点）。
func (c *MenuController) Index(ctx http.Context) http.Response {
	tree, err := c.service.Tree(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tree)
}

// Simple 查询精简菜单树，供「选择上级菜单」与角色授权勾选使用。
// 无入参；返回菜单树。
func (c *MenuController) Simple(ctx http.Context) http.Response {
	tree, err := c.service.Tree(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tree)
}

// Routes 按当前登录者可见的菜单生成前端动态路由（后端权限模式）。
// 无入参，用户身份取自令牌；返回菜单路由树。
func (c *MenuController) Routes(ctx http.Context) http.Response {
	routes, err := c.service.Routes(ctx.Context(), types.ID(support.UserID(ctx.Context())))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, routes)
}

// Show 查询单个菜单详情，供编辑弹窗回填。
// 入参：路径 id；返回菜单详情，不存在返回业务错误码。
func (c *MenuController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	menu, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, menu)
}

// Store 新增菜单或按钮权限点，并刷新 rbac 缓存版本。
// 入参：name、permission、type、parentId、path、component、status 等；返回新菜单 ID。
func (c *MenuController) Store(ctx http.Context) http.Response {
	var req request.MenuCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.MenuSaveInput{
		Name:          req.Name,
		Permission:    req.Permission,
		Type:          req.Type,
		Sort:          req.Sort,
		ParentID:      req.ParentID,
		Path:          req.Path,
		Icon:          req.Icon,
		Component:     req.Component,
		ComponentName: req.ComponentName,
		Status:        req.Status,
		Visible:       req.Visible,
		KeepAlive:     req.KeepAlive,
		AlwaysShow:    req.AlwaysShow,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑菜单的名称、权限点、路由路径、组件与显隐配置，并刷新 rbac 缓存版本。
// 入参：路径 id + 菜单字段；返回 null。
func (c *MenuController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MenuUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.MenuSaveInput{
		Name:          req.Name,
		Permission:    req.Permission,
		Type:          req.Type,
		Sort:          req.Sort,
		ParentID:      req.ParentID,
		Path:          req.Path,
		Icon:          req.Icon,
		Component:     req.Component,
		ComponentName: req.ComponentName,
		Status:        req.Status,
		Visible:       req.Visible,
		KeepAlive:     req.KeepAlive,
		AlwaysShow:    req.AlwaysShow,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除菜单（软删）；存在子菜单时由服务层拒绝，避免出现孤立节点。
// 入参：路径 id；返回 null。
func (c *MenuController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
