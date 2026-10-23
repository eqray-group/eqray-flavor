package service

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payErrs "eqray-flavor/app/modules/pay/errs"
	"eqray-flavor/app/modules/pay/model"
	"eqray-flavor/app/modules/pay/repository"
)

const (
	// payOrderSubjectMaxLength 商品标题长度上限，与 pay_order.subject 对齐
	payOrderSubjectMaxLength = 128
	// payOrderExpireMinutesDefault 支付单默认有效期（分钟）
	payOrderExpireMinutesDefault = 30
	// payOrderExpireMinutesMax 支付单有效期上限（分钟），避免长期未付的单一直占着商户订单号
	payOrderExpireMinutesMax = 24 * 60
	// payOrderChannelErrorMsgMax 拓展单里渠道报错信息的长度上限，与 pay_order_extension.channel_error_msg 对齐
	payOrderChannelErrorMsgMax = 512
)

// PayOrderListQuery 是后台支付单列表的筛选条件。
type PayOrderListQuery struct {
	// Keyword 支付单号 / 商户订单号 / 渠道订单号关键字
	Keyword string
	// AppID 应用编号，0 表示不限
	AppID types.ID
	// ChannelCode 渠道编码
	ChannelCode string
	// Status 支付状态，nil 表示不限
	Status *payEnums.PayOrderStatusEnum
}

// PayOrderCreateInput 是创建支付单并入渠道下单的入参。
// 金额由业务侧试算后传入（单位分），服务端不再接受前端直传的金额调整。
type PayOrderCreateInput struct {
	// AppID 应用编号
	AppID types.ID
	// ChannelCode 渠道编码，决定用哪个渠道配置下单
	ChannelCode string
	// UserID 付款用户编号
	UserID types.ID
	// UserType 用户类型，与会员模块对应
	UserType int8
	// MerchantOrderID 业务侧订单号，用于下单幂等
	MerchantOrderID string
	// Subject 商品标题
	Subject string
	// Body 商品描述
	Body string
	// Price 支付金额（分）
	Price int64
	// NotifyURL 业务回调地址，留空时取应用配置的支付回调地址
	NotifyURL string
	// OpenID 微信 JSAPI / 小程序支付必填
	OpenID string
	// AuthCode 支付宝条码支付必填
	AuthCode string
	// UserIP 用户 IP，渠道风控使用
	UserIP string
	// ExpireMinutes 支付单有效期（分钟），0 表示用默认 30 分钟
	ExpireMinutes int
}

// PayOrderSubmitResult 是下单成功后返回给前端的调起参数。
type PayOrderSubmitResult struct {
	// OrderID 支付单主键
	OrderID types.ID `json:"orderId"`
	// No 支付单号
	No string `json:"no"`
	// ChannelCode 渠道编码
	ChannelCode string `json:"channelCode"`
	// ChannelExtras 渠道额外参数 JSON，如二维码链接
	ChannelExtras string `json:"channelExtras"`
	// PayParams 前端调起支付所需参数，按渠道不同
	PayParams map[string]any `json:"payParams"`
}

// PayOrderRow 是支付单列表 / 详情输出，补上状态与渠道文案。
type PayOrderRow struct {
	model.PayOrderModel
	// ChannelName 渠道中文名，如「微信小程序」
	ChannelName string `json:"channelName"`
	// StatusName 状态文案，如「待支付」
	StatusName string `json:"statusName"`
}

// PayOrderDetail 是支付单详情，附带每次渠道调用的拓展记录。
type PayOrderDetail struct {
	PayOrderRow
	// Extensions 渠道调用历史，按主键升序
	Extensions []model.PayOrderExtensionModel `json:"extensions"`
}

// PayOrderService 负责支付单的创建、调起、查询与关闭，对齐 yudao PayOrderService。
// 只负责与渠道交互和支付单状态，业务订单状态由调用方根据回调推进。
type PayOrderService struct {
	orders     *repository.PayOrderRepository
	extensions *repository.PayOrderExtensionRepository
	channels   *PayChannelService
	// notifyTasks 补单成功后登记业务通知任务
	notifyTasks *PayNotifyTaskService
	// wallet 钱包余额渠道的支付实现：渠道编码为 wallet 时直接扣减余额
	wallet *PayWalletService
}

// NewPayOrderService 构造支付单服务。
func NewPayOrderService() *PayOrderService {
	return &PayOrderService{
		orders:      repository.NewPayOrderRepository(),
		extensions:  repository.NewPayOrderExtensionRepository(),
		channels:    NewPayChannelService(),
		notifyTasks: NewPayNotifyTaskService(),
		wallet:      NewPayWalletService(),
	}
}

