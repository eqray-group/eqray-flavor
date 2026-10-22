// Package repository 提供门店域的数据库访问，只做查询与写入，不含业务分支。
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

const storeTable = model.StoreTable

// StoreFilter 是门店列表的筛选条件，零值表示不限制。
type StoreFilter struct {
	Keyword string
	Type    *enums.StoreTypeEnum
	Status  *enums.StoreStatusEnum
	// IDs 限定结果集合，由 Service 按组织树算出，空表示不限
	IDs []types.ID
}

// StoreRepository 是 store 表的仓储。
type StoreRepository struct{}

// NewStoreRepository 构造门店仓储。
func NewStoreRepository() *StoreRepository {
	return &StoreRepository{}
}

// Create 插入门店；编码唯一冲突转换为业务错误码。
func (r *StoreRepository) Create(ctx context.Context, m *model.StoreModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeTable).Insert(map[string]any{
		"id":               m.ID,
		"parent_id":        m.ParentID,
		"type":             m.Type,
		"code":             m.Code,
		"name":             m.Name,
		"short_name":       m.ShortName,
		"status":           m.Status,
		"contact_name":     m.ContactName,
		"contact_mobile":   m.ContactMobile,
		"district_id":      m.DistrictID,
		"address":          m.Address,
		"longitude":        m.Longitude,
		"latitude":         m.Latitude,
		"open_time":        m.OpenTime,
		"close_time":       m.CloseTime,
		"delivery_rule":    m.DeliveryRule,
		"auto_accept":      m.AutoAccept,
		"online_enabled":   m.OnlineEnabled,
		"delivery_enabled": m.DeliveryEnabled,
		"sort":             m.Sort,
		"remark":           m.Remark,
		"creator":          m.Creator,
		"created_at":       m.CreatedAt,
		"updater":          m.Updater,
		"updated_at":       m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StoreCodeExists, m.Code)
	}

	return err
}

// Update 更新门店可变字段；编码创建后不可修改，因此不在更新集合里。
func (r *StoreRepository) Update(ctx context.Context, m *model.StoreModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storeTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"parent_id":        m.ParentID,
			"type":             m.Type,
			"name":             m.Name,
			"short_name":       m.ShortName,
			"status":           m.Status,
			"contact_name":     m.ContactName,
			"contact_mobile":   m.ContactMobile,
			"district_id":      m.DistrictID,
			"address":          m.Address,
			"longitude":        m.Longitude,
			"latitude":         m.Latitude,
			"open_time":        m.OpenTime,
			"close_time":       m.CloseTime,
			"delivery_rule":    m.DeliveryRule,
			"auto_accept":      m.AutoAccept,
			"online_enabled":   m.OnlineEnabled,
			"delivery_enabled": m.DeliveryEnabled,
			"sort":             m.Sort,
			"remark":           m.Remark,
			"updater":          m.Updater,
			"updated_at":       m.UpdatedAt,
		})

	return err
}

// Delete 软删除门店。
func (r *StoreRepository) Delete(ctx context.Context, m *model.StoreModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(storeTable).
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
		return baseErrs.New(storeErrs.StoreNotExists, m.Name)
	}

	return nil
}

// FindByID 按主键查询未删除的门店；未命中返回 (nil, nil)。
func (r *StoreRepository) FindByID(ctx context.Context, id types.ID) (*model.StoreModel, error) {
	var m model.StoreModel

	if err := facades.DB().WithContext(ctx).Table(storeTable).
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

// ExistsByCode 判断门店编码是否被占用，excludeID 用于编辑时排除自身。
func (r *StoreRepository) ExistsByCode(ctx context.Context, code string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(storeTable).
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

// CountChildren 统计直接下级节点数量，用于删除前保护。
func (r *StoreRepository) CountChildren(ctx context.Context, id types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(storeTable).
		Where("parent_id", id).
		WhereNull("deleted_at").
		Count()
}

// CountByType 统计指定类型的未删除节点数量，excludeID 用于编辑时排除自身。
// 品牌总部唯一性校验依赖该查询，新增传 0。
func (r *StoreRepository) CountByType(ctx context.Context, typ enums.StoreTypeEnum, excludeID types.ID) (int64, error) {
	query := facades.DB().WithContext(ctx).Table(storeTable).
		Where("type", typ).
		WhereNull("deleted_at")
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	return query.Count()
}

// All 返回全部门店节点，按层级与排序升序，供组织树组装。
func (r *StoreRepository) All(ctx context.Context) ([]model.StoreModel, error) {
	list := make([]model.StoreModel, 0)

	if err := facades.DB().WithContext(ctx).Table(storeTable).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// FindByIDs 批量查询未删除的门店节点，供员工列表补齐门店名称。
// 入参：ids 门店 ID 集合；返回门店列表，未命中的 ID 直接忽略。
func (r *StoreRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.StoreModel, error) {
	list := make([]model.StoreModel, 0)
	values := storeIDValues(ids)
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(storeTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ActiveShopsByIDs 从给定 ID 中筛出未删除的门店类型节点，供可见门店集合使用。
// 入参：ids 候选门店 ID；返回仍有效的门店列表。
func (r *StoreRepository) ActiveShopsByIDs(ctx context.Context, ids []types.ID) ([]model.StoreModel, error) {
	list := make([]model.StoreModel, 0)
	values := storeIDValues(ids)
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(storeTable).
		WhereIn("id", values).
		Where("type", enums.StoreTypeShop).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// AllShops 返回全部未删除的门店类型节点，供总部账号与超管加载门店切换器。
func (r *StoreRepository) AllShops(ctx context.Context) ([]model.StoreModel, error) {
	list := make([]model.StoreModel, 0)

	if err := facades.DB().WithContext(ctx).Table(storeTable).
		Where("type", enums.StoreTypeShop).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 按条件分页查询门店。
func (r *StoreRepository) Paginate(ctx context.Context, filter StoreFilter, pageNo, pageSize int) ([]model.StoreModel, int64, error) {
	list := make([]model.StoreModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(storeTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		// 关键字命中名称、编码或简称，分组包裹避免与类型/状态条件串成 OR
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("name", keyword).
				OrWhereLike("code", keyword).
				OrWhereLike("short_name", keyword)
		})
	}
	if filter.Type != nil {
		query = query.Where("type", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}
	if values := storeIDValues(filter.IDs); len(values) > 0 {
		query = query.WhereIn("id", values)
	}

	if err := query.OrderBy("sort").OrderBy("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// storeIDValues 把门店 ID 集合转换为查询构造器可用的 any 切片，并过滤非正数。
func storeIDValues(ids []types.ID) []any {
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			values = append(values, id)
		}
	}

	return values
}
