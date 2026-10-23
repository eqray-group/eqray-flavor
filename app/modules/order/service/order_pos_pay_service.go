// order_pos_pay_service.go 实现 POS 的扫码 / 条码收款。
// 全项目只跑一套支付：支付应用由后端自动解析（要求恰好启用一个），收银端不传 appId。
package service

import (
	"context"
	"strings"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
	orderModel "eqray-flavor/app/modules/order/model"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payModel "eqray-flavor/app/modules/pay/model"
	payRepository "eqray-flavor/app/modules/pay/repository"
	payService "eqray-flavor/app/modules/pay/service"
)

// PosPayType 是 POS 结账的收款方式。
type PosPayType int8

const (
	// PosPayTypeWechat 微信扫码：收银台展示二维码，由顾客用微信扫
	PosPayTypeWechat PosPayType = 1
	// PosPayTypeAlipay 支付宝条码：收银员扫顾客付款码，通常同步返回结果
	PosPayTypeAlipay PosPayType = 2
)

// String 返回收款方式的中文描述。
func (t PosPayType) String() string {
	switch t {
	case PosPayTypeWechat:
		return "微信扫码"
	case PosPayTypeAlipay:
		return "支付宝条码"
	default:
		return "未知方式"
	}
}

// ChannelCode 返回收款方式对应的支付渠道编码；不支持的取值返回空串。
// 微信没有付款码渠道，POS 只能走 Native 扫码；支付宝支持条码与扫码两种，这里取条码。
func (t PosPayType) ChannelCode() string {
	switch t {
	case PosPayTypeWechat:
		return payEnums.ChannelWxNative
	case PosPayTypeAlipay:
		return payEnums.ChannelAlipayBar
	default:
		return ""
	}
}

// PosPayInput 是 POS 扫码收款的入参。
type PosPayInput struct {
	// OrderID 订单 ID
	OrderID types.ID
	// StoreID 门店 ID，由请求头带入
	StoreID types.ID
	// UserID 收银员后台用户 ID
	UserID types.ID
	// PayType 收款方式
	PayType PosPayType
	// AuthCode 顾客付款码，支付宝条码必填，微信扫码留空
	AuthCode string
}

// PosPayResult 是扫码收款的结果。
type PosPayResult struct {
	// OrderNo 订单号
	OrderNo string `json:"orderNo"`
	// PayAmount 应收金额（分）
	PayAmount int64 `json:"payAmount"`
	// PayOrderNo 支付单号
	PayOrderNo string `json:"payOrderNo"`
	// ChannelCode 实际使用的渠道编码
	ChannelCode string `json:"channelCode"`
	// Paid 是否已确认收款成功；false 时前端展示二维码并轮询状态
	Paid bool `json:"paid"`
	// PayParams 渠道调起参数，微信为二维码链接
	PayParams map[string]any `json:"payParams"`
}

// PosPayStatusInput 是查询收款状态的入参。
type PosPayStatusInput struct {
	// OrderID 订单 ID
	OrderID types.ID
	// StoreID 门店 ID
	StoreID types.ID
	// UserID 收银员后台用户 ID
	UserID types.ID
}

// PosPayStatusResult 是收款状态查询结果。
type PosPayStatusResult struct {
	// OrderNo 订单号
	OrderNo string `json:"orderNo"`
	// Paid 是否已收款成功
	Paid bool `json:"paid"`
	// PayStatus 支付状态：0 未支付 / 10 已支付 / 20 已退款
	PayStatus int8 `json:"payStatus"`
	// WaitScan 是否仍在等待顾客扫码
	WaitScan bool `json:"waitScan"`
}

