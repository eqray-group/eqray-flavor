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

const dishSpuTable = model.DishSpuTable

// DishSpuFilter 是菜品列表的筛选条件，零值表示不限制。
type DishSpuFilter struct {
	// Keyword 菜品名称或编码关键字
	Keyword string
	// CategoryIDs 分类范围（含下级分类），空表示全部分类
	CategoryIDs []types.ID
	// Status 上下架状态，nil 表示不限
	Status *int8
}

// DishSpuRepository 是 dish_spu 表的仓储。
type DishSpuRepository struct{}

// NewDishSpuRepository 构造菜品仓储。
func NewDishSpuRepository() *DishSpuRepository {
	return &DishSpuRepository{}
}

// Paginate 按筛选条件分页查询菜品，按分类与排序升序。
func (r *DishSpuRepository) Paginate(ctx context.Context, filter DishSpuFilter, pageNo, pageSize int) ([]model.DishSpuModel, int64, error) {
	list := make([]model.DishSpuModel, 0)
	var total int64

	query := r.applyFilter(facades.DB().WithContext(ctx).Table(dishSpuTable).WhereNull("deleted_at"), filter)
	if err := query.OrderBy("category_id").OrderBy("sort").OrderBy("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindIDsByFilter 返回命中筛选条件的菜品 ID，供门店菜品列表先按菜品过滤再分页规格。
func (r *DishSpuRepository) FindIDsByFilter(ctx context.Context, filter DishSpuFilter) ([]types.ID, error) {
	rows := make([]model.DishSpuModel, 0)
	query := r.applyFilter(facades.DB().WithContext(ctx).Table(dishSpuTable).Select("id").WhereNull("deleted_at"), filter)
	if err := query.OrderBy("category_id").OrderBy("sort").OrderBy("id").Get(&rows); err != nil {
		return nil, err
	}

	ids := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	return ids, nil
}

// FindByID 按主键查询未删除的菜品；未命中返回 (nil, nil)。
func (r *DishSpuRepository) FindByID(ctx context.Context, id types.ID) (*model.DishSpuModel, error) {
	var m model.DishSpuModel
	if err := facades.DB().WithContext(ctx).Table(dishSpuTable).
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

// FindByIDs 批量查询未删除的菜品。
func (r *DishSpuRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.DishSpuModel, error) {
	list := make([]model.DishSpuModel, 0)
	values := idValues(uniqueIDs(ids))
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishSpuTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ExistsByName 判断同一分类下菜品名称是否已占用；excludeID > 0 时排除自身。
func (r *DishSpuRepository) ExistsByName(ctx context.Context, categoryID types.ID, name string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(dishSpuTable).
		Where("category_id", categoryID).
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

// ExistsByCode 判断菜品编码是否已占用；空编码不参与判重。
func (r *DishSpuRepository) ExistsByCode(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	if code == "" {
		return false, nil
	}

	query := facades.DB().WithContext(ctx).Table(dishSpuTable).
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

// CountByCategoryIDs 统计各分类下的菜品数量，返回分类 ID 到数量的映射。
func (r *DishSpuRepository) CountByCategoryIDs(ctx context.Context, categoryIDs []types.ID) (map[types.ID]int64, error) {
	rows := make([]struct {
		CategoryID types.ID `db:"category_id"`
		Total      int64    `db:"total"`
	}, 0)

	values := idValues(uniqueIDs(categoryIDs))
	if len(values) == 0 {
		return map[types.ID]int64{}, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishSpuTable).
		Select("category_id", "COUNT(*) AS total").
		WhereIn("category_id", values).
		WhereNull("deleted_at").
		GroupBy("category_id").
		Get(&rows); err != nil {
		return nil, err
	}

	counts := make(map[types.ID]int64, len(rows))
	for _, row := range rows {
		counts[row.CategoryID] = row.Total
	}

	return counts, nil
}

// CountByCategoryID 统计指定分类下的菜品数量，删除分类前校验用。
func (r *DishSpuRepository) CountByCategoryID(ctx context.Context, categoryID types.ID) (int64, error) {
	total, err := facades.DB().WithContext(ctx).Table(dishSpuTable).
		Where("category_id", categoryID).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return 0, err
	}

	return total, nil
}

// Create 在事务内插入菜品；同一分类下名称冲突转换为业务错误码。
func (r *DishSpuRepository) Create(ctx context.Context, tx db.Tx, m *model.DishSpuModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(dishSpuTable).Insert(map[string]any{
		"id":          m.ID,
		"category_id": m.CategoryID,
		"name":        m.Name,
		"code":        m.Code,
		"unit":        m.Unit,
		"description": m.Description,
		"image":       m.Image,
		"tags":        m.Tags,
		"allergens":   m.Allergens,
		"status":      m.Status,
		"sort":        m.Sort,
		"remark":      m.Remark,
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishSpuNameExists, m.Name)
	}

	return err
}

// Update 在事务内更新菜品基础信息与上下架状态。
func (r *DishSpuRepository) Update(ctx context.Context, tx db.Tx, m *model.DishSpuModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(dishSpuTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"category_id": m.CategoryID,
			"name":        m.Name,
			"code":        m.Code,
			"unit":        m.Unit,
			"description": m.Description,
			"image":       m.Image,
			"tags":        m.Tags,
			"allergens":   m.Allergens,
			"status":      m.Status,
			"sort":        m.Sort,
			"remark":      m.Remark,
			"updater":     m.Updater,
			"updated_at":  m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishSpuNameExists, m.Name)
	}

	return err
}

// Delete 物理删除菜品，SKU 与门店配置由服务层在同一事务内级联清理。
func (r *DishSpuRepository) Delete(ctx context.Context, tx db.Tx, id types.ID) error {
	res, err := tx.Table(dishSpuTable).
		Where("id", id).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(dishErrs.DishSpuNotExists)
	}

	return nil
}

// applyFilter 把关键字、分类范围与状态条件应用到查询构造器。
func (r *DishSpuRepository) applyFilter(query db.Query, filter DishSpuFilter) db.Query {
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		// 关键字命中菜品名称或编码，分组包裹避免与分类/状态条件串成 OR
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("name", like).OrWhereLike("code", like)
		})
	}
	if len(filter.CategoryIDs) > 0 {
		query = query.WhereIn("category_id", idValues(uniqueIDs(filter.CategoryIDs)))
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	return query
}
