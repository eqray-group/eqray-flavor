// order_admin_service.go 实现订单的后台经营操作：备注、改价、派单与小票重打、导出。
// 这些动作都不改变订单状态机，只补齐店员日常经营需要的字段与视图。
package service

import (
	"context"
	"strconv"
	"strings"
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
	storeEnums "eqray-flavor/app/modules/store/enums"
	storeRepository "eqray-flavor/app/modules/store/repository"
)

const (
	// orderRemarkMaxLength 是备注长度上限，与数据库列宽一致。
	orderRemarkMaxLength = 255
)

// UpdateRemarkInput 是修改订单备注的入参。
type UpdateRemarkInput struct {
	// OrderID 订单 ID
	OrderID types.ID
	// Remark 顾客备注，空字符串表示清空
	Remark string
	// MerchantRemark 商家备注，空字符串表示清空
	MerchantRemark string
	// UserID 当前后台用户，用于门店数据权限校验
	UserID types.ID
}

// AdjustPriceInput 是订单改价的入参，金额单位为分。
type AdjustPriceInput struct {
	// OrderID 订单 ID
	OrderID types.ID
	// Delta 改价差额（分）：正数加价、负数减免
	Delta int64
	// UserID 当前后台用户，用于门店数据权限校验
	UserID types.ID
}

// AssignRiderInput 是订单派单的入参。
type AssignRiderInput struct {
	// OrderID 订单 ID
	OrderID types.ID
	// RiderID 骑手 ID
	RiderID types.ID
	// UserID 当前后台用户，用于门店数据权限校验
	UserID types.ID
}

// OrderTicketAddon 是小票上的一条加料。
type OrderTicketAddon struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// OrderTicketItem 是小票上的一条菜品。
type OrderTicketItem struct {
	Name     string             `json:"name"`
	SkuName  string             `json:"skuName"`
	Quantity int                `json:"quantity"`
	UnitPrice int64             `json:"unitPrice"`
	Subtotal int64              `json:"subtotal"`
	Remark   string             `json:"remark"`
	Addons   []OrderTicketAddon `json:"addons"`
}

// OrderTicket 是打印小票所需的完整数据，前端或打印插件据此渲染。
type OrderTicket struct {
	// StoreName 门店名称
	StoreName string `json:"storeName"`
	// StoreShortName 门店简称，小票抬头用
	StoreShortName string `json:"storeShortName"`
	// OrderNo 订单号
	OrderNo string `json:"orderNo"`
	// PickupCode 取餐码
	PickupCode string `json:"pickupCode"`
	// TableName 桌台名称，非堂食为空
	TableName string `json:"tableName"`
	// TypeText 订单类型文案
	TypeText string `json:"typeText"`
	// ChannelText 下单渠道文案
	ChannelText string `json:"channelText"`
	// ReceiverName 收货人，外卖单使用
	ReceiverName string `json:"receiverName"`
	// ReceiverPhone 收货电话
	ReceiverPhone string `json:"receiverPhone"`
	// ReceiverAddress 收货地址
	ReceiverAddress string `json:"receiverAddress"`
	// Remark 顾客备注
	Remark string `json:"remark"`
	// MerchantRemark 商家备注
	MerchantRemark string `json:"merchantRemark"`
	// Items 菜品明细
	Items []OrderTicketItem `json:"items"`
	// ItemAmount 商品金额（分）
	ItemAmount int64 `json:"itemAmount"`
	// AddonAmount 加料费（分）
	AddonAmount int64 `json:"addonAmount"`
	// DiscountAmount 优惠合计（分）
	DiscountAmount int64 `json:"discountAmount"`
	// AdjustAmount 改价差额（分）
	AdjustAmount int64 `json:"adjustAmount"`
	// PayAmount 应付金额（分）
	PayAmount int64 `json:"payAmount"`
	// PayStatusText 支付状态文案
	PayStatusText string `json:"payStatusText"`
	// CreatedAt 下单时间
	CreatedAt string `json:"createdAt"`
	// PayTime 支付时间，未支付为空
	PayTime string `json:"payTime"`
}

