package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/model"
	"eqray-flavor/app/modules/pay/repository"
)

const (
	// payNotifyBatchSize 单次调度最多处理的任务数，避免一次抢占过多连接
	payNotifyBatchSize = 50
	// payNotifyResponseMaxLength 日志里保存的响应内容上限，防止业务方返回大报文撑爆表
	payNotifyResponseMaxLength = 512
	// payNotifySuccessText 业务方约定：处理成功时响应体返回 success
	payNotifySuccessText = "success"
	// payNotifyHTTPTimeout 单次通知的超时时间
	payNotifyHTTPTimeout = 10 * time.Second
)

// payNotifyBackoff 是 yudao PayNotifyTaskDO.NOTIFY_FREQUENCY 的等价实现，单位为秒：
// 15、15、30、180、1800、1800、1800、3600，算上首次通知一共 9 次。
// 最后一项对后续次数一直生效；次数用尽后任务转最终失败，不再被调度捞出。
var payNotifyBackoff = []time.Duration{
	15 * time.Second,
	15 * time.Second,
	30 * time.Second,
	3 * time.Minute,
	30 * time.Minute,
	30 * time.Minute,
	30 * time.Minute,
	time.Hour,
}

// payNotifyMaxTimes 单条任务的最大通知次数，与 yudao 一致：重试档位数 + 首次 = 9。
var payNotifyMaxTimes = int8(len(payNotifyBackoff) + 1)

// PayNotifyTaskRow 是通知任务列表输出，补上类型与状态文案。
type PayNotifyTaskRow struct {
	model.PayNotifyTaskModel
	// TypeName 类型文案，如「支付单」
	TypeName string `json:"typeName"`
	// StatusName 状态文案，如「等待通知」
	StatusName string `json:"statusName"`
}

// PayNotifyTaskService 负责向业务方派发支付 / 退款结果通知，并在失败后按间隔重试。
// 约定：业务方处理成功后响应体返回纯文本 success；签名算法见 signPayload。
type PayNotifyTaskService struct {
	notify  *repository.PayNotifyRepository
	apps    *repository.PayAppRepository
	orders  *repository.PayOrderRepository
	refunds *repository.PayRefundRepository
}

// NewPayNotifyTaskService 构造通知任务服务。
func NewPayNotifyTaskService() *PayNotifyTaskService {
	return &PayNotifyTaskService{
		notify:  repository.NewPayNotifyRepository(),
		apps:    repository.NewPayAppRepository(),
		orders:  repository.NewPayOrderRepository(),
		refunds: repository.NewPayRefundRepository(),
	}
}

// Page 分页查询通知任务，支持应用、类型、状态与商户订单号关键字筛选。
// 入参：筛选条件与分页；返回补好文案的任务列表与总数。
func (s *PayNotifyTaskService) Page(ctx context.Context, filter repository.PayNotifyTaskFilter, pageNo, pageSize int) ([]PayNotifyTaskRow, int64, error) {
	list, total, err := s.notify.PaginateTasks(ctx, filter, pageNo, pageSize)
	if err != nil {
		return nil, 0, err
	}

	rows := make([]PayNotifyTaskRow, 0, len(list))
	for _, item := range list {
		rows = append(rows, PayNotifyTaskRow{
			PayNotifyTaskModel: item,
			TypeName:           item.Type.String(),
			StatusName:         item.Status.String(),
		})
	}

	return rows, total, nil
}

// Logs 分页查询某条任务的通知日志，用于排查业务方是否收到、响应了什么。
// 入参：任务 ID 与分页；返回日志列表与总数。
func (s *PayNotifyTaskService) Logs(ctx context.Context, taskID types.ID, pageNo, pageSize int) ([]model.PayNotifyLogModel, int64, error) {
	return s.notify.PaginateLogs(ctx, taskID, pageNo, pageSize)
}

// CreateOrderTask 在支付成功的事务里登记一条支付通知任务；应用未配置回调地址时不登记。
// 入参：事务、支付单；返回错误表示落库失败（会导致支付回调整体重试）。
func (s *PayNotifyTaskService) CreateOrderTask(ctx context.Context, tx db.Tx, order *model.PayOrderModel) error {
	app, err := s.apps.FindByID(ctx, order.AppID)
	if err != nil {
		return err
	}
	if app == nil || strings.TrimSpace(app.OrderNotifyURL) == "" {
		return nil
	}

	return s.notify.CreateTask(ctx, tx, &model.PayNotifyTaskModel{
		AppID:           order.AppID,
		Type:            payEnums.PayNotifyTypeOrder,
		DataID:          order.ID,
		MerchantOrderID: order.MerchantOrderID,
		NotifyURL:       app.OrderNotifyURL,
		Status:          payEnums.PayNotifyStatusWaiting,
		NextNotifyTime:  types.DateTime(time.Now()),
		MaxNotifyTimes:  payNotifyMaxTimes,
	})
}

