package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payErrs "eqray-flavor/app/modules/pay/errs"
	"eqray-flavor/app/modules/pay/model"
	"eqray-flavor/app/modules/pay/repository"
)

const (
	// payWechatAckSuccess 微信支付 v3 回调成功应答
	payWechatAckSuccess = `{"code":"SUCCESS","message":"成功"}`
	// payAlipayAckSuccess 支付宝异步通知成功应答，必须是小写 success
	payAlipayAckSuccess = "success"
	// payAlipayAckFail 支付宝异步通知失败应答，返回后支付宝会按策略重试
	payAlipayAckFail = "fail"
)

// PayCallbackAck 是回调处理结果中要回给渠道的应答。
type PayCallbackAck struct {
	// ContentType 应答 Content-Type
	ContentType string
	// Body 应答正文
	Body []byte
}

// PayNotifyService 负责渠道异步回调的验签、幂等推进与应答，对齐 yudao PayNotifyService 的入口部分。
// 只推进支付单状态，业务订单状态由业务侧感知后自行处理。
type PayNotifyService struct {
	orders     *repository.PayOrderRepository
	extensions *repository.PayOrderExtensionRepository
	channels   *repository.PayChannelRepository
	refunds    *repository.PayRefundRepository
	refundSvc  *PayRefundService
	// notifyTasks 支付成功后登记业务通知任务
	notifyTasks *PayNotifyTaskService
	// walletSvc 充值单支付成功后入账
	walletSvc *PayWalletService
}

// NewPayNotifyService 构造支付回调服务。
func NewPayNotifyService() *PayNotifyService {
	return &PayNotifyService{
		orders:      repository.NewPayOrderRepository(),
		extensions:  repository.NewPayOrderExtensionRepository(),
		channels:    repository.NewPayChannelRepository(),
		refunds:     repository.NewPayRefundRepository(),
		refundSvc:   NewPayRefundService(),
		notifyTasks: NewPayNotifyTaskService(),
		walletSvc:   NewPayWalletService(),
	}
}

// HandleOrderCallback 处理渠道支付回调：验签解密 → 校验金额 → 幂等置为已支付。
// 入参：渠道主键、原始请求体与请求头；返回给渠道的应答（失败应答会触发渠道重试）。
// 约束：同一笔回调重复推送时只有第一次会更新支付单，其余直接返回成功应答。
func (s *PayNotifyService) HandleOrderCallback(ctx context.Context, channelID types.ID, body []byte, headers map[string]string) (*PayCallbackAck, error) {
	channel, err := s.channels.FindByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, baseErrs.New(payErrs.PayChannelNotExists)
	}

	gateway, err := NewChannelGateway(channel)
	if err != nil {
		return payOrderAck(channel.Code, false, trimMessage(err.Error())), err
	}

	notify, err := gateway.ParseNotify(ctx, body, headers, nil)
	if err != nil {
		return payOrderAck(channel.Code, false, trimMessage(err.Error())), err
	}
	// 退款回调走另一个地址，避免两类回调互相覆盖状态
	if notify.RefundNo != "" {
		failure := baseErrs.WithMsg(baseErrs.BadRequest, "收到退款回调，请检查渠道 notify_url 配置")

		return payOrderAck(channel.Code, false, failure.Msg()), failure
	}

	order, err := s.orders.FindByNo(ctx, notify.OutTradeNo)
	if err != nil {
		return payOrderAck(channel.Code, false, "支付单查询失败"), err
	}
	if order == nil {
		failure := baseErrs.New(payErrs.PayOrderNotExists)

		return payOrderAck(channel.Code, false, failure.Msg()), failure
	}
	// 渠道与应用一一对应，跨应用的支付单不允许被该渠道推进
	if order.AppID != channel.AppID {
		failure := baseErrs.WithMsg(baseErrs.BadRequest, "回调渠道与支付单所属应用不一致")

		return payOrderAck(channel.Code, false, failure.Msg()), failure
	}
	if notify.Price != order.Price {
		failure := baseErrs.New(payErrs.PayOrderNotifyPriceMismatch, notify.Price, order.Price)

		return payOrderAck(channel.Code, false, "回调金额与支付单金额不一致"), failure
	}
	// 非成功状态（如支付中、已关闭）不需要改单，直接应答成功避免渠道反复推送
	if !notify.Success {
		return payOrderAck(channel.Code, true, ""), nil
	}
	// 已支付或已退款的单说明回调重复推送，幂等返回成功
	if order.Status != payEnums.PayOrderStatusWaiting {
		return payOrderAck(channel.Code, true, ""), nil
	}

	extension, err := s.latestExtension(ctx, order.ID)
	if err != nil {
		return payOrderAck(channel.Code, false, "渠道调用记录查询失败"), err
	}

	affected, err := s.markOrderPaid(ctx, order, extension, notify)
	if err != nil {
		return payOrderAck(channel.Code, false, "支付单更新失败"), err
	}
	// 并发回调里另一次已改单成功，本次同样按成功应答
	if affected == 0 {
		return payOrderAck(channel.Code, true, ""), nil
	}

	return payOrderAck(channel.Code, true, ""), nil
}

