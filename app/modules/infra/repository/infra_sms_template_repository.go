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

const smsTemplateTable = model.InfraSmsTemplateTable

// InfraSmsTemplateRepository 是 infra_sms_template 表的仓储。
type InfraSmsTemplateRepository struct{}

func NewInfraSmsTemplateRepository() *InfraSmsTemplateRepository {
	return &InfraSmsTemplateRepository{}
}

// Create 新增模板；模板代码冲突转换为业务错误码。
func (r *InfraSmsTemplateRepository) Create(ctx context.Context, m *model.InfraSmsTemplateModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(smsTemplateTable).Insert(map[string]any{
		"id":              m.ID,
		"channel_id":      m.ChannelID,
		"code":            m.Code,
		"name":            m.Name,
		"type":            m.Type,
		"api_template_id": m.APITemplateID,
		"content":         m.Content,
		"params":          m.Params,
		"status":          m.Status,
		"remark":          m.Remark,
		"creator":         m.Creator,
		"created_at":      m.CreatedAt,
		"updater":         m.Updater,
		"updated_at":      m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(infraErrs.InfraSmsTemplateCodeExists, m.Code)
	}

	return err
}

// Update 更新模板。
func (r *InfraSmsTemplateRepository) Update(ctx context.Context, m *model.InfraSmsTemplateModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(smsTemplateTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"channel_id":      m.ChannelID,
			"name":            m.Name,
			"type":            m.Type,
			"api_template_id": m.APITemplateID,
			"content":         m.Content,
			"params":          m.Params,
			"status":          m.Status,
			"remark":          m.Remark,
			"updater":         m.Updater,
			"updated_at":      m.UpdatedAt,
		})

	return err
}

// Delete 软删除模板。
func (r *InfraSmsTemplateRepository) Delete(ctx context.Context, m *model.InfraSmsTemplateModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(smsTemplateTable).
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
		return baseErrs.New(infraErrs.InfraSmsTemplateNotExists, m.ID)
	}

	return nil
}

// FindByID 按主键查询未软删的模板；未命中返回 (nil, nil)。
func (r *InfraSmsTemplateRepository) FindByID(ctx context.Context, id types.ID) (*model.InfraSmsTemplateModel, error) {
	var m model.InfraSmsTemplateModel

	if err := facades.DB().WithContext(ctx).Table(smsTemplateTable).
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

// FindEnabledByCode 按模板代码查询启用中的模板，供发送时使用。
func (r *InfraSmsTemplateRepository) FindEnabledByCode(ctx context.Context, code string) (*model.InfraSmsTemplateModel, error) {
	var m model.InfraSmsTemplateModel

	if err := facades.DB().WithContext(ctx).Table(smsTemplateTable).
		Where("code", code).
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

// ExistsByCode 判断模板代码是否已被占用，excludeID 用于编辑时排除自身。
func (r *InfraSmsTemplateRepository) ExistsByCode(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(smsTemplateTable).
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

// Paginate 分页查询模板。
func (r *InfraSmsTemplateRepository) Paginate(ctx context.Context, filter InfraSmsTemplateFilter, pageNo, pageSize int) ([]model.InfraSmsTemplateModel, int64, error) {
	list := make([]model.InfraSmsTemplateModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(smsTemplateTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
	}
	if filter.ChannelID > 0 {
		query = query.Where("channel_id", filter.ChannelID)
	}
	if filter.Type != nil {
		query = query.Where("type", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// InfraSmsTemplateFilter 是模板列表的筛选条件。
type InfraSmsTemplateFilter struct {
	Keyword   string
	ChannelID types.ID
	Type      *int8
	Status    *baseEnums.CommonStatusEnum
}
