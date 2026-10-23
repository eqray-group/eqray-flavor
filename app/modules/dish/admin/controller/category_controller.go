package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/dish/admin/request"
	"eqray-flavor/app/modules/dish/service"
)

// CategoryController 提供菜品分类树的查询与维护接口。
type CategoryController struct {
	service *service.CategoryService
}

// NewCategoryController 构造菜品分类控制器。
func NewCategoryController() *CategoryController {
	return &CategoryController{service: service.NewCategoryService()}
}

// Tree 返回完整菜品分类树，节点带直属菜品数，供分类管理与菜品表单选择使用。
func (c *CategoryController) Tree(ctx http.Context) http.Response {
	tree, err := c.service.Tree(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tree)
}

// Store 新增菜品分类；同一上级下名称唯一，上级为 0 时表示根分类。
// 入参：parentId、name、image、sort、status、remark；返回新分类 ID。
func (c *CategoryController) Store(ctx http.Context) http.Response {
	var req request.CategorySaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toCategorySaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑菜品分类；上级不能是自身或自身下级，同级不能重名。
// 入参：路径 id + 分类字段；返回 null。
func (c *CategoryController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.CategorySaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toCategorySaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除菜品分类；仍有子分类或菜品时拒绝。
// 入参：路径 id；返回 null。
func (c *CategoryController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toCategorySaveInput 把请求体转换为服务层入参。
func toCategorySaveInput(req request.CategorySaveRequest) service.CategorySaveInput {
	return service.CategorySaveInput{
		ParentID: req.ParentID,
		Name:     req.Name,
		Image:    req.Image,
		Sort:     req.Sort,
		Status:   baseStatus(req.Status),
		Remark:   req.Remark,
	}
}
