package controller

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/admin/request"
	"eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/service"
)

// PayChannelController 提供支付渠道的增删改查与渠道编码下拉接口。
type PayChannelController struct {
	service *service.PayChannelService
}

// NewPayChannelController 构造支付渠道控制器。
func NewPayChannelController() *PayChannelController {
	return &PayChannelController{service: service.NewPayChannelService()}
}

// Index 分页查询支付渠道，支持应用、渠道编码与开启状态筛选。
// 入参：appId、code、status、pageNo、pageSize（查询串）；返回脱敏后的渠道列表与总数。
func (c *PayChannelController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	appID, err := queryID(ctx, "appId")
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.PayChannelListQuery{
		AppID: appID,
		Code:  ctx.Request().Query("code"),
	}
	// status 的合法值包含 0（开启），必须按参数是否出现判断
	if raw := ctx.Request().Query("status"); raw != "" {
		status := int8(ctx.Request().QueryInt("status", 0))
		q.Status = &status
	}

	list, total, err := c.service.Page(ctx.Context(), q, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Options 返回支持的渠道编码清单，供渠道配置页与代下单页面的下拉使用。
// 无入参；返回 code / name 列表。
func (c *PayChannelController) Options(ctx http.Context) http.Response {
	return result.Success(ctx, enums.PayChannelItems())
}

// Show 查询渠道详情，config 中的密钥只返回掩码。
// 入参：路径 id；返回渠道详情，不存在返回业务错误码。
func (c *PayChannelController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	row, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}

// Store 新增支付渠道，同一应用下渠道编码不可重复。
// 入参：渠道字段；返回新渠道主键。
func (c *PayChannelController) Store(ctx http.Context) http.Response {
	var req request.PayChannelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toPayChannelSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑支付渠道，应用与渠道编码创建后不可修改。
// 入参：路径 id + 渠道字段；返回 null。
func (c *PayChannelController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PayChannelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toPayChannelSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除支付渠道，历史支付单仍保留渠道快照。
// 入参：路径 id；返回 null。
func (c *PayChannelController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toPayChannelSaveInput 把请求体转换为服务层入参。
// 入参：已通过校验的请求体；返回服务层可直接使用的渠道字段。
func toPayChannelSaveInput(req request.PayChannelSaveRequest) service.PayChannelSaveInput {
	return service.PayChannelSaveInput{
		AppID:   req.AppID,
		Code:    req.Code,
		Status:  enums.PayEnableStatusEnum(req.Status),
		Remark:  req.Remark,
		FeeRate: req.FeeRate,
		Config:  req.Config,
	}
}

// queryID 解析查询串里的雪花 ID，未传返回 0，格式非法返回 400 业务错误。
// 入参：上下文与查询参数名；返回解析结果。
func queryID(ctx http.Context, key string) (types.ID, error) {
	raw := ctx.Request().Query(key)
	if raw == "" {
		return 0, nil
	}
	// 雪花 ID 超过 JS 安全整数，前端按字符串提交，这里同样用字符串解析
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < 0 {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "查询参数 "+key+" 不合法")
	}

	return types.ID(parsed), nil
}
