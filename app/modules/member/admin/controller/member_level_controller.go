package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/member/admin/request"
	"eqray-flavor/app/modules/member/service"
)

// MemberLevelController 提供会员等级的增删改查接口。
type MemberLevelController struct {
	service *service.MemberLevelService
}

// NewMemberLevelController 构造会员等级控制器。
func NewMemberLevelController() *MemberLevelController {
	return &MemberLevelController{service: service.NewMemberLevelService()}
}

// Index 分页查询会员等级，支持按名称模糊与状态筛选。
// 入参：keyword、status、pageNo、pageSize（查询串）；返回等级列表与总数。
func (c *MemberLevelController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.MemberLevelListQuery{
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

// Options 查询启用中的等级，供会员表单下拉选择。
// 无入参；返回等级列表。
func (c *MemberLevelController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show 查询单个会员等级的详情。
// 入参：路径 id；返回等级详情，不存在返回业务错误码。
func (c *MemberLevelController) Show(ctx http.Context) http.Response {
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

// Store 新增会员等级；等级序号唯一，重复时返回业务错误码。
// 入参：level、name、icon、experience、discountPercent、description、status、sort；返回新等级 ID。
func (c *MemberLevelController) Store(ctx http.Context) http.Response {
	var req request.MemberLevelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.MemberLevelSaveInput{
		Level:           req.Level,
		Name:            req.Name,
		Icon:            req.Icon,
		Experience:      req.Experience,
		DiscountPercent: req.DiscountPercent,
		Description:     req.Description,
		Status:          req.Status,
		Sort:            req.Sort,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑会员等级；等级序号创建后不可修改，入参中的 level 会被忽略。
// 入参：路径 id + 等级配置字段；返回 null。
func (c *MemberLevelController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberLevelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.MemberLevelSaveInput{
		Level:           req.Level,
		Name:            req.Name,
		Icon:            req.Icon,
		Experience:      req.Experience,
		DiscountPercent: req.DiscountPercent,
		Description:     req.Description,
		Status:          req.Status,
		Sort:            req.Sort,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除会员等级；仍有会员处于该等级时拒绝删除。
// 入参：路径 id；返回 null。
func (c *MemberLevelController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
