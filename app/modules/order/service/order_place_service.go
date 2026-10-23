// order_place_service.go 实现下单（幂等）、状态流转、超时关单与支付对接。
// 支付对接方式与 yudao 一致：订单号作为支付单的商户订单号（merchant_order_id），
// 支付成功后 pay 模块按应用配置的回调地址推送签名通知，订单侧验签 + 幂等更新状态。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/errs"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
	orderModel "eqray-flavor/app/modules/order/model"
	orderRepository "eqray-flavor/app/modules/order/repository"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payRepository "eqray-flavor/app/modules/pay/repository"
	payService "eqray-flavor/app/modules/pay/service"
)

// payUserTypeMember 是 C 端会员在 pay 模块的用户类型，与支付客户端约定一致。
const payUserTypeMember int8 = 1

// orderPayExpire 是订单待支付的超时时间。
const orderPayExpire = 15 * time.Minute

// CreateOrderItemInput 是下单输入里的一行明细。
type CreateOrderItemInput struct {
	// SkuID 菜品规格 ID
	SkuID types.ID
	// Quantity 数量
	Quantity int
	// Addons 已选加料
	Addons []EstimateAddonInput
	// Orderer 点菜人，堂食多人点菜时使用
	Orderer string
	// Remark 明细备注
	Remark string
}

// CreateOrderInput 是下单输入。
type CreateOrderInput struct {
	// StoreID 门店 ID
	StoreID types.ID
	// MemberID 会员 ID，0 表示散客（POS）
	MemberID types.ID
	// Type 订单类型：1 堂食 / 2 外卖 / 3 自提
	Type orderEnums.OrderType
	// Channel 下单渠道：1 小程序 / 2 POS / 3 自助机 / 4 外卖平台
	Channel orderEnums.OrderChannel
	// TableID 堂食桌台 ID
	TableID types.ID
	// RequestNo 幂等请求号，客户端缺省时服务端生成
	RequestNo string
	// Remark 订单备注
	Remark string
	// ReceiverName / ReceiverPhone / ReceiverAddress 外卖收货信息
	ReceiverName    string
	ReceiverPhone   string
	ReceiverAddress string
	// Items 明细行
	Items []CreateOrderItemInput
	// IsMember 是否会员（决定会员价）
	IsMember bool
}

// CreateResult 是下单结果：订单与是否幂等命中。
type CreateResult struct {
	// Order 订单（含明细加载前的金额拆分）
	Order *orderModelRef
	// Items 明细快照
	Items []orderItemRef
	// IdempotentHit 是否命中幂等：true 表示同请求号返回的是旧订单
	IdempotentHit bool
}

// orderModelRef 是对外输出的订单字段，避免直接暴露 model 指针。
type orderModelRef struct {
	ID         types.ID       `json:"id"`
	No         string         `json:"no"`
	Status     int16          `json:"status"`
	StatusText string         `json:"statusText"`
	PayAmount  int64          `json:"payAmount"`
	PickupCode string         `json:"pickupCode"`
	ExpireTime types.DateTime `json:"expireTime"`
}

// orderItemRef 是对外输出的明细字段。
type orderItemRef struct {
	ID        types.ID `json:"id"`
	SpuName   string   `json:"spuName"`
	SkuName   string   `json:"skuName"`
	UnitPrice int64    `json:"unitPrice"`
	Quantity  int      `json:"quantity"`
	Subtotal  int64    `json:"subtotal"`
}

