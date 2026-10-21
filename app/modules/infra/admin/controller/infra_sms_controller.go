package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/infra/admin/request"
	"eqray-flavor/app/modules/infra/service"
)

// InfraSmsChannelController 提供短信渠道配置的增删改查。
type InfraSmsChannelController struct {
	service *service.InfraSmsChannelService
}

func NewInfraSmsChannelController() *InfraSmsChannelController {
	return &InfraSmsChannelController{service: service.NewInfraSmsChannelService()}
}

// Index 分页查询短信渠道，支持关键字、渠道编码与状态筛选，列表不返回 apiSecret。
// 入参：keyword、code、status、pageNo、pageSize（查询串）；返回渠道列表与总数。
func (c *InfraSmsChannelController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.InfraSmsChannelListQuery{
		Keyword:  ctx.Request().Query("keyword"),
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

// Options 查询启用中的短信渠道，供短信模板绑定渠道时下拉选择。
// 无入参；返回渠道列表。
func (c *InfraSmsChannelController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show 查询单个短信渠道详情，apiSecret 不返回。
// 入参：路径 id；返回渠道详情，不存在返回业务错误码。
func (c *InfraSmsChannelController) Show(ctx http.Context) http.Response {
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

// Store 新增短信渠道；code 唯一，取值 DebugLog / AliyunSms / TencentSms。
// 入参：name、code、apiKey、apiSecret、signature、extra（JSON）、status、remark；返回新渠道 ID。
func (c *InfraSmsChannelController) Store(ctx http.Context) http.Response {
	var req request.InfraSmsChannelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.InfraSmsChannelSaveInput{
		Name:      req.Name,
		Code:      req.Code,
		APIKey:    req.APIKey,
		APISecret: req.APISecret,
		Signature: req.Signature,
		Extra:     req.Extra,
		Status:    req.Status,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑短信渠道；apiSecret 留空表示不修改。
// 入参：路径 id + 渠道字段；返回 null。
func (c *InfraSmsChannelController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.InfraSmsChannelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.InfraSmsChannelSaveInput{
		Name:      req.Name,
		Code:      req.Code,
		APIKey:    req.APIKey,
		APISecret: req.APISecret,
		Signature: req.Signature,
		Extra:     req.Extra,
		Status:    req.Status,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除短信渠道；渠道下仍有模板时由服务层拒绝删除。
// 入参：路径 id；返回 null。
func (c *InfraSmsChannelController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// InfraSmsTemplateController 提供短信模板的增删改查。
type InfraSmsTemplateController struct {
	service *service.InfraSmsTemplateService
}

func NewInfraSmsTemplateController() *InfraSmsTemplateController {
	return &InfraSmsTemplateController{service: service.NewInfraSmsTemplateService()}
}

// Index 分页查询短信模板，支持关键字、绑定渠道、模板类型与状态筛选。
// 入参：keyword、channelId、type、status、pageNo、pageSize（查询串）；返回模板列表与总数。
func (c *InfraSmsTemplateController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.InfraSmsTemplateListQuery{
		Keyword:   ctx.Request().Query("keyword"),
		ChannelID: types.ID(ctx.Request().QueryInt("channelId", 0)),
		PageNo:    page.PageNo,
		PageSize:  page.PageSize,
	}
	if templateType := ctx.Request().QueryInt("type", 0); templateType > 0 {
		v := int8(templateType)
		q.Type = &v
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

// Show 查询单个短信模板详情。
// 入参：路径 id；返回模板详情，不存在返回业务错误码。
func (c *InfraSmsTemplateController) Show(ctx http.Context) http.Response {
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

// Store 新增短信模板并绑定发送渠道；模板编码唯一。
// 入参：channelId、code、name、type、apiTemplateId、content、params、status、remark；返回新模板 ID。
func (c *InfraSmsTemplateController) Store(ctx http.Context) http.Response {
	var req request.InfraSmsTemplateSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.InfraSmsTemplateSaveInput{
		ChannelID:     req.ChannelID,
		Code:          req.Code,
		Name:          req.Name,
		Type:          req.Type,
		APITemplateID: req.APITemplateID,
		Content:       req.Content,
		Params:        req.Params,
		Status:        req.Status,
		Remark:        req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑短信模板内容、变量与绑定渠道。
// 入参：路径 id + 模板字段；返回 null。
func (c *InfraSmsTemplateController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.InfraSmsTemplateSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.InfraSmsTemplateSaveInput{
		ChannelID:     req.ChannelID,
		Code:          req.Code,
		Name:          req.Name,
		Type:          req.Type,
		APITemplateID: req.APITemplateID,
		Content:       req.Content,
		Params:        req.Params,
		Status:        req.Status,
		Remark:        req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除短信模板（软删）；删除后对应场景的验证码将无法发送。
// 入参：路径 id；返回 null。
func (c *InfraSmsTemplateController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
