package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/member/admin/request"
	"eqray-flavor/app/modules/member/service"
)

// MemberSocialClientController 提供第三方登录渠道配置的增删改查。
type MemberSocialClientController struct {
	service *service.MemberSocialClientService
}

// NewMemberSocialClientController 构造第三方登录渠道配置控制器。
func NewMemberSocialClientController() *MemberSocialClientController {
	return &MemberSocialClientController{service: service.NewMemberSocialClientService()}
}

// Index 分页查询第三方登录渠道配置，列表不返回 client_secret。
// 入参：keyword、status、pageNo、pageSize（查询串）；返回配置列表与总数。
func (c *MemberSocialClientController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.MemberSocialClientListQuery{
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

// Show 查询单个渠道配置详情，client_secret 不返回。
// 入参：路径 id；返回配置详情。
func (c *MemberSocialClientController) Show(ctx http.Context) http.Response {
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

// Store 新增第三方登录渠道配置；同一 socialType 只能有一条。
// 入参：name、socialType、clientId、clientSecret、agentId、status；返回新配置 ID。
func (c *MemberSocialClientController) Store(ctx http.Context) http.Response {
	var req request.MemberSocialClientSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.MemberSocialClientSaveInput{
		Name:         req.Name,
		SocialType:   req.SocialType,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		AgentID:      req.AgentID,
		Status:       req.Status,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑渠道配置；clientSecret 留空表示不修改。
// 入参：路径 id + 配置字段；返回 null。
func (c *MemberSocialClientController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberSocialClientSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.MemberSocialClientSaveInput{
		Name:         req.Name,
		SocialType:   req.SocialType,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		AgentID:      req.AgentID,
		Status:       req.Status,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除第三方登录渠道配置，删除后该渠道无法再登录。
// 入参：路径 id；返回 null。
func (c *MemberSocialClientController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
