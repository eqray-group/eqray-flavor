package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/model"
)

// payChannelTable 是支付渠道表名，对应 yudao pay_channel。
const payChannelTable = model.PayChannelTable

// PayChannelRepository 负责支付渠道的读写，对齐 yudao PayChannelService 的持久化部分。
type PayChannelRepository struct{}

// NewPayChannelRepository 构造支付渠道仓储。
func NewPayChannelRepository() *PayChannelRepository {
	return &PayChannelRepository{}
}

// PayChannelFilter 是渠道列表的筛选条件。
type PayChannelFilter struct {
	// AppID 应用编号，0 表示不限
	AppID types.ID
	// Code 渠道编码精确匹配
	Code string
	// Status 开启状态（0 开启 / 1 关闭），nil 表示不限
	Status *int8
}

// Paginate 按条件分页查询渠道，按主键倒序。
// 入参：筛选条件与分页；返回渠道列表与总数。
func (r *PayChannelRepository) Paginate(ctx context.Context, filter PayChannelFilter, pageNo, pageSize int) ([]model.PayChannelModel, int64, error) {
	list := make([]model.PayChannelModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payChannelTable).WhereNull("deleted_at")
	if filter.AppID > 0 {
		query = query.Where("app_id", filter.AppID)
	}
	if filter.Code != "" {
		query = query.Where("code", filter.Code)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// AllByApp 返回指定应用下全部渠道，供后台下拉与配置展示。
// 入参：应用编号；返回该应用的渠道列表，按主键升序。
func (r *PayChannelRepository) AllByApp(ctx context.Context, appID types.ID) ([]model.PayChannelModel, error) {
	list := make([]model.PayChannelModel, 0)
	if err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("app_id", appID).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// FindByID 按主键查询未删除的渠道；未命中返回 (nil, nil)。
func (r *PayChannelRepository) FindByID(ctx context.Context, id types.ID) (*model.PayChannelModel, error) {
	var m model.PayChannelModel
	if err := facades.DB().WithContext(ctx).Table(payChannelTable).
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

// FindByAppAndCode 按「应用 + 渠道编码」查询；未命中返回 (nil, nil)。
// 入参：应用编号与渠道编码；下单时据此定位渠道配置。
func (r *PayChannelRepository) FindByAppAndCode(ctx context.Context, appID types.ID, code string) (*model.PayChannelModel, error) {
	var m model.PayChannelModel
	if err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("app_id", appID).
		Where("code", code).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ExistsByAppAndCode 判断同一应用下渠道编码是否已存在；excludeID > 0 时排除自身。
func (r *PayChannelRepository) ExistsByAppAndCode(ctx context.Context, appID types.ID, code string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("app_id", appID).
		Where("code", code).
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

// Create 在事务内插入渠道。
// 入参：事务与渠道模型（config 为已校验的 JSON 字符串）；返回错误表示落库失败。
func (r *PayChannelRepository) Create(ctx context.Context, tx db.Tx, m *model.PayChannelModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payChannelTable).Insert(map[string]any{
		"id":         m.ID,
		"app_id":     m.AppID,
		"code":       m.Code,
		"status":     int8(m.Status),
		"remark":     m.Remark,
		"fee_rate":   m.FeeRate,
		"config":     m.Config,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// Update 在事务内更新渠道；应用与渠道编码不可变，由服务层保持原值。
func (r *PayChannelRepository) Update(ctx context.Context, tx db.Tx, m *model.PayChannelModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(payChannelTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":     int8(m.Status),
			"remark":     m.Remark,
			"fee_rate":   m.FeeRate,
			"config":     m.Config,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// Delete 软删除渠道，历史支付单仍保留 channel_id / channel_code 快照。
// 入参：事务与渠道主键；返回错误表示删除失败。
func (r *PayChannelRepository) Delete(ctx context.Context, tx db.Tx, id types.ID) error {
	_, err := tx.Table(payChannelTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updated_at": types.Now(),
		})

	return err
}
