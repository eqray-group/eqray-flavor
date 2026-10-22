package repository

import (
	"context"
	"time"

	frameworkDB "github.com/goravel/framework/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
)

const storeRiderTable = model.StoreRiderTable

// riderDateLayout 是 store_rider.stat_date 的存储格式，与 MySQL DATE 对齐。
const riderDateLayout = "2006-01-02"

// StoreRiderFilter 是骑手列表的筛选条件，零值表示不限制。
type StoreRiderFilter struct {
	// StoreID 请求上下文指定的门店，优先级高于 StoreIDs
	StoreID types.ID
	// StoreIDs 可见门店集合；RestrictStores=true 时空集合表示无可见门店
	StoreIDs []types.ID
	// RestrictStores 是否按可见门店集合限制查询
	RestrictStores bool
	// EmployeeIDs 关键字命中的门店员工 ID 集合
	EmployeeIDs []types.ID
	// Status 骑手状态筛选
	Status *enums.RiderStatusEnum
}

// StoreRiderRepository 是 store_rider 骑手档案表的仓储。
// 接单与释放名额使用条件更新，避免「先查再改」造成超接。
type StoreRiderRepository struct{}

// NewStoreRiderRepository 构造骑手档案仓储。
func NewStoreRiderRepository() *StoreRiderRepository {
	return &StoreRiderRepository{}
}

// Create 插入骑手档案；employee_id 唯一冲突转换为业务错误码。
func (r *StoreRiderRepository) Create(ctx context.Context, m *model.StoreRiderModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeRiderTable).Insert(map[string]any{
		"id":             m.ID,
		"store_id":       m.StoreID,
		"employee_id":    m.EmployeeID,
		"status":         m.Status,
		"max_orders":     m.MaxOrders,
		"current_orders": m.CurrentOrders,
		"today_orders":   m.TodayOrders,
		"stat_date":      riderDateValue(m.StatDate),
		"remark":         m.Remark,
		"creator":        m.Creator,
		"created_at":     m.CreatedAt,
		"updater":        m.Updater,
		"updated_at":     m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StoreRiderExists)
	}

	return err
}

