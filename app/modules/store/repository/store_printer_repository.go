package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
)

const storePrinterTable = model.StorePrinterTable

// StorePrinterFilter 是打印设备列表的筛选条件，零值表示不限制。
type StorePrinterFilter struct {
	// StoreID 请求上下文指定的门店，优先级高于 StoreIDs
	StoreID types.ID
	// StoreIDs 可见门店集合；RestrictStores=true 时空集合表示无可见门店
	StoreIDs []types.ID
	// RestrictStores 是否按可见门店集合限制查询
	RestrictStores bool
	// Keyword 设备名称 / 设备号关键字
	Keyword string
	// Type 设备类型筛选
	Type *enums.PrinterTypeEnum
	// AccessMode 接入方式筛选
	AccessMode *enums.PrinterAccessModeEnum
	// Status 启停状态筛选；0 是合法筛选值，不能用零值判断未传
	Status *baseEnums.CommonStatusEnum
}

// StorePrinterRepository 是 store_printer 打印设备表的仓储。
type StorePrinterRepository struct{}

// NewStorePrinterRepository 构造打印设备仓储。
func NewStorePrinterRepository() *StorePrinterRepository {
	return &StorePrinterRepository{}
}

// Create 插入打印设备；同一门店内设备号冲突转换为业务错误码。
func (r *StorePrinterRepository) Create(ctx context.Context, m *model.StorePrinterModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storePrinterTable).Insert(map[string]any{
		"id":          m.ID,
		"store_id":    m.StoreID,
		"name":        m.Name,
		"device_no":   m.DeviceNo,
		"type":        m.Type,
		"access_mode": m.AccessMode,
		"config":      m.Config,
		"secret":      m.Secret,
		"secret_mask": m.SecretMask,
		"copies":      m.Copies,
		"status":      m.Status,
		"remark":      m.Remark,
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StorePrinterDeviceExists, m.DeviceNo)
	}

	return err
}

// Update 更新打印设备的全部可编辑字段；设备不能跨门店移动，因此不更新 store_id。
func (r *StorePrinterRepository) Update(ctx context.Context, m *model.StorePrinterModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storePrinterTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":        m.Name,
			"device_no":   m.DeviceNo,
			"type":        m.Type,
			"access_mode": m.AccessMode,
			"config":      m.Config,
			"secret":      m.Secret,
			"secret_mask": m.SecretMask,
			"copies":      m.Copies,
			"status":      m.Status,
			"remark":      m.Remark,
			"updater":     m.Updater,
			"updated_at":  m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StorePrinterDeviceExists, m.DeviceNo)
	}

	return err
}

// Delete 物理删除打印设备，避免软删除记录占用（store_id, device_no）唯一键。
func (r *StorePrinterRepository) Delete(ctx context.Context, m *model.StorePrinterModel) error {
	res, err := facades.DB().WithContext(ctx).Table(storePrinterTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(storeErrs.StorePrinterNotExists)
	}

	return nil
}

// FindByID 按主键查询未删除的打印设备；未命中返回 (nil, nil)。
func (r *StorePrinterRepository) FindByID(ctx context.Context, id types.ID) (*model.StorePrinterModel, error) {
	var m model.StorePrinterModel

	if err := facades.DB().WithContext(ctx).Table(storePrinterTable).
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

// ExistsByDeviceNo 判断同一门店内设备号是否已占用；excludeID > 0 时排除自身。
func (r *StorePrinterRepository) ExistsByDeviceNo(ctx context.Context, storeID types.ID, deviceNo string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(storePrinterTable).
		Where("store_id", storeID).
		Where("device_no", deviceNo).
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

// Paginate 按筛选条件分页查询打印设备，按门店与主键倒序返回。
func (r *StorePrinterRepository) Paginate(ctx context.Context, filter StorePrinterFilter, pageNo, pageSize int) ([]model.StorePrinterModel, int64, error) {
	list := make([]model.StorePrinterModel, 0)
	var total int64

	query := r.applyFilter(facades.DB().WithContext(ctx).Table(storePrinterTable).WhereNull("deleted_at"), filter)
	if err := query.OrderBy("store_id").OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// applyFilter 把门店范围与打印设备筛选条件应用到查询构造器。
// 入参：query 基础查询、filter 筛选条件；返回追加条件后的查询。
func (r *StorePrinterRepository) applyFilter(query db.Query, filter StorePrinterFilter) db.Query {
	if filter.StoreID > 0 {
		query = query.Where("store_id", filter.StoreID)
	} else if filter.RestrictStores {
		if len(filter.StoreIDs) == 0 {
			return query.WhereIn("store_id", []any{0})
		}
		query = query.WhereIn("store_id", storeIDValues(filter.StoreIDs))
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("name", like).OrWhereLike("device_no", like)
		})
	}
	if filter.Type != nil {
		query = query.Where("type", *filter.Type)
	}
	if filter.AccessMode != nil {
		query = query.Where("access_mode", *filter.AccessMode)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	return query
}
