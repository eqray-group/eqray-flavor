package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/model"
)

// payOrderExtensionTable 是支付订单拓展表名，对应 yudao pay_order_extension。
const payOrderExtensionTable = model.PayOrderExtensionTable

// PayOrderExtensionRepository 负责支付订单拓展单的读写。
// 每次调用渠道下单写一条，保存渠道额外参数与报错，便于定位下单失败原因。
type PayOrderExtensionRepository struct{}

// NewPayOrderExtensionRepository 构造支付订单拓展仓储。
func NewPayOrderExtensionRepository() *PayOrderExtensionRepository {
	return &PayOrderExtensionRepository{}
}

// FindByID 按主键查询拓展单；未命中返回 (nil, nil)。
func (r *PayOrderExtensionRepository) FindByID(ctx context.Context, id types.ID) (*model.PayOrderExtensionModel, error) {
	var m model.PayOrderExtensionModel
	if err := facades.DB().WithContext(ctx).Table(payOrderExtensionTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ListByOrderID 查询某个支付单下的全部拓展单，按主键升序。
// 入参：支付单 ID；返回拓展单列表，用于后台查看渠道调用历史。
func (r *PayOrderExtensionRepository) ListByOrderID(ctx context.Context, orderID types.ID) ([]model.PayOrderExtensionModel, error) {
	list := make([]model.PayOrderExtensionModel, 0)
	if err := facades.DB().WithContext(ctx).Table(payOrderExtensionTable).
		Where("order_id", orderID).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Create 在事务内插入拓展单。
// 入参：事务与拓展单模型；返回错误表示落库失败。
func (r *PayOrderExtensionRepository) Create(ctx context.Context, tx db.Tx, m *model.PayOrderExtensionModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payOrderExtensionTable).Insert(map[string]any{
		"id":                  m.ID,
		"no":                  m.No,
		"order_id":            m.OrderID,
		"channel_id":          m.ChannelID,
		"channel_code":        m.ChannelCode,
		"user_ip":             m.UserIP,
		"status":              int8(m.Status),
		"channel_extras":      m.ChannelExtras,
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

// UpdateResult 回写渠道调用结果：额外参数、错误码与错误信息。
// 入参：事务与拓展单模型；返回错误表示更新失败。
func (r *PayOrderExtensionRepository) UpdateResult(ctx context.Context, tx db.Tx, m *model.PayOrderExtensionModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(payOrderExtensionTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":              int8(m.Status),
			"channel_extras":      m.ChannelExtras,
			"channel_error_code":  m.ChannelErrorCode,
			"channel_error_msg":   m.ChannelErrorMsg,
			"channel_notify_data": m.ChannelNotifyData,
			"updater":             m.Updater,
			"updated_at":          m.UpdatedAt,
		})

	return err
}
