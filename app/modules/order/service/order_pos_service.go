// order_pos_service.go 实现 POS 收银端的代客下单与现金结账。
// POS 是散客优先的场景：不要求会员身份，也可以关联会员享受会员价与积分；
// 渠道固定为 POS，不接受前端传入，避免绕过「C 端必须登录」的约束。
package service

import (
	"context"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberService "eqray-flavor/app/modules/member/service"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
)

// PosCreateInput 是收银员代客下单的入参。
// 与 CreateOrderInput 的差别只有两点：渠道固定 POS、会员可选。
type PosCreateInput struct {
	// StoreID 门店 ID，由请求头门店上下文带入
	StoreID types.ID
	// MemberID 会员 ID，0 表示散客
	MemberID types.ID
	// Type 订单类型：1 堂食 / 2 外卖 / 3 自提
	Type orderEnums.OrderType
	// TableID 堂食桌台 ID
	TableID types.ID
	// RequestNo 幂等请求号，收银端生成一次后重试复用
	RequestNo string
	// Remark 订单备注
	Remark string
	// ReceiverName / ReceiverPhone / ReceiverAddress 外卖收货信息
	ReceiverName    string
	ReceiverPhone   string
	ReceiverAddress string
	// Items 明细行
	Items []CreateOrderItemInput
	// CouponID 使用的优惠券 ID，0 表示不用券
	CouponID types.ID
	// UsePoints 是否使用积分抵扣
	UsePoints bool
}

// CreateByStaff 收银员代客下单：渠道固定为 POS，会员可选。
// 入参：门店、可选会员与明细；返回订单与是否幂等命中。
// 说明：指定会员时会先校验会员存在，避免后续扣积分时才发现会员不存在导致整单回滚。
func (s *OrderService) CreateByStaff(ctx context.Context, in PosCreateInput) (*CreateResult, error) {
	if in.StoreID == 0 {
		return nil, errs.New(orderErrs.OrderPosStoreRequired)
	}
	if in.MemberID > 0 {
		if _, err := memberService.NewMemberService().Detail(ctx, in.MemberID); err != nil {
			return nil, err
		}
	}

	return s.Create(ctx, CreateOrderInput{
		StoreID:         in.StoreID,
		MemberID:        in.MemberID,
		Type:            in.Type,
		Channel:         orderEnums.OrderChannelPOS,
		TableID:         in.TableID,
		RequestNo:       in.RequestNo,
		Remark:          in.Remark,
		ReceiverName:    in.ReceiverName,
		ReceiverPhone:   in.ReceiverPhone,
		ReceiverAddress: in.ReceiverAddress,
		Items:           in.Items,
		IsMember:        in.MemberID > 0,
		CouponID:        in.CouponID,
		UsePoints:       in.UsePoints,
	})
}

// PosSettleInput 是 POS 现金结账的入参。
type PosSettleInput struct {
	// OrderID 订单 ID
	OrderID types.ID
	// StoreID 门店 ID，由请求头带入，用于校验订单归属
	StoreID types.ID
	// UserID 收银员后台用户 ID，用于数据权限与操作留痕
	UserID types.ID
}

// SettleCash 现金结账：不生成支付单，直接把订单置为已支付并推进状态。
// 入参：订单、门店与收银员；返回错误表示订单不可收款或已收款。
// 说明：现金没有渠道流水，对账口径是「pay_status 已支付且 pay_order_id = 0」，
// 与微信 / 支付宝（走支付单）区分开；后续做交班时按这个口径统计现金。
func (s *OrderService) SettleCash(ctx context.Context, in PosSettleInput) error {
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
	if in.UserID > 0 {
		if err := s.ensureStoreVisible(ctx, in.UserID, order.StoreID); err != nil {
			return err
		}
	}
	if order.Status != orderEnums.OrderStatusPendingPay {
		return errs.New(orderErrs.OrderCashSettleStatusInvalid, order.Status.String())
	}
	if order.PayStatus != orderEnums.PayStatusUnpaid {
		return errs.New(orderErrs.OrderAlreadyPaid)
	}

	return s.applyPaid(ctx, order, nil, paidLog{
		OperatorType: int8(orderEnums.OperatorTypeStaff),
		OperatorID:   in.UserID,
		OperatorName: "收银员",
		Reason:       "现金收款",
	})
}
