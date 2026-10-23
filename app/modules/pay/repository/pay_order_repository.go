package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"
	frameworkDB "github.com/goravel/framework/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/model"
)

// payOrderTable 是支付单表名，对应 yudao pay_order。
const payOrderTable = model.PayOrderTable

// PayOrderRepository 负责支付单的读写，对齐 yudao PayOrderService 的持久化部分。
type PayOrderRepository struct{}

// NewPayOrderRepository 构造支付单仓储。
func NewPayOrderRepository() *PayOrderRepository {
	return &PayOrderRepository{}
}

// PayOrderFilter 是支付单列表的筛选条件。
type PayOrderFilter struct {
	// Keyword 支付单号 / 商户订单号 / 渠道订单号关键字
	Keyword string
	// AppID 应用编号，0 表示不限
	AppID types.ID
	// ChannelCode 渠道编码
	ChannelCode string
	// Status 支付状态
	Status *int8
	// UserID 付款用户编号，C 端按令牌强制填充
	UserID types.ID
	// CreatedFrom 创建时间起（含）
	CreatedFrom *types.DateTime
	// CreatedTo 创建时间止（含）
	CreatedTo *types.DateTime
}

// Paginate 按条件分页查询支付单，按主键倒序。
// 入参：筛选条件与分页；返回支付单列表与总数。
func (r *PayOrderRepository) Paginate(ctx context.Context, filter PayOrderFilter, pageNo, pageSize int) ([]model.PayOrderModel, int64, error) {
	list := make([]model.PayOrderModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payOrderTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("no", like).
				OrWhereLike("merchant_order_id", like).
				OrWhereLike("channel_order_no", like)
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
	if filter.UserID > 0 {
		query = query.Where("user_id", filter.UserID)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filter.CreatedTo)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByID 按主键查询未删除的支付单；未命中返回 (nil, nil)。
func (r *PayOrderRepository) FindByID(ctx context.Context, id types.ID) (*model.PayOrderModel, error) {
	return r.find(ctx, facades.DB().WithContext(ctx).Table(payOrderTable).Where("id", id))
}

// FindByNo 按支付单号查询；未命中返回 (nil, nil)。
// 入参：对外支付单号；返回支付单，回调处理时据此定位。
func (r *PayOrderRepository) FindByNo(ctx context.Context, no string) (*model.PayOrderModel, error) {
	return r.find(ctx, facades.DB().WithContext(ctx).Table(payOrderTable).Where("no", no))
}

// FindByMerchantOrderID 按「应用 + 商户订单号」查询支付单，不区分状态；未命中返回 (nil, nil)。
// 入参：应用编号与商户订单号；用于下单幂等。与 yudao 一致，同一应用下商户订单号全局唯一，
// 已关闭的单也算占用（需要重新发起支付时走 ReSubmit 或换业务单号）。
func (r *PayOrderRepository) FindByMerchantOrderID(ctx context.Context, appID types.ID, merchantOrderID string) (*model.PayOrderModel, error) {
	query := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("app_id", appID).
		Where("merchant_order_id", merchantOrderID)

	return r.find(ctx, query)
}

// FindLatestByMerchantOrderID 按商户订单号查询最近一张支付单，不限应用；未命中返回 (nil, nil)。
// 入参：商户订单号；返回支付单与错误。
// 说明：业务侧（订单）关单前需要确认该单是否已收款，而业务侧只知道自己这一侧的订单号，
// 不掌握应用编号，因此这里放宽应用维度，按创建时间倒序取最近一张。
func (r *PayOrderRepository) FindLatestByMerchantOrderID(ctx context.Context, merchantOrderID string) (*model.PayOrderModel, error) {
	query := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("merchant_order_id", merchantOrderID).
		WhereNull("deleted_at").
		OrderByDesc("id")

	return r.find(ctx, query)
}

// Create 在事务内插入支付单。
// 入参：事务与支付单模型；返回错误表示落库失败（no 唯一键冲突由服务层保证）。
func (r *PayOrderRepository) Create(ctx context.Context, tx db.Tx, m *model.PayOrderModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payOrderTable).Insert(map[string]any{
		"id":                m.ID,
		"app_id":            m.AppID,
		"channel_id":        m.ChannelID,
		"channel_code":      m.ChannelCode,
		"user_id":           m.UserID,
		"user_type":         m.UserType,
		"merchant_order_id": m.MerchantOrderID,
		"subject":           m.Subject,
		"body":              m.Body,
		"notify_url":        m.NotifyURL,
		"price":             m.Price,
		"channel_fee_rate":  m.ChannelFeeRate,
		"channel_fee_price": m.ChannelFeePrice,
		"status":            int8(m.Status),
		"refund_price":      m.RefundPrice,
		"user_ip":           m.UserIP,
		"expire_time":       m.ExpireTime,
		"success_time":      m.SuccessTime,
		"extension_id":      m.ExtensionID,
		"no":                m.No,
		"channel_order_no":  m.ChannelOrderNo,
		"channel_user_id":   m.ChannelUserID,
		"creator":           m.Creator,
		"created_at":        m.CreatedAt,
		"updater":           m.Updater,
		"updated_at":        m.UpdatedAt,
	})

	return err
}

// UpdateChannel 在事务内把支付单切到新渠道并回写费率快照。
// 入参：事务、支付单 ID、渠道 ID / 编码、费率与手续费；返回错误表示更新失败。
func (r *PayOrderRepository) UpdateChannel(ctx context.Context, tx db.Tx, id, channelID types.ID, channelCode string, feeRate float64, feePrice int64) error {
	_, err := tx.Table(payOrderTable).
		Where("id", id).
		Where("status", int8(payEnums.PayOrderStatusWaiting)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"channel_id":        channelID,
			"channel_code":      channelCode,
			"channel_fee_rate":  feeRate,
			"channel_fee_price": feePrice,
			"updated_at":        types.Now(),
		})

	return err
}

