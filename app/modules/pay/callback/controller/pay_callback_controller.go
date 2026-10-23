// Package controller 接收渠道异步回调。回调不带后台令牌，安全由渠道签名与金额校验保证。
package controller

import (
	"io"
	"strconv"

	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/service"
)

const (
	// payCallbackMaxBodyBytes 回调报文大小上限，正常报文远小于该值，超出直接拒绝
	payCallbackMaxBodyBytes = 1 << 20
	// payCallbackJSONContentType 通知渠道应答的 JSON 类型
	payCallbackJSONContentType = "application/json"
)

// PayCallbackController 提供渠道支付 / 退款回调入口。
type PayCallbackController struct {
	service *service.PayNotifyService
}

// NewPayCallbackController 构造支付回调控制器。
func NewPayCallbackController() *PayCallbackController {
	return &PayCallbackController{service: service.NewPayNotifyService()}
}

// Order 处理渠道支付结果回调：验签解密后幂等推进支付单，并按渠道要求应答。
// 入参：路径 channelId（渠道主键）、原始报文与请求头；返回渠道约定的应答正文。
// 说明：应答正文必须是渠道规定格式，否则渠道会持续重推，因此失败场景也返回渠道格式而不是统一响应体。
func (c *PayCallbackController) Order(ctx http.Context) http.Response {
	channelID, err := parseChannelID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	body, headers, err := readCallback(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	ack, handleErr := c.service.HandleOrderCallback(ctx.Context(), channelID, body, headers)
	if handleErr != nil {
		// 回调失败要进错误日志：先标记失败业务码，再返回渠道格式的失败应答
		result.MarkFail(ctx, baseErrs.InternalServerError)
	}
	if ack == nil {
		return ctx.Response().Data(http.StatusNotFound, payCallbackJSONContentType,
			[]byte(`{"code":"FAIL","message":"支付渠道不存在"}`))
	}

	return ctx.Response().Data(http.StatusOK, ack.ContentType, ack.Body)
}

// Refund 处理渠道退款结果回调：验签解密后幂等推进退款单，并按渠道要求应答。
// 入参：路径 channelId（渠道主键）、原始报文与请求头；返回渠道约定的应答正文。
func (c *PayCallbackController) Refund(ctx http.Context) http.Response {
	channelID, err := parseChannelID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	body, headers, err := readCallback(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	ack, handleErr := c.service.HandleRefundCallback(ctx.Context(), channelID, body, headers)
	if handleErr != nil {
		result.MarkFail(ctx, baseErrs.InternalServerError)
	}
	if ack == nil {
		return ctx.Response().Data(http.StatusNotFound, payCallbackJSONContentType,
			[]byte(`{"code":"FAIL","message":"支付渠道不存在"}`))
	}

	return ctx.Response().Data(http.StatusOK, ack.ContentType, ack.Body)
}

// readCallback 读取回调原始报文与请求头，报文过大直接拒绝。
// 入参：上下文；返回报文、头映射与读取错误。
func readCallback(ctx http.Context) ([]byte, map[string]string, error) {
	body, err := io.ReadAll(io.LimitReader(ctx.Request().Origin().Body, payCallbackMaxBodyBytes))
	if err != nil {
		return nil, nil, baseErrs.WithMsg(baseErrs.BadRequest, "回调报文读取失败")
	}

	return body, service.NormalizeCallbackHeaders(ctx.Request().Headers()), nil
}

// parseChannelID 解析路径里的渠道主键，雪花 ID 必须按字符串转整数，避免精度丢失。
// 入参：上下文；返回渠道主键，非法时返回 400 业务错误。
func parseChannelID(ctx http.Context) (types.ID, error) {
	raw := ctx.Request().Route("channelId")
	if raw == "" {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "渠道 ID 不能为空")
	}
	// 这里不用 param.PathID：它固定读名为 id 的路径参数，复用会读到空值
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "渠道 ID 不合法")
	}

	return types.ID(parsed), nil
}
