package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
)

const smsChannelTable = model.InfraSmsChannelTable

// InfraSmsChannelRepository 是 infra_sms_channel 表的仓储。
type InfraSmsChannelRepository struct{}

func NewInfraSmsChannelRepository() *InfraSmsChannelRepository {
	return &InfraSmsChannelRepository{}
}

// Create 新增渠道；渠道代码冲突转换为业务错误码。
func (r *InfraSmsChannelRepository) Create(ctx context.Context, m *model.InfraSmsChannelModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(smsChannelTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"code":       m.Code,
		"api_key":    m.APIKey,
		"api_secret": m.APISecret,
		"signature":  m.Signature,
		"extra":      m.Extra,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(infraErrs.InfraSmsChannelCodeExists, m.Code)
	}

	return err
}

// Update 更新渠道；api_secret 传空时沿用原值，避免列表页脱敏后回写覆盖。
func (r *InfraSmsChannelRepository) Update(ctx context.Context, m *model.InfraSmsChannelModel) error {
	m.FillUpdate(ctx)

	data := map[string]any{
		"name":       m.Name,
		"api_key":    m.APIKey,
		"signature":  m.Signature,
		"extra":      m.Extra,
		"status":     m.Status,
		"remark":     m.Remark,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	}
	if m.APISecret != "" {
		data["api_secret"] = m.APISecret
	}

	_, err := facades.DB().WithContext(ctx).Table(smsChannelTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(data)

	return err
}

// Delete 软删除渠道。
func (r *InfraSmsChannelRepository) Delete(ctx context.Context, m *model.InfraSmsChannelModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(smsChannelTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(infraErrs.InfraSmsChannelNotExists, m.ID)
	}

	return nil
}

// FindByID 按主键查询未软删的渠道；未命中返回 (nil, nil)。
func (r *InfraSmsChannelRepository) FindByID(ctx context.Context, id types.ID) (*model.InfraSmsChannelModel, error) {
	var m model.InfraSmsChannelModel

	if err := facades.DB().WithContext(ctx).Table(smsChannelTable).
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

// FindEnabledByID 查询启用中的渠道，供发送时使用。
func (r *InfraSmsChannelRepository) FindEnabledByID(ctx context.Context, id types.ID) (*model.InfraSmsChannelModel, error) {
	var m model.InfraSmsChannelModel

	if err := facades.DB().WithContext(ctx).Table(smsChannelTable).
		Where("id", id).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ExistsByCode 判断渠道代码是否已被占用，excludeID 用于编辑时排除自身。
func (r *InfraSmsChannelRepository) ExistsByCode(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(smsChannelTable).
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

// CountTemplatesByChannelID 统计渠道下的模板数量，用于删除前保护。
func (r *InfraSmsChannelRepository) CountTemplatesByChannelID(ctx context.Context, channelID types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(model.InfraSmsTemplateTable).
		Where("channel_id", channelID).
		WhereNull("deleted_at").
		Count()
}

// Enabled 返回全部启用渠道，按创建时间升序，供模板表单下拉。
func (r *InfraSmsChannelRepository) Enabled(ctx context.Context) ([]model.InfraSmsChannelModel, error) {
	list := make([]model.InfraSmsChannelModel, 0)

	if err := facades.DB().WithContext(ctx).Table(smsChannelTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 分页查询渠道配置。
func (r *InfraSmsChannelRepository) Paginate(ctx context.Context, filter InfraSmsChannelFilter, pageNo, pageSize int) ([]model.InfraSmsChannelModel, int64, error) {
	list := make([]model.InfraSmsChannelModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(smsChannelTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
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

// InfraSmsChannelFilter 是渠道列表的筛选条件。
type InfraSmsChannelFilter struct {
	Keyword string
	Code    string
	Status  *baseEnums.CommonStatusEnum
}