// CreateRefundTask 在退款成功的事务里登记一条退款通知任务；应用未配置回调地址时不登记。
// 入参：事务、退款单；返回错误表示落库失败（会导致退款回调整体重试）。
func (s *PayNotifyTaskService) CreateRefundTask(ctx context.Context, tx db.Tx, refund *model.PayRefundModel) error {
	app, err := s.apps.FindByID(ctx, refund.AppID)
	if err != nil {
		return err
	}
	if app == nil || strings.TrimSpace(app.RefundNotifyURL) == "" {
		return nil
	}

	return s.notify.CreateTask(ctx, tx, &model.PayNotifyTaskModel{
		AppID:           refund.AppID,
		Type:            payEnums.PayNotifyTypeRefund,
		DataID:          refund.ID,
		MerchantOrderID: refund.MerchantOrderID,
		NotifyURL:       app.RefundNotifyURL,
		Status:          payEnums.PayNotifyStatusWaiting,
		NextNotifyTime:  types.DateTime(time.Now()),
		MaxNotifyTimes:  payNotifyMaxTimes,
	})
}

// ExecuteDueTasks 扫描到点的通知任务并逐条派发，供定时任务调用。
// 入参：上下文；返回本次处理的任务数与首个落库错误（单条通知失败只记录日志，不中断后续任务）。
func (s *PayNotifyTaskService) ExecuteDueTasks(ctx context.Context) (int, error) {
	tasks, err := s.notify.ListDueTasks(ctx, types.Now(), payNotifyBatchSize)
	if err != nil {
		return 0, err
	}

	handled := 0
	var firstErr error
	for index := range tasks {
		if err := s.dispatch(ctx, &tasks[index]); err != nil && firstErr == nil {
			firstErr = err
		}
		handled++
	}

	return handled, firstErr
}

// dispatch 组装并发送一条通知，随后写通知日志并回写任务状态与下次通知时间。
// 入参：任务；返回错误仅表示本地落库异常，通知失败本身记录在任务与日志里。
func (s *PayNotifyTaskService) dispatch(ctx context.Context, task *model.PayNotifyTaskModel) error {
	app, err := s.apps.FindByID(ctx, task.AppID)
	if err != nil {
		return err
	}

	payload, buildErr := s.buildPayload(ctx, task, app)
	task.NotifyTimes++
	task.LastExecuteTime = types.Now()

	// 状态口径对齐 yudao：请求发不出去记「请求失败」，发出去但业务方不认账记
	// 「请求成功但结果失败」，两者都重试；只有业务方返回 success 才是终态成功。
	status := payEnums.PayNotifyStatusSuccess
	response := ""
	if buildErr != nil {
		status = payEnums.PayNotifyStatusRequestFail
		response = "组装通知内容失败：" + trimMessage(buildErr.Error())
	} else {
		var body string
		status, body = s.post(task.NotifyURL, app, payload)
		response = body
	}

	task.Status = status
	switch {
	case status == payEnums.PayNotifyStatusSuccess:
		task.NextNotifyTime = types.Now()
	case task.NotifyTimes >= task.MaxNotifyTimes:
		// 次数用尽转最终失败，避免调度每轮都把它捞出来重试
		task.Status = payEnums.PayNotifyStatusFail
		task.NextNotifyTime = types.Now()
	default:
		task.NextNotifyTime = types.DateTime(time.Now().Add(backoffDuration(task.NotifyTimes)))
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.notify.UpdateTaskResult(ctx, tx, task); err != nil {
			return err
		}

		return s.notify.CreateLog(ctx, tx, &model.PayNotifyLogModel{
			TaskID:      task.ID,
			NotifyTimes: task.NotifyTimes,
			Response:    trimTo(response, payNotifyResponseMaxLength),
			Status:      task.Status,
		})
	})
}

