package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	dishErrs "eqray-flavor/app/modules/dish/errs"
	"eqray-flavor/app/modules/dish/model"
)

const dishCategoryTable = model.DishCategoryTable

// DishCategoryRepository 是 dish_category 表的仓储，负责分类树的读写。
type DishCategoryRepository struct{}

// NewDishCategoryRepository 构造菜品分类仓储。
func NewDishCategoryRepository() *DishCategoryRepository {
	return &DishCategoryRepository{}
}

// All 返回全部未删除的分类，按排序与主键升序，供服务层在内存中组树。
func (r *DishCategoryRepository) All(ctx context.Context) ([]model.DishCategoryModel, error) {
	list := make([]model.DishCategoryModel, 0)
	if err := facades.DB().WithContext(ctx).Table(dishCategoryTable).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// FindByID 按主键查询未删除的分类；未命中返回 (nil, nil)。
func (r *DishCategoryRepository) FindByID(ctx context.Context, id types.ID) (*model.DishCategoryModel, error) {
	var m model.DishCategoryModel
	if err := facades.DB().WithContext(ctx).Table(dishCategoryTable).
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

// FindByIDs 批量查询未删除的分类，供菜品列表补齐分类名称。
func (r *DishCategoryRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.DishCategoryModel, error) {
	list := make([]model.DishCategoryModel, 0)
	values := idValues(uniqueIDs(ids))
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishCategoryTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ExistsByName 判断同一上级下分类名称是否已占用；excludeID > 0 时排除自身。
func (r *DishCategoryRepository) ExistsByName(ctx context.Context, parentID types.ID, name string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(dishCategoryTable).
		Where("parent_id", parentID).
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

// HasChildren 判断分类下是否仍有未删除的子分类。
func (r *DishCategoryRepository) HasChildren(ctx context.Context, id types.ID) (bool, error) {
	total, err := facades.DB().WithContext(ctx).Table(dishCategoryTable).
		Where("parent_id", id).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return false, err
	}

	return total > 0, nil
}

// Create 在事务内插入分类；同一上级下名称冲突转换为业务错误码。
func (r *DishCategoryRepository) Create(ctx context.Context, tx db.Tx, m *model.DishCategoryModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(dishCategoryTable).Insert(map[string]any{
		"id":         m.ID,
		"parent_id":  m.ParentID,
		"name":       m.Name,
		"image":      m.Image,
		"sort":       m.Sort,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishCategoryExists, m.Name)
	}

	return err
}

// Update 在事务内更新分类的名称、图标、排序、状态与备注；不改动上级关系之外的字段。
func (r *DishCategoryRepository) Update(ctx context.Context, tx db.Tx, m *model.DishCategoryModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(dishCategoryTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"parent_id":  m.ParentID,
			"name":       m.Name,
			"image":      m.Image,
			"sort":       m.Sort,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishCategoryExists, m.Name)
	}

	return err
}

// Delete 物理删除分类，避免软删除记录占用 (parent_id, name) 唯一键。
func (r *DishCategoryRepository) Delete(ctx context.Context, tx db.Tx, id types.ID) error {
	res, err := tx.Table(dishCategoryTable).
		Where("id", id).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(dishErrs.DishCategoryNotExists)
	}

	return nil
}