// Page 分页查询支付单，支持关键字、应用、渠道与状态筛选。
// 入参：筛选条件与分页；返回补好文案的支付单列表与总数。
func (s *PayOrderService) Page(ctx context.Context, q PayOrderListQuery, pageNo, pageSize int) ([]PayOrderRow, int64, error) {
	filter := repository.PayOrderFilter{
		Keyword:     strings.TrimSpace(q.Keyword),
		AppID:       q.AppID,
		ChannelCode: strings.TrimSpace(q.ChannelCode),
	}
	if q.Status != nil {
		status := int8(*q.Status)
		filter.Status = &status
	}

	rows, total, err := s.orders.Paginate(ctx, filter, pageNo, pageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]PayOrderRow, 0, len(rows))
	for _, row := range rows {
		list = append(list, decoratePayOrder(row))
	}

	return list, total, nil
}

// Detail 查询支付单详情，并带出全部渠道调用记录。
// 入参：支付单主键；返回详情，不存在时报「支付单不存在」。
func (s *PayOrderService) Detail(ctx context.Context, id types.ID) (*PayOrderDetail, error) {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}

	extensions, err := s.extensions.ListByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &PayOrderDetail{PayOrderRow: decoratePayOrder(*m), Extensions: extensions}, nil
}

// Create 创建支付单并立即调用渠道下单。
// 入参：应用、渠道、商户订单号、金额（分）等；返回前端调起参数。
// 约束：同一应用下商户订单号已有未关闭的支付单时直接拒绝，保证下单幂等。
// 说明：渠道编码为 wallet 时是内置的钱包余额渠道，不查 pay_channel、不走渠道网关，
// 直接转交钱包服务扣减余额并落一张已支付的支付单，返回的调起参数为空。
func (s *PayOrderService) Create(ctx context.Context, in PayOrderCreateInput) (*PayOrderSubmitResult, error) {
	if err := normalizePayOrderInput(&in); err != nil {
		return nil, err
	}

	channel, app, err := s.channels.ResolveEnabled(ctx, in.AppID, in.ChannelCode)
	if err != nil {
		return nil, err
	}

	exists, err := s.orders.FindByMerchantOrderID(ctx, in.AppID, in.MerchantOrderID)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, baseErrs.New(payErrs.PayOrderMerchantOrderIDDuplicate, in.MerchantOrderID)
	}

	// 钱包余额是内置渠道：不需要渠道回调地址，也不需要网关，扣款成功后支付单即已支付
	if channel.Code == payEnums.ChannelWallet {
		return s.submitToWallet(ctx, in, app)
	}

	// 渠道配置与回调地址有问题尽早失败，避免先落一张必然支付不了的支付单
	if _, err := NewChannelGateway(channel); err != nil {
		return nil, err
	}
	if channelNotifyURL(channel.ID) == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "未配置支付回调基础地址 PAY_NOTIFY_BASE_URL")
	}

	order := &model.PayOrderModel{
		AppID:           in.AppID,
		ChannelID:       channel.ID,
		ChannelCode:     channel.Code,
		UserID:          in.UserID,
		UserType:        in.UserType,
		MerchantOrderID: in.MerchantOrderID,
		Subject:         in.Subject,
		Body:            in.Body,
		NotifyURL:       resolveNotifyURL(in.NotifyURL, app),
		Price:           in.Price,
		ChannelFeeRate:  channel.FeeRate,
		ChannelFeePrice: calcChannelFeePrice(in.Price, channel.FeeRate),
		Status:          payEnums.PayOrderStatusWaiting,
		UserIP:          in.UserIP,
		ExpireTime:      types.DateTime(time.Now().Add(time.Duration(in.ExpireMinutes) * time.Minute)),
		No:              newPayOrderNo(),
	}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.orders.Create(ctx, tx, order)
	}); err != nil {
		return nil, err
	}

	return s.submitToChannel(ctx, order, channel, in.OpenID, in.AuthCode)
}