// buildPayload 按任务类型组装通知业务方的 JSON。
// 入参：任务与应用；返回 JSON 字符串，业务数据缺失时返回错误。
func (s *PayNotifyTaskService) buildPayload(ctx context.Context, task *model.PayNotifyTaskModel, app *model.PayAppModel) (string, error) {
	if app == nil {
		return "", errors.New("支付应用不存在")
	}

	payload := map[string]any{
		"type":            notifyTypeText(task.Type),
		"appKey":          app.AppKey,
		"merchantOrderId": task.MerchantOrderID,
	}

	switch task.Type {
	case payEnums.PayNotifyTypeOrder:
		order, err := s.orders.FindByID(ctx, task.DataID)
		if err != nil {
			return "", err
		}
		if order == nil {
			return "", errors.New("支付单不存在")
		}
		payload["orderNo"] = order.No
		payload["channelCode"] = order.ChannelCode
		payload["channelOrderNo"] = order.ChannelOrderNo
		payload["price"] = order.Price
		payload["refundPrice"] = order.RefundPrice
		payload["successTime"] = order.SuccessTime

	case payEnums.PayNotifyTypeRefund:
		refund, err := s.refunds.FindByID(ctx, task.DataID)
		if err != nil {
			return "", err
		}
		if refund == nil {
			return "", errors.New("退款单不存在")
		}
		payload["refundNo"] = refund.No
		payload["orderNo"] = refund.OrderNo
		payload["channelCode"] = refund.ChannelCode
		payload["refundPrice"] = refund.RefundPrice
		payload["successTime"] = refund.SuccessTime

	default:
		return "", errors.New("暂不支持的通知类型")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// post 带签名请求业务方回调地址，返回通知状态与响应内容。
// 判定口径与 yudao 一致：请求发不出去（密钥缺失 / 网络异常 / HTTP 非 2xx）算「请求失败」；
// 请求成功但响应体不是 success 算「请求成功但结果失败」；响应 success 才算「通知成功」。
func (s *PayNotifyTaskService) post(notifyURL string, app *model.PayAppModel, payload string) (payEnums.PayNotifyStatusEnum, string) {
	if app == nil || app.AppSecret == "" || notifyURL == "" {
		return payEnums.PayNotifyStatusRequestFail, "应用密钥或回调地址缺失"
	}

	timestamp := time.Now().Unix()
	nonce := support.NewID().String()
	headers := map[string]string{
		"Content-Type":    "application/json",
		"X-Pay-AppKey":    app.AppKey,
		"X-Pay-Timestamp": strconv.FormatInt(timestamp, 10),
		"X-Pay-Nonce":     nonce,
		"X-Pay-Sign":      signPayload(app.AppSecret, app.AppKey, timestamp, nonce, payload),
	}

	ctx, cancel := context.WithTimeout(context.Background(), payNotifyHTTPTimeout)
	defer cancel()

	response, err := facades.Http().WithContext(ctx).
		WithHeaders(headers).
		Post(notifyURL, strings.NewReader(payload))
	if err != nil {
		return payEnums.PayNotifyStatusRequestFail, trimMessage("请求失败：" + err.Error())
	}

	body, err := response.Body()
	if err != nil {
		return payEnums.PayNotifyStatusRequestFail, "读取响应失败"
	}
	if !response.Successful() {
		return payEnums.PayNotifyStatusRequestFail, trimTo(body, payNotifyResponseMaxLength)
	}
	if strings.EqualFold(strings.TrimSpace(body), payNotifySuccessText) {
		return payEnums.PayNotifyStatusSuccess, trimTo(body, payNotifyResponseMaxLength)
	}

	return payEnums.PayNotifyStatusRequestSuccess, trimTo(body, payNotifyResponseMaxLength)
}

// signPayload 计算回调签名：HMAC-SHA256(appSecret, appKey + "\n" + timestamp + "\n" + nonce + "\n" + body)，输出小写十六进制。
// 入参：应用密钥、应用标识、秒级时间戳、随机串与请求体；返回签名字符串。
func signPayload(secret, appKey string, timestamp int64, nonce, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(appKey))
	mac.Write([]byte("\n"))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("\n"))
	mac.Write([]byte(nonce))
	mac.Write([]byte("\n"))
	mac.Write([]byte(payload))

	return hex.EncodeToString(mac.Sum(nil))
}

// backoffDuration 按已通知次数取下一次重试间隔，超过预设档位后沿用最后一档。
// 入参：已通知次数（从 1 开始）；返回等待时长。
func backoffDuration(notifyTimes int8) time.Duration {
	index := int(notifyTimes) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(payNotifyBackoff) {
		index = len(payNotifyBackoff) - 1
	}

	return payNotifyBackoff[index]
}

// notifyTypeText 把通知类型转成业务方可读的类型名。
// 入参：通知类型；返回 order / refund。
func notifyTypeText(notifyType payEnums.PayNotifyTypeEnum) string {
	switch notifyType {
	case payEnums.PayNotifyTypeOrder:
		return "order"
	case payEnums.PayNotifyTypeRefund:
		return "refund"
	default:
		return "unknown"
	}
}

// trimTo 截断字符串到指定字符数，用于日志与响应落库。
// 入参：原文与最大字符数；返回截断后的文本。
func trimTo(text string, maxLength int) string {
	if len([]rune(text)) <= maxLength {
		return text
	}

	return string([]rune(text)[:maxLength])
}
