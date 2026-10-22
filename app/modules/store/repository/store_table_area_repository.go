package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
)

const storeTableAreaTable = model.StoreTableAreaTable

// StoreTableAreaFilter 是桌台区域列表的筛选条件，零值表示不限制。
type StoreTableAreaFilter struct {
	// StoreID 请求上下文指定的门店，优先级高于 StoreIDs
	StoreID types.ID
	// StoreIDs 可见门店集合；RestrictStores=true 时空集合表示无可见门店
	StoreIDs []types.ID
	// RestrictStores 是否按可见门店集合限制查询
	RestrictStores bool
	// Keyword 区域名称关键字
	Keyword string
}

// StoreTableAreaRepository 是 store_table_area 表的仓储。
type StoreTableAreaRepository struct{}

// NewStoreTableAreaRepository 构造桌台区域仓储。
func NewStoreTableAreaRepository() *StoreTableAreaRepository {
	return &StoreTableAreaRepository{}
}

// Create 插入桌台区域；同一门店内名称冲突转换为业务错误码。
func (r *StoreTableAreaRepository) Create(ctx context.Context, m *model.StoreTableAreaModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeTableAreaTable).Insert(map[string]any{
		"id":         m.ID,
		"store_id":   m.StoreID,
		"name":       m.Name,
		"sort":       m.Sort,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StoreTableAreaExists, m.Name)
	}

	return err
}

// Update 更新区域名称、排序与备注；区域不能跨门店移动，因此不更新 store_id。
func (r *StoreTableAreaRepository) Update(ctx context.Context, m *model.StoreTableAreaModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeTableAreaTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"sort":       m.Sort,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StoreTableAreaExists, m.Name)
	}

	return err
}

// Delete 物理删除桌台区域，避免软删除记录占用（store_id, name）唯一键。
func (r *StoreTableAreaRepository) Delete(ctx context.Context, m *model.StoreTableAreaModel) error {
	res, err := facades.DB().WithContext(ctx).Table(storeTableAreaTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(storeErrs.StoreTableAreaNotExists)
	}

	return nil
}

// FindByID 按主键查询未删除的桌台区域；未命中返回 (nil, nil)。
func (r *StoreTableAreaRepository) FindByID(ctx context.Context, id types.ID) (*model.StoreTableAreaModel, error) {
	var m model.StoreTableAreaModel

	if err := facades.DB().WithContext(ctx).Table(storeTableAreaTable).
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

// FindByIDs 批量查询未删除的桌台区域，供桌台列表补齐区域名称。
func (r *StoreTableAreaRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.StoreTableAreaModel, error) {
	list := make([]model.StoreTableAreaModel, 0)
	values := storeIDValues(ids)
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(storeTableAreaTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ExistsByName 判断同一门店内区域名称是否已占用；excludeID > 0 时排除自身。
func (r *StoreTableAreaRepository) ExistsByName(ctx context.Context, storeID types.ID, name string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(storeTableAreaTable).
		Where("store_id", storeID).
		Where("name", name).
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

// All 返回筛选范围内的全部桌台区域，按门店与排序升序，供下拉选择使用。
func (r *StoreTableAreaRepository) All(ctx context.Context, filter StoreTableAreaFilter) ([]model.StoreTableAreaModel, error) {
	list := make([]model.StoreTableAreaModel, 0)

	query := r.applyFilter(facades.DB().WithContext(ctx).Table(storeTableAreaTable).WhereNull("deleted_at"), filter)
	if err := query.OrderBy("store_id").OrderBy("sort").OrderBy("id").Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 按筛选条件分页查询桌台区域。
func (r *StoreTableAreaRepository) Paginate(ctx context.Context, filter StoreTableAreaFilter, pageNo, pageSize int) ([]model.StoreTableAreaModel, int64, error) {
	list := make([]model.StoreTableAreaModel, 0)
	var total int64

	query := r.applyFilter(facades.DB().WithContext(ctx).Table(storeTableAreaTable).WhereNull("deleted_at"), filter)
	if err := query.OrderBy("store_id").OrderBy("sort").OrderBy("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// applyFilter 把门店范围与关键字条件应用到查询构造器。
// 入参：query 基础查询、filter 筛选条件；返回追加条件后的查询。
func (r *StoreTableAreaRepository) applyFilter(query db.Query, filter StoreTableAreaFilter) db.Query {
	if filter.StoreID > 0 {
		query = query.Where("store_id", filter.StoreID)
	} else if filter.RestrictStores {
		if len(filter.StoreIDs) == 0 {
			return query.WhereIn("store_id", []any{0})
		}
		query = query.WhereIn("store_id", storeIDValues(filter.StoreIDs))
	}
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
	}

	return query
}
