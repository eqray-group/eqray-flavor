package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
)

const districtTable = model.InfraDistrictTable

// DistrictFilter 是行政区划列表的筛选条件，零值表示不限制。
type DistrictFilter struct {
	// Keyword 命中名称、简称或行政编码
	Keyword string
	// Level 行政级别，0 表示不限
	Level int
	// Pid 上级区划 ID，nil 表示不限
	Pid *types.ID
}

// InfraDistrictRepository 是 infra_district 表的仓储。
type InfraDistrictRepository struct{}

// NewInfraDistrictRepository 构造行政区划仓储。
func NewInfraDistrictRepository() *InfraDistrictRepository {
	return &InfraDistrictRepository{}
}

// Create 插入行政区划；行政编码重复转换为业务错误码。
func (r *InfraDistrictRepository) Create(ctx context.Context, m *model.InfraDistrictModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(districtTable).Insert(map[string]any{
		"id":               m.ID,
		"pid":              m.Pid,
		"area_name":        m.AreaName,
		"short_name":       m.ShortName,
		"area_code":        m.AreaCode,
		"parent_area_code": m.ParentAreaCode,
		"level":            m.Level,
		"sort":             m.Sort,
		"creator":          m.Creator,
		"created_at":       m.CreatedAt,
		"updater":          m.Updater,
		"updated_at":       m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(infraErrs.InfraDistrictCodeExists, m.AreaCode)
	}

	return err
}

// Update 更新名称、简称与排序；上级与行政编码创建后不可改，层级因此不会变化。
func (r *InfraDistrictRepository) Update(ctx context.Context, m *model.InfraDistrictModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(districtTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"area_name":  m.AreaName,
			"short_name": m.ShortName,
			"sort":       m.Sort,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(infraErrs.InfraDistrictNotExists, m.AreaName)
	}

	return nil
}

// Delete 软删除行政区划。
func (r *InfraDistrictRepository) Delete(ctx context.Context, m *model.InfraDistrictModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(districtTable).
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
		return baseErrs.New(infraErrs.InfraDistrictNotExists, m.AreaName)
	}

	return nil
}

// FindByID 按主键查询未删除的区划；未命中返回 (nil, nil)。
func (r *InfraDistrictRepository) FindByID(ctx context.Context, id types.ID) (*model.InfraDistrictModel, error) {
	var m model.InfraDistrictModel

	if err := facades.DB().WithContext(ctx).Table(districtTable).
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

// ExistsByCode 判断行政编码是否被占用，excludeID 用于编辑时排除自身。
func (r *InfraDistrictRepository) ExistsByCode(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(districtTable).
		Where("area_code", code).
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

// CountChildren 统计直接下级数量，用于删除前保护。
func (r *InfraDistrictRepository) CountChildren(ctx context.Context, pid types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(districtTable).
		Where("pid", pid).
		WhereNull("deleted_at").
		Count()
}

// All 返回全部区划，按级别与排序升序，供组织树组装。
func (r *InfraDistrictRepository) All(ctx context.Context) ([]model.InfraDistrictModel, error) {
	list := make([]model.InfraDistrictModel, 0)

	if err := facades.DB().WithContext(ctx).Table(districtTable).
		WhereNull("deleted_at").
		OrderBy("level").
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Children 返回指定上级的直接下级，按排序升序；pid 传 0 表示顶级。
func (r *InfraDistrictRepository) Children(ctx context.Context, pid types.ID) ([]model.InfraDistrictModel, error) {
	list := make([]model.InfraDistrictModel, 0)

	if err := facades.DB().WithContext(ctx).Table(districtTable).
		Where("pid", pid).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 按条件分页查询区划。
func (r *InfraDistrictRepository) Paginate(ctx context.Context, filter DistrictFilter, pageNo, pageSize int) ([]model.InfraDistrictModel, int64, error) {
	list := make([]model.InfraDistrictModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(districtTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		// 关键字命中名称、简称或编码，分组包裹避免与级别条件串成 OR
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("area_name", keyword).
				OrWhereLike("short_name", keyword).
				OrWhereLike("area_code", keyword)
		})
	}
	if filter.Level > 0 {
		query = query.Where("level", filter.Level)
	}
	if filter.Pid != nil {
		query = query.Where("pid", *filter.Pid)
	}

	if err := query.OrderBy("level").OrderBy("sort").OrderBy("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
