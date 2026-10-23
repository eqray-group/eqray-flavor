// order_pos_append_service.go 实现堂食加菜与并台。
// 两者的本质是同一个动作：把一批明细并入一张未支付订单，整单重新计价后重写明细。
package service

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
	orderModel "eqray-flavor/app/modules/order/model"
	orderRepository "eqray-flavor/app/modules/order/repository"
)

// AppendOrderItemsInput 是加菜入参。
type AppendOrderItemsInput struct {
	// OrderID 目标订单
	OrderID types.ID
	// StoreID 门店，用于校验与数据权限
	StoreID types.ID
	// UserID 收银员
	UserID types.ID
	// Items 新增明细
	Items []CreateOrderItemInput
}

// AppendOrderItemsResult 是加菜结果。
type AppendOrderItemsResult struct {
	// OrderID 订单 ID
	OrderID types.ID `json:"id"`
	// No 订单号
	No string `json:"no"`
	// AddedLines 本次新增的明细行数
	AddedLines int `json:"addedLines"`
	// PayAmount 加菜后的整单应付（分）
	PayAmount int64 `json:"payAmount"`
}

// MergeOrdersInput 是并台入参。
type MergeOrdersInput struct {
	// TargetOrderID 保留的订单（被并入方）
	TargetOrderID types.ID
	// SourceOrderID 被并入的订单（并入后取消）
	SourceOrderID types.ID
	// StoreID / UserID 门店上下文与操作人
	StoreID types.ID
	UserID  types.ID
}

// MergeOrdersResult 是并台结果。
type MergeOrdersResult struct {
	// OrderID 保留的订单 ID
	OrderID types.ID `json:"id"`
	// No 保留的订单号
	No string `json:"no"`
	// PayAmount 合并后的整单应付（分）
	PayAmount int64 `json:"payAmount"`
	// SourceOrderID 被取消的订单
	SourceOrderID types.ID `json:"sourceOrderId"`
	// SourceTableID 被并入方的桌台，供前端清台
	SourceTableID types.ID `json:"sourceTableId"`
}

