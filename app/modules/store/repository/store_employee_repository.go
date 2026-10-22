package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
)

const storeEmployeeTable = model.StoreEmployeeTable

// StoreEmployeeFilter 是门店员工列表的筛选条件，零值表示不限制。
type StoreEmployeeFilter struct {
	// StoreID 请求上下文指定的门店，优先级高于 StoreIDs
	StoreID types.ID
	// StoreIDs 可见门店集合；RestrictStores=true 时空集合表示无可见门店
	StoreIDs []types.ID
	// RestrictStores 是否按可见门店集合限制查询
	RestrictStores bool
	// UserIDs 关键字命中的后台账号 ID 集合
	UserIDs []types.ID
	// Role 门店角色筛选
	Role *enums.StoreEmployeeRoleEnum
	// Status 启停状态筛选
	Status *baseEnums.CommonStatusEnum
}

// StoreEmployeeRepository 是 store_employee 关联表的仓储。
type StoreEmployeeRepository struct{}

// NewStoreEmployeeRepository 构造门店员工仓储。
func NewStoreEmployeeRepository() *StoreEmployeeRepository {
	return &StoreEmployeeRepository{}
}

// Create 插入门店员工关联；唯一键冲突转换为业务错误码。
// 关联表采用物理删除，唯一键不会被软删除记录占用。
func (r *StoreEmployeeRepository) Create(ctx context.Context, m *model.StoreEmployeeModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeEmployeeTable).Insert(map[string]any{
		"id":         m.ID,
		"store_id":   m.StoreID,
		"user_id":    m.UserID,
		"role":       m.Role,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StoreEmployeeExists)
	}

	return err
}

// Update 更新门店员工的所属门店、账号、角色、状态与备注。
// 唯一键冲突转换为业务错误码；服务层已先行加载记录。
func (r *StoreEmployeeRepository) Update(ctx context.Context, m *model.StoreEmployeeModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"store_id":   m.StoreID,
			"user_id":    m.UserID,
			"role":       m.Role,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StoreEmployeeExists)
	}

	return err
}

// Delete 物理删除门店员工关联，避免占用 (store_id, user_id) 唯一键。
func (r *StoreEmployeeRepository) Delete(ctx context.Context, m *model.StoreEmployeeModel) error {
	res, err := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(storeErrs.StoreEmployeeNotExists)
	}

	return nil
}

// FindByID 按主键查询门店员工关联；未命中返回 (nil, nil)。
func (r *StoreEmployeeRepository) FindByID(ctx context.Context, id types.ID) (*model.StoreEmployeeModel, error) {
	var m model.StoreEmployeeModel

	if err := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
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

// FindByStoreUser 查询门店与账号的绑定记录；未命中返回 (nil, nil)。
// 骑手开通时用它复用已存在的员工关联，避免重复分配。
func (r *StoreEmployeeRepository) FindByStoreUser(ctx context.Context, storeID, userID types.ID) (*model.StoreEmployeeModel, error) {
	var m model.StoreEmployeeModel

	if err := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
		Where("store_id", storeID).
		Where("user_id", userID).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByIDs 批量查询未删除的门店员工关联，供骑手列表补齐账号与门店信息。
func (r *StoreEmployeeRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.StoreEmployeeModel, error) {
	list := make([]model.StoreEmployeeModel, 0)
	values := employeeIDValues(ids)
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ListByUserIDs 查询指定账号在门店内的员工关联，供骑手列表按账号关键字筛选。
// storeID > 0 时限定门店；userIDs 为空时直接返回空列表。
func (r *StoreEmployeeRepository) ListByUserIDs(ctx context.Context, storeID types.ID, userIDs []types.ID) ([]model.StoreEmployeeModel, error) {
	list := make([]model.StoreEmployeeModel, 0)
	values := employeeIDValues(userIDs)
	if len(values) == 0 {
		return list, nil
	}

	query := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
		WhereIn("user_id", values).
		WhereNull("deleted_at")
	if storeID > 0 {
		query = query.Where("store_id", storeID)
	}
	if err := query.Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ExistsByStoreUser 判断门店与账号是否已绑定；excludeID > 0 时排除自身。
func (r *StoreEmployeeRepository) ExistsByStoreUser(ctx context.Context, storeID, userID, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
		Where("store_id", storeID).
		Where("user_id", userID).
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

// HasAnyByUser 统计用户是否存在任何门店关联（含禁用记录）。
// 该结果用于区分总部账号（无关联，跨店可见）与门店员工（必须按可见集合隔离）。
func (r *StoreEmployeeRepository) HasAnyByUser(ctx context.Context, userID types.ID) (bool, error) {
	total, err := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return false, err
	}

	return total > 0, nil
}

// ListStoreIDsByUser 返回用户关联的门店 ID；status 非空时按启停状态过滤。
// 不做门店有效性过滤，由服务层再用 store 表筛出仍有效的门店节点。
func (r *StoreEmployeeRepository) ListStoreIDsByUser(ctx context.Context, userID types.ID, status *baseEnums.CommonStatusEnum) ([]types.ID, error) {
	rows := make([]struct {
		StoreID types.ID `db:"store_id"`
	}, 0)

	query := facades.DB().WithContext(ctx).Table(storeEmployeeTable).
		Select("store_id").
		Where("user_id", userID).
		WhereNull("deleted_at")
	if status != nil {
		query = query.Where("status", *status)
	}
	if err := query.Get(&rows); err != nil {
		return nil, err
	}

	ids := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		if row.StoreID > 0 {
			ids = append(ids, row.StoreID)
		}
	}

	return ids, nil
}

// Paginate 按筛选条件分页查询门店员工关联，按主键倒序返回。
func (r *StoreEmployeeRepository) Paginate(ctx context.Context, filter StoreEmployeeFilter, pageNo, pageSize int) ([]model.StoreEmployeeModel, int64, error) {
	list := make([]model.StoreEmployeeModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(storeEmployeeTable).WhereNull("deleted_at")
	if filter.StoreID > 0 {
		query = query.Where("store_id", filter.StoreID)
	} else if filter.RestrictStores {
		if len(filter.StoreIDs) == 0 {
			return list, 0, nil
		}
		query = query.WhereIn("store_id", employeeIDValues(filter.StoreIDs))
	}
	if len(filter.UserIDs) > 0 {
		query = query.WhereIn("user_id", employeeIDValues(filter.UserIDs))
	}
	if filter.Role != nil {
		query = query.Where("role", *filter.Role)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// employeeIDValues 把 ID 集合转为查询构造器可用的 any 切片，并过滤非正数。
func employeeIDValues(ids []types.ID) []any {
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			values = append(values, id)
		}
	}

	return values
}
