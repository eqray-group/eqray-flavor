package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
)

const storeTableTable = model.StoreTableTable

// StoreTableFilter 是桌台列表的筛选条件，零值表示不限制。
type StoreTableFilter struct {
	// StoreID 请求上下文指定的门店，优先级高于 StoreIDs
	StoreID types.ID
	// StoreIDs 可见门店集合；RestrictStores=true 时空集合表示无可见门店
	StoreIDs []types.ID
	// RestrictStores 是否按可见门店集合限制查询
	RestrictStores bool
	// Keyword 桌号 / 桌码关键字
	Keyword string
	// AreaID 所属区域筛选
	AreaID types.ID
	// Status 桌台状态筛选
	Status *enums.TableStatusEnum
}

// StoreTableRepository 是 store_table 表的仓储。
type StoreTableRepository struct{}

// NewStoreTableRepository 构造桌台仓储。
func NewStoreTableRepository() *StoreTableRepository {
	return &StoreTableRepository{}
}

// Create 插入桌台；桌号或桌码冲突转换为业务错误码。
func (r *StoreTableRepository) Create(ctx context.Context, m *model.StoreTableModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeTableTable).Insert(storeTableInsertData(m))
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StoreTableExists)
	}

	return err
}

// BatchCreate 在事务内批量插入桌台，任意一条冲突时整批回滚。
// 入参：已生成名称与桌码的桌台模型；返回错误表示事务已回滚。
func (r *StoreTableRepository) BatchCreate(ctx context.Context, list []*model.StoreTableModel) error {
	if len(list) == 0 {
		return nil
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		for _, m := range list {
			m.FillCreate(ctx)
			if _, err := tx.Table(storeTableTable).Insert(storeTableInsertData(m)); err != nil {
				if support.IsDuplicate(err) {
					return baseErrs.New(storeErrs.StoreTableBatchExists, m.Name)
				}
				return err
			}
		}

		return nil
	})
}

// Update 更新桌号、区域、座位数、排序与备注；桌码创建后不可修改。
func (r *StoreTableRepository) Update(ctx context.Context, m *model.StoreTableModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeTableTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"area_id":    m.AreaID,
			"name":       m.Name,
			"seats":      m.Seats,
			"sort":       m.Sort,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StoreTableNameExists, m.Name)
	}

	return err
}

// Delete 物理删除桌台；调用方需先确认桌台空闲且没有台位单。
func (r *StoreTableRepository) Delete(ctx context.Context, m *model.StoreTableModel) error {
	res, err := facades.DB().WithContext(ctx).Table(storeTableTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(storeErrs.StoreTableNotExists)
	}

	return nil
}

// FindByID 按主键查询未删除的桌台；未命中返回 (nil, nil)。
func (r *StoreTableRepository) FindByID(ctx context.Context, id types.ID) (*model.StoreTableModel, error) {
	var m model.StoreTableModel

	if err := facades.DB().WithContext(ctx).Table(storeTableTable).
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

// ExistsByName 判断同一门店内桌号是否已占用；excludeID > 0 时排除自身。
func (r *StoreTableRepository) ExistsByName(ctx context.Context, storeID types.ID, name string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(storeTableTable).
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

// ExistsByCode 判断桌码是否已占用；桌码全局唯一，excludeID > 0 时排除自身。
func (r *StoreTableRepository) ExistsByCode(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(storeTableTable).
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

// ExistingNames 批量查询门店内已存在的桌号，供批量生成前做冲突提示。
func (r *StoreTableRepository) ExistingNames(ctx context.Context, storeID types.ID, names []string) (map[string]bool, error) {
	result := make(map[string]bool, len(names))
	values := make([]any, 0, len(names))
	for _, name := range names {
		if name != "" {
			values = append(values, name)
		}
	}
	if len(values) == 0 {
		return result, nil
	}

	rows := make([]struct {
		Name string `db:"name"`
	}, 0)
	if err := facades.DB().WithContext(ctx).Table(storeTableTable).
		Select("name").
		Where("store_id", storeID).
		WhereIn("name", values).
		WhereNull("deleted_at").
		Get(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.Name] = true
	}

	return result, nil
}

// ExistingCodes 批量查询已存在的桌码，供批量生成前做冲突提示。
func (r *StoreTableRepository) ExistingCodes(ctx context.Context, codes []string) (map[string]bool, error) {
	result := make(map[string]bool, len(codes))
	values := make([]any, 0, len(codes))
	for _, code := range codes {
		if code != "" {
			values = append(values, code)
		}
	}
	if len(values) == 0 {
		return result, nil
	}

	rows := make([]struct {
		Code string `db:"code"`
	}, 0)
	if err := facades.DB().WithContext(ctx).Table(storeTableTable).
		Select("code").
		WhereIn("code", values).
		WhereNull("deleted_at").
		Get(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.Code] = true
	}

	return result, nil
}

// CountByAreaID 统计区域下的未删除桌台数量，用于删除区域前保护。
func (r *StoreTableRepository) CountByAreaID(ctx context.Context, areaID types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(storeTableTable).
		Where("area_id", areaID).
		WhereNull("deleted_at").
		Count()
}

// CountByAreaIDs 按区域统计桌台数量，供区域列表展示桌台数。
func (r *StoreTableRepository) CountByAreaIDs(ctx context.Context, areaIDs []types.ID) (map[types.ID]int64, error) {
	result := make(map[types.ID]int64, len(areaIDs))
	values := storeIDValues(areaIDs)
	if len(values) == 0 {
		return result, nil
	}

	rows := make([]struct {
		AreaID types.ID `db:"area_id"`
		Total  int64    `db:"total"`
	}, 0)
	if err := facades.DB().WithContext(ctx).Table(storeTableTable).
		Select("area_id, COUNT(*) AS total").
		WhereIn("area_id", values).
		WhereNull("deleted_at").
		GroupBy("area_id").
		Get(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.AreaID] = row.Total
	}

	return result, nil
}