// UpdateRemark 修改订单的顾客备注与商家备注。
// 入参：订单、备注内容与操作人；返回错误表示订单不存在或越权。
// 说明：只允许未取消、未退款的订单改备注；两个备注分列存放，互不覆盖。
func (s *OrderService) UpdateRemark(ctx context.Context, in UpdateRemarkInput) error {
	order, err := s.adminOrder(ctx, in.OrderID, in.UserID)
	if err != nil {
		return err
	}
	if order.Status.IsTerminal() {
		return errs.New(orderErrs.OrderStatusNotAdjustable, order.Status.String())
	}

	remark := strings.TrimSpace(in.Remark)
	merchantRemark := strings.TrimSpace(in.MerchantRemark)
	if len([]rune(remark)) > orderRemarkMaxLength || len([]rune(merchantRemark)) > orderRemarkMaxLength {
		return errs.WithMsg(errs.BadRequest, "备注不能超过 255 个字符")
	}

	_, err = s.orders.UpdateRemark(ctx, order.ID, remark, merchantRemark)

	return err
}

// AdjustPrice 对未支付订单改价：加收或减免，应付金额同步调整。
// 入参：订单、差额（分）与操作人；返回错误表示状态不允许或改价后金额为负。
// 约束：已支付订单不改价，避免与支付单金额不一致；差额累计写入 adjust_amount 便于对账。
func (s *OrderService) AdjustPrice(ctx context.Context, in AdjustPriceInput) error {
	if in.Delta == 0 {
		return errs.New(orderErrs.OrderAdjustAmountInvalid)
	}

	order, err := s.adminOrder(ctx, in.OrderID, in.UserID)
	if err != nil {
		return err
	}
	if order.PayStatus != orderEnums.PayStatusUnpaid {
		return errs.New(orderErrs.OrderPaidCannotAdjust)
	}
	if !order.Status.CanAdjust() {
		return errs.New(orderErrs.OrderStatusNotAdjustable, order.Status.String())
	}
	if order.PayAmount+in.Delta < 0 {
		return errs.New(orderErrs.OrderAdjustAmountInvalid)
	}

	rows, err := s.orders.AdjustPayAmount(ctx, order.ID, in.Delta)
	if err != nil {
		return err
	}
	if rows == 0 {
		// 并发下订单刚被支付，条件更新会失配
		return errs.New(orderErrs.OrderPaidCannotAdjust)
	}

	return nil
}

// AssignRider 给外卖订单派单：占用骑手接单额度并写入订单骑手。
// 入参：订单、骑手与操作人；返回错误表示订单类型/状态不允许、骑手不可用或已派单。
// 约束：骑手额度占用与订单更新在同一事务，任一失败整体回滚，不会出现「占了额度没派单」。
func (s *OrderService) AssignRider(ctx context.Context, in AssignRiderInput) error {
	order, err := s.adminOrder(ctx, in.OrderID, in.UserID)
	if err != nil {
		return err
	}
	if order.Type != orderEnums.OrderTypeTakeout {
		return errs.New(orderErrs.OrderNotDeliverable)
	}
	if !order.Status.CanDispatch() {
		return errs.New(orderErrs.OrderStatusNotDispatchable, order.Status.String())
	}
	if order.RiderID > 0 {
		return errs.New(orderErrs.OrderAlreadyDispatched)
	}

	rider, err := s.riders.FindByID(ctx, in.RiderID)
	if err != nil {
		return err
	}
	if rider == nil || rider.StoreID != order.StoreID {
		return errs.New(orderErrs.OrderRiderNotExists)
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		accepted, err := s.riders.TryAccept(ctx, rider.ID, time.Now())
		if err != nil {
			return err
		}
		if accepted == 0 {
			return errs.New(orderErrs.OrderRiderUnavailable)
		}

		rows, err := s.orders.AssignRider(ctx, order.ID, rider.ID, time.Now())
		if err != nil {
			return err
		}
		if rows == 0 {
			return errs.New(orderErrs.OrderAlreadyDispatched)
		}

		return nil
	})
}