// PosPay 发起扫码收款：建支付单并（条码支付时）同步确认结果。
// 入参：订单、门店、收银员、收款方式与付款码；返回支付单号与调起参数。
// 说明：微信走 Native 二维码，需前端轮询 PosPayStatus；支付宝条码建单后立刻查一次渠道。
func (s *OrderService) PosPay(ctx context.Context, in PosPayInput) (*PosPayResult, error) {
	order, err := s.posPayableOrder(ctx, in.OrderID, in.StoreID, in.UserID)
	if err != nil {
		return nil, err
	}

	authCode := strings.TrimSpace(in.AuthCode)
	if in.PayType == PosPayTypeAlipay && authCode == "" {
		return nil, errs.New(orderErrs.OrderPosPayAuthCodeRequired)
	}

	channelCode := in.PayType.ChannelCode()
	if channelCode == "" {
		return nil, errs.New(orderErrs.OrderPosPayChannelMissing, in.PayType.String())
	}

	app, err := s.resolveSinglePayApp(ctx)
	if err != nil {
		return nil, err
	}

	payOrders := payRepository.NewPayOrderRepository()
	existing, err := payOrders.FindLatestByMerchantOrderID(ctx, order.No)
	if err != nil {
		return nil, err
	}

	// 已成功的支付单：直接把订单补成已支付，避免收银员重复收款
	if existing != nil && existing.Status == payEnums.PayOrderStatusSuccess {
		if err := s.settleFromPayOrder(ctx, order, existing, in.UserID); err != nil {
			return nil, err
		}

		return &PosPayResult{
			OrderNo:     order.No,
			PayAmount:   order.PayAmount,
			PayOrderNo:  existing.No,
			ChannelCode: existing.ChannelCode,
			Paid:        true,
		}, nil
	}

	var result *payService.PayOrderSubmitResult
	if existing != nil && existing.Status == payEnums.PayOrderStatusWaiting {
		// 待支付的单复用原支付单，避免同一订单堆出多张单
		result, err = payService.NewPayOrderService().ReSubmit(ctx, existing.ID, "", authCode)
	} else {
		result, err = payService.NewPayOrderService().Create(ctx, payService.PayOrderCreateInput{
			AppID:           app.ID,
			ChannelCode:     channelCode,
			UserID:          order.MemberID,
			MerchantOrderID: order.No,
			Subject:         "餐饮订单 " + order.No,
			Body:            order.TableTitle,
			Price:           order.PayAmount,
			AuthCode:        authCode,
		})
	}
	if err != nil {
		return nil, err
	}

	out := &PosPayResult{
		OrderNo:     order.No,
		PayAmount:   order.PayAmount,
		PayOrderNo:  result.No,
		ChannelCode: result.ChannelCode,
		PayParams:   result.PayParams,
	}

	// 条码支付由渠道同步返回结果，这里立刻查一次；微信扫码等回调，交给轮询
	if authCode != "" {
		paid, confirmErr := s.confirmPayOrder(ctx, order, result.OrderID, in.UserID)
		if confirmErr != nil {
			return nil, confirmErr
		}
		out.Paid = paid
	}

	return out, nil
}

// PosPayStatus 查询收款状态：POS 展示二维码后轮询，成功即把订单推进为已支付。
// 入参：订单、门店与收银员；返回是否已收款与是否仍在等待扫码。
func (s *OrderService) PosPayStatus(ctx context.Context, in PosPayStatusInput) (*PosPayStatusResult, error) {
	order, err := s.posOrder(ctx, in.OrderID, in.StoreID, in.UserID)
	if err != nil {
		return nil, err
	}

	out := &PosPayStatusResult{
		OrderNo:   order.No,
		PayStatus: int8(order.PayStatus),
	}
	if order.PayStatus == orderEnums.PayStatusPaid {
		out.Paid = true

		return out, nil
	}

	payOrder, err := payRepository.NewPayOrderRepository().FindLatestByMerchantOrderID(ctx, order.No)
	if err != nil {
		return nil, err
	}
	if payOrder == nil {
		return out, nil
	}

	switch payOrder.Status {
	case payEnums.PayOrderStatusSuccess:
		if err := s.settleFromPayOrder(ctx, order, payOrder, in.UserID); err != nil {
			return nil, err
		}
		out.Paid = true
		out.PayStatus = int8(orderEnums.PayStatusPaid)
	case payEnums.PayOrderStatusWaiting:
		out.WaitScan = true
	}

	return out, nil
}

