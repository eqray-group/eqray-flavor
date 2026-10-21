package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

const dictDataTable = model.SystemDictDataTable

// DictDataFilter 是字典数据列表的筛选条件。
type DictDataFilter struct {
	DictType string
	Label    string
	Status   *baseEnums.CommonStatusEnum
}

// DictDataRepository 是 system_dict_data 表的仓储。
type DictDataRepository struct{}

func NewDictDataRepository() *DictDataRepository {
	return &DictDataRepository{}
}

// Paginate 分页查询字典数据。
func (r *DictDataRepository) Paginate(ctx context.Context, filter DictDataFilter, pageNo, pageSize int) ([]model.SystemDictDataModel, int64, error) {
	list := make([]model.SystemDictDataModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(dictDataTable).WhereNull("deleted_at")
	if filter.DictType != "" {
		query = query.Where("dict_type", filter.DictType)
	}
	if filter.Label != "" {
		query = query.WhereLike("label", "%"+filter.Label+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderBy("sort").OrderBy("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ListEnabledByType 返回某个字典类型下的启用项，按排序输出。
func (r *DictDataRepository) ListEnabledByType(ctx context.Context, dictType string) ([]model.SystemDictDataModel, error) {
	list := make([]model.SystemDictDataModel, 0)

	err := facades.DB().WithContext(ctx).Table(dictDataTable).
		Where("dict_type", dictType).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// FindByID 按主键查询。
func (r *DictDataRepository) FindByID(ctx context.Context, id types.ID) (*model.SystemDictDataModel, error) {
	var m model.SystemDictDataModel

	err := facades.DB().WithContext(ctx).Table(dictDataTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m)
	if err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ExistsByValue 判断同一字典类型下键值是否重复；excludeID > 0 时排除自身。
func (r *DictDataRepository) ExistsByValue(ctx context.Context, dictType, value string, excludeID types.ID) (bool, error) {
	if dictType == "" || value == "" {
		return false, nil
	}

	var m model.SystemDictDataModel
	query := facades.DB().WithContext(ctx).Table(dictDataTable).
		Where("dict_type", dictType).
		Where("value", value)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	if err := query.First(&m); err != nil {
		return false, err
	}

	return m.ID != 0, nil
}

// Create 插入字典数据。
func (r *DictDataRepository) Create(ctx context.Context, m *model.SystemDictDataModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(dictDataTable).Insert(map[string]any{
		"id":         m.ID,
		"sort":       m.Sort,
		"label":      m.Label,
		"value":      m.Value,
		"dict_type":  m.DictType,
		"status":     m.Status,
		"color_type": m.ColorType,
		"css_class":  m.CSSClass,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// Update 更新字典数据。
func (r *DictDataRepository) Update(ctx context.Context, m *model.SystemDictDataModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(dictDataTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"sort":       m.Sort,
			"label":      m.Label,
			"value":      m.Value,
			"dict_type":  m.DictType,
			"status":     m.Status,
			"color_type": m.ColorType,
			"css_class":  m.CSSClass,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// Delete 物理删除字典数据，允许重新创建同键值。
func (r *DictDataRepository) Delete(ctx context.Context, id types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(dictDataTable).Where("id", id).Delete()

	return err
}

// DeleteByType 物理删除某个字典类型下的全部数据。
func (r *DictDataRepository) DeleteByType(ctx context.Context, tx db.Tx, dictType string) error {
	_, err := tx.Table(dictDataTable).Where("dict_type", dictType).Delete()

	return err
}