// Create 下单。幂等靠 (store_id, request_no) 唯一索引兜底：
// 同请求号重复提交返回旧订单并标记 IdempotentHit，不重复创建。
func (s *OrderService) Create(ctx context.Context, in CreateOrderInput) (*CreateResult, error) {
	if len(in.Items) == 0 {
		return nil, errs.New(orderErrs.OrderItemsEmpty)
	}
	if in.Type != orderEnums.OrderTypeDineIn && in.Type != orderEnums.OrderTypeTakeout && in.Type != orderEnums.OrderTypePickup {
		return nil, errs.New(orderErrs.OrderTypeInvalid)
	}
	if in.Type == orderEnums.OrderTypeTakeout && (in.ReceiverName == "" || in.ReceiverPhone == "" || in.ReceiverAddress == "") {
		return nil, errs.New(orderErrs.OrderDeliveryInfoRequired)
	}

	// 会员单必须有会员身份：身份只从令牌取，请求体不传会员 ID（C 端）。
	// C 端控制器把渠道固定为小程序，因此这里等价于「C 端下单必须登录」；
	// POS / 自助机渠道是散客场景，不受此限制。
	if in.MemberID == 0 && in.Channel == orderEnums.OrderChannelMiniProgram {
		return nil, errs.New(orderErrs.OrderMemberRequired)
	}

	requestNo := in.RequestNo
	if requestNo == "" {
		requestNo = "REQ" + support.NewID().String()
	}

	// 幂等第一道：同请求号已有订单直接返回
	existing, err := s.orders.FindByIdempotency(ctx, in.StoreID, requestNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.buildCreateResult(existing, nil, true), nil
	}

	// 堂食桌台快照
	tableTitle := ""
	if in.Type == orderEnums.OrderTypeDineIn {
		if in.TableID == 0 {
			return nil, errs.New(orderErrs.OrderTableRequired)
		}
		table, err := s.tables.FindByID(ctx, in.TableID)
		if err != nil {
			return nil, err
		}
		if table == nil {
			return nil, errs.New(orderErrs.OrderTableRequired)
		}
		tableTitle = table.Name
	}

	// 装配明细（复用试算口径）并计价
	estimateItems := make([]EstimateItemInput, 0, len(in.Items))
	for _, item := range in.Items {
		estimateItems = append(estimateItems, EstimateItemInput{
			SkuID: item.SkuID, Quantity: item.Quantity, Addons: item.Addons,
		})
	}
	lines, err := s.resolveLines(ctx, in.StoreID, in.IsMember, estimateItems)
	if err != nil {
		return nil, err
	}
	pricing := CalculatePricing(lines, PricingRuleInput{IsMember: in.IsMember})

	// 组装订单与快照
	order := &orderModel.CaterOrderModel{
		No:                 newOrderNo(),
		StoreID:            in.StoreID,
		MemberID:           in.MemberID,
		Type:               in.Type,
		Channel:            in.Channel,
		TableID:            in.TableID,
		TableTitle:         tableTitle,
		RequestNo:          requestNo,
		Status:             orderEnums.OrderStatusPendingPay,
		PayStatus:          orderEnums.PayStatusUnpaid,
		ItemAmount:         pricing.ItemAmount,
		AddonAmount:        pricing.AddonAmount,
		LineDiscountAmount: pricing.LineDiscountAmount,
		ActivityAmount:     pricing.ActivityAmount,
		CouponAmount:       pricing.CouponAmount,
		PointsAmount:       pricing.PointsAmount,
		PayAmount:          pricing.PayAmount,
		DiscountSnapshot:   marshalSnapshot(pricing),
		PickupCode:         newPickupCode(),
		ReceiverName:       in.ReceiverName,
		ReceiverPhone:      in.ReceiverPhone,
		ReceiverAddress:    in.ReceiverAddress,
		Remark:             in.Remark,
		ExpireTime:         types.DateTime(time.Now().Add(orderPayExpire)),
	}

	items := make([]*orderModel.CaterOrderItemModel, 0, len(pricing.Lines))
	addons := make([]*orderModel.CaterOrderItemAddonModel, 0)
	for index, line := range pricing.Lines {
		item := &orderModel.CaterOrderItemModel{
			OrderID:            order.ID,
			SpuID:              line.SpuID,
			SkuID:              line.SkuID,
			SpuName:            line.SpuName,
			SkuName:            line.SkuName,
			OriginalPrice:      line.UnitPrice,
			UnitPrice:          line.UnitPrice,
			Quantity:           line.Quantity,
			AddonAmount:        line.AddonAmount,
			ItemDiscountAmount: line.OrderDiscount,
			Subtotal:           line.Subtotal,
			Status:             int8(orderEnums.ItemStatusNormal),
			Orderer:            in.Items[index].Orderer,
			Remark:             in.Items[index].Remark,
		}
		items = append(items, item)

		for _, addon := range line.Addons {
			addons = append(addons, &orderModel.CaterOrderItemAddonModel{
				OrderID:     order.ID,
				OrderItemID: item.ID,
				AddonItemID: addon.AddonItemID,
				AddonName:   addon.Name,
				Price:       addon.Price,
				Quantity:    addon.Quantity,
			})
		}
	}

	// 事务写入：幂等第二道靠 (store_id, request_no) 唯一索引，冲突时回落查旧单
	var conflict *orderModel.CaterOrderModel
	txErr := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		dup, err := s.orders.FindByIdempotency(ctx, in.StoreID, requestNo)
		if err != nil {
			return err
		}
		if dup != nil {
			conflict = dup

			return nil
		}

		if err := s.orders.CreateOrder(ctx, tx, order); err != nil {
			return err
		}
		if err := s.orders.CreateItems(ctx, tx, items); err != nil {
			return err
		}
		if err := s.orders.CreateItemAddons(ctx, tx, addons); err != nil {
			return err
		}

		return s.orders.CreateLog(ctx, tx, orderRepository.NewOrderLogModel(
			order.ID, 0, orderEnums.OrderStatusPendingPay,
			int8(orderEnums.OperatorTypeSystem), 0, "创建订单", "下单",
		))
	})
	if txErr != nil {
		return nil, txErr
	}
	if conflict != nil {
		return s.buildCreateResult(conflict, nil, true), nil
	}

	return s.buildCreateResult(order, items, false), nil
}