// UpdateConfig 只更新骑手可编辑的状态、并发上限与备注，不改写接单负载字段。
// 通过条件更新保证并发安全：切到休息 / 停用时当前负载必须为 0，且新上限不能低于当前负载。
// 返回影响行数；0 可能是「没有字段变化」或「条件不满足」，由服务层重新查询区分。
func (r *StoreRiderRepository) UpdateConfig(ctx context.Context, m *model.StoreRiderModel) (int64, error) {
	m.FillUpdate(ctx)

	query := facades.DB().WithContext(ctx).Table(storeRiderTable).
		Where("id", m.ID).
		WhereNull("deleted_at")
	if m.Status.BlocksOrders() {
		query = query.Where("current_orders", 0)
	}
	query = query.WhereRaw("current_orders <= ?", []any{m.MaxOrders})

	res, err := query.Update(map[string]any{
		"status":     m.Status,
		"max_orders": m.MaxOrders,
		"remark":     m.Remark,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return 0, baseErrs.New(storeErrs.StoreRiderExists)
	}
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// Restore 恢复已软删除的骑手档案并重置负载，供删除后重新绑定同一员工。
// 入参：m 需要恢复的档案；store_id 以员工当前门店为准，employee_id 沿用原记录。
func (r *StoreRiderRepository) Restore(ctx context.Context, m *model.StoreRiderModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(storeRiderTable).
		Where("id", m.ID).
		Update(map[string]any{
			"store_id":       m.StoreID,
			"status":         m.Status,
			"max_orders":     m.MaxOrders,
			"current_orders": 0,
			"today_orders":   0,
			"stat_date":      riderDateValue(m.StatDate),
			"remark":         m.Remark,
			"deleted_at":     nil,
			"updater":        m.Updater,
			"updated_at":     m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(storeErrs.StoreRiderNotExists)
	}

	return nil
}

// Delete 软删除骑手档案，保留历史记录。
// 条件 current_orders=0 与接单占用互斥，避免服务层校验后被并发接单插入进行中订单。
// 返回影响行数；0 可能是已删除或仍有进行中订单，由服务层重新查询区分。
func (r *StoreRiderRepository) Delete(ctx context.Context, m *model.StoreRiderModel) (int64, error) {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(storeRiderTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Where("current_orders", 0).
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// FindByID 按主键查询未删除的骑手档案；未命中返回 (nil, nil)。
func (r *StoreRiderRepository) FindByID(ctx context.Context, id types.ID) (*model.StoreRiderModel, error) {
	var m model.StoreRiderModel

	if err := facades.DB().WithContext(ctx).Table(storeRiderTable).
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

// FindByEmployeeID 查询员工当前有效的骑手档案；未命中返回 (nil, nil)。
func (r *StoreRiderRepository) FindByEmployeeID(ctx context.Context, employeeID types.ID) (*model.StoreRiderModel, error) {
	var m model.StoreRiderModel

	if err := facades.DB().WithContext(ctx).Table(storeRiderTable).
		Where("employee_id", employeeID).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByEmployeeIDAny 查询员工包含软删除在内的骑手档案。
// 删除后重新绑定同一员工时用它恢复原记录，避免唯一键冲突。
func (r *StoreRiderRepository) FindByEmployeeIDAny(ctx context.Context, employeeID types.ID) (*model.StoreRiderModel, error) {
	var m model.StoreRiderModel

	if err := facades.DB().WithContext(ctx).Table(storeRiderTable).
		Where("employee_id", employeeID).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// Paginate 按筛选条件分页查询骑手档案，按主键倒序返回。
func (r *StoreRiderRepository) Paginate(ctx context.Context, filter StoreRiderFilter, pageNo, pageSize int) ([]model.StoreRiderModel, int64, error) {
	list := make([]model.StoreRiderModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(storeRiderTable).WhereNull("deleted_at")
	if filter.StoreID > 0 {
		query = query.Where("store_id", filter.StoreID)
	} else if filter.RestrictStores {
		storeIDs := riderIDValues(filter.StoreIDs)
		if len(storeIDs) == 0 {
			return list, 0, nil
		}
		query = query.WhereIn("store_id", storeIDs)
	}
	if len(filter.EmployeeIDs) > 0 {
		employeeIDs := riderIDValues(filter.EmployeeIDs)
		if len(employeeIDs) == 0 {
			return list, 0, nil
		}
		query = query.WhereIn("employee_id", employeeIDs)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// TryAccept 以条件更新抢占一个接单名额。
// 条件：档案未删除、状态为接单中、当前负载小于并发上限；跨日时自动把今日单量重置为 1。
// 返回影响行数：1 表示抢占成功，0 表示状态不符或已达上限，由服务层查询具体原因。
func (r *StoreRiderRepository) TryAccept(ctx context.Context, id types.ID, today time.Time) (int64, error) {
	date := today.Format(riderDateLayout)

	res, err := facades.DB().WithContext(ctx).Table(storeRiderTable).
		Where("id", id).
		WhereNull("deleted_at").
		Where("status", enums.RiderStatusAccepting).
		WhereRaw("current_orders < max_orders", []any{}).
		Update(map[string]any{
			"current_orders": frameworkDB.Raw("current_orders + 1"),
			"today_orders":   frameworkDB.Raw("CASE WHEN stat_date = ? THEN today_orders + 1 ELSE 1 END", date),
			"stat_date":      date,
			"updated_at":     types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// Release 释放一个接单名额，当前负载为 0 时保持幂等。
// 条件更新保证并发释放不会把 current_orders 减成负数。
func (r *StoreRiderRepository) Release(ctx context.Context, id types.ID) (int64, error) {
	res, err := facades.DB().WithContext(ctx).Table(storeRiderTable).
		Where("id", id).
		WhereNull("deleted_at").
		WhereRaw("current_orders > 0", []any{}).
		Update(map[string]any{
			"current_orders": frameworkDB.Raw("current_orders - 1"),
			"updated_at":     types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// riderDateValue 把日期指针转成 DATE 列可写入的字符串；nil 落库为 NULL。
func riderDateValue(value *types.DateTime) any {
	if value == nil || value.IsZero() {
		return nil
	}

	return value.Time().Format(riderDateLayout)
}

// riderIDValues 把 ID 集合转为查询构造器可用的 any 切片，并过滤非正数。
func riderIDValues(ids []types.ID) []any {
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			values = append(values, id)
		}
	}

	return values
}
