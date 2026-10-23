// Package controller 接收支付模块推送的业务通知。通知不带会员令牌，
// 安全由 pay 侧的签名校验与本侧的金额核对保证。
package controller

import (
	"io"

	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/order/service"
	payService "eqray-flavor/app/modules/pay/service"
)

const (
	// orderCallbackMaxBodyBytes 通知报文大小上限，正常报文远小于该值，超出直接拒绝
	orderCallbackMaxBodyBytes = 1 << 20
	// orderCallbackContentType 通知应答的文本类型
	orderCallbackContentType = "text/plain; charset=utf-8"
)

// OrderCallbackController 提供支付结果通知入口。
type OrderCallbackController struct {
	service *service.OrderService
}

// NewOrderCallbackController 构造订单回调控制器。
func NewOrderCallbackController() *OrderCallbackController {
	return &OrderCallbackController{service: service.NewOrderService()}
}

// Pay 处理支付结果通知：验签 → 金额核对 → 幂等推进订单状态。
// 入参：原始请求体与请求头；返回业务侧应答文本（success / error）。
// 说明：pay 侧把非 success 的应答视为失败并按重试策略继续推送，因此本方法必须保持幂等。
func (c *OrderCallbackController) Pay(ctx http.Context) http.Response {
	body, err := io.ReadAll(io.LimitReader(ctx.Request().Origin().Body, orderCallbackMaxBodyBytes))
	if err != nil {
		result.MarkFail(ctx, baseErrs.BadRequest)

		return ctx.Response().Data(http.StatusBadRequest, orderCallbackContentType, []byte("error"))
	}

	ack, handleErr := c.service.HandlePayNotify(ctx.Context(), payService.NormalizeCallbackHeaders(ctx.Request().Headers()), body)
	if handleErr != nil {
		// 通知失败要进错误日志，但应答体必须是纯文本 success / error，不能用统一响应体
		result.MarkFail(ctx, baseErrs.InternalServerError)
	}

	return ctx.Response().Data(http.StatusOK, orderCallbackContentType, []byte(ack))
}
