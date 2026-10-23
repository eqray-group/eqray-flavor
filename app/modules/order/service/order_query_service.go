// order_query_service.go 实现订单查询：管理端分页列表与详情（明细、加料、日志）。
package service

import (
	"context"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
	orderModel "eqray-flavor/app/modules/order/model"
	orderRepository "eqray-flavor/app/modules/order/repository"
	storeService "eqray-flavor/app/modules/store/service"
)

// OrderLogRow 是订单日志的输出行。
type OrderLogRow struct {
	ID           types.ID       `json:"id"`
	FromStatus   int16          `json:"fromStatus"`
	ToStatus     int16          `json:"toStatus"`
	OperatorType int8           `json:"operatorType"`
	OperatorName string         `json:"operatorName"`
	Reason       string         `json:"reason"`
	CreatedAt    types.DateTime `json:"createdAt"`
}

// OrderAddonRow 是明细加料的输出行。
type OrderAddonRow struct {
	ID          types.ID `json:"id"`
	OrderItemID types.ID `json:"orderItemId"`
	AddonItemID types.ID `json:"addonItemId"`
	AddonName   string   `json:"addonName"`
	Price       int64    `json:"price"`
	Quantity    int      `json:"quantity"`
}

// OrderItemRow 是订单明细的输出行。
type OrderItemRow struct {
	ID                 types.ID        `json:"id"`
	SpuID              types.ID        `json:"spuId"`
	SkuID              types.ID        `json:"skuId"`
	SpuName            string          `json:"spuName"`
	SkuName            string          `json:"skuName"`
	OriginalPrice      int64           `json:"originalPrice"`
	UnitPrice          int64           `json:"unitPrice"`
	Quantity           int             `json:"quantity"`
	AddonAmount        int64           `json:"addonAmount"`
	ItemDiscountAmount int64           `json:"itemDiscountAmount"`
	Subtotal           int64           `json:"subtotal"`
	Status             int8            `json:"status"`
	Orderer            string          `json:"orderer"`
	Remark             string          `json:"remark"`
	Addons             []OrderAddonRow `json:"addons"`
}

// OrderRow 是订单的输出行（列表与详情共用，详情多明细与日志）。
type OrderRow struct {
	ID                 types.ID        `json:"id"`
	No                 string          `json:"no"`
	StoreID            types.ID        `json:"storeId"`
	MemberID           types.ID        `json:"memberId"`
	Type               int8            `json:"type"`
	TypeText           string          `json:"typeText"`
	Channel            int8            `json:"channel"`
	ChannelText        string          `json:"channelText"`
	TableID            types.ID        `json:"tableId"`
	TableName          string          `json:"tableName"`
	Status             int16           `json:"status"`
	StatusText         string          `json:"statusText"`
	PayStatus          int8            `json:"payStatus"`
	PayStatusText      string          `json:"payStatusText"`
	ItemAmount         int64           `json:"itemAmount"`
	AddonAmount        int64           `json:"addonAmount"`
	LineDiscountAmount int64           `json:"lineDiscountAmount"`
	ActivityAmount     int64           `json:"activityAmount"`
	CouponAmount       int64           `json:"couponAmount"`
	PointsAmount       int64           `json:"pointsAmount"`
	PayAmount          int64           `json:"payAmount"`
	RefundAmount       int64           `json:"refundAmount"`
	DiscountSnapshot   string          `json:"discountSnapshot"`
	PickupCode         string          `json:"pickupCode"`
	ReceiverName       string          `json:"receiverName"`
	ReceiverPhone      string          `json:"receiverPhone"`
	ReceiverAddress    string          `json:"receiverAddress"`
	Remark             string          `json:"remark"`
	MerchantRemark     string          `json:"merchantRemark"`
	AdjustAmount       int64           `json:"adjustAmount"`
	RiderID            types.ID        `json:"riderId"`
	DispatchTime       *types.DateTime `json:"dispatchTime"`
	PayTime            *types.DateTime `json:"payTime"`
	AcceptTime         *types.DateTime `json:"acceptTime"`
	CompleteTime       *types.DateTime `json:"completeTime"`
	CancelTime         *types.DateTime `json:"cancelTime"`
	CancelReason       string          `json:"cancelReason"`
	ExpireTime         types.DateTime  `json:"expireTime"`
	CreatedAt          types.DateTime  `json:"createdAt"`
}

