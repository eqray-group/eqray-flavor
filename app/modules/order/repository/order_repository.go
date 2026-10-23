// Package repository 定义 order 模块的数据访问层。
// 订单写路径全部走事务；状态流转用「带原状态的条件更新」防并发，影响 0 行即视为流转失败。
package repository

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/database/db"
	frameworkDB "github.com/goravel/framework/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	orderEnums "eqray-flavor/app/modules/order/enums"
	"eqray-flavor/app/modules/order/model"
)

// CaterOrderRepository 封装订单主表、明细、加料与日志的读写。
type CaterOrderRepository struct{}

// NewCaterOrderRepository 构造订单仓储。
func NewCaterOrderRepository() *CaterOrderRepository {
	return &CaterOrderRepository{}
}

// toAny 把 ID 列表转为查询构造器需要的 []any。
func toAny(ids []types.ID) []any {
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}

	return values
}

// CreateOrder 在事务内写入订单主表。
func (r *CaterOrderRepository) CreateOrder(ctx context.Context, tx db.Tx, m *model.CaterOrderModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(model.CaterOrderTable).Insert(map[string]any{
		"id":                   m.ID,
		"no":                   m.No,
		"store_id":             m.StoreID,
		"member_id":            m.MemberID,
		"type":                 int8(m.Type),
		"channel":              int8(m.Channel),
		"table_id":             m.TableID,
		"table_name":           m.TableTitle,
		"request_no":           m.RequestNo,
		"status":               int16(m.Status),
		"pay_status":           int8(m.PayStatus),
		"item_amount":          m.ItemAmount,
		"addon_amount":         m.AddonAmount,
		"line_discount_amount": m.LineDiscountAmount,
		"activity_amount":      m.ActivityAmount,
		"coupon_amount":        m.CouponAmount,
		"points_amount":        m.PointsAmount,
		"points_used":          m.PointsUsed,
		"pay_amount":           m.PayAmount,
		"refund_amount":        m.RefundAmount,
		"discount_snapshot":    m.DiscountSnapshot,
		"pickup_code":          m.PickupCode,
		"receiver_name":        m.ReceiverName,
		"receiver_phone":       m.ReceiverPhone,
		"receiver_address":     m.ReceiverAddress,
		"remark":               m.Remark,
		"pay_order_id":         m.PayOrderID,
		"pay_time":             m.PayTime,
		"accept_time":          m.AcceptTime,
		"complete_time":        m.CompleteTime,
		"cancel_time":          m.CancelTime,
		"cancel_reason":        m.CancelReason,
		"expire_time":          m.ExpireTime,
		"creator":              m.Creator,
		"created_at":           m.CreatedAt,
		"updater":              m.Updater,
		"updated_at":           m.UpdatedAt,
	})

	return err
}

// CreateItems 在事务内批量写入订单明细。
func (r *CaterOrderRepository) CreateItems(ctx context.Context, tx db.Tx, items []*model.CaterOrderItemModel) error {
	for _, item := range items {
		item.FillCreate(ctx)
		if _, err := tx.Table(model.CaterOrderItemTable).Insert(map[string]any{
			"id":                   item.ID,
			"order_id":             item.OrderID,
			"spu_id":               item.SpuID,
			"sku_id":               item.SkuID,
			"spu_name":             item.SpuName,
			"sku_name":             item.SkuName,
			"original_price":       item.OriginalPrice,
			"unit_price":           item.UnitPrice,
			"quantity":             item.Quantity,
			"addon_amount":         item.AddonAmount,
			"item_discount_amount": item.ItemDiscountAmount,
			"subtotal":             item.Subtotal,
			"status":               item.Status,
			"orderer":              item.Orderer,
			"remark":               item.Remark,
			"creator":              item.Creator,
			"created_at":           item.CreatedAt,
			"updater":              item.Updater,
			"updated_at":           item.UpdatedAt,
		}); err != nil {
			return err
		}
	}

	return nil
}

// CreateItemAddons 在事务内批量写入明细加料快照。
func (r *CaterOrderRepository) CreateItemAddons(ctx context.Context, tx db.Tx, addons []*model.CaterOrderItemAddonModel) error {
	for _, addon := range addons {
		addon.FillCreate(ctx)
		if _, err := tx.Table(model.CaterOrderItemAddonTable).Insert(map[string]any{
			"id":            addon.ID,
			"order_id":      addon.OrderID,
			"order_item_id": addon.OrderItemID,
			"addon_item_id": addon.AddonItemID,
			"addon_name":    addon.AddonName,
			"price":         addon.Price,
			"quantity":      addon.Quantity,
			"creator":       addon.Creator,
			"created_at":    addon.CreatedAt,
			"updater":       addon.Updater,
			"updated_at":    addon.UpdatedAt,
		}); err != nil {
			return err
		}
	}

	return nil
}