// Paginate 按筛选条件分页查询桌台，按门店、区域与排序升序返回。
func (r *StoreTableRepository) Paginate(ctx context.Context, filter StoreTableFilter, pageNo, pageSize int) ([]model.StoreTableModel, int64, error) {
	list := make([]model.StoreTableModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(storeTableTable).WhereNull("deleted_at")
	if filter.StoreID > 0 {
		query = query.Where("store_id", filter.StoreID)
	} else if filter.RestrictStores {
		if len(filter.StoreIDs) == 0 {
			return list, 0, nil
		}
		query = query.WhereIn("store_id", storeIDValues(filter.StoreIDs))
	}
	if filter.AreaID > 0 {
		query = query.Where("area_id", filter.AreaID)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("name", like).OrWhereLike("code", like)
		})
	}

	if err := query.OrderBy("store_id").OrderBy("area_id").OrderBy("sort").OrderBy("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Open 把空闲桌台改为使用中并记录开台时间与台位单。
// 采用条件更新，避免两个收银员并发开同一张桌。
func (r *StoreTableRepository) Open(ctx context.Context, m *model.StoreTableModel, orderID types.ID) error {
	m.FillUpdate(ctx)
	openedAt := types.Now()

	res, err := facades.DB().WithContext(ctx).Table(storeTableTable).
		Where("id", m.ID).
		Where("status", enums.TableStatusIdle).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":           enums.TableStatusDining,
			"current_order_id": orderID,
			"opened_at":        openedAt,
			"updater":          m.Updater,
			"updated_at":       m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(storeErrs.StoreTableNotIdle)
	}

	m.Status = enums.TableStatusDining
	m.CurrentOrderID = orderID
	m.OpenedAt = &openedAt

	return nil
}

// Clear 把使用中或待清台的桌台恢复为空闲，并清空台位单与开台时间。
func (r *StoreTableRepository) Clear(ctx context.Context, m *model.StoreTableModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(storeTableTable).
		Where("id", m.ID).
		WhereIn("status", []any{enums.TableStatusDining, enums.TableStatusCleaning}).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":           enums.TableStatusIdle,
			"current_order_id": 0,
			"opened_at":        nil,
			"updater":          m.Updater,
			"updated_at":       m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(storeErrs.StoreTableStatusInvalid)
	}

	m.Status = enums.TableStatusIdle
	m.CurrentOrderID = 0
	m.OpenedAt = nil

	return nil
}

// storeTableInsertData 把桌台模型转换为 Insert 使用的字段集合。
func storeTableInsertData(m *model.StoreTableModel) map[string]any {
	return map[string]any{
		"id":               m.ID,
		"store_id":         m.StoreID,
		"area_id":          m.AreaID,
		"name":             m.Name,
		"code":             m.Code,
		"seats":            m.Seats,
		"sort":             m.Sort,
		"status":           m.Status,
		"current_order_id": m.CurrentOrderID,
		"opened_at":        m.OpenedAt,
		"remark":           m.Remark,
		"creator":          m.Creator,
		"created_at":       m.CreatedAt,
		"updater":          m.Updater,
		"updated_at":       m.UpdatedAt,
	}
}
