package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/model"
)

// payRefundTable 是退款单表名，对应 yudao pay_refund。
const payRefundTable = model.PayRefundTable

// PayRefundRepository 负责退款单的读写，对齐 yudao PayRefundService 的持久化部分。
type PayRefundRepository struct{}

// NewPayRefundRepository 构造退款单仓储。
func NewPayRefundRepository() *PayRefundRepository {
	return &PayRefundRepository{}
}

// PayRefundFilter 是退款单列表的筛选条件。
type PayRefundFilter struct {
	// Keyword 退款单号 / 支付单号 / 渠道退款单号关键字
	Keyword string
	// AppID 应用编号，0 表示不限
	AppID types.ID
	// OrderID 支付单编号
	OrderID types.ID
	// ChannelCode 渠道编码
	ChannelCode string
	// Status 退款状态
	Status *int8
	// CreatedFrom 创建时间起（含）
	CreatedFrom *types.DateTime
	// CreatedTo 创建时间止（含）
	CreatedTo *types.DateTime
}

// Paginate 按条件分页查询退款单，按主键倒序。
// 入参：筛选条件与分页；返回退款单列表与总数。
func (r *PayRefundRepository) Paginate(ctx context.Context, filter PayRefundFilter, pageNo, pageSize int) ([]model.PayRefundModel, int64, error) {
	list := make([]model.PayRefundModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payRefundTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("no", like).
				OrWhereLike("order_no", like).
				OrWhereLike("channel_refund_no", like)
		})
	}
	if filter.AppID > 0 {
		query = query.Where("app_id", filter.AppID)
	}
	if filter.OrderID > 0 {
		query = query.Where("order_id", filter.OrderID)
	}
	if filter.ChannelCode != "" {
		query = query.Where("channel_code", filter.ChannelCode)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("created_at", ">=", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("created_at", "<=", *filter.CreatedTo)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByID 按主键查询未删除的退款单；未命中返回 (nil, nil)。
func (r *PayRefundRepository) FindByID(ctx context.Context, id types.ID) (*model.PayRefundModel, error) {
	return r.find(ctx, facades.DB().WithContext(ctx).Table(payRefundTable).Where("id", id))
}

// FindByNo 按退款单号查询；未命中返回 (nil, nil)。
func (r *PayRefundRepository) FindByNo(ctx context.Context, no string) (*model.PayRefundModel, error) {
	return r.find(ctx, facades.DB().WithContext(ctx).Table(payRefundTable).Where("no", no))
}

// FindByMerchantRefundID 按「应用 + 商户退款单号」查询；未命中返回 (nil, nil)。
// 入参：应用编号与商户退款单号；用于退款幂等。
func (r *PayRefundRepository) FindByMerchantRefundID(ctx context.Context, appID types.ID, merchantRefundID string) (*model.PayRefundModel, error) {
	return r.find(ctx, facades.DB().WithContext(ctx).Table(payRefundTable).
		Where("app_id", appID).
		Where("merchant_refund_id", merchantRefundID))
}

// SumRefundPrice 统计某个支付单已占用（等待中 + 已成功）的退款金额。
// 入参：支付单 ID；返回已退金额（分），空集返回 0。
func (r *PayRefundRepository) SumRefundPrice(ctx context.Context, orderID types.ID) (int64, error) {
	row := struct {
		Total int64 `db:"total"`
	}{}

	if err := facades.DB().WithContext(ctx).Table(payRefundTable).
		Select("COALESCE(SUM(refund_price), 0) AS total").
		Where("order_id", orderID).
		WhereIn("status", []any{int8(payEnums.PayRefundStatusCreate), int8(payEnums.PayRefundStatusSuccess)}).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, err
	}

	return row.Total, nil
}

// Create 在事务内插入退款单。
// 入参：事务与退款单模型；返回错误表示落库失败。
func (r *PayRefundRepository) Create(ctx context.Context, tx db.Tx, m *model.PayRefundModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payRefundTable).Insert(map[string]any{
		"id":                  m.ID,
		"no":                  m.No,
		"app_id":              m.AppID,
		"channel_id":          m.ChannelID,
		"channel_code":        m.ChannelCode,
		"order_id":            m.OrderID,
		"order_no":            m.OrderNo,
		"merchant_order_id":   m.MerchantOrderID,
		"merchant_refund_id":  m.MerchantRefundID,
		"notify_url":          m.NotifyURL,
		"status":              int8(m.Status),
		"pay_price":           m.PayPrice,
		"refund_price":        m.RefundPrice,
		"reason":              m.Reason,
		"user_ip":             m.UserIP,
		"channel_order_no":    m.ChannelOrderNo,
		"channel_refund_no":   m.ChannelRefundNo,
		"success_time":        m.SuccessTime,
		"channel_error_code":  m.ChannelErrorCode,
		"channel_error_msg":   m.ChannelErrorMsg,
		"channel_notify_data": m.ChannelNotifyData,
		"creator":             m.Creator,
		"created_at":          m.CreatedAt,
		"updater":             m.Updater,
		"updated_at":          m.UpdatedAt,
	})

	return err
}

// UpdateResult 以条件更新回写退款结果，只命中仍在等待退款的单。
// 入参：事务与退款单模型（status / 渠道退款单号 / 成功时间 / 错误信息）；返回影响行数。
func (r *PayRefundRepository) UpdateResult(ctx context.Context, tx db.Tx, m *model.PayRefundModel) (int64, error) {
	m.FillUpdate(ctx)

	res, err := tx.Table(payRefundTable).
		Where("id", m.ID).
		Where("status", int8(payEnums.PayRefundStatusCreate)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":              int8(m.Status),
			"channel_refund_no":   m.ChannelRefundNo,
			"success_time":        m.SuccessTime,
			"channel_error_code":  m.ChannelErrorCode,
			"channel_error_msg":   m.ChannelErrorMsg,
			"channel_notify_data": m.ChannelNotifyData,
			"updater":             m.Updater,
			"updated_at":          m.UpdatedAt,
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// find 执行单条查询，未命中时返回 (nil, nil)。
func (r *PayRefundRepository) find(ctx context.Context, query db.Query) (*model.PayRefundModel, error) {
	var m model.PayRefundModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}