// CreateLog 在事务内写入订单状态日志，只增不改。
func (r *CaterOrderRepository) CreateLog(ctx context.Context, tx db.Tx, m *model.CaterOrderLogModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(model.CaterOrderLogTable).Insert(map[string]any{
		"id":            m.ID,
		"order_id":      m.OrderID,
		"from_status":   m.FromStatus,
		"to_status":     m.ToStatus,
		"operator_type": m.OperatorType,
		"operator_id":   m.OperatorID,
		"operator_name": m.OperatorName,
		"reason":        m.Reason,
		"creator":       m.Creator,
		"created_at":    m.CreatedAt,
	})

	return err
}

// find 按 query 取一条订单，未命中返回 (nil, nil)。
func (r *CaterOrderRepository) find(ctx context.Context, query db.Query) (*model.CaterOrderModel, error) {
	var m model.CaterOrderModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByID 按主键查询订单。
func (r *CaterOrderRepository) FindByID(ctx context.Context, id types.ID) (*model.CaterOrderModel, error) {
	query := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).Where("id", id)

	return r.find(ctx, query)
}

// FindByNo 按订单号查询订单。
func (r *CaterOrderRepository) FindByNo(ctx context.Context, no string) (*model.CaterOrderModel, error) {
	query := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).Where("no", no)

	return r.find(ctx, query)
}

// FindByIdempotency 按「门店 + 请求号」查询已有订单，用于下单幂等：命中直接返回旧单。
func (r *CaterOrderRepository) FindByIdempotency(ctx context.Context, storeID types.ID, requestNo string) (*model.CaterOrderModel, error) {
	query := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Where("store_id", storeID).
		Where("request_no", requestNo)

	return r.find(ctx, query)
}

// Transit 把订单从原状态流转到新状态，带原状态条件更新防并发；返回是否流转成功。
// 入参：事务、订单 ID、原状态、新状态与附加更新字段（如 pay_time / cancel_reason）。
func (r *CaterOrderRepository) Transit(ctx context.Context, tx db.Tx, id types.ID, from, to orderEnums.OrderStatus, extra map[string]any) (bool, error) {
	updates := map[string]any{
		"status":     int16(to),
		"updated_at": types.Now(),
	}
	for key, value := range extra {
		updates[key] = value
	}

	res, err := tx.Table(model.CaterOrderTable).
		Where("id", id).
		Where("status", int16(from)).
		WhereNull("deleted_at").
		Update(updates)
	if err != nil {
		return false, err
	}

	return res.RowsAffected > 0, nil
}

// ListExpiredPendingPay 查询已超时的待支付订单，供超时关单任务使用。
func (r *CaterOrderRepository) ListExpiredPendingPay(ctx context.Context, now time.Time, limit uint64) ([]model.CaterOrderModel, error) {
	list := make([]model.CaterOrderModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Where("status", int16(orderEnums.OrderStatusPendingPay)).
		Where("expire_time < ?", types.DateTime(now)).
		WhereNull("deleted_at").
		OrderBy("id").
		Limit(limit).
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// OrderListFilter 是订单管理端列表的筛选条件。
type OrderListFilter struct {
	// StoreID 门店，nil 表示当前用户可见全部门店
	StoreID *types.ID
	// StoreIDs 可见门店集合，非空时只返回这些门店的订单；空集合表示不限
	StoreIDs []types.ID
	// Status 订单状态，nil 表示不限
	Status *orderEnums.OrderStatus
	// PayStatus 支付状态，nil 表示不限
	PayStatus *orderEnums.PayStatus
	// Type 订单类型，nil 表示不限
	Type *orderEnums.OrderType
	// Keyword 订单号 / 取餐码 / 收货人电话模糊匹配
	Keyword string
	// MemberID 会员，nil 表示不限
	MemberID *types.ID
}

// Page 分页查询订单，返回列表与总数；明细不在此处加载。
func (r *CaterOrderRepository) Page(ctx context.Context, filter OrderListFilter, pageNo, pageSize int) ([]model.CaterOrderModel, int64, error) {
	query := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).WhereNull("deleted_at")
	if filter.StoreID != nil && *filter.StoreID > 0 {
		query = query.Where("store_id", *filter.StoreID)
	}
	// 门店数据权限：显式指定门店时以其为准，否则按可见门店集合过滤
	if filter.StoreID == nil || *filter.StoreID <= 0 {
		if len(filter.StoreIDs) > 0 {
			query = query.WhereIn("store_id", toAny(filter.StoreIDs))
		} else if filter.StoreIDs != nil {
			// 可见门店为空：说明该账号没有任何门店权限，不返回任何订单
			query = query.Where("id", 0)
		}
	}
	if filter.Status != nil {
		query = query.Where("status", int16(*filter.Status))
	}
	if filter.PayStatus != nil {
		query = query.Where("pay_status", int8(*filter.PayStatus))
	}
	if filter.Type != nil {
		query = query.Where("type", int8(*filter.Type))
	}
	if filter.MemberID != nil && *filter.MemberID > 0 {
		query = query.Where("member_id", *filter.MemberID)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("no", like).
				OrWhereLike("pickup_code", like).
				OrWhereLike("receiver_phone", like)
		})
	}

	list := make([]model.CaterOrderModel, 0)
	var total int64
	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ListItemsByOrderIDs 批量加载订单明细，按 order_id 分组返回。