// resolveSinglePayApp 取系统内启用的支付应用。
// 返回：恰好一个时直接返回；0 个或多个都报错——本项目按「全局一套支付」运行，多应用属于配置错误。
func (s *OrderService) resolveSinglePayApp(ctx context.Context) (*payModel.PayAppModel, error) {
	apps, err := payRepository.NewPayAppRepository().AllEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if len(apps) == 0 {
		return nil, errs.New(orderErrs.OrderPosPayAppNotConfigured)
	}
	if len(apps) > 1 {
		return nil, errs.New(orderErrs.OrderPosPayAppNotUnique, len(apps))
	}

	return &apps[0], nil
}

// confirmPayOrder 主动查一次渠道，成功则落地订单。
// 入参：订单、支付单 ID 与收银员；返回是否已支付。
func (s *OrderService) confirmPayOrder(ctx context.Context, order *orderModel.CaterOrderModel, payOrderID, userID types.ID) (bool, error) {
	row, err := payService.NewPayOrderService().Sync(ctx, payOrderID)
	if err != nil {
		return false, err
	}
	if row == nil || row.Status != payEnums.PayOrderStatusSuccess {
		return false, nil
	}

	if err := s.settleFromPayOrder(ctx, order, &row.PayOrderModel, userID); err != nil {
		return false, err
	}

	return true, nil
}

// settleFromPayOrder 用成功的支付单把订单推进为已支付。
// 入参：订单、支付单与收银员；返回错误表示金额不一致或状态流转失败。
// 说明：金额红线与支付回调一致——支付单金额必须等于订单应付，避免少收漏收。
func (s *OrderService) settleFromPayOrder(ctx context.Context, order *orderModel.CaterOrderModel, payOrder *payModel.PayOrderModel, userID types.ID) error {
	if payOrder.Price != order.PayAmount {
		return errs.New(orderErrs.OrderPayAmountMismatch, payOrder.Price, order.PayAmount)
	}
	// 幂等：回调与轮询可能同时到达，已支付的直接返回
	if order.PayStatus == orderEnums.PayStatusPaid {
		return nil
	}

	return s.applyPaid(ctx, order, payOrder, paidLog{
		OperatorType: int8(orderEnums.OperatorTypeStaff),
		OperatorID:   userID,
		OperatorName: "收银员",
		Reason:       "POS 收款 " + payOrder.No,
	})
}

// posOrder 读取订单并校验门店与数据权限，不做状态判断。
func (s *OrderService) posOrder(ctx context.Context, orderID, storeID, userID types.ID) (*orderModel.CaterOrderModel, error) {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errs.New(orderErrs.OrderNotExists)
	}
	if storeID > 0 && order.StoreID != storeID {
		return nil, errs.New(orderErrs.OrderStoreMismatch)
	}
	if userID > 0 {
		if err := s.ensureStoreVisible(ctx, userID, order.StoreID); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// posPayableOrder 读取订单并校验可收款：待支付、未收款。
func (s *OrderService) posPayableOrder(ctx context.Context, orderID, storeID, userID types.ID) (*orderModel.CaterOrderModel, error) {
	order, err := s.posOrder(ctx, orderID, storeID, userID)
	if err != nil {
		return nil, err
	}
	if order.Status != orderEnums.OrderStatusPendingPay {
		return nil, errs.New(orderErrs.OrderCashSettleStatusInvalid, order.Status.String())
	}
	if order.PayStatus != orderEnums.PayStatusUnpaid {
		return nil, errs.New(orderErrs.OrderAlreadyPaid)
	}

	return order, nil
}