// AppendOrderItems 给未支付的堂食订单加菜。
// 入参：订单、门店、操作人与新增明细；返回加菜后的整单应付。
// 约束：只有未支付、未用券的订单能加菜——已支付的单金额已入账，改金额要走退款；
// 已用券的单重新计价会让券的分摊对不上，宁可拒绝也不产生错账。
func (s *OrderService) AppendOrderItems(ctx context.Context, in AppendOrderItemsInput) (*AppendOrderItemsResult, error) {
	if len(in.Items) == 0 {
		return nil, errs.New(orderErrs.OrderItemsEmpty)
	}

	order, err := s.posOrder(ctx, in.OrderID, in.StoreID, in.UserID)
	if err != nil {
		return nil, err
	}
	if err := ensureAppendable(order); err != nil {
		return nil, err
	}

	// 把库里已有的明细还原成下单入参，与新明细一起整单重算
	existing, err := s.loadOrderItemsAsInput(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	lines := make([]CreateOrderItemInput, 0, len(existing)+len(in.Items))
	lines = append(lines, existing...)
	lines = append(lines, in.Items...)

	payAmount, err := s.replaceOrderItems(ctx, order, lines, int8(orderEnums.OperatorTypeStaff), in.UserID, "POS 加菜")
	if err != nil {
		return nil, err
	}

	return &AppendOrderItemsResult{
		OrderID:    order.ID,
		No:         order.No,
		AddedLines: len(in.Items),
		PayAmount:  payAmount,
	}, nil
}

// MergeOrders 并台：把源单明细并入目标单，随后取消源单。
// 入参：目标单、源单、门店与操作人；返回合并后的应付与源台（供前端清台）。
// 说明：两边都必须是未支付堂食单且同店；整个过程在一个事务内，
// 源单取消失败会连同明细合并一起回滚，不会出现「菜并过去了、单还在」的双份。
func (s *OrderService) MergeOrders(ctx context.Context, in MergeOrdersInput) (*MergeOrdersResult, error) {
	if in.TargetOrderID == in.SourceOrderID {
		return nil, errs.New(orderErrs.OrderMergeNotAllowed)
	}

	target, err := s.posOrder(ctx, in.TargetOrderID, in.StoreID, in.UserID)
	if err != nil {
		return nil, err
	}
	source, err := s.posOrder(ctx, in.SourceOrderID, in.StoreID, in.UserID)
	if err != nil {
		return nil, err
	}
	if err := ensureAppendable(target); err != nil {
		return nil, err
	}
	if err := ensureAppendable(source); err != nil {
		return nil, err
	}
	if target.StoreID != source.StoreID || target.Type != orderEnums.OrderTypeDineIn {
		return nil, errs.New(orderErrs.OrderMergeNotAllowed)
	}

	targetLines, err := s.loadOrderItemsAsInput(ctx, target.ID)
	if err != nil {
		return nil, err
	}
	sourceLines, err := s.loadOrderItemsAsInput(ctx, source.ID)
	if err != nil {
		return nil, err
	}
	if len(sourceLines) == 0 {
		return nil, errs.New(orderErrs.OrderItemsEmpty)
	}

	lines := make([]CreateOrderItemInput, 0, len(targetLines)+len(sourceLines))
	lines = append(lines, targetLines...)
	lines = append(lines, sourceLines...)

	// 并台与加菜共用重算逻辑，额外多一步取消源单，因此放在同一事务里执行
	var payAmount int64
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		amount, err := s.replaceOrderItemsTx(ctx, tx, target, lines, 0, int8(orderEnums.OperatorTypeStaff), in.UserID, "并台")
		if err != nil {
			return err
		}
		payAmount = amount

		moved, err := s.orders.Transit(ctx, tx, source.ID, source.Status, orderEnums.OrderStatusCancelled, nil)
		if err != nil {
			return err
		}
		if !moved {
			return errs.New(orderErrs.OrderStatusInvalidTransition, source.Status.String(), orderEnums.OrderStatusCancelled.String())
		}

		return s.orders.CreateLog(ctx, tx, orderRepository.NewOrderLogModel(
			source.ID, source.Status, orderEnums.OrderStatusCancelled,
			int8(orderEnums.OperatorTypeStaff), in.UserID, "收银员", "并台到 "+target.No,
		))
	}); err != nil {
		return nil, err
	}

	return &MergeOrdersResult{
		OrderID:       target.ID,
		No:            target.No,
		PayAmount:     payAmount,
		SourceOrderID: source.ID,
		SourceTableID: source.TableID,
	}, nil
}

// ensureAppendable 校验订单是否允许加菜 / 并台。
func ensureAppendable(order *orderModel.CaterOrderModel) error {
	if order.PayStatus != orderEnums.PayStatusUnpaid {
		return errs.New(orderErrs.OrderAlreadyPaid)
	}
	if order.Status != orderEnums.OrderStatusPendingPay {
		return errs.New(orderErrs.OrderAppendNotAllowed, order.Status.String())
	}
	if order.CouponAmount > 0 {
		return errs.New(orderErrs.OrderAppendCouponLocked)
	}

	return nil
}

// loadOrderItemsAsInput 把已落库的明细还原成下单入参，用于整单重算。
// 入参：订单 ID；返回按原顺序排列的明细（含加料与备注）。
func (s *OrderService) loadOrderItemsAsInput(ctx context.Context, orderID types.ID) ([]CreateOrderItemInput, error) {
	grouped, err := s.orders.ListItemsByOrderIDs(ctx, []types.ID{orderID})
	if err != nil {
		return nil, err
	}
	items := grouped[orderID]
	if len(items) == 0 {
		return nil, nil
	}

	addonGrouped, err := s.orders.ListAddonsByOrderIDs(ctx, []types.ID{orderID})
	if err != nil {
		return nil, err
	}
	addonsByItem := make(map[types.ID][]orderModel.CaterOrderItemAddonModel, len(items))
	for _, addon := range addonGrouped[orderID] {
		addonsByItem[addon.OrderItemID] = append(addonsByItem[addon.OrderItemID], addon)
	}

	lines := make([]CreateOrderItemInput, 0, len(items))
	for _, item := range items {
		addons := make([]EstimateAddonInput, 0, len(addonsByItem[item.ID]))
		for _, addon := range addonsByItem[item.ID] {
			addons = append(addons, EstimateAddonInput{
				AddonItemID: addon.AddonItemID,
				Quantity:    addon.Quantity,
			})
		}
		lines = append(lines, CreateOrderItemInput{
			SkuID:    item.SkuID,
			Quantity: item.Quantity,
			Addons:   addons,
			Orderer:  item.Orderer,
			Remark:   item.Remark,
		})
	}

	return lines, nil
}

