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

const storePrintTemplateTable = model.StorePrintTemplateTable

// StorePrintTemplateFilter 是打印模板列表的筛选条件，零值表示不限制。
type StorePrintTemplateFilter struct {
	// StoreID 请求上下文指定的门店，优先级高于 StoreIDs
	StoreID types.ID
	// StoreIDs 可见门店集合；RestrictStores=true 时空集合表示无可见门店
	StoreIDs []types.ID
	// RestrictStores 是否按可见门店集合限制查询
	RestrictStores bool
	// Keyword 模板名称关键字
	Keyword string
	// PrinterType 适用设备类型筛选
	PrinterType *enums.PrinterTypeEnum
	// PrintType 业务打印类型筛选
	PrintType *enums.PrintTypeEnum
	// Status 启停状态筛选；0 是合法筛选值，不能用零值判断未传
	Status *baseEnums.CommonStatusEnum
}

// StorePrintTemplateRepository 是 store_print_template 打印模板表的仓储。
type StorePrintTemplateRepository struct{}

// NewStorePrintTemplateRepository 构造打印模板仓储。
func NewStorePrintTemplateRepository() *StorePrintTemplateRepository {
	return &StorePrintTemplateRepository{}
}

// Create 插入打印模板；同一门店下设备类型与打印类型冲突时转换为业务错误码。
func (r *StorePrintTemplateRepository) Create(ctx context.Context, m *model.StorePrintTemplateModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storePrintTemplateTable).Insert(map[string]any{
		"id":           m.ID,
		"store_id":     m.StoreID,
		"name":         m.Name,
		"printer_type": m.PrinterType,
		"print_type":   m.PrintType,
		"paper_width":  m.PaperWidth,
		"content":      m.Content,
		"version":      m.Version,
		"status":       m.Status,
		"remark":       m.Remark,
		"creator":      m.Creator,
		"created_at":   m.CreatedAt,
		"updater":      m.Updater,
		"updated_at":   m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StorePrintTemplateKindExists, m.PrinterType.String(), m.PrintType.String())
	}

	return err
}

// Update 更新打印模板的可编辑字段与版本号；模板不能跨门店移动，因此不更新 store_id。
func (r *StorePrintTemplateRepository) Update(ctx context.Context, m *model.StorePrintTemplateModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(storePrintTemplateTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":         m.Name,
			"printer_type": m.PrinterType,
			"print_type":   m.PrintType,
			"paper_width":  m.PaperWidth,
			"content":      m.Content,
			"version":      m.Version,
			"status":       m.Status,
			"remark":       m.Remark,
			"updater":      m.Updater,
			"updated_at":   m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(storeErrs.StorePrintTemplateKindExists, m.PrinterType.String(), m.PrintType.String())
	}

	return err
}

// Delete 物理删除打印模板；打印任务生成时已保存模板版本与渲染快照，不依赖本表历史。
func (r *StorePrintTemplateRepository) Delete(ctx context.Context, m *model.StorePrintTemplateModel) error {
	res, err := facades.DB().WithContext(ctx).Table(storePrintTemplateTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(storeErrs.StorePrintTemplateNotExists)
	}

	return nil
}

// FindByID 按主键查询未删除的打印模板；未命中返回 (nil, nil)。
func (r *StorePrintTemplateRepository) FindByID(ctx context.Context, id types.ID) (*model.StorePrintTemplateModel, error) {
	var m model.StorePrintTemplateModel

	if err := facades.DB().WithContext(ctx).Table(storePrintTemplateTable).
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

// ExistsByKind 判断同一门店下「设备类型 + 打印类型」是否已有模板；excludeID > 0 时排除自身。
func (r *StorePrintTemplateRepository) ExistsByKind(ctx context.Context, storeID types.ID, printerType enums.PrinterTypeEnum, printType enums.PrintTypeEnum, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(storePrintTemplateTable).
		Where("store_id", storeID).
		Where("printer_type", printerType).
		Where("print_type", printType).
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

// Paginate 按筛选条件分页查询打印模板，按门店、设备类型与打印类型排序。
func (r *StorePrintTemplateRepository) Paginate(ctx context.Context, filter StorePrintTemplateFilter, pageNo, pageSize int) ([]model.StorePrintTemplateModel, int64, error) {
	list := make([]model.StorePrintTemplateModel, 0)
	var total int64

	query := r.applyFilter(facades.DB().WithContext(ctx).Table(storePrintTemplateTable).WhereNull("deleted_at"), filter)
	if err := query.OrderBy("store_id").OrderBy("printer_type").OrderBy("print_type").OrderByDesc("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// applyFilter 把门店范围与打印模板筛选条件应用到查询构造器。
// 入参：query 基础查询、filter 筛选条件；返回追加条件后的查询。
func (r *StorePrintTemplateRepository) applyFilter(query db.Query, filter StorePrintTemplateFilter) db.Query {
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
	if filter.PrinterType != nil {
		query = query.Where("printer_type", *filter.PrinterType)
	}
	if filter.PrintType != nil {
		query = query.Where("print_type", *filter.PrintType)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	return query
}
