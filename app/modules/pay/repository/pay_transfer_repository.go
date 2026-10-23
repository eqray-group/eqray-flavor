package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/model"
)

// payTransferTable 是转账单表名，对应 yudao pay_transfer。
const payTransferTable = model.PayTransferTable

// PayTransferRepository 负责转账单的读写，对齐 yudao PayTransferService 的持久化部分。
type PayTransferRepository struct{}

// NewPayTransferRepository 构造转账单仓储。
func NewPayTransferRepository() *PayTransferRepository {
	return &PayTransferRepository{}
}

// PayTransferFilter 是转账单列表的筛选条件。
type PayTransferFilter struct {
	// Keyword 转账单号 / 商户转账单号关键字
	Keyword string
	// AppID 应用编号，0 表示不限
	AppID types.ID
	// ChannelCode 渠道编码
	ChannelCode string
	// Status 转账状态
	Status *int8
}

// Paginate 按条件分页查询转账单，按主键倒序。
// 入参：筛选条件与分页；返回转账单列表与总数。
func (r *PayTransferRepository) Paginate(ctx context.Context, filter PayTransferFilter, pageNo, pageSize int) ([]model.PayTransferModel, int64, error) {
	list := make([]model.PayTransferModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payTransferTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("no", like).OrWhereLike("merchant_transfer_id", like)
		})
	}
	if filter.AppID > 0 {
		query = query.Where("app_id", filter.AppID)
	}
	if filter.ChannelCode != "" {
		query = query.Where("channel_code", filter.ChannelCode)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByID 按主键查询未删除的转账单；未命中返回 (nil, nil)。
func (r *PayTransferRepository) FindByID(ctx context.Context, id types.ID) (*model.PayTransferModel, error) {
	return r.find(ctx, facades.DB().WithContext(ctx).Table(payTransferTable).Where("id", id))
}

// FindByNo 按转账单号查询；未命中返回 (nil, nil)。
func (r *PayTransferRepository) FindByNo(ctx context.Context, no string) (*model.PayTransferModel, error) {
	return r.find(ctx, facades.DB().WithContext(ctx).Table(payTransferTable).Where("no", no))
}

// FindByMerchantTransferID 按「应用 + 商户转账单号」查询；未命中返回 (nil, nil)。
// 入参：应用编号与商户转账单号；用于转账幂等。
func (r *PayTransferRepository) FindByMerchantTransferID(ctx context.Context, appID types.ID, merchantTransferID string) (*model.PayTransferModel, error) {
	return r.find(ctx, facades.DB().WithContext(ctx).Table(payTransferTable).
		Where("app_id", appID).
		Where("merchant_transfer_id", merchantTransferID))
}

// Create 在事务内插入转账单。
// 入参：事务与转账单模型；返回错误表示落库失败。
func (r *PayTransferRepository) Create(ctx context.Context, tx db.Tx, m *model.PayTransferModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payTransferTable).Insert(map[string]any{
		"id":                   m.ID,
		"no":                   m.No,
		"app_id":               m.AppID,
		"channel_id":           m.ChannelID,
		"channel_code":         m.ChannelCode,
		"user_id":              m.UserID,
		"user_type":            m.UserType,
		"merchant_transfer_id": m.MerchantTransferID,
		"subject":              m.Subject,
		"price":                m.Price,
		"channel_fee_rate":     m.ChannelFeeRate,
		"channel_fee_price":    m.ChannelFeePrice,
		"status":               int8(m.Status),
		"success_time":         m.SuccessTime,
		"channel_extras":       m.ChannelExtras,
		"channel_error_code":   m.ChannelErrorCode,
		"channel_error_msg":    m.ChannelErrorMsg,
		"channel_notify_data":  m.ChannelNotifyData,
		"creator":              m.Creator,
		"created_at":           m.CreatedAt,
		"updater":              m.Updater,
		"updated_at":           m.UpdatedAt,
	})

	return err
}

// UpdateResult 以条件更新回写转账结果，只命中仍在等待转账的单。
// 入参：事务与转账单模型；返回影响行数，0 表示状态已变化。
func (r *PayTransferRepository) UpdateResult(ctx context.Context, tx db.Tx, m *model.PayTransferModel) (int64, error) {
	m.FillUpdate(ctx)

	res, err := tx.Table(payTransferTable).
		Where("id", m.ID).
		Where("status", int8(payEnums.PayTransferStatusWaiting)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":              int8(m.Status),
			"success_time":        m.SuccessTime,
			"channel_extras":      m.ChannelExtras,
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
func (r *PayTransferRepository) find(ctx context.Context, query db.Query) (*model.PayTransferModel, error) {
	var m model.PayTransferModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}