// submitToWallet 走钱包余额完成支付：直接落一张已支付的支付单并扣减余额。
// 入参：下单入参与已校验的支付应用；返回支付单号与支付后余额，余额不足时整体回滚。
// 说明：调起参数为空，前端拿到结果即可认为支付完成，无需再拉起任何收银台。
func (s *PayOrderService) submitToWallet(ctx context.Context, in PayOrderCreateInput, app *model.PayAppModel) (*PayOrderSubmitResult, error) {
	if in.UserID == 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "钱包余额支付必须指定付款用户")
	}

	paid, err := s.wallet.PayByBalance(ctx, PayWalletPayInput{
		AppID:           in.AppID,
		UserID:          in.UserID,
		UserType:        in.UserType,
		MerchantOrderID: in.MerchantOrderID,
		Subject:         in.Subject,
		Body:            in.Body,
		Price:           in.Price,
		NotifyURL:       resolveNotifyURL(in.NotifyURL, app),
		UserIP:          in.UserIP,
	})
	if err != nil {
		return nil, err
	}

	return &PayOrderSubmitResult{
		OrderID:     paid.OrderID,
		No:          paid.No,
		ChannelCode: payEnums.ChannelWallet,
		PayParams: map[string]any{
			"balance": paid.Balance,
			"status":  paid.Status,
		},
	}, nil
}

// ReSubmit 对仍待支付且未过期的支付单重新调起渠道。
// 入参：支付单主键、openid 与付款码；返回新的调起参数，每次调用都会新增一条拓展记录。
func (s *PayOrderService) ReSubmit(ctx context.Context, id types.ID, openID, authCode string) (*PayOrderSubmitResult, error) {
	order, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.Status != payEnums.PayOrderStatusWaiting {
		return nil, baseErrs.New(payErrs.PayOrderStatusNotWaiting, order.No)
	}
	if !order.ExpireTime.Time().After(time.Now()) {
		return nil, baseErrs.New(payErrs.PayOrderExpired, order.No)
	}

	channel, err := s.channels.EnabledByID(ctx, order.ChannelID)
	if err != nil {
		return nil, err
	}

	return s.submitToChannel(ctx, order, channel, openID, authCode)
}

// Sync 主动查询渠道订单状态并补齐本地状态，用于回调丢失时的人工补单。
// 入参：支付单主键；返回最新支付单，非待支付状态直接返回本地数据。
func (s *PayOrderService) Sync(ctx context.Context, id types.ID) (*PayOrderRow, error) {
	order, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 只有待支付的单需要补单，其余状态以本地为准，避免覆盖退款结果
	if order.Status != payEnums.PayOrderStatusWaiting {
		row := decoratePayOrder(*order)

		return &row, nil
	}
	// 钱包余额支付没有渠道记录，也不可能停留在待支付，直接返回本地状态
	if order.ChannelCode == payEnums.ChannelWallet {
		row := decoratePayOrder(*order)

		return &row, nil
	}

	channel, err := s.channels.ByID(ctx, order.ChannelID)
	if err != nil {
		return nil, err
	}
	gateway, err := NewChannelGateway(channel)
	if err != nil {
		return nil, err
	}

	result, err := gateway.Query(ctx, QueryRequest{No: order.No})
	if err != nil {
		return nil, err
	}
	if result.Paid {
		if err := s.markPaidFromQuery(ctx, order, result); err != nil {
			return nil, err
		}

		reloaded, err := s.findByID(ctx, id)
		if err != nil {
			return nil, err
		}
		row := decoratePayOrder(*reloaded)

		return &row, nil
	}

	row := decoratePayOrder(*order)

	return &row, nil
}

// Close 关闭待支付的支付单：先查单确认渠道未收款，再关渠道订单，最后改本地状态。
// 入参：支付单主键；返回错误表示状态不允许关闭或渠道关单失败。
// 约束：渠道已收款时必须先补成已支付并报错，不能直接置关闭，否则用户付款会丢单。
func (s *PayOrderService) Close(ctx context.Context, id types.ID) error {
	order, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status == payEnums.PayOrderStatusClosed {
		return nil
	}
	if order.Status != payEnums.PayOrderStatusWaiting {
		return baseErrs.New(payErrs.PayOrderStatusNotWaiting, order.No)
	}

	channel, err := s.channels.ByID(ctx, order.ChannelID)
	if err != nil {
		return err
	}
	gateway, err := NewChannelGateway(channel)
	if err != nil {
		return err
	}

	// 关单前必须查单：查单失败就不关，避免把已收款的单置为关闭
	result, err := gateway.Query(ctx, QueryRequest{No: order.No})
	if err != nil {
		return err
	}
	if result.Paid {
		if err := s.markPaidFromQuery(ctx, order, result); err != nil {
			return err
		}

		return baseErrs.New(payErrs.PayOrderStatusNotWaiting, order.No)
	}

	if err := gateway.Close(ctx, QueryRequest{No: order.No}); err != nil {
		return err
	}

	var affected int64
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		rows, err := s.orders.MarkClosed(ctx, tx, id)
		affected = rows

		return err
	}); err != nil {
		return err
	}
	if affected == 0 {
		return baseErrs.New(payErrs.PayOrderStatusNotWaiting, order.No)
	}

	return nil
}