// markOrderPaid 在事务内把支付单置为已支付，并回写拓展单的状态与回调原文。
// 入参：支付单、拓展单（可为空）与回调结果；返回影响行数，0 表示状态已被其它回调推进。
func (s *PayNotifyService) markOrderPaid(ctx context.Context, order *model.PayOrderModel, extension *model.PayOrderExtensionModel, notify *NotifyResult) (int64, error) {
	var (
		affected    int64
		extensionID types.ID
	)
	if extension != nil {
		extensionID = extension.ID
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		rows, err := s.orders.MarkPaid(ctx, tx, order.ID, extensionID, notify.ChannelOrderNo, notify.ChannelUserID, types.Now())
		if err != nil {
			return err
		}
		affected = rows
		if rows == 0 {
			return nil
		}

		// 充值单支付成功后同步入账，非充值单在内部直接跳过
		if err := s.walletSvc.HandleRechargePaid(ctx, tx, order); err != nil {
			return err
		}

		// 支付成功即登记业务通知任务，应用未配置回调地址时内部跳过
		if err := s.notifyTasks.CreateOrderTask(ctx, tx, order); err != nil {
			return err
		}
		if extension == nil {
			return nil
		}

		extension.Status = payEnums.PayOrderStatusSuccess
		extension.ChannelNotifyData = notify.RawData

		return s.extensions.UpdateResult(ctx, tx, extension)
	}); err != nil {
		return 0, err
	}

	return affected, nil
}

// HandleRefundCallback 处理渠道退款回调：验签解密 → 幂等把退款单置为成功并累计支付单已退金额。
// 入参：渠道主键、原始请求体与请求头；返回给渠道的应答（失败应答会触发渠道重试）。
func (s *PayNotifyService) HandleRefundCallback(ctx context.Context, channelID types.ID, body []byte, headers map[string]string) (*PayCallbackAck, error) {
	channel, err := s.channels.FindByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, baseErrs.New(payErrs.PayChannelNotExists)
	}

	gateway, err := NewChannelGateway(channel)
	if err != nil {
		return payOrderAck(channel.Code, false, trimMessage(err.Error())), err
	}

	notify, err := gateway.ParseNotify(ctx, body, headers, nil)
	if err != nil {
		return payOrderAck(channel.Code, false, trimMessage(err.Error())), err
	}
	// 退款回调必须带商户退款单号，否则说明通知打错了地址
	if notify.RefundNo == "" {
		failure := baseErrs.WithMsg(baseErrs.BadRequest, "不是退款回调")

		return payOrderAck(channel.Code, false, failure.Msg()), failure
	}

	refund, err := s.refunds.FindByNo(ctx, notify.RefundNo)
	if err != nil {
		return payOrderAck(channel.Code, false, "退款单查询失败"), err
	}
	if refund == nil {
		failure := baseErrs.New(payErrs.PayRefundNotExists)

		return payOrderAck(channel.Code, false, failure.Msg()), failure
	}
	if refund.AppID != channel.AppID {
		failure := baseErrs.WithMsg(baseErrs.BadRequest, "回调渠道与退款单所属应用不一致")

		return payOrderAck(channel.Code, false, failure.Msg()), failure
	}
	// 退款关闭 / 异常等非成功状态不改单，直接应答成功避免渠道反复推送
	if !notify.RefundSuccess {
		return payOrderAck(channel.Code, true, ""), nil
	}
	// 已成功的退款单说明回调重复推送，幂等返回成功
	if refund.Status != payEnums.PayRefundStatusCreate {
		return payOrderAck(channel.Code, true, ""), nil
	}
	if notify.RefundPrice != 0 && notify.RefundPrice != refund.RefundPrice {
		failure := baseErrs.WithMsg(baseErrs.BadRequest, "回调退款金额与退款单金额不一致")

		return payOrderAck(channel.Code, false, failure.Msg()), failure
	}

	// 渠道退款单号在发起退款时已落库，这里只用回调原文与成功时间，不覆盖已有单号
	if _, err := s.refundSvc.ConfirmRefund(ctx, refund, "", notify.RawData); err != nil {
		return payOrderAck(channel.Code, false, "退款单更新失败"), err
	}

	return payOrderAck(channel.Code, true, ""), nil
}

// latestExtension 取支付单最近一次渠道调用记录，回调时用于回写渠道订单号与原文。
// 入参：支付单主键；返回最近一条拓展单，没有记录时返回 (nil, nil)。
func (s *PayNotifyService) latestExtension(ctx context.Context, orderID types.ID) (*model.PayOrderExtensionModel, error) {
	list, err := s.extensions.ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}

	return &list[len(list)-1], nil
}

// payOrderAck 按渠道生成回调应答：微信返回 JSON，支付宝返回纯文本。
// 入参：渠道编码、是否成功与失败原因；返回应答的 Content-Type 与正文。
func payOrderAck(channelCode string, success bool, message string) *PayCallbackAck {
	if payEnums.IsWechatChannel(channelCode) {
		if success {
			return &PayCallbackAck{ContentType: "application/json", Body: []byte(payWechatAckSuccess)}
		}

		return &PayCallbackAck{ContentType: "application/json", Body: []byte(failureAckJSON(message))}
	}

	if success {
		return &PayCallbackAck{ContentType: "text/plain", Body: []byte(payAlipayAckSuccess)}
	}

	return &PayCallbackAck{ContentType: "text/plain", Body: []byte(payAlipayAckFail)}
}

// failureAckJSON 生成微信失败应答，消息里的引号与换行必须转义，否则渠道无法解析。
// 入参：失败原因；返回 JSON 应答正文。
func failureAckJSON(message string) string {
	if message == "" {
		message = "处理失败"
	}

	payload, err := json.Marshal(map[string]string{"code": "FAIL", "message": message})
	if err != nil {
		return `{"code":"FAIL","message":"处理失败"}`
	}

	return string(payload)
}

// NormalizeCallbackHeaders 把框架的请求头转成忽略大小写的键值对，供渠道验签使用。
// 入参：原始请求头；返回同名小写的映射。
func NormalizeCallbackHeaders(headers map[string][]string) map[string]string {
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		result[strings.TrimSpace(key)] = values[0]
	}

	return result
}
