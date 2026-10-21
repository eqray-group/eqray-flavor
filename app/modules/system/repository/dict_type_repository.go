package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

const dictTypeTable = model.SystemDictTypeTable

// DictTypeFilter 是字典类型列表的筛选条件。
type DictTypeFilter struct {
	Name   string
	Type   string
	Status *baseEnums.CommonStatusEnum
}

// DictTypeRepository 是 system_dict_type 表的仓储。
type DictTypeRepository struct{}

func NewDictTypeRepository() *DictTypeRepository {
	return &DictTypeRepository{}
}

// Paginate 分页查询字典类型。
func (r *DictTypeRepository) Paginate(ctx context.Context, filter DictTypeFilter, pageNo, pageSize int) ([]model.SystemDictTypeModel, int64, error) {
	list := make([]model.SystemDictTypeModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(dictTypeTable).WhereNull("deleted_at")
	if filter.Name != "" {
		query = query.WhereLike("name", "%"+filter.Name+"%")
	}
	if filter.Type != "" {
		query = query.WhereLike("type", "%"+filter.Type+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ListEnabled 返回全部启用字典类型，供下拉选择。
func (r *DictTypeRepository) ListEnabled(ctx context.Context) ([]model.SystemDictTypeModel, error) {
	list := make([]model.SystemDictTypeModel, 0)

	err := facades.DB().WithContext(ctx).Table(dictTypeTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderByDesc("id").
		Get(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// FindByID 按主键查询。
func (r *DictTypeRepository) FindByID(ctx context.Context, id types.ID) (*model.SystemDictTypeModel, error) {
	var m model.SystemDictTypeModel

	err := facades.DB().WithContext(ctx).Table(dictTypeTable).
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

// FindByType 按类型编码查询。
func (r *DictTypeRepository) FindByType(ctx context.Context, code string) (*model.SystemDictTypeModel, error) {
	var m model.SystemDictTypeModel

	err := facades.DB().WithContext(ctx).Table(dictTypeTable).
		Where("type", code).
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

// ExistsByType 判断类型编码是否被占用；excludeID > 0 时排除自身。
func (r *DictTypeRepository) ExistsByType(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	if code == "" {
		return false, nil
	}

	var m model.SystemDictTypeModel
	query := facades.DB().WithContext(ctx).Table(dictTypeTable).
		Where("type", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	if err := query.First(&m); err != nil {
		return false, err
	}

	return m.ID != 0, nil
}

// Create 插入字典类型。
func (r *DictTypeRepository) Create(ctx context.Context, m *model.SystemDictTypeModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(dictTypeTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"type":       m.Type,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// Update 更新字典类型；类型编码创建后不可修改。
func (r *DictTypeRepository) Update(ctx context.Context, m *model.SystemDictTypeModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(dictTypeTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// Delete 物理删除字典类型；字典属于配置数据，删除后允许重新创建同编码。
func (r *DictTypeRepository) Delete(ctx context.Context, tx db.Tx, id types.ID) error {
	_, err := tx.Table(dictTypeTable).Where("id", id).Delete()

	return err
}