// CloseExpired 关闭已超过有效期仍待支付的支付单，供定时任务调用。
// 入参：单次最多处理条数；返回成功关闭数量与首个错误（单笔失败不影响其余单）。
func (s *PayOrderService) CloseExpired(ctx context.Context, limit uint64) (int, error) {
	list, err := s.orders.ListExpired(ctx, types.Now(), limit)
	if err != nil {
		return 0, err
	}

	closed := 0
	var firstErr error
	for index := range list {
		if err := s.Close(ctx, list[index].ID); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		closed++
	}

	return closed, firstErr
}

// markPaidFromQuery 依据查单结果把待支付的单补成已支付，并登记业务通知任务。
// 入参：支付单与查单结果；返回错误表示落库失败，条件不满足时静默返回。
func (s *PayOrderService) markPaidFromQuery(ctx context.Context, order *model.PayOrderModel, result *QueryResult) error {
	extension, err := s.latestExtension(ctx, order.ID)
	if err != nil {
		return err
	}

	var extensionID types.ID
	if extension != nil {
		extensionID = extension.ID
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		rows, err := s.orders.MarkPaid(ctx, tx, order.ID, extensionID, result.ChannelOrderNo, result.ChannelUserID, types.Now())
		if err != nil || rows == 0 {
			return err
		}

		return s.notifyTasks.CreateOrderTask(ctx, tx, order)
	})
}

// latestExtension 取支付单最近一次渠道调用记录，补单时用于回写拓展单编号。
// 入参：支付单主键；返回最近一条拓展单，没有记录时返回 (nil, nil)。
func (s *PayOrderService) latestExtension(ctx context.Context, orderID types.ID) (*model.PayOrderExtensionModel, error) {
	list, err := s.extensions.ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}

	return &list[len(list)-1], nil
}

// submitToChannel 新建一条拓展记录并调用渠道下单，成功回写调起参数、失败回写渠道报错。
// 入参：支付单、渠道、openid 与付款码；返回前端所需的调起参数。
func (s *PayOrderService) submitToChannel(ctx context.Context, order *model.PayOrderModel, channel *model.PayChannelModel, openID, authCode string) (*PayOrderSubmitResult, error) {
	gateway, err := NewChannelGateway(channel)
	if err != nil {
		return nil, err
	}

	// 渠道回调必须回到本模块，业务侧回调地址由 app.order_notify_url 在通知任务里使用
	callbackURL := channelNotifyURL(channel.ID)
	if callbackURL == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "未配置支付回调基础地址 PAY_NOTIFY_BASE_URL")
	}

	extension := &model.PayOrderExtensionModel{
		No:          order.No,
		OrderID:     order.ID,
		ChannelID:   channel.ID,
		ChannelCode: channel.Code,
		UserIP:      order.UserIP,
		Status:      payEnums.PayOrderStatusWaiting,
	}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.extensions.Create(ctx, tx, extension)
	}); err != nil {
		return nil, err
	}

	result, submitErr := gateway.Submit(ctx, SubmitRequest{
		No:        order.No,
		Subject:   order.Subject,
		Body:      order.Body,
		Price:     order.Price,
		NotifyURL: callbackURL,
		OpenID:    openID,
		AuthCode:  authCode,
		ClientIP:  order.UserIP,
		Attach:    order.MerchantOrderID,
	})
	if submitErr != nil {
		s.recordChannelError(ctx, extension, submitErr)

		return nil, submitErr
	}

	extension.ChannelExtras = result.Extras
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.extensions.UpdateResult(ctx, tx, extension)
	}); err != nil {
		return nil, err
	}

	return &PayOrderSubmitResult{
		OrderID:       order.ID,
		No:            order.No,
		ChannelCode:   order.ChannelCode,
		ChannelExtras: result.Extras,
		PayParams:     result.PayParams,
	}, nil
}