// buildCreateResult 组装下单结果。
func (s *OrderService) buildCreateResult(order *orderModel.CaterOrderModel, items []*orderModel.CaterOrderItemModel, hit bool) *CreateResult {
	result := &CreateResult{
		Order: &orderModelRef{
			ID:         order.ID,
			No:         order.No,
			Status:     int16(order.Status),
			StatusText: order.Status.String(),
			PayAmount:  order.PayAmount,
			PickupCode: order.PickupCode,
			ExpireTime: order.ExpireTime,
		},
		IdempotentHit: hit,
		Items:         make([]orderItemRef, 0, len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, orderItemRef{
			ID: item.ID, SpuName: item.SpuName, SkuName: item.SkuName,
			UnitPrice: item.UnitPrice, Quantity: item.Quantity, Subtotal: item.Subtotal,
		})
	}

	return result
}

// TransitInput 是店员 / 系统状态流转的输入。
type TransitInput struct {
	// OrderID 订单 ID
	OrderID types.ID
	// To 目标状态
	To orderEnums.OrderStatus
	// Reason 原因，如拒单理由
	Reason string
	// OperatorType 操作方：1 顾客 / 2 店员 / 3 系统
	OperatorType int8
	// OperatorID 操作人 ID
	OperatorID types.ID
	// OperatorName 操作人名称
	OperatorName string
	// StoreID 门店上下文，>0 时校验归属
	StoreID types.ID
	// UserID 当前后台用户，>0 时按门店数据权限校验归属
	UserID types.ID
}

