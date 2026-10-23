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