// recordChannelError 把渠道下单失败的错误码与文案写进拓展单，便于后台排查。
// 入参：拓展单与渠道错误；写入失败只忽略，不影响向上返回原始错误。
func (s *PayOrderService) recordChannelError(ctx context.Context, extension *model.PayOrderExtensionModel, cause error) {
	extension.ChannelErrorCode = "0"
	extension.ChannelErrorMsg = trimMessage(cause.Error())

	var businessErr *baseErrs.BusinessError
	if errors.As(cause, &businessErr) {
		extension.ChannelErrorCode = strconv.Itoa(businessErr.Code())
		extension.ChannelErrorMsg = businessErr.Msg()
	}
	if len([]rune(extension.ChannelErrorMsg)) > payOrderChannelErrorMsgMax {
		extension.ChannelErrorMsg = string([]rune(extension.ChannelErrorMsg)[:payOrderChannelErrorMsgMax])
	}

	_ = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.extensions.UpdateResult(ctx, tx, extension)
	})
}

// findByID 按主键取支付单，未命中转换成固定业务错误。
func (s *PayOrderService) findByID(ctx context.Context, id types.ID) (*model.PayOrderModel, error) {
	m, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayOrderNotExists)
	}

	return m, nil
}

// decoratePayOrder 为支付单补齐渠道中文名与状态文案。
func decoratePayOrder(m model.PayOrderModel) PayOrderRow {
	return PayOrderRow{
		PayOrderModel: m,
		ChannelName:   payEnums.ChannelLabel(m.ChannelCode),
		StatusName:    m.Status.String(),
	}
}

// normalizePayOrderInput 清洗并校验下单入参：商户订单号、标题、金额与有效期。
// 入参：待落库的下单入参指针，会被原地清洗；返回业务错误表示校验不通过。
func normalizePayOrderInput(in *PayOrderCreateInput) error {
	in.ChannelCode = strings.TrimSpace(in.ChannelCode)
	in.MerchantOrderID = strings.TrimSpace(in.MerchantOrderID)
	in.Subject = strings.TrimSpace(in.Subject)
	in.Body = strings.TrimSpace(in.Body)
	in.NotifyURL = strings.TrimSpace(in.NotifyURL)
	in.OpenID = strings.TrimSpace(in.OpenID)
	in.AuthCode = strings.TrimSpace(in.AuthCode)
	in.UserIP = strings.TrimSpace(in.UserIP)

	if in.MerchantOrderID == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "商户订单号不能为空")
	}
	if in.Subject == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "商品标题不能为空")
	}
	if len([]rune(in.Subject)) > payOrderSubjectMaxLength {
		return baseErrs.WithMsg(baseErrs.BadRequest, "商品标题不能超过 128 个字符")
	}
	// 金额只认服务端试算结果，0 或负数一律拒绝，避免 0 元支付单
	if in.Price <= 0 {
		return baseErrs.New(payErrs.PayOrderPriceInvalid)
	}
	if in.ExpireMinutes <= 0 {
		in.ExpireMinutes = payOrderExpireMinutesDefault
	}
	if in.ExpireMinutes > payOrderExpireMinutesMax {
		in.ExpireMinutes = payOrderExpireMinutesMax
	}

	return nil
}

// resolveNotifyURL 取业务回调地址，业务未指定时回落到应用配置的支付回调地址。
// 入参：业务侧回调地址与应用；返回最终下发给渠道的 notify_url。
func resolveNotifyURL(requestURL string, app *model.PayAppModel) string {
	if requestURL != "" {
		return requestURL
	}
	if app == nil {
		return ""
	}

	return app.OrderNotifyURL
}

// calcChannelFeePrice 按渠道费率（百分比）算出渠道手续费，单位分，四舍五入。
// 入参：支付金额（分）与费率；返回手续费（分）。
func calcChannelFeePrice(price int64, feeRate float64) int64 {
	if price <= 0 || feeRate <= 0 {
		return 0
	}

	return int64(math.Round(float64(price) * feeRate / 100))
}

// channelNotifyURL 生成渠道异步通知地址：{pay.notify_base_url}/api/v1/pay/callbacks/order/{渠道主键}。
// 入参：渠道主键；返回完整地址，未配置基础地址时返回空串。
func channelNotifyURL(channelID types.ID) string {
	base := strings.TrimRight(facades.Config().GetString("pay.notify_base_url", ""), "/")
	if base == "" {
		return ""
	}

	return base + "/api/v1/pay/callbacks/order/" + channelID.String()
}

// newPayOrderNo 生成支付单号：P + 秒级时间 + 雪花 ID 末 6 位，兼顾可读性与唯一性。
// 入参：无；返回形如 P20261022153000123456 的支付单号。
func newPayOrderNo() string {
	id := support.NewID().String()

	return "P" + time.Now().Format("20060102150405") + id[len(id)-6:]
}
