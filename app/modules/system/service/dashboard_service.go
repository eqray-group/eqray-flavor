package service

import (
	"context"
	"runtime"
	"time"

	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/repository"
)

const (
	// dashboardTrendDays 是趋势图的统计天数。
	dashboardTrendDays = 7
	// dashboardRecentLogins 是最近登录记录的条数。
	dashboardRecentLogins = 6
	// dashboardTopProducts 是热销榜返回的菜品条数。
	dashboardTopProducts = 5
)

// DashboardCounts 是控制台的统计卡片数据。
type DashboardCounts struct {
	User         int64 `json:"user"`
	Role         int64 `json:"role"`
	Menu         int64 `json:"menu"`
	File         int64 `json:"file"`
	FileSize     int64 `json:"fileSize"`
	TodayLogin   int64 `json:"todayLogin"`
	TodayRequest int64 `json:"todayRequest"`
	TodayError   int64 `json:"todayError"`
}

// DashboardTrendItem 是趋势图的一个数据点。
type DashboardTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// DashboardLoginItem 是最近登录记录。
type DashboardLoginItem struct {
	Username  string         `json:"username"`
	UserIP    string         `json:"userIp"`
	Result    int8           `json:"result"`
	CreatedAt types.DateTime `json:"createdAt"`
}

// DashboardSystem 是系统运行信息。
type DashboardSystem struct {
	GoVersion  string `json:"goVersion"`
	ServerTime string `json:"serverTime"`
}

// DashboardBusinessTrendItem 是经营趋势图的一个数据点。
type DashboardBusinessTrendItem struct {
	// Date 月-日，如 10-19
	Date string `json:"date"`
	// OrderCount 当日下单量，含未支付
	OrderCount int64 `json:"orderCount"`
	// SaleAmount 当日营业额（分），只统计已支付订单
	SaleAmount int64 `json:"saleAmount"`
}

// DashboardTopProductItem 是热销榜的一条菜品数据。
type DashboardTopProductItem struct {
	// Name 菜品名称
	Name string `json:"name"`
	// Quantity 销量
	Quantity int64 `json:"quantity"`
	// Amount 销售额（分）
	Amount int64 `json:"amount"`
}

// DashboardBusiness 是控制台的经营看板数据。
// 与 Counts 里的运维指标分开返回，首页据此渲染营业额、客单价与热销榜。
type DashboardBusiness struct {
	// TodayOrderCount 今日下单量，含未支付
	TodayOrderCount int64 `json:"todayOrderCount"`
	// TodayPaidCount 今日已支付订单数
	TodayPaidCount int64 `json:"todayPaidCount"`
	// TodaySaleAmount 今日营业额（分）
	TodaySaleAmount int64 `json:"todaySaleAmount"`
	// TodayRefundAmount 今日退款额（分）
	TodayRefundAmount int64 `json:"todayRefundAmount"`
	// TodayAvgAmount 今日客单价（分）= 营业额 / 已支付订单数
	TodayAvgAmount int64 `json:"todayAvgAmount"`
	// OrderTrend 近 7 天订单量与营业额
	OrderTrend []DashboardBusinessTrendItem `json:"orderTrend"`
	// TopProducts 近 7 天热销菜品
	TopProducts []DashboardTopProductItem `json:"topProducts"`
}

// DashboardOverview 是控制台首页的聚合数据。
type DashboardOverview struct {
	Counts       DashboardCounts      `json:"counts"`
	Business     DashboardBusiness    `json:"business"`
	LoginTrend   []DashboardTrendItem `json:"loginTrend"`
	RequestTrend []DashboardTrendItem `json:"requestTrend"`
	RecentLogins []DashboardLoginItem `json:"recentLogins"`
	System       DashboardSystem      `json:"system"`
}

// DashboardService 汇总控制台数据。
type DashboardService struct {
	repo *repository.DashboardRepository
}

func NewDashboardService() *DashboardService {
	return &DashboardService{repo: repository.NewDashboardRepository()}
}

