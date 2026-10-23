// order_pos_report_service.go 提供门店端经营报表。
// 口径与首页概览保持一致：只统计已支付订单，金额单位为分。
package service

import (
	"context"
	"fmt"
	"time"

	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
	orderErrs "eqray-flavor/app/modules/order/errs"
)

// 报表区间与排行的取值范围，防止一次拉出整年的数据把库拖住。
const (
	// reportMinDays 是最短统计天数
	reportMinDays = 1
	// reportMaxDays 是最长统计天数
	reportMaxDays = 31
	// reportMinTop 是排行榜最少条数
	reportMinTop = 1
	// reportMaxTop 是排行榜最多条数
	reportMaxTop = 50
	// reportRefundLimit 是退款明细最多返回条数
	reportRefundLimit = 50
)

// PosTrendRow 是报表里的一天。
type PosTrendRow struct {
	// Date 日期，格式 yyyy-MM-dd
	Date string `json:"date"`
	// Orders 当日已支付订单数
	Orders int64 `json:"orders"`
	// Revenue 当日营收（分）
	Revenue int64 `json:"revenue"`
}

// PosDishRow 是菜品销量排行的一行。
type PosDishRow struct {
	SpuName  string `json:"spuName"`
	SkuName  string `json:"skuName"`
	Quantity int64  `json:"quantity"`
	Amount   int64  `json:"amount"`
}

// PosHourRow 是按小时统计的一行。
type PosHourRow struct {
	// Hour 小时，两位数字，如 08
	Hour string `json:"hour"`
	// Orders 该小时订单数
	Orders int64 `json:"orders"`
	// Revenue 该小时营收（分）
	Revenue int64 `json:"revenue"`
}

// PosRefundRow 是退款明细的一行。
type PosRefundRow struct {
	OrderNo string `json:"orderNo"`
	// PayAmount 订单应收（分）
	PayAmount int64 `json:"payAmount"`
	// RefundAmount 已退金额（分）
	RefundAmount int64 `json:"refundAmount"`
	Status       int16 `json:"status"`
	StatusText   string `json:"statusText"`
	// UpdatedAt 最后更新时间；订单表没有退款时间字段，以它作为近似
	UpdatedAt string `json:"updatedAt"`
}

// PosReport 是门店经营报表。
type PosReport struct {
	// Days 统计天数
	Days int `json:"days"`
	// From / To 统计区间（yyyy-MM-dd），To 为今天
	From string `json:"from"`
	To   string `json:"to"`
	// Revenue 区间总营收（分）
	Revenue int64 `json:"revenue"`
	// Orders 区间已支付订单数
	Orders int64 `json:"orders"`
	// AvgPrice 客单价（分）
	AvgPrice int64 `json:"avgPrice"`
	// RefundAmount 区间退款额（分）
	RefundAmount int64 `json:"refundAmount"`
	// CashAmount / CashOrders 现金收款金额与单量
	CashAmount int64 `json:"cashAmount"`
	CashOrders int64 `json:"cashOrders"`
	// OnlineAmount / OnlineOrders 线上渠道收款金额与单量
	OnlineAmount int64 `json:"onlineAmount"`
	OnlineOrders int64 `json:"onlineOrders"`
	// Trend 按日趋势，缺失的日期补 0，保证前端图表连续
	Trend []PosTrendRow `json:"trend"`
	// Hourly 按时段分布，固定 24 条，没有订单的小时补 0
	Hourly []PosHourRow `json:"hourly"`
	// TopDishes 菜品销量排行
	TopDishes []PosDishRow `json:"topDishes"`
	// Refunds 区间内有退款的订单，最多 50 条
	Refunds []PosRefundRow `json:"refunds"`
}

