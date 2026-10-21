package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
)

const configTable = model.InfraConfigTable

// ConfigFilter 是参数配置列表的筛选条件，零值表示不限制。
type ConfigFilter struct {
	Category string
	Keyword  string
}

// ConfigRepository 是 infra_config 表的仓储。
type ConfigRepository struct{}

func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{}
}

// Create 插入参数配置；config_key 冲突转换为统一的业务错误码。
func (r *ConfigRepository) Create(ctx context.Context, m *model.InfraConfigModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(configTable).Insert(map[string]any{
		"id":         m.ID,
		"category":   m.Category,
		"name":       m.Name,
		"config_key": m.ConfigKey,
		"value":      m.Value,
		"type":       m.Type,
		"visible":    m.Visible,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(infraErrs.InfraConfigKeyExists, m.ConfigKey)
	}

	return err
}

// Update 更新可变字段；config_key 与 type 创建后不可改，因此不在更新集合里。
func (r *ConfigRepository) Update(ctx context.Context, m *model.InfraConfigModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(configTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"category":   m.Category,
			"name":       m.Name,
			"value":      m.Value,
			"visible":    m.Visible,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(infraErrs.InfraConfigNotExists, m.ConfigKey)
	}

	return nil
}

// Delete 软删除参数配置。
func (r *ConfigRepository) Delete(ctx context.Context, m *model.InfraConfigModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(configTable).
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
		return baseErrs.New(infraErrs.InfraConfigNotExists, m.ConfigKey)
	}

	return nil
}

// FindByID 按主键查询未软删的参数配置；未命中返回 (nil, nil)。
func (r *ConfigRepository) FindByID(ctx context.Context, id types.ID) (*model.InfraConfigModel, error) {
	var m model.InfraConfigModel

	if err := facades.DB().WithContext(ctx).Table(configTable).
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

// FindByKey 按参数键查询未软删的参数配置；未命中返回 (nil, nil)。
func (r *ConfigRepository) FindByKey(ctx context.Context, key string) (*model.InfraConfigModel, error) {
	var m model.InfraConfigModel

	if err := facades.DB().WithContext(ctx).Table(configTable).
		Where("config_key", key).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ExistsByKey 判断参数键是否已被占用，excludeID 用于编辑时排除自身。
func (r *ConfigRepository) ExistsByKey(ctx context.Context, key string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(configTable).
		Where("config_key", key).
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

// Paginate 按条件分页查询参数配置。
func (r *ConfigRepository) Paginate(ctx context.Context, filter ConfigFilter, pageNo, pageSize int) ([]model.InfraConfigModel, int64, error) {
	list := make([]model.InfraConfigModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(configTable).WhereNull("deleted_at")
	if filter.Category != "" {
		query = query.Where("category", filter.Category)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.WhereLike("name", keyword)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