// Transit 执行一次状态流转：状态机校验 + 条件更新防并发 + 日志留痕。
func (s *OrderService) Transit(ctx context.Context, in TransitInput) error {
	order, err := s.orders.FindByID(ctx, in.OrderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errs.New(orderErrs.OrderNotExists)
	}
	if in.StoreID > 0 && order.StoreID != in.StoreID {
		return errs.New(orderErrs.OrderStoreMismatch)
	}
	// 门店数据权限：店员只能操作自己可见门店的订单
	if in.UserID > 0 {
		if err := s.ensureStoreVisible(ctx, in.UserID, order.StoreID); err != nil {
			return errs.New(orderErrs.OrderStoreMismatch)
		}
	}
	if !order.Status.CanTransit(in.To) {
		return errs.New(orderErrs.OrderStatusInvalidTransition, order.Status.String(), in.To.String())
	}

	extra := map[string]any{}
	now := types.Now()
	switch in.To {
	case orderEnums.OrderStatusPendingAccept:
		extra["accept_time"] = now
	case orderEnums.OrderStatusCompleted:
		extra["complete_time"] = now
	case orderEnums.OrderStatusCancelled:
		extra["cancel_time"] = now
		extra["cancel_reason"] = in.Reason
	case orderEnums.OrderStatusRefunding:
		extra["cancel_reason"] = in.Reason
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		moved, err := s.orders.Transit(ctx, tx, order.ID, order.Status, in.To, extra)
		if err != nil {
			return err
		}
		if !moved {
			// 并发下另一请求已先流转，按非法流转拒绝
			return errs.New(orderErrs.OrderStatusInvalidTransition, order.Status.String(), in.To.String())
		}

		return s.orders.CreateLog(ctx, tx, orderRepository.NewOrderLogModel(
			order.ID, order.Status, in.To, in.OperatorType, in.OperatorID, in.OperatorName, in.Reason,
		))
	})
}

// Cancel 顾客取消订单：只允许待支付状态；已支付订单走退款流程。
func (s *OrderService) Cancel(ctx context.Context, orderID, memberID types.ID, reason string) error {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errs.New(orderErrs.OrderNotExists)
	}
	if order.Status != orderEnums.OrderStatusPendingPay {
		return errs.New(orderErrs.OrderCancelNotAllowed, order.Status.String())
	}
	// C 端会员只能取消自己的订单
	if memberID > 0 && order.MemberID != memberID {
		return errs.New(orderErrs.OrderNotExists)
	}

	return s.Transit(ctx, TransitInput{
		OrderID: orderID, To: orderEnums.OrderStatusCancelled,
		Reason: reason, OperatorType: int8(orderEnums.OperatorTypeMember),
		OperatorID: memberID, OperatorName: "顾客",
	})
}

// CloseExpired 关闭一批超时的待支付订单，返回关单数；由定时任务调用。
// 关单顺序：先让 pay 侧查渠道确认未收款并关掉支付单，确认关闭后才取消订单。
// 渠道已收款的单不取消（由支付回调推进），避免出现「已收款但订单被取消」。
func (s *OrderService) CloseExpired(ctx context.Context, limit uint64) (int, error) {
	expired, err := s.orders.ListExpiredPendingPay(ctx, time.Now(), limit)
	if err != nil {
		return 0, err
	}

	closed := 0
	for i := range expired {
		order := &expired[i]

		// 关支付单：内部会查渠道确认未收款；已收款的单返回支付成功状态，这里不取消订单
		payOrder, closeErr := payService.NewPayOrderService().CloseByMerchantOrderID(ctx, order.No)
		if closeErr != nil {
			return closed, closeErr
		}
		if payOrder != nil && payOrder.Status == payEnums.PayOrderStatusSuccess {
			// 渠道已收款：订单交给支付回调推进，本轮不关单
			continue
		}

		err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
			moved, err := s.orders.Transit(ctx, tx, order.ID, order.Status, orderEnums.OrderStatusCancelled, map[string]any{
				"cancel_time":   types.Now(),
				"cancel_reason": "超时未支付，系统自动关单",
			})
			if err != nil || !moved {
				return err
			}

			return s.orders.CreateLog(ctx, tx, orderRepository.NewOrderLogModel(
				order.ID, order.Status, orderEnums.OrderStatusCancelled,
				int8(orderEnums.OperatorTypeSystem), 0, "超时关单", "超时未支付",
			))
		})
		if err != nil {
			return closed, err
		}
		closed++
	}

	return closed, nil
}