// OrderDetail 是订单详情：订单 + 明细（含加料）+ 状态日志。
type OrderDetail struct {
	OrderRow
	Items []OrderItemRow `json:"items"`
	Logs  []OrderLogRow  `json:"logs"`
}

// PageQuery 是订单列表查询参数。
type PageQuery struct {
	StoreID   *types.ID
	Status    *orderEnums.OrderStatus
	PayStatus *orderEnums.PayStatus
	Type      *orderEnums.OrderType
	Keyword   string
	MemberID  *types.ID
	// UserID 当前后台用户，用于计算门店数据权限；0 表示不限制（内部调用）
	UserID types.ID
}

// Page 分页查询订单列表。
// 入参：筛选条件与分页参数；返回订单列表与总数。
// 说明：门店数据权限在此收口——按用户可见门店过滤，显式指定门店时以其为准但仍需校验可见性。
func (s *OrderService) Page(ctx context.Context, q PageQuery, pageNo, pageSize int) ([]OrderRow, int64, error) {
	filter := orderRepository.OrderListFilter{
		StoreID: q.StoreID, Status: q.Status, PayStatus: q.PayStatus,
		Type: q.Type, Keyword: q.Keyword, MemberID: q.MemberID,
	}

	if q.StoreID == nil || *q.StoreID <= 0 {
		scope, err := s.scope(ctx, q.UserID)
		if err != nil {
			return nil, 0, err
		}
		if !scope.All {
			filter.StoreIDs = scope.IDs
		}
	} else {
		// 显式指定门店时校验归属，防止越权查看别家门店的订单
		if q.UserID > 0 {
			if err := s.ensureStoreVisible(ctx, q.UserID, *q.StoreID); err != nil {
				return nil, 0, err
			}
		}
	}

	list, total, err := s.orders.Page(ctx, filter, pageNo, pageSize)
	if err != nil {
		return nil, 0, err
	}

	rows := make([]OrderRow, 0, len(list))
	for i := range list {
		rows = append(rows, buildOrderRow(&list[i]))
	}

	return rows, total, nil
}

// scope 计算当前用户可见的门店范围；内部调用（UserID 为 0）时不限制。
func (s *OrderService) scope(ctx context.Context, userID types.ID) (storeService.StoreScope, error) {
	if userID <= 0 {
		return storeService.StoreScope{All: true}, nil
	}

	return storeService.NewStoreService().Scope(ctx, userID)
}

// ensureStoreVisible 校验门店在用户可见范围内。
func (s *OrderService) ensureStoreVisible(ctx context.Context, userID, storeID types.ID) error {
	if userID <= 0 {
		return nil
	}

	return storeService.NewStoreService().EnsureStoreVisible(ctx, userID, storeID)
}

// MemberDetail 查询会员自己的订单详情。
// 入参：订单 ID 与会员 ID；非本人订单按不存在处理，避免枚举探测。
func (s *OrderService) MemberDetail(ctx context.Context, id, memberID types.ID) (*OrderDetail, error) {
	if memberID == 0 {
		return nil, errs.New(orderErrs.OrderNotExists)
	}

	order, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil || order.MemberID != memberID {
		return nil, errs.New(orderErrs.OrderNotExists)
	}

	return s.buildDetail(ctx, order)
}

// Detail 查询订单详情，含明细（含加料快照）与状态日志。
func (s *OrderService) Detail(ctx context.Context, id types.ID) (*OrderDetail, error) {
	order, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errs.New(orderErrs.OrderNotExists)
	}

	return s.buildDetail(ctx, order)
}

// DetailForUser 查询订单详情，并按门店数据权限校验归属。
// 入参：订单 ID 与当前用户 ID；越权访问按「订单不属于当前门店」拒绝。
func (s *OrderService) DetailForUser(ctx context.Context, id, userID types.ID) (*OrderDetail, error) {
	order, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errs.New(orderErrs.OrderNotExists)
	}
	if err := s.ensureStoreVisible(ctx, userID, order.StoreID); err != nil {
		return nil, errs.New(orderErrs.OrderStoreMismatch)
	}

	return s.buildDetail(ctx, order)
}

