// order_pos_overview_service.go 提供门店端首页的经营概览。
// 与系统看板的区别：只统计当前门店、只统计当天，口径面向店员而不是管理层。
package service

import (
	"context"
	"time"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
)

// PosOverview 是门店端首页的经营概览，金额单位为分。
type PosOverview struct {
	// Date 统计日期（yyyy-MM-dd）
	Date string `json:"date"`
	// Revenue 今日营收
	Revenue int64 `json:"revenue"`
	// PaidOrders 今日已支付订单数
	PaidOrders int64 `json:"paidOrders"`
	// AvgPrice 客单价，无单时为 0
	AvgPrice int64 `json:"avgPrice"`
	// RefundAmount 今日退款额
	RefundAmount int64 `json:"refundAmount"`
	// PendingPay 待支付单量
	PendingPay int64 `json:"pendingPay"`
	// PendingAccept 待接单量
	PendingAccept int64 `json:"pendingAccept"`
	// Making 制作中单量
	Making int64 `json:"making"`
	// WaitingPickup 待取餐单量
	WaitingPickup int64 `json:"waitingPickup"`
}

// PosOverview 汇总当前门店今日的经营数据。
// 入参：门店与操作人（用于数据权限校验）；返回营收、客单价、退款与各状态单量。
// 说明：只统计当天 0 点起的订单，跨天营业的门店需要另做营业日口径，一期不处理。
func (s *OrderService) PosOverview(ctx context.Context, storeID, userID types.ID) (*PosOverview, error) {
	if storeID == 0 {
		return nil, errs.New(orderErrs.OrderPosStoreRequired)
	}
	if userID > 0 {
		if err := s.ensureStoreVisible(ctx, userID, storeID); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	stats, err := s.orders.StoreDailyStats(ctx, storeID, start, start.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	out := &PosOverview{
		Date:          start.Format("2006-01-02"),
		Revenue:       stats.Revenue,
		PaidOrders:    stats.PaidOrders,
		RefundAmount:  stats.RefundAmount,
		PendingPay:    stats.StatusCounts[int16(orderEnums.OrderStatusPendingPay)],
		PendingAccept: stats.StatusCounts[int16(orderEnums.OrderStatusPendingAccept)],
		Making:        stats.StatusCounts[int16(orderEnums.OrderStatusMaking)],
		WaitingPickup: stats.StatusCounts[int16(orderEnums.OrderStatusWaitingPickup)],
	}
	// 客单价只按已支付单算，避免待支付单把均值拉低
	if stats.PaidOrders > 0 {
		out.AvgPrice = stats.Revenue / stats.PaidOrders
	}

	return out, nil
}