func (r *CaterOrderRepository) ListItemsByOrderIDs(ctx context.Context, orderIDs []types.ID) (map[types.ID][]model.CaterOrderItemModel, error) {
	items := make([]model.CaterOrderItemModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.CaterOrderItemTable).
		WhereIn("order_id", toAny(orderIDs)).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&items); err != nil {
		return nil, err
	}

	grouped := make(map[types.ID][]model.CaterOrderItemModel, len(orderIDs))
	for _, item := range items {
		grouped[item.OrderID] = append(grouped[item.OrderID], item)
	}

	return grouped, nil
}

// ListAddonsByOrderIDs 批量加载订单明细加料快照，按 order_item_id 分组返回。
func (r *CaterOrderRepository) ListAddonsByOrderIDs(ctx context.Context, orderIDs []types.ID) (map[types.ID][]model.CaterOrderItemAddonModel, error) {
	addons := make([]model.CaterOrderItemAddonModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.CaterOrderItemAddonTable).
		WhereIn("order_id", toAny(orderIDs)).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&addons); err != nil {
		return nil, err
	}

	grouped := make(map[types.ID][]model.CaterOrderItemAddonModel, len(orderIDs))
	for _, addon := range addons {
		grouped[addon.OrderItemID] = append(grouped[addon.OrderItemID], addon)
	}

	return grouped, nil
}

// ListLogsByOrderID 加载订单状态日志，按时间正序。
func (r *CaterOrderRepository) ListLogsByOrderID(ctx context.Context, orderID types.ID) ([]model.CaterOrderLogModel, error) {
	logs := make([]model.CaterOrderLogModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.CaterOrderLogTable).
		Where("order_id", orderID).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// BindPayOrder 回写订单关联的支付单 ID。
// 入参：订单 ID 与支付单 ID；返回错误表示落库失败。
func (r *CaterOrderRepository) BindPayOrder(ctx context.Context, id, payOrderID types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"pay_order_id": payOrderID,
			"updated_at":   types.Now(),
		})

	return err
}