// PosReport 汇总门店最近 days 天的经营报表。
// 入参：门店、操作人、统计天数与排行条数；返回趋势、收款构成与菜品排行。
// 说明：区间为「今天往前推 days-1 天的 0 点」到「明天 0 点」，含今天整天。
func (s *OrderService) PosReport(ctx context.Context, storeID, userID types.ID, days, top int) (*PosReport, error) {
	if storeID == 0 {
		return nil, errs.New(orderErrs.OrderPosStoreRequired)
	}
	if userID > 0 {
		if err := s.ensureStoreVisible(ctx, userID, storeID); err != nil {
			return nil, err
		}
	}

	days = clampInt(days, reportMinDays, reportMaxDays)
	top = clampInt(top, reportMinTop, reportMaxTop)

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	from := today.AddDate(0, 0, -(days - 1))
	to := today.AddDate(0, 0, 1)

	stats, err := s.orders.StoreDailyStats(ctx, storeID, from, to)
	if err != nil {
		return nil, err
	}
	trendRows, err := s.orders.DailyTrend(ctx, storeID, from, to)
	if err != nil {
		return nil, err
	}
	mix, err := s.orders.PaymentMix(ctx, storeID, from, to)
	if err != nil {
		return nil, err
	}
	dishes, err := s.orders.DishSales(ctx, storeID, from, to, top)
	if err != nil {
		return nil, err
	}
	hourlyRows, err := s.orders.HourlyStats(ctx, storeID, from, to)
	if err != nil {
		return nil, err
	}
	refundRows, err := s.orders.RefundOrders(ctx, storeID, from, to, reportRefundLimit)
	if err != nil {
		return nil, err
	}

	report := &PosReport{
		Days:         days,
		From:         from.Format("2006-01-02"),
		To:           today.Format("2006-01-02"),
		Revenue:      stats.Revenue,
		Orders:       stats.PaidOrders,
		RefundAmount: stats.RefundAmount,
		CashAmount:   mix.CashAmount,
		CashOrders:   mix.CashOrders,
		OnlineAmount: mix.OnlineAmount,
		OnlineOrders: mix.OnlineOrders,
		Trend:        make([]PosTrendRow, 0, days),
		Hourly:       make([]PosHourRow, 0, 24),
		TopDishes:    make([]PosDishRow, 0, len(dishes)),
		Refunds:      make([]PosRefundRow, 0, len(refundRows)),
	}
	if stats.PaidOrders > 0 {
		report.AvgPrice = stats.Revenue / stats.PaidOrders
	}

	// 没有订单的日期也要出现在趋势里，否则折线 / 柱子会缺格
	byDay := make(map[string]trendRow, len(trendRows))
	for _, row := range trendRows {
		byDay[row.Day] = trendRow{Orders: row.Orders, Revenue: row.Revenue}
	}
	for i := 0; i < days; i++ {
		day := from.AddDate(0, 0, i).Format("2006-01-02")
		hit := byDay[day]
		report.Trend = append(report.Trend, PosTrendRow{Date: day, Orders: hit.Orders, Revenue: hit.Revenue})
	}

	for _, dish := range dishes {
		report.TopDishes = append(report.TopDishes, PosDishRow{
			SpuName:  dish.SpuName,
			SkuName:  dish.SkuName,
			Quantity: dish.Quantity,
			Amount:   dish.Amount,
		})
	}

	// 时段固定输出 24 格，前端按格画柱，不需要自己补空
	byHour := make(map[string]hourRow, len(hourlyRows))
	for _, row := range hourlyRows {
		byHour[row.Hour] = hourRow{Orders: row.Orders, Revenue: row.Revenue}
	}
	for i := 0; i < 24; i++ {
		hour := fmt.Sprintf("%02d", i)
		hit := byHour[hour]
		report.Hourly = append(report.Hourly, PosHourRow{Hour: hour, Orders: hit.Orders, Revenue: hit.Revenue})
	}

	for _, row := range refundRows {
		status := orderEnums.OrderStatus(row.Status)
		report.Refunds = append(report.Refunds, PosRefundRow{
			OrderNo:      row.No,
			PayAmount:    row.PayAmount,
			RefundAmount: row.RefundAmount,
			Status:       row.Status,
			StatusText:   status.String(),
			UpdatedAt:    row.UpdatedAt.String(),
		})
	}

	return report, nil
}

// trendRow 是补齐趋势时用的中间结构。
type trendRow struct {
	Orders  int64
	Revenue int64
}

// hourRow 是补齐时段分布时用的中间结构。
type hourRow struct {
	Orders  int64
	Revenue int64
}

// clampInt 把数值限制在 [min, max] 区间内。
func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}

	return value
}
