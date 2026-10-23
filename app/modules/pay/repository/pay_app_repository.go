package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/model"
)

// payAppTable 是支付应用表名，对应 yudao pay_app。
const payAppTable = model.PayAppTable

// PayAppRepository 负责支付应用的读写，对齐 yudao PayAppService 的持久化部分。
type PayAppRepository struct{}

// NewPayAppRepository 构造支付应用仓储。
func NewPayAppRepository() *PayAppRepository {
	return &PayAppRepository{}
}

// PayAppFilter 是应用列表的筛选条件。
type PayAppFilter struct {
	// Keyword 应用名或 appKey 关键字
	Keyword string
	// Status 开启状态（0 开启 / 1 关闭），nil 表示不限
	Status *int8
}

// Paginate 按条件分页查询应用，按主键倒序。
// 入参：筛选条件与分页；返回应用列表与总数。
func (r *PayAppRepository) Paginate(ctx context.Context, filter PayAppFilter, pageNo, pageSize int) ([]model.PayAppModel, int64, error) {
	list := make([]model.PayAppModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payAppTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("name", like).OrWhereLike("app_key", like)
		})
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// AllEnabled 返回全部开启状态的应用，供渠道编辑与下拉选择。
// 入参：无；返回 status=0 的应用列表。
func (r *PayAppRepository) AllEnabled(ctx context.Context) ([]model.PayAppModel, error) {
	list := make([]model.PayAppModel, 0)
	if err := facades.DB().WithContext(ctx).Table(payAppTable).
		Where("status", 0).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// FindByID 按主键查询未删除的应用；未命中返回 (nil, nil)。
func (r *PayAppRepository) FindByID(ctx context.Context, id types.ID) (*model.PayAppModel, error) {
	return r.first(ctx, facades.DB().WithContext(ctx).Table(payAppTable).Where("id", id))
}

// FindByAppKey 按 appKey 查询未删除的应用；未命中返回 (nil, nil)。
// 入参：应用标识；返回应用，用于回调验签时定位应用。
func (r *PayAppRepository) FindByAppKey(ctx context.Context, appKey string) (*model.PayAppModel, error) {
	return r.first(ctx, facades.DB().WithContext(ctx).Table(payAppTable).Where("app_key", appKey))
}

// ExistsByAppKey 判断 appKey 是否已被占用；excludeID > 0 时排除自身。
func (r *PayAppRepository) ExistsByAppKey(ctx context.Context, appKey string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(payAppTable).
		Where("app_key", appKey).
		WhereNull("deleted_at")
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	total, err := query.Count()
	if err != nil {
		return false, err
	}

	return total > 0, nil
}

// Create 在事务内插入应用。
// 入参：事务与应用模型；返回错误表示落库失败（appKey 唯一键冲突由服务层提前校验）。
func (r *PayAppRepository) Create(ctx context.Context, tx db.Tx, m *model.PayAppModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payAppTable).Insert(map[string]any{
		"id":                  m.ID,
		"name":                m.Name,
		"status":              int8(m.Status),
		"remark":              m.Remark,
		"order_notify_url":    m.OrderNotifyURL,
		"refund_notify_url":   m.RefundNotifyURL,
		"transfer_notify_url": m.TransferNotifyURL,
		"app_key":             m.AppKey,
		"app_secret":          m.AppSecret,
		"creator":             m.Creator,
		"created_at":          m.CreatedAt,
		"updater":             m.Updater,
		"updated_at":          m.UpdatedAt,
	})

	return err
}

// Update 在事务内更新应用；appKey 不可修改，由服务层保持原值。
func (r *PayAppRepository) Update(ctx context.Context, tx db.Tx, m *model.PayAppModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(payAppTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":                m.Name,
			"status":              int8(m.Status),
			"remark":              m.Remark,
			"order_notify_url":    m.OrderNotifyURL,
			"refund_notify_url":   m.RefundNotifyURL,
			"transfer_notify_url": m.TransferNotifyURL,
			"app_secret":          m.AppSecret,
			"updater":             m.Updater,
			"updated_at":          m.UpdatedAt,
		})

	return err
}

// Delete 软删除应用，保留历史订单可追溯。
// 入参：事务与应用主键；返回错误表示删除失败。
func (r *PayAppRepository) Delete(ctx context.Context, tx db.Tx, id types.ID) error {
	_, err := tx.Table(payAppTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updated_at": types.Now(),
		})

	return err
}

// first 执行单条查询，未命中时返回 (nil, nil) 而不是 sql.ErrNoRows。
// 入参：已构造的查询；返回应用或 nil。
func (r *PayAppRepository) first(ctx context.Context, query db.Query) (*model.PayAppModel, error) {
	var m model.PayAppModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}
