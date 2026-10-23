// Package repository 中的数据访问实现。
package repository

import (
	"context"
	"fmt"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
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
	// 经营看板只读交易域订单表，与上面的跨模块只读口径一致
	dashboardOrderTable     = "cater_order"
	dashboardOrderItemTable = "cater_order_item"
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

// BusinessSummaryRow 是订单经营汇总的聚合行。
type BusinessSummaryRow struct {
	// OrderCount 区间内下单量，含未支付
	OrderCount int64 `db:"order_count"`
	// PaidCount 区间内已支付订单数
	PaidCount int64 `db:"paid_count"`
	// SaleAmount 区间内营业额（分），只统计已支付订单
	SaleAmount int64 `db:"sale_amount"`
	// RefundAmount 区间内退款额（分）
	RefundAmount int64 `db:"refund_amount"`
}

// BusinessTrendRow 是按日经营趋势行。
type BusinessTrendRow struct {
	// Day 日期，格式 2006-01-02
	Day string `db:"day"`
	// Total 当日下单量，含未支付
	Total int64 `db:"total"`
	// Amount 当日营业额（分），只统计已支付订单
	Amount int64 `db:"amount"`
}

// TopProductRow 是热销菜品聚合行。
type TopProductRow struct {
	// Name 菜品名称快照
	Name string `db:"name"`
	// Quantity 销量
	Quantity int64 `db:"quantity"`
	// Amount 销售额（分）
	Amount int64 `db:"amount"`
}

// BusinessSummary 统计时间区间内的下单量、已支付单量、营业额与退款额。
// 入参：区间起点（含）与终点（不含）；返回聚合结果，空区间时各项为 0。
// 说明：营业额只认已支付订单（pay_status = 10），与财务对账口径一致。
func (r *DashboardRepository) BusinessSummary(ctx context.Context, start, end time.Time) (*BusinessSummaryRow, error) {
	paid := int8(orderEnums.PayStatusPaid)

	// Select 只接受列名表达式，带参数的占位符无法下传，这里把状态常量拼进语句
	var row BusinessSummaryRow
	err := facades.DB().WithContext(ctx).Table(dashboardOrderTable).
		Select(fmt.Sprintf(
			"COUNT(*) AS order_count, "+
				"COALESCE(SUM(CASE WHEN pay_status = %d THEN 1 ELSE 0 END), 0) AS paid_count, "+
				"COALESCE(SUM(CASE WHEN pay_status = %d THEN pay_amount ELSE 0 END), 0) AS sale_amount, "+
				"COALESCE(SUM(refund_amount), 0) AS refund_amount",
			paid, paid,
		)).
		WhereNull("deleted_at").
		Where("created_at >= ?", start).
		Where("created_at < ?", end).
		First(&row)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

// BusinessTrend 按天统计下单量与营业额，供经营趋势图使用。
func (r *DashboardRepository) BusinessTrend(ctx context.Context, since time.Time) ([]BusinessTrendRow, error) {
	paid := int8(orderEnums.PayStatusPaid)

	rows := make([]BusinessTrendRow, 0)
	err := facades.DB().WithContext(ctx).Table(dashboardOrderTable).
		Select(fmt.Sprintf(
			"DATE_FORMAT(created_at, '%%Y-%%m-%%d') AS day, COUNT(*) AS total, "+
				"COALESCE(SUM(CASE WHEN pay_status = %d THEN pay_amount ELSE 0 END), 0) AS amount",
			paid,
		)).
		WhereNull("deleted_at").
		Where("created_at >= ?", since).
		GroupBy("day").
		OrderBy("day").
		Get(&rows)

	return rows, err
}

// TopProducts 统计已支付订单里的热销菜品，按销量倒序。
// 入参：统计起点、条数；返回菜品名称、销量与销售额。
func (r *DashboardRepository) TopProducts(ctx context.Context, since time.Time, limit int) ([]TopProductRow, error) {
	paid := int8(orderEnums.PayStatusPaid)

	rows := make([]TopProductRow, 0)
	err := facades.DB().WithContext(ctx).Table(dashboardOrderItemTable+" AS i").
		Select("i.spu_name AS name, COALESCE(SUM(i.quantity), 0) AS quantity, COALESCE(SUM(i.subtotal), 0) AS amount").
		Join("cater_order AS o ON o.id = i.order_id").
		WhereNull("i.deleted_at").
		WhereNull("o.deleted_at").
		Where("o.pay_status", paid).
		Where("o.created_at >= ?", since).
		GroupBy("i.spu_name").
		OrderByDesc("quantity").
		Limit(uint64(limit)).
		Get(&rows)

	return rows, err
}