// RefundAccumulated 累加订单已退金额（分），返回影响行数；用于逆向流程记账。
func (r *CaterOrderRepository) RefundAccumulated(ctx context.Context, tx db.Tx, id types.ID, amount int64) (int64, error) {
	res, err := tx.Table(model.CaterOrderTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"refund_amount": frameworkDB.Raw("refund_amount + ?", amount),
			"updated_at":    types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// UpdateRemark 更新订单的顾客备注与商家备注，返回影响行数。
// 入参：订单 ID、顾客备注与商家备注；两者分列存放，店员改动不覆盖顾客原始诉求。
func (r *CaterOrderRepository) UpdateRemark(ctx context.Context, id types.ID, remark, merchantRemark string) (int64, error) {
	res, err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"remark":          remark,
			"merchant_remark": merchantRemark,
			"updated_at":      types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// AdjustPayAmount 对未支付订单改价：应付金额与改价差额同时累加，返回影响行数。
// 入参：订单 ID 与差额（分，正加价 / 负减免）。
// 约束：条件里带上 pay_status = 0，已支付订单不会被改动，返回 0 行由调用方报错。
func (r *CaterOrderRepository) AdjustPayAmount(ctx context.Context, id types.ID, delta int64) (int64, error) {
	res, err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Where("id", id).
		Where("pay_status", int8(orderEnums.PayStatusUnpaid)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"pay_amount":    frameworkDB.Raw("pay_amount + ?", delta),
			"adjust_amount": frameworkDB.Raw("adjust_amount + ?", delta),
			"updated_at":    types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// AssignRider 给订单派单：写入骑手与派单时间，返回影响行数。
// 入参：订单 ID、骑手 ID 与派单时间。
// 约束：条件里带上 rider_id = 0，重复派单返回 0 行，避免两名骑手同时接到同一单。
func (r *CaterOrderRepository) AssignRider(ctx context.Context, id, riderID types.ID, dispatchTime time.Time) (int64, error) {
	res, err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Where("id", id).
		Where("rider_id", 0).
		WhereNull("deleted_at").
		Update(map[string]any{
			"rider_id":      riderID,
			"dispatch_time": types.DateTime(dispatchTime),
			"updated_at":    types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// ClearRider 清空订单的骑手与派单时间，用于订单完成或取消后释放配送关系。
// 入参：订单 ID；返回影响行数。
func (r *CaterOrderRepository) ClearRider(ctx context.Context, id types.ID) (int64, error) {
	res, err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Where("id", id).
		Where("rider_id > 0").
		WhereNull("deleted_at").
		Update(map[string]any{
			"rider_id":      0,
			"dispatch_time": nil,
			"updated_at":    types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// ListForExport 按筛选条件取一批订单用于导出，不分页。
// 入参：筛选条件与最大条数；返回按 ID 倒序的订单列表。
// 说明：导出是后台运维动作，上限由调用方控制，避免一次拉空整张表。
func (r *CaterOrderRepository) ListForExport(ctx context.Context, filter OrderListFilter, limit int) ([]model.CaterOrderModel, error) {
	query := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).WhereNull("deleted_at")
	if filter.StoreID != nil && *filter.StoreID > 0 {
		query = query.Where("store_id", *filter.StoreID)
	}
	if filter.StoreID == nil || *filter.StoreID <= 0 {
		if len(filter.StoreIDs) > 0 {
			query = query.WhereIn("store_id", toAny(filter.StoreIDs))
		} else if filter.StoreIDs != nil {
			query = query.Where("id", 0)
		}
	}
	if filter.Status != nil {
		query = query.Where("status", int16(*filter.Status))
	}
	if filter.PayStatus != nil {
		query = query.Where("pay_status", int8(*filter.PayStatus))
	}
	if filter.Type != nil {
		query = query.Where("type", int8(*filter.Type))
	}
	if filter.MemberID != nil && *filter.MemberID > 0 {
		query = query.Where("member_id", *filter.MemberID)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("no", like).
				OrWhereLike("pickup_code", like).
				OrWhereLike("receiver_phone", like)
		})
	}

	list := make([]model.CaterOrderModel, 0)
	if err := query.OrderByDesc("id").Limit(uint64(limit)).Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// StoreDailyStats 是门店在指定区间内的经营汇总，供门店端首页展示。
type StoreDailyStats struct {
	// PaidOrders 已支付订单数
	PaidOrders int64
	// Revenue 已收金额（分）
	Revenue int64
	// RefundAmount 已退金额（分）
	RefundAmount int64
	// StatusCounts 各状态的订单数，键为状态值
	StatusCounts map[int16]int64
}

// StoreDailyStats 统计门店在 [from, to) 区间内的已收金额与各状态订单数。
// 入参：门店、区间起止；返回汇总，金额为分。
// 说明：营收口径只认 pay_status = 已支付，退款金额按 refund_amount 累加；
// 已整单退款的订单货币状态会变成已退款，自然被排除在营收之外，不会重复计算。
func (r *CaterOrderRepository) StoreDailyStats(ctx context.Context, storeID types.ID, from, to time.Time) (*StoreDailyStats, error) {
	stats := &StoreDailyStats{StatusCounts: make(map[int16]int64, 8)}

	rangeWhere := func(query db.Query) db.Query {
		return query.Where("store_id", storeID).
			WhereBetween("created_at", types.DateTime(from), types.DateTime(to)).
			WhereNull("deleted_at")
	}

	// 待处理状态：待支付 / 待接单 / 制作中 / 待取餐
	pendingStatuses := []any{
		int16(orderEnums.OrderStatusPendingPay),
		int16(orderEnums.OrderStatusPendingAccept),
		int16(orderEnums.OrderStatusMaking),
		int16(orderEnums.OrderStatusWaitingPickup),
	}

	paid := struct {
		PaidOrders int64 `db:"paid_orders"`
		Revenue    int64 `db:"revenue"`
		Refund     int64 `db:"refund_amount"`
	}{}
	if err := rangeWhere(facades.DB().WithContext(ctx).Table(model.CaterOrderTable)).
		Select("COUNT(*) AS paid_orders, COALESCE(SUM(pay_amount), 0) AS revenue, COALESCE(SUM(refund_amount), 0) AS refund_amount").
		Where("pay_status", int8(orderEnums.PayStatusPaid)).
		First(&paid); err != nil {
		return nil, err
	}
	stats.PaidOrders = paid.PaidOrders
	stats.Revenue = paid.Revenue
	stats.RefundAmount = paid.Refund

	rows := make([]struct {
		Status int16 `db:"status"`
		Count  int64 `db:"cnt"`
	}, 0)
	if err := rangeWhere(facades.DB().WithContext(ctx).Table(model.CaterOrderTable)).
		Select("status, COUNT(*) AS cnt").
		WhereIn("status", pendingStatuses).
		GroupBy("status").
		Get(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		stats.StatusCounts[row.Status] = row.Count
	}

	return stats, nil
}

// DailyTrendRow 是按日聚合的营收行。
type DailyTrendRow struct {
	// Day 日期，格式 yyyy-MM-dd
	Day string `db:"day"`
	// Orders 当日已支付订单数
	Orders int64 `db:"orders"`
	// Revenue 当日营收（分）
	Revenue int64 `db:"revenue"`
}

// DishSalesRow 是菜品销量聚合行，按「菜品 + 规格」归并。
type DishSalesRow struct {
	SpuName  string `db:"spu_name"`
	SkuName  string `db:"sku_name"`
	Quantity int64  `db:"quantity"`
	Amount   int64  `db:"amount"`
}

// PaymentMix 是收款方式构成：现金与线上渠道各占多少。
type PaymentMix struct {
	CashAmount   int64
	CashOrders   int64
	OnlineAmount int64
	OnlineOrders int64
}

// DailyTrend 按天统计门店营收与单量。
// 入参：门店、区间起止；返回按日期升序的行，没有订单的日期不会出现在结果里，由调用方补齐。
func (r *CaterOrderRepository) DailyTrend(ctx context.Context, storeID types.ID, from, to time.Time) ([]DailyTrendRow, error) {
	rows := make([]DailyTrendRow, 0)
	// 用 DATE_FORMAT 而不是 DATE()：DATE 在 parseTime 驱动下会扫成 time.Time，写进 string 会失败
	err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') AS day, COUNT(*) AS orders, COALESCE(SUM(pay_amount), 0) AS revenue").
		Where("store_id", storeID).
		Where("pay_status", int8(orderEnums.PayStatusPaid)).
		WhereBetween("created_at", types.DateTime(from), types.DateTime(to)).
		WhereNull("deleted_at").
		GroupBy("DATE_FORMAT(created_at, '%Y-%m-%d')").
		OrderBy("day").
		Get(&rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// DishSales 统计门店菜品销量排行。
// 入参：门店、区间、取前 N 名；返回按销量降序的行。
// 说明：只统计已支付订单的正常明细，已退明细不计销量，避免退货把排行刷高。
func (r *CaterOrderRepository) DishSales(ctx context.Context, storeID types.ID, from, to time.Time, limit int) ([]DishSalesRow, error) {
	rows := make([]DishSalesRow, 0)
	// 主表不加别名，避免构造器把 "表 AS 别名" 整体加引号；join 片段直接写原始 SQL
	err := facades.DB().WithContext(ctx).Table(model.CaterOrderItemTable).
		Join("cater_order AS o ON o.id = cater_order_item.order_id").
		Select("cater_order_item.spu_name AS spu_name, cater_order_item.sku_name AS sku_name, " +
			"SUM(cater_order_item.quantity) AS quantity, COALESCE(SUM(cater_order_item.subtotal), 0) AS amount").
		Where("o.store_id", storeID).
		Where("o.pay_status", int8(orderEnums.PayStatusPaid)).
		Where("cater_order_item.status", int8(1)).
		WhereBetween("o.created_at", types.DateTime(from), types.DateTime(to)).
		WhereNull("o.deleted_at").
		WhereNull("cater_order_item.deleted_at").
		GroupBy("cater_order_item.spu_name, cater_order_item.sku_name").
		OrderByDesc("quantity").
		Limit(uint64(limit)).
		Get(&rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// PaymentMix 统计门店现金与线上渠道的收款构成。
// 入参：门店、区间；返回两种方式的金额与单量。
// 说明：现金收款不生成支付单，pay_order_id 为 0，据此与线上渠道区分。
func (r *CaterOrderRepository) PaymentMix(ctx context.Context, storeID types.ID, from, to time.Time) (*PaymentMix, error) {
	mix := &PaymentMix{}

	sum := func(cashOnly bool) (int64, int64, error) {
		row := struct {
			Orders int64 `db:"orders"`
			Amount int64 `db:"amount"`
		}{}
		query := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
			Select("COUNT(*) AS orders, COALESCE(SUM(pay_amount), 0) AS amount").
			Where("store_id", storeID).
			Where("pay_status", int8(orderEnums.PayStatusPaid)).
			WhereBetween("created_at", types.DateTime(from), types.DateTime(to)).
			WhereNull("deleted_at")
		if cashOnly {
			query = query.Where("pay_order_id", 0)
		} else {
			query = query.Where("pay_order_id > 0")
		}

		if err := query.First(&row); err != nil {
			return 0, 0, err
		}

		return row.Orders, row.Amount, nil
	}

	var err error
	if mix.CashOrders, mix.CashAmount, err = sum(true); err != nil {
		return nil, err
	}
	if mix.OnlineOrders, mix.OnlineAmount, err = sum(false); err != nil {
		return nil, err
	}

	return mix, nil
}

// HourlyRow 是按小时聚合的营收行。
type HourlyRow struct {
	// Hour 小时，两位数字，如 08、23
	Hour string `db:"h"`
	// Orders 该小时已支付订单数
	Orders int64 `db:"orders"`
	// Revenue 该小时营收（分）
	Revenue int64 `db:"revenue"`
}

// RefundRow 是有退款的订单行，供门店核对退给了谁、退了多少。
type RefundRow struct {
	No           string         `db:"no"`
	PayAmount    int64          `db:"pay_amount"`
	RefundAmount int64          `db:"refund_amount"`
	Status       int16          `db:"status"`
	UpdatedAt    types.DateTime `db:"updated_at"`
}

// HourlyStats 按小时统计门店营收，用于看营业高峰时段。
// 入参：门店、区间；返回按小时升序的行，没有订单的小时不会出现。
func (r *CaterOrderRepository) HourlyStats(ctx context.Context, storeID types.ID, from, to time.Time) ([]HourlyRow, error) {
	rows := make([]HourlyRow, 0)
	err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Select("DATE_FORMAT(created_at, '%H') AS h, COUNT(*) AS orders, COALESCE(SUM(pay_amount), 0) AS revenue").
		Where("store_id", storeID).
		Where("pay_status", int8(orderEnums.PayStatusPaid)).
		WhereBetween("created_at", types.DateTime(from), types.DateTime(to)).
		WhereNull("deleted_at").
		GroupBy("DATE_FORMAT(created_at, '%H')").
		OrderBy("h").
		Get(&rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// RefundOrders 列出区间内有退款的订单。
// 入参：门店、区间、返回条数上限；返回按下单时间倒序的行。
// 说明：按订单维度统计而不是退款单维度——现金退款不生成退款单，只查退款单会漏；
// 订单表没有独立的退款时间，展示时用 updated_at 作为「最后更新时间」。
func (r *CaterOrderRepository) RefundOrders(ctx context.Context, storeID types.ID, from, to time.Time, limit int) ([]RefundRow, error) {
	rows := make([]RefundRow, 0)
	err := facades.DB().WithContext(ctx).Table(model.CaterOrderTable).
		Select("no, pay_amount, refund_amount, status, updated_at").
		Where("store_id", storeID).
		Where("refund_amount > 0").
		WhereBetween("created_at", types.DateTime(from), types.DateTime(to)).
		WhereNull("deleted_at").
		OrderByDesc("updated_at").
		Limit(uint64(limit)).
		Get(&rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// NewOrderLogModel 组装一条订单状态日志。
func NewOrderLogModel(orderID types.ID, from, to orderEnums.OrderStatus, operatorType int8, operatorID types.ID, operatorName, reason string) *model.CaterOrderLogModel {
	return &model.CaterOrderLogModel{
		OrderID:      orderID,
		FromStatus:   int16(from),
		ToStatus:     int16(to),
		OperatorType: operatorType,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Reason:       reason,
	}
}