// replaceOrderItems 开事务重算并重写整单明细。
func (s *OrderService) replaceOrderItems(ctx context.Context, order *orderModel.CaterOrderModel, lines []CreateOrderItemInput, operatorType int8, operatorID types.ID, reason string) (int64, error) {
	var payAmount int64
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		amount, err := s.replaceOrderItemsTx(ctx, tx, order, lines, 0, operatorType, operatorID, reason)
		if err != nil {
			return err
		}
		payAmount = amount

		return nil
	}); err != nil {
		return 0, err
	}

	return payAmount, nil
}

// replaceOrderItemsTx 在调用方事务内完成「整单重新计价 + 明细重写 + 金额回写 + 日志」。
// 入参：事务、订单、完整明细、券 ID、操作人信息与日志原因；返回新的应付金额。
// 说明：换活动优惠会随整单金额变化，因此活动名额不在这里重新占用——
// 一期只在没有用券的单上做加菜，避免了券分摊与名额回收的复杂度。
func (s *OrderService) replaceOrderItemsTx(ctx context.Context, tx db.Tx, order *orderModel.CaterOrderModel, lines []CreateOrderItemInput, couponID types.ID, operatorType int8, operatorID types.ID, reason string) (int64, error) {
	estimateItems := make([]EstimateItemInput, 0, len(lines))
	for _, line := range lines {
		estimateItems = append(estimateItems, EstimateItemInput{
			SkuID: line.SkuID, Quantity: line.Quantity, Addons: line.Addons,
		})
	}

	resolved, err := s.resolveLines(ctx, order.StoreID, order.MemberID > 0, estimateItems)
	if err != nil {
		return 0, err
	}
	marketing, err := s.buildPricingRule(ctx, order.StoreID, order.MemberID, int8(order.Channel), couponID, order.PointsAmount > 0, resolved)
	if err != nil {
		return 0, err
	}
	pricing := CalculatePricing(resolved, marketing.Rule)

	pointsUsed := 0
	if pricing.PointsAmount > 0 && marketing.Output.ExchangeRate > 0 {
		pointsUsed = int(pricing.PointsAmount * int64(marketing.Output.ExchangeRate) / 100)
	}

	// 明细与计价结果按行对应；行数对不上时只保留金额，备注与点菜人不硬塞
	items := make([]*orderModel.CaterOrderItemModel, 0, len(pricing.Lines))
	addons := make([]*orderModel.CaterOrderItemAddonModel, 0)
	aligned := len(pricing.Lines) == len(lines)
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
		}
		if aligned {
			item.Orderer = lines[index].Orderer
			item.Remark = lines[index].Remark
		}
		// 显式分配主键：加料要引用明细 ID，不能等插入时才生成
		item.ID = support.NewID()
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

	if err := s.orders.ReplaceItems(ctx, tx, order.ID, items, addons); err != nil {
		return 0, err
	}

	rows, err := s.orders.UpdateAmounts(ctx, tx, order.ID, map[string]any{
		"item_amount":          pricing.ItemAmount,
		"addon_amount":         pricing.AddonAmount,
		"line_discount_amount": pricing.LineDiscountAmount,
		"activity_amount":      pricing.ActivityAmount,
		"coupon_amount":        pricing.CouponAmount,
		"points_amount":        pricing.PointsAmount,
		"points_used":          pointsUsed,
		"pay_amount":           pricing.PayAmount,
		"discount_snapshot":    marshalSnapshot(pricing),
	})
	if err != nil {
		return 0, err
	}
	// 并发收款把订单置为已支付时这里会更新到 0 行，整笔作废避免改了明细却没改金额
	if rows == 0 {
		return 0, errs.New(orderErrs.OrderAlreadyPaid)
	}

	if err := s.orders.CreateLog(ctx, tx, orderRepository.NewOrderLogModel(
		order.ID, order.Status, order.Status, operatorType, operatorID, "收银员", reason,
	)); err != nil {
		return 0, err
	}

	return pricing.PayAmount, nil
}
