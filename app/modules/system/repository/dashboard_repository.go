// Package repository 中的数据访问实现。
package repository

import (
	"context"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
)

// 控制台是聚合视图，允许跨模块只读查询；写操作仍归各自模块。
const (
	// 注意用户表是复数 system_users，角色与菜单是单数
	dashboardUserTable      = "system_users"
	dashboardRoleTable      = "system_role"
	dashboardMenuTable      = "system_menu"
	dashboardFileTable      = "infra_file"
	dashboardLoginLogTable  = "auth_login_log"
	dashboardAccessLogTable = "infra_api_access_log"
	dashboardErrorLogTable  = "infra_api_error_log"
)

// TrendRow 是趋势查询的原始行。
type TrendRow struct {
	Day   string `db:"day" json:"day"`
	Total int64  `db:"total" json:"total"`
}

// RecentLogin 是最近登录记录。
type RecentLogin struct {
	Username  string         `db:"username" json:"username"`
	UserIP    string         `db:"user_ip" json:"userIp"`
	Result    int8           `db:"result" json:"result"`
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
}

// fileStatsRow 是文件占用空间的聚合结果行。
type fileStatsRow struct {
	Size int64 `db:"size" json:"size"`
}

// DashboardRepository 提供控制台统计查询。
type DashboardRepository struct{}

// NewDashboardRepository 构造控制台统计仓储；聚合查询无状态，不需要注入依赖。
func NewDashboardRepository() *DashboardRepository {
	return &DashboardRepository{}
}

// CountUsers 统计未删除的用户数。
func (r *DashboardRepository) CountUsers(ctx context.Context) (int64, error) {
	return facades.DB().WithContext(ctx).Table(dashboardUserTable).WhereNull("deleted_at").Count()
}

// CountRoles 统计未删除的角色数。
func (r *DashboardRepository) CountRoles(ctx context.Context) (int64, error) {
	return facades.DB().WithContext(ctx).Table(dashboardRoleTable).WhereNull("deleted_at").Count()
}

// CountMenus 统计未删除的菜单数。
func (r *DashboardRepository) CountMenus(ctx context.Context) (int64, error) {
	return facades.DB().WithContext(ctx).Table(dashboardMenuTable).WhereNull("deleted_at").Count()
}

// CountLoginsSince 统计指定时间之后的成功登录次数。
func (r *DashboardRepository) CountLoginsSince(ctx context.Context, since time.Time) (int64, error) {
	return facades.DB().WithContext(ctx).Table(dashboardLoginLogTable).
		Where("type", 1).
		Where("result", 1).
		Where("created_at >= ?", since).
		Count()
}

// CountRequestsSince 统计指定时间之后的 API 请求数。
func (r *DashboardRepository) CountRequestsSince(ctx context.Context, since time.Time) (int64, error) {
	return facades.DB().WithContext(ctx).Table(dashboardAccessLogTable).
		Where("created_at >= ?", since).
		Count()
}

// CountErrorsSince 统计指定时间之后的 API 错误数。
func (r *DashboardRepository) CountErrorsSince(ctx context.Context, since time.Time) (int64, error) {
	return facades.DB().WithContext(ctx).Table(dashboardErrorLogTable).
		Where("created_at >= ?", since).
		Count()
}

// FileStats 返回文件数量与占用空间（字节）。
func (r *DashboardRepository) FileStats(ctx context.Context) (int64, int64, error) {
	count, err := facades.DB().WithContext(ctx).Table(dashboardFileTable).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return 0, 0, err
	}

	// 空表时 SUM(size) 返回 NULL，直接扫进 int64 会报错，用 COALESCE 兜底成 0。
	// 聚合结果只有一行，用 First 按结构体扫描。
	var row fileStatsRow
	if err := facades.DB().WithContext(ctx).Table(dashboardFileTable).
		WhereNull("deleted_at").
		Select("COALESCE(SUM(size), 0) AS size").
		First(&row); err != nil {
		return 0, 0, err
	}

	return count, row.Size, nil
}

// LoginTrend 按天统计成功登录次数。
func (r *DashboardRepository) LoginTrend(ctx context.Context, since time.Time) ([]TrendRow, error) {
	rows := make([]TrendRow, 0)
	err := facades.DB().WithContext(ctx).Table(dashboardLoginLogTable).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') AS day, COUNT(*) AS total").
		Where("type", 1).
		Where("result", 1).
		Where("created_at >= ?", since).
		GroupBy("day").
		OrderBy("day").
		Get(&rows)

	return rows, err
}

// RequestTrend 按天统计 API 请求次数。
func (r *DashboardRepository) RequestTrend(ctx context.Context, since time.Time) ([]TrendRow, error) {
	rows := make([]TrendRow, 0)
	err := facades.DB().WithContext(ctx).Table(dashboardAccessLogTable).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') AS day, COUNT(*) AS total").
		Where("created_at >= ?", since).
		GroupBy("day").
		OrderBy("day").
		Get(&rows)

	return rows, err
}

// RecentLogins 返回最近若干条登录记录。
func (r *DashboardRepository) RecentLogins(ctx context.Context, limit int) ([]RecentLogin, error) {
	rows := make([]RecentLogin, 0)
	err := facades.DB().WithContext(ctx).Table(dashboardLoginLogTable).
		Select("username, user_ip, result, created_at").
		Where("type", 1).
		OrderByDesc("id").
		Limit(uint64(limit)).
		Get(&rows)

	return rows, err
}