// CreatePaymentInput 是为订单发起支付的输入。
type CreatePaymentInput struct {
	// OrderID 订单 ID
	OrderID types.ID
	// AppID 支付应用编号，C 端按品牌配置传入
	AppID types.ID
	// ChannelCode 渠道编码，如 wechat_mini / alipay_wap
	ChannelCode string
	// OpenID 微信 JSAPI / 小程序支付必填
	OpenID string
	// MemberID 会员身份，从令牌取
	MemberID types.ID
}

// CreatePayment 为待支付订单发起支付：金额与订单号在 pay 侧做幂等，
// 重复调用会返回既有支付单的调起参数。
func (s *OrderService) CreatePayment(ctx context.Context, in CreatePaymentInput) (*payService.PayOrderSubmitResult, error) {
	order, err := s.orders.FindByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errs.New(orderErrs.OrderNotExists)
	}
	if in.MemberID > 0 && order.MemberID != in.MemberID {
		return nil, errs.New(orderErrs.OrderNotExists)
	}
	if order.Status != orderEnums.OrderStatusPendingPay {
		return nil, errs.New(orderErrs.OrderStatusInvalidTransition, order.Status.String(), "支付")
	}
	if order.PayAmount <= 0 {
		return nil, errs.New(orderErrs.OrderPayAmountMissing)
	}

	subject := order.No
	if len(order.Remark) > 0 {
		subject = subject + " " + order.Remark
	}

	// 幂等：同一订单号在 pay 侧只允许一张有效支付单，重复发起复用既有单并重新调起渠道
	existing, err := payRepository.NewPayOrderRepository().FindByMerchantOrderID(ctx, in.AppID, order.No)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Status != payEnums.PayOrderStatusWaiting {
			return nil, errs.New(orderErrs.OrderStatusInvalidTransition, order.Status.String(), "支付")
		}

		return payService.NewPayOrderService().ReSubmit(ctx, existing.ID, in.OpenID, "")
	}

	result, err := payService.NewPayOrderService().Create(ctx, payService.PayOrderCreateInput{
		AppID:           in.AppID,
		ChannelCode:     in.ChannelCode,
		UserID:          order.MemberID,
		UserType:        payUserTypeMember,
		MerchantOrderID: order.No,
		Subject:         subject,
		Price:           order.PayAmount,
		OpenID:          in.OpenID,
	})
	if err != nil {
		return nil, err
	}

	// 回写支付单关联，便于后台从订单直达支付单
	if err := s.orders.BindPayOrder(ctx, order.ID, result.OrderID); err != nil {
		return nil, err
	}

	return result, nil
}

// PayNotifyPayload 是 pay 模块推送给业务方的通知报文，字段与 buildPayload 对应。
type PayNotifyPayload struct {
	// Type 通知类型：order / refund
	Type string `json:"type"`
	// AppKey 应用标识
	AppKey string `json:"appKey"`
	// MerchantOrderID 商户订单号，这里即订单号
	MerchantOrderID string `json:"merchantOrderId"`
	// OrderNo 支付单号
	OrderNo string `json:"orderNo"`
	// ChannelCode 支付渠道
	ChannelCode string `json:"channelCode"`
	// ChannelOrderNo 渠道单号
	ChannelOrderNo string `json:"channelOrderNo"`
	// Price 支付金额（分）
	Price int64 `json:"price"`
	// SuccessTime 支付成功时间
	SuccessTime string `json:"successTime"`
}

