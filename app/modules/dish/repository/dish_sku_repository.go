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

const (
	dishSkuTable        = model.DishSkuTable
	dishStoreSkuTable   = model.DishStoreSkuTable
	dishStorePriceTable = model.DishStorePriceTable
)

// DishSkuRepository 是 dish_sku 表的仓储；门店可售与门店价格的级联清理也放在这里。
type DishSkuRepository struct{}

// NewDishSkuRepository 构造菜品规格仓储。
func NewDishSkuRepository() *DishSkuRepository {
	return &DishSkuRepository{}
}

// DishSkuFilter 是规格列表的筛选条件。
type DishSkuFilter struct {
	// Keyword 规格名称或编码关键字
	Keyword string
	// SpuIDs 菜品范围；RestrictSpu=true 时空集合表示无结果
	SpuIDs []types.ID
	// RestrictSpu 是否按菜品范围限制查询
	RestrictSpu bool
}

// Paginate 按筛选条件分页查询规格，按菜品与排序升序。
func (r *DishSkuRepository) Paginate(ctx context.Context, filter DishSkuFilter, pageNo, pageSize int) ([]model.DishSkuModel, int64, error) {
	list := make([]model.DishSkuModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(dishSkuTable).WhereNull("deleted_at")
	if filter.RestrictSpu {
		if len(filter.SpuIDs) == 0 {
			return list, 0, nil
		}
		query = query.WhereIn("spu_id", idValues(uniqueIDs(filter.SpuIDs)))
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("name", like).OrWhereLike("code", like)
		})
	}

	if err := query.OrderBy("spu_id").OrderBy("sort").OrderBy("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByID 按主键查询未删除的规格；未命中返回 (nil, nil)。
func (r *DishSkuRepository) FindByID(ctx context.Context, id types.ID) (*model.DishSkuModel, error) {
	var m model.DishSkuModel
	if err := facades.DB().WithContext(ctx).Table(dishSkuTable).
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

// FindByIDs 批量查询规格；供订单试算与下单按 SKU 集合取价使用。
func (r *DishSkuRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.DishSkuModel, error) {
	list := make([]model.DishSkuModel, 0)
	if len(ids) == 0 {
		return list, nil
	}
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}
	if err := facades.DB().WithContext(ctx).Table(dishSkuTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// FindBySpuIDs 批量查询菜品下的规格，按 SPU 与排序升序。
func (r *DishSkuRepository) FindBySpuIDs(ctx context.Context, spuIDs []types.ID) ([]model.DishSkuModel, error) {
	list := make([]model.DishSkuModel, 0)
	values := idValues(uniqueIDs(spuIDs))
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishSkuTable).
		WhereIn("spu_id", values).
		WhereNull("deleted_at").
		OrderBy("spu_id").
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// CountBySpuIDs 统计各菜品下的规格数量，返回 SPU ID 到数量的映射。
func (r *DishSkuRepository) CountBySpuIDs(ctx context.Context, spuIDs []types.ID) (map[types.ID]int64, error) {
	rows := make([]struct {
		SpuID types.ID `db:"spu_id"`
		Total int64    `db:"total"`
	}, 0)

	values := idValues(uniqueIDs(spuIDs))
	if len(values) == 0 {
		return map[types.ID]int64{}, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishSkuTable).
		Select("spu_id", "COUNT(*) AS total").
		WhereIn("spu_id", values).
		WhereNull("deleted_at").
		GroupBy("spu_id").
		Get(&rows); err != nil {
		return nil, err
	}

	counts := make(map[types.ID]int64, len(rows))
	for _, row := range rows {
		counts[row.SpuID] = row.Total
	}

	return counts, nil
}

// ExistsByName 判断同一菜品下规格名称是否已占用；excludeID > 0 时排除自身。
func (r *DishSkuRepository) ExistsByName(ctx context.Context, spuID types.ID, name string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(dishSkuTable).
		Where("spu_id", spuID).
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

// ExistsByCode 判断规格编码是否已占用；空编码不参与判重。
func (r *DishSkuRepository) ExistsByCode(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	if code == "" {
		return false, nil
	}

	query := facades.DB().WithContext(ctx).Table(dishSkuTable).
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

// Create 在事务内批量插入规格；同名校验由服务层完成，这里只兜底唯一键冲突。
func (r *DishSkuRepository) Create(ctx context.Context, tx db.Tx, list []*model.DishSkuModel) error {
	for _, m := range list {
		m.FillCreate(ctx)
		if _, err := tx.Table(dishSkuTable).Insert(map[string]any{
			"id":             m.ID,
			"spu_id":         m.SpuID,
			"name":           m.Name,
			"code":           m.Code,
			"standard_price": m.StandardPrice,
			"member_price":   m.MemberPrice,
			"cost_price":     m.CostPrice,
			"status":         m.Status,
			"sort":           m.Sort,
			"remark":         m.Remark,
			"creator":        m.Creator,
			"created_at":     m.CreatedAt,
			"updater":        m.Updater,
			"updated_at":     m.UpdatedAt,
		}); err != nil {
			if support.IsDuplicate(err) {
				return baseErrs.New(dishErrs.DishSkuNameExists, m.Name)
			}
			return err
		}
	}

	return nil
}

// Update 在事务内更新规格；SPU 归属不允许修改。
func (r *DishSkuRepository) Update(ctx context.Context, tx db.Tx, m *model.DishSkuModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(dishSkuTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":           m.Name,
			"code":           m.Code,
			"standard_price": m.StandardPrice,
			"member_price":   m.MemberPrice,
			"cost_price":     m.CostPrice,
			"status":         m.Status,
			"sort":           m.Sort,
			"remark":         m.Remark,
			"updater":        m.Updater,
			"updated_at":     m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishSkuNameExists, m.Name)
	}

	return err
}

// DeleteByIDs 在事务内物理删除指定规格。
func (r *DishSkuRepository) DeleteByIDs(ctx context.Context, tx db.Tx, ids []types.ID) error {
	values := idValues(uniqueIDs(ids))
	if len(values) == 0 {
		return nil
	}

	_, err := tx.Table(dishSkuTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Delete()

	return err
}

// DeleteBySpuID 在事务内删除某个菜品下的全部规格。
func (r *DishSkuRepository) DeleteBySpuID(ctx context.Context, tx db.Tx, spuID types.ID) error {
	_, err := tx.Table(dishSkuTable).
		Where("spu_id", spuID).
		WhereNull("deleted_at").
		Delete()

	return err
}

// DeleteStoreConfigs 在事务内清理规格的门店可售与门店价格配置，避免菜品删除后留下孤儿数据。
func (r *DishSkuRepository) DeleteStoreConfigs(ctx context.Context, tx db.Tx, skuIDs []types.ID) error {
	values := idValues(uniqueIDs(skuIDs))
	if len(values) == 0 {
		return nil
	}

	if _, err := tx.Table(dishStoreSkuTable).WhereIn("sku_id", values).Delete(); err != nil {
		return err
	}

	_, err := tx.Table(dishStorePriceTable).WhereIn("sku_id", values).Delete()

	return err
}