// CancelDispatch 取消订单派单：清空订单骑手并归还骑手接单额度。
// 入参：订单与操作人；返回错误表示订单不存在或越权。未派单的订单直接返回成功。
func (s *OrderService) CancelDispatch(ctx context.Context, orderID, userID types.ID) error {
	order, err := s.adminOrder(ctx, orderID, userID)
	if err != nil {
		return err
	}
	if order.RiderID == 0 {
		return nil
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		rows, err := s.orders.ClearRider(ctx, order.ID)
		if err != nil {
			return err
		}
		if rows == 0 {
			return nil
		}

		_, err = s.riders.Release(ctx, order.RiderID)

		return err
	})
}

// OrderRiderOption 是可派单骑手的下拉项。
type OrderRiderOption struct {
	// RiderID 骑手档案 ID
	RiderID types.ID `json:"riderId"`
	// Name 骑手姓名，取自后台账号昵称
	Name string `json:"name"`
	// CurrentOrders 当前进行中配送单量
	CurrentOrders int `json:"currentOrders"`
	// MaxOrders 接单并发上限
	MaxOrders int `json:"maxOrders"`
	// Available 是否还有接单额度
	Available bool `json:"available"`
}

// DispatchableRiders 返回该订单门店下可接单的骑手，供派单弹窗选择。
// 入参：订单与当前用户；返回骑手列表（含当前负载），仅包含在线接单中的骑手。
func (s *OrderService) DispatchableRiders(ctx context.Context, orderID, userID types.ID) ([]OrderRiderOption, error) {
	order, err := s.adminOrder(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}
	if order.Type != orderEnums.OrderTypeTakeout {
		return nil, errs.New(orderErrs.OrderNotDeliverable)
	}

	status := storeEnums.RiderStatusAccepting
	riders, _, err := s.riders.Paginate(ctx, storeRepository.StoreRiderFilter{
		StoreID: order.StoreID,
		Status:  &status,
	}, 1, 100)
	if err != nil {
		return nil, err
	}

	options := make([]OrderRiderOption, 0, len(riders))
	if len(riders) == 0 {
		return options, nil
	}

	employeeIDs := make([]types.ID, 0, len(riders))
	for _, rider := range riders {
		employeeIDs = append(employeeIDs, rider.EmployeeID)
	}
	employees, err := s.employees.FindByIDs(ctx, employeeIDs)
	if err != nil {
		return nil, err
	}

	userIDs := make([]types.ID, 0, len(employees))
	for _, employee := range employees {
		userIDs = append(userIDs, employee.UserID)
	}
	users, err := support.FindUserSummaries(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	userNames := make(map[types.ID]string, len(users))
	for _, user := range users {
		userNames[user.ID] = user.Nickname
		if userNames[user.ID] == "" {
			userNames[user.ID] = user.Username
		}
	}

	employeeUsers := make(map[types.ID]types.ID, len(employees))
	for _, employee := range employees {
		employeeUsers[employee.ID] = employee.UserID
	}

	for _, rider := range riders {
		options = append(options, OrderRiderOption{
			RiderID:       rider.ID,
			Name:          userNames[employeeUsers[rider.EmployeeID]],
			CurrentOrders: rider.CurrentOrders,
			MaxOrders:     rider.MaxOrders,
			Available:     rider.CurrentOrders < rider.MaxOrders,
		})
	}

	return options, nil
}

// Ticket 组装打印小票数据，用于前台补打/重打。
// 入参：订单与当前用户；返回门店抬头、明细（含加料）与金额拆分。
func (s *OrderService) Ticket(ctx context.Context, orderID, userID types.ID) (*OrderTicket, error) {
	detail, err := s.DetailForUser(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}

	store, err := s.stores.FindByID(ctx, detail.StoreID)
	if err != nil {
		return nil, err
	}

	ticket := &OrderTicket{
		OrderNo:         detail.No,
		PickupCode:      detail.PickupCode,
		TableName:       detail.TableName,
		TypeText:        detail.TypeText,
		ChannelText:     detail.ChannelText,
		ReceiverName:    detail.ReceiverName,
		ReceiverPhone:   detail.ReceiverPhone,
		ReceiverAddress: detail.ReceiverAddress,
		Remark:          detail.Remark,
		MerchantRemark:  detail.MerchantRemark,
		ItemAmount:      detail.ItemAmount,
		AddonAmount:     detail.AddonAmount,
		DiscountAmount:  detail.LineDiscountAmount + detail.ActivityAmount + detail.CouponAmount + detail.PointsAmount,
		AdjustAmount:    detail.AdjustAmount,
		PayAmount:       detail.PayAmount,
		PayStatusText:   detail.PayStatusText,
		CreatedAt:       detail.CreatedAt.String(),
		Items:           make([]OrderTicketItem, 0, len(detail.Items)),
	}
	if store != nil {
		ticket.StoreName = store.Name
		ticket.StoreShortName = store.ShortName
	}
	if detail.PayTime != nil {
		ticket.PayTime = detail.PayTime.String()
	}

	for _, item := range detail.Items {
		row := OrderTicketItem{
			Name:      item.SpuName,
			SkuName:   item.SkuName,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal:  item.Subtotal,
			Remark:    item.Remark,
			Addons:    make([]OrderTicketAddon, 0, len(item.Addons)),
		}
		for _, addon := range item.Addons {
			row.Addons = append(row.Addons, OrderTicketAddon{Name: addon.AddonName, Quantity: addon.Quantity})
		}
		ticket.Items = append(ticket.Items, row)
	}

	return ticket, nil
}

// Export 按筛选条件导出订单为 CSV（UTF-8 BOM，Excel 可直接打开）。
// 入参：筛选条件；返回文件字节与建议文件名，超过上限时只导出最近的订单。
func (s *OrderService) Export(ctx context.Context, q PageQuery) ([]byte, error) {
	filter := orderRepository.OrderListFilter{
		StoreID: q.StoreID, Status: q.Status, PayStatus: q.PayStatus,
		Type: q.Type, Keyword: q.Keyword, MemberID: q.MemberID,
	}

	if q.StoreID == nil || *q.StoreID <= 0 {
		scope, err := s.scope(ctx, q.UserID)
		if err != nil {
			return nil, err
		}
		if !scope.All {
			filter.StoreIDs = scope.IDs
		}
	} else if q.UserID > 0 {
		if err := s.ensureStoreVisible(ctx, q.UserID, *q.StoreID); err != nil {
			return nil, err
		}
	}

	list, err := s.orders.ListForExport(ctx, filter, support.ExportRowLimit)
	if err != nil {
		return nil, err
	}

	header := []string{
		"订单号", "门店ID", "订单类型", "下单渠道", "订单状态", "支付状态",
		"桌台/收货人", "收货电话", "商品金额(元)", "加料费(元)", "优惠合计(元)",
		"改价(元)", "应付(元)", "已退(元)", "顾客备注", "商家备注", "下单时间", "支付时间",
	}
	rows := make([][]string, 0, len(list))
	for i := range list {
		order := &list[i]
		rows = append(rows, []string{
			order.No,
			strconv.FormatInt(int64(order.StoreID), 10),
			order.Type.String(),
			order.Channel.String(),
			order.Status.String(),
			order.PayStatus.String(),
			pickupOrReceiver(order),
			order.ReceiverPhone,
			support.FormatYuan(order.ItemAmount),
			support.FormatYuan(order.AddonAmount),
			support.FormatYuan(order.LineDiscountAmount + order.ActivityAmount + order.CouponAmount + order.PointsAmount),
			support.FormatYuan(order.AdjustAmount),
			support.FormatYuan(order.PayAmount),
			support.FormatYuan(order.RefundAmount),
			order.Remark,
			order.MerchantRemark,
			order.CreatedAt.String(),
			formatTime(order.PayTime),
		})
	}

	return support.BuildCSV(header, rows)
}

// adminOrder 读取订单并校验门店数据权限，供后台写操作共用。
func (s *OrderService) adminOrder(ctx context.Context, id, userID types.ID) (*orderModel.CaterOrderModel, error) {
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

	return order, nil
}

// pickupOrReceiver 返回小票/导出里的「桌台或收货人」列。
func pickupOrReceiver(order *orderModel.CaterOrderModel) string {
	if order.TableTitle != "" {
		return order.TableTitle
	}

	return order.ReceiverName
}

// formatTime 格式化可空时间，未支付时留空。
func formatTime(t *types.DateTime) string {
	if t == nil || t.IsZero() {
		return ""
	}

	return t.String()
}
