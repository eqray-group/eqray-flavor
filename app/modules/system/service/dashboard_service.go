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

// DashboardOverview 是控制台首页的聚合数据。
type DashboardOverview struct {
	Counts       DashboardCounts      `json:"counts"`
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
		LoginTrend:   fillTrend(loginRows, dashboardTrendDays),
		RequestTrend: fillTrend(requestRows, dashboardTrendDays),
		RecentLogins: toLoginItems(recentRows),
		System: DashboardSystem{
			GoVersion:  runtime.Version(),
			ServerTime: now.Format("2006-01-02 15:04:05"),
		},
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