// buildDetail 按订单组装详情：明细（含加料快照）与状态日志。
func (s *OrderService) buildDetail(ctx context.Context, order *orderModel.CaterOrderModel) (*OrderDetail, error) {
	items, err := s.orders.ListItemsByOrderIDs(ctx, []types.ID{order.ID})
	if err != nil {
		return nil, err
	}
	addons, err := s.orders.ListAddonsByOrderIDs(ctx, []types.ID{order.ID})
	if err != nil {
		return nil, err
	}
	logs, err := s.orders.ListLogsByOrderID(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	detail := &OrderDetail{OrderRow: buildOrderRow(order), Items: make([]OrderItemRow, 0), Logs: make([]OrderLogRow, 0, len(logs))}
	for _, item := range items[order.ID] {
		row := OrderItemRow{
			ID: item.ID, SpuID: item.SpuID, SkuID: item.SkuID,
			SpuName: item.SpuName, SkuName: item.SkuName,
			OriginalPrice: item.OriginalPrice, UnitPrice: item.UnitPrice,
			Quantity: item.Quantity, AddonAmount: item.AddonAmount,
			ItemDiscountAmount: item.ItemDiscountAmount,
			Subtotal:           item.Subtotal, Status: item.Status,
			Orderer: item.Orderer, Remark: item.Remark,
			Addons: make([]OrderAddonRow, 0),
		}
		for _, addon := range addons[item.ID] {
			row.Addons = append(row.Addons, OrderAddonRow{
				ID: addon.ID, OrderItemID: addon.OrderItemID,
				AddonItemID: addon.AddonItemID, AddonName: addon.AddonName,
				Price: addon.Price, Quantity: addon.Quantity,
			})
		}
		detail.Items = append(detail.Items, row)
	}
	for _, log := range logs {
		detail.Logs = append(detail.Logs, OrderLogRow{
			ID: log.ID, FromStatus: log.FromStatus, ToStatus: log.ToStatus,
			OperatorType: log.OperatorType, OperatorName: log.OperatorName,
			Reason: log.Reason, CreatedAt: log.CreatedAt,
		})
	}

	return detail, nil
}

// buildOrderRow 把订单模型转成输出行。
func buildOrderRow(m *orderModel.CaterOrderModel) OrderRow {
	return OrderRow{
		ID: m.ID, No: m.No, StoreID: m.StoreID, MemberID: m.MemberID,
		Type: int8(m.Type), TypeText: m.Type.String(),
		Channel: int8(m.Channel), ChannelText: m.Channel.String(),
		TableID: m.TableID, TableName: m.TableTitle,
		Status: int16(m.Status), StatusText: m.Status.String(),
		PayStatus: int8(m.PayStatus), PayStatusText: m.PayStatus.String(),
		ItemAmount: m.ItemAmount, AddonAmount: m.AddonAmount,
		LineDiscountAmount: m.LineDiscountAmount,
		ActivityAmount:     m.ActivityAmount, CouponAmount: m.CouponAmount,
		PointsAmount: m.PointsAmount, PayAmount: m.PayAmount,
		RefundAmount: m.RefundAmount, DiscountSnapshot: m.DiscountSnapshot,
		PickupCode:   m.PickupCode,
		ReceiverName: m.ReceiverName, ReceiverPhone: m.ReceiverPhone, ReceiverAddress: m.ReceiverAddress,
		Remark:         m.Remark,
		MerchantRemark: m.MerchantRemark,
		AdjustAmount:   m.AdjustAmount,
		RiderID:        m.RiderID,
		DispatchTime:   m.DispatchTime,
		PayTime:        m.PayTime, AcceptTime: m.AcceptTime,
		CompleteTime: m.CompleteTime, CancelTime: m.CancelTime,
		CancelReason: m.CancelReason, ExpireTime: m.ExpireTime,
		CreatedAt: m.CreatedAt,
	}
}