// HandlePayNotify 处理支付成功的业务通知：验签 → 金额核对 → 幂等流转订单状态。
// 入参：请求头（签名四件套）与原始请求体；返回业务侧应答文本（success / error）。
// 返回 error 时 pay 侧会按重试策略继续推送，因此这里必须保持幂等。
func (s *OrderService) HandlePayNotify(ctx context.Context, headers map[string]string, body []byte) (string, error) {
	var payload PayNotifyPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return "error", fmt.Errorf("通知报文解析失败: %w", err)
	}
	if payload.Type != "order" || payload.MerchantOrderID == "" {
		return "success", nil // 非订单通知或字段缺失：按已处理应答，避免无意义重试
	}

	// 验签：应用密钥按 AppKey 取
	app, err := payRepository.NewPayAppRepository().FindByAppKey(ctx, payload.AppKey)
	if err != nil {
		return "error", err
	}
	if app == nil {
		return "error", fmt.Errorf("未知应用 %s", payload.AppKey)
	}
	if !payService.VerifyNotifySignature(app.AppSecret, headers, body) {
		return "error", fmt.Errorf("通知验签失败")
	}

	order, err := s.orders.FindByNo(ctx, payload.MerchantOrderID)
	if err != nil {
		return "error", err
	}
	if order == nil {
		return "error", fmt.Errorf("订单 %s 不存在", payload.MerchantOrderID)
	}

	// 定位支付单：通知里的 orderNo 是支付单号，用于回写关联关系
	payOrder, err := payRepository.NewPayOrderRepository().FindByNo(ctx, payload.OrderNo)
	if err != nil {
		return "error", err
	}
	if payOrder == nil {
		return "error", fmt.Errorf("支付单 %s 不存在", payload.OrderNo)
	}

	// 金额红线：以支付单金额为准核对订单应付，不一致直接拒绝并让 pay 持续重试（人工介入）
	if payOrder.Price != order.PayAmount {
		return "error", fmt.Errorf("订单 %s 金额不一致：支付单 %d ≠ 应付 %d", order.No, payOrder.Price, order.PayAmount)
	}
	if payload.Price != 0 && payload.Price != payOrder.Price {
		return "error", fmt.Errorf("支付单 %s 通知金额 %d ≠ 支付单金额 %d", payOrder.No, payload.Price, payOrder.Price)
	}

	// 幂等：非待支付状态说明已处理过，直接应答成功
	if order.Status != orderEnums.OrderStatusPendingPay {
		return "success", nil
	}

	// 流转：堂食支付即制作（100→300），外卖 / 自提进入待接单（100→200）
	target := orderEnums.OrderStatusPendingAccept
	if order.Type == orderEnums.OrderTypeDineIn {
		target = orderEnums.OrderStatusMaking
	}

	txErr := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		moved, err := s.orders.Transit(ctx, tx, order.ID, order.Status, target, map[string]any{
			"pay_status":   int8(orderEnums.PayStatusPaid),
			"pay_time":     types.Now(),
			"pay_order_id": payOrder.ID,
		})
		if err != nil || !moved {
			return err
		}

		return s.orders.CreateLog(ctx, tx, orderRepository.NewOrderLogModel(
			order.ID, order.Status, target,
			int8(orderEnums.OperatorTypeSystem), 0, "支付回调", "支付成功 "+payOrder.No,
		))
	})
	if txErr != nil {
		return "error", txErr
	}

	return "success", nil
}

// newOrderNo 生成订单号：O + 时间戳 + 雪花尾 6 位。
func newOrderNo() string {
	id := support.NewID().String()

	return "O" + time.Now().Format("20060102150405") + id[len(id)-6:]
}

// newPickupCode 生成 4 位取餐码；门店内偶发重复可接受，不设唯一索引。
func newPickupCode() string {
	return fmt.Sprintf("%04d", 1000+rand.Intn(9000))
}

// marshalSnapshot 把计价结果序列化成优惠快照 JSON。
func marshalSnapshot(pricing *PricingResult) string {
	body, err := json.Marshal(map[string]any{
		"itemAmount":     pricing.ItemAmount,
		"addonAmount":    pricing.AddonAmount,
		"lineDiscount":   pricing.LineDiscountAmount,
		"activityAmount": pricing.ActivityAmount,
		"couponAmount":   pricing.CouponAmount,
		"pointsAmount":   pricing.PointsAmount,
		"payAmount":      pricing.PayAmount,
		"steps":          pricing.Steps,
	})
	if err != nil {
		return ""
	}

	return string(body)
}
