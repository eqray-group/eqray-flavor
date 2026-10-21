package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/member/admin/request"
	"eqray-flavor/app/modules/member/service"
)

// MemberGroupController 提供会员分组的增删改查接口。
type MemberGroupController struct {
	service *service.MemberGroupService
}

// NewMemberGroupController 构造会员分组控制器。
func NewMemberGroupController() *MemberGroupController {
	return &MemberGroupController{service: service.NewMemberGroupService()}
}

// Index 分页查询会员分组，支持按名称模糊与状态筛选。
// 入参：keyword、status、pageNo、pageSize（查询串）；返回分组列表与总数。
func (c *MemberGroupController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.MemberGroupListQuery{
		Keyword:  ctx.Request().Query("keyword"),
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

// Options 查询启用中的分组，供会员表单下拉。
// 无入参；返回分组列表。
func (c *MemberGroupController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Store 新增会员分组。
// 入参：name、status、remark；返回新分组 ID。
func (c *MemberGroupController) Store(ctx http.Context) http.Response {
	var req request.MemberGroupSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.MemberGroupSaveInput{
		Name:   req.Name,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑会员分组的名称、状态与备注。
// 入参：路径 id + name、status、remark；返回 null。
func (c *MemberGroupController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberGroupSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, service.MemberGroupSaveInput{
		Name:   req.Name,
		Status: req.Status,
		Remark: req.Remark,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除会员分组（软删）。
// 入参：路径 id；返回 null。
func (c *MemberGroupController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