// MarkPaid 以条件更新把支付单置为已支付，只有待支付的单会被命中。
// 入参：事务、支付单 ID、拓展单 ID、渠道订单号、渠道用户标识、成功时间；
// 返回影响行数，0 表示状态已变化（并发回调已处理）。
func (r *PayOrderRepository) MarkPaid(ctx context.Context, tx db.Tx, id, extensionID types.ID, channelOrderNo, channelUserID string, successTime types.DateTime) (int64, error) {
	res, err := tx.Table(payOrderTable).
		Where("id", id).
		Where("status", int8(payEnums.PayOrderStatusWaiting)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":           int8(payEnums.PayOrderStatusSuccess),
			"extension_id":     extensionID,
			"channel_order_no": channelOrderNo,
			"channel_user_id":  channelUserID,
			"success_time":     successTime,
			"updated_at":       types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// MarkRefunded 累加支付单的已退金额，并保证累计不超过支付金额。
// 入参：事务、支付单 ID、本次退款金额；返回影响行数，0 表示支付单不存在或累计会超退。
// 说明：与 yudao 的 updateOrderRefundPrice 校验等价 —— yudao 先查再比，这里用条件更新
// （refund_price + 本次 <= price）把并发退款也挡住；全额退款不改变支付单状态，
// 仍是「支付成功」，是否退完由 refund_price >= price 判定。
func (r *PayOrderRepository) MarkRefunded(ctx context.Context, tx db.Tx, id types.ID, refundPrice int64) (int64, error) {
	res, err := tx.Table(payOrderTable).
		Where("id", id).
		WhereRaw("refund_price + ? <= price", []any{refundPrice}).
		WhereNull("deleted_at").
		Update(map[string]any{
			"refund_price": frameworkDB.Raw("refund_price + ?", refundPrice),
			"updated_at":   types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// MarkClosed 以条件更新关闭支付单，已支付 / 已退款的单不会被命中。
// 入参：事务与支付单 ID；返回影响行数，0 表示状态不允许关闭。
func (r *PayOrderRepository) MarkClosed(ctx context.Context, tx db.Tx, id types.ID) (int64, error) {
	res, err := tx.Table(payOrderTable).
		Where("id", id).
		Where("status", int8(payEnums.PayOrderStatusWaiting)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":     int8(payEnums.PayOrderStatusClosed),
			"updated_at": types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// ListExpired 查询已过期但仍待支付的支付单，供关单任务使用。
// 入参：当前时间与单次最大条数；返回待关闭的支付单列表。
func (r *PayOrderRepository) ListExpired(ctx context.Context, now types.DateTime, limit uint64) ([]model.PayOrderModel, error) {
	list := make([]model.PayOrderModel, 0)
	if err := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("status", int8(payEnums.PayOrderStatusWaiting)).
		Where("expire_time <= ?", now).
		WhereNull("deleted_at").
		OrderBy("id").
		Limit(limit).
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// find 执行单条查询，未命中时返回 (nil, nil)。
func (r *PayOrderRepository) find(ctx context.Context, query db.Query) (*model.PayOrderModel, error) {
	var m model.PayOrderModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}