// Overview 返回控制台首页所需的全部统计。
func (s *DashboardService) Overview(ctx context.Context) (*DashboardOverview, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	trendStart := todayStart.AddDate(0, 0, -(dashboardTrendDays - 1))

	userCount, err := s.repo.CountUsers(ctx)
	if err != nil {
		return nil, err
	}

	roleCount, err := s.repo.CountRoles(ctx)
	if err != nil {
		return nil, err
	}

	menuCount, err := s.repo.CountMenus(ctx)
	if err != nil {
		return nil, err
	}

	fileCount, fileSize, err := s.repo.FileStats(ctx)
	if err != nil {
		return nil, err
	}

	todayLogin, err := s.repo.CountLoginsSince(ctx, todayStart)
	if err != nil {
		return nil, err
	}

	todayRequest, err := s.repo.CountRequestsSince(ctx, todayStart)
	if err != nil {
		return nil, err
	}

	todayError, err := s.repo.CountErrorsSince(ctx, todayStart)
	if err != nil {
		return nil, err
	}

	loginRows, err := s.repo.LoginTrend(ctx, trendStart)
	if err != nil {
		return nil, err
	}

	requestRows, err := s.repo.RequestTrend(ctx, trendStart)
	if err != nil {
		return nil, err
	}

	recentRows, err := s.repo.RecentLogins(ctx, dashboardRecentLogins)
	if err != nil {
		return nil, err
	}

	business, err := s.business(ctx, todayStart, trendStart)
	if err != nil {
		return nil, err
	}

	return &DashboardOverview{
		Counts: DashboardCounts{
			User:         userCount,
			Role:         roleCount,
			Menu:         menuCount,
			File:         fileCount,
			FileSize:     fileSize,
			TodayLogin:   todayLogin,
			TodayRequest: todayRequest,
			TodayError:   todayError,
		},
		Business:     *business,
		LoginTrend:   fillTrend(loginRows, dashboardTrendDays),
		RequestTrend: fillTrend(requestRows, dashboardTrendDays),
		RecentLogins: toLoginItems(recentRows),
		System: DashboardSystem{
			GoVersion:  runtime.Version(),
			ServerTime: now.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// business 汇总经营看板数据：今日营业额、客单价、近 7 天趋势与热销榜。
// 入参：今日零点与趋势起点；返回各项统计，查询失败时返回错误。
// 说明：营业额只统计已支付订单，客单价 = 营业额 / 已支付订单数，除零时返回 0。
func (s *DashboardService) business(ctx context.Context, todayStart, trendStart time.Time) (*DashboardBusiness, error) {
	summary, err := s.repo.BusinessSummary(ctx, todayStart, todayStart.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	trendRows, err := s.repo.BusinessTrend(ctx, trendStart)
	if err != nil {
		return nil, err
	}

	topRows, err := s.repo.TopProducts(ctx, trendStart, dashboardTopProducts)
	if err != nil {
		return nil, err
	}

	avgAmount := int64(0)
	if summary.PaidCount > 0 {
		avgAmount = summary.SaleAmount / summary.PaidCount
	}

	topProducts := make([]DashboardTopProductItem, 0, len(topRows))
	for _, row := range topRows {
		topProducts = append(topProducts, DashboardTopProductItem{
			Name:     row.Name,
			Quantity: row.Quantity,
			Amount:   row.Amount,
		})
	}

	return &DashboardBusiness{
		TodayOrderCount:   summary.OrderCount,
		TodayPaidCount:    summary.PaidCount,
		TodaySaleAmount:   summary.SaleAmount,
		TodayRefundAmount: summary.RefundAmount,
		TodayAvgAmount:    avgAmount,
		OrderTrend:        fillBusinessTrend(trendRows, dashboardTrendDays),
		TopProducts:       topProducts,
	}, nil
}

// fillTrend 把稀疏的按日统计补齐为连续日期序列，缺失日期记 0。
func fillTrend(rows []repository.TrendRow, days int) []DashboardTrendItem {
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Day] = row.Total
	}

	items := make([]DashboardTrendItem, 0, days)
	for i := days - 1; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i)
		items = append(items, DashboardTrendItem{
			Date:  day.Format("01-02"),
			Count: counts[day.Format("2006-01-02")],
		})
	}

	return items
}

// fillBusinessTrend 把稀疏的按日经营统计补齐为连续日期序列，缺失日期记 0。
// 入参：仓储返回的按日聚合行与天数；返回按日期升序的趋势点。
func fillBusinessTrend(rows []repository.BusinessTrendRow, days int) []DashboardBusinessTrendItem {
	orderCounts := make(map[string]int64, len(rows))
	saleAmounts := make(map[string]int64, len(rows))
	for _, row := range rows {
		orderCounts[row.Day] = row.Total
		saleAmounts[row.Day] = row.Amount
	}

	items := make([]DashboardBusinessTrendItem, 0, days)
	for i := days - 1; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i)
		key := day.Format("2006-01-02")
		items = append(items, DashboardBusinessTrendItem{
			Date:       day.Format("01-02"),
			OrderCount: orderCounts[key],
			SaleAmount: saleAmounts[key],
		})
	}

	return items
}

// toLoginItems 转换最近登录记录。
func toLoginItems(rows []repository.RecentLogin) []DashboardLoginItem {
	items := make([]DashboardLoginItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, DashboardLoginItem{
			Username:  row.Username,
			UserIP:    row.UserIP,
			Result:    row.Result,
			CreatedAt: row.CreatedAt,
		})
	}

	return items
}
