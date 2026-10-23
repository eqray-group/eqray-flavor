package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"
	frameworkDB "github.com/goravel/framework/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/model"
)

// printTaskStoreIDs 把门店 ID 列表转成 WhereIn 需要的 []any。
func printTaskStoreIDs(ids []types.ID) []any {
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}

	return values
}

// storePrintTaskTable 是打印任务表名。
const storePrintTaskTable = model.StorePrintTaskTable

// StorePrintTaskFilter 是打印任务列表的筛选条件，零值表示不限制。
type StorePrintTaskFilter struct {
	// StoreID 请求上下文指定的门店，优先级高于 StoreIDs
	StoreID types.ID
	// StoreIDs 可见门店集合；RestrictStores=true 时空集合表示无可见门店
	StoreIDs []types.ID
	// RestrictStores 是否按可见门店集合限制查询
	RestrictStores bool
	// PrinterID 设备编号
	PrinterID types.ID
	// PrintType 打印类型
	PrintType *enums.PrintTypeEnum
	// Status 任务状态；0 是合法筛选值，不能用零值判断未传
	Status *enums.PrintTaskStatusEnum
	// OrderNo 订单号关键字
	OrderNo string
}

// StorePrintTaskRepository 是 store_print_task 打印任务表的仓储。
type StorePrintTaskRepository struct{}

// NewStorePrintTaskRepository 构造打印任务仓储。
func NewStorePrintTaskRepository() *StorePrintTaskRepository {
	return &StorePrintTaskRepository{}
}

// CreateBatch 在一个事务内批量写入打印任务，空列表直接返回。
// 入参：事务与任务列表；返回错误表示落库失败，调用方按整体失败处理。
func (r *StorePrintTaskRepository) CreateBatch(ctx context.Context, tx db.Tx, tasks []model.StorePrintTaskModel) error {
	if len(tasks) == 0 {
		return nil
	}

	rows := make([]map[string]any, 0, len(tasks))
	for i := range tasks {
		m := &tasks[i]
		m.FillCreate(ctx)
		rows = append(rows, map[string]any{
			"id":               m.ID,
			"store_id":         m.StoreID,
			"printer_id":       m.PrinterID,
			"printer_name":     m.PrinterName,
			"template_id":      m.TemplateID,
			"template_version": m.TemplateVersion,
			"print_type":       int8(m.PrintType),
			"order_id":         m.OrderID,
			"order_no":         m.OrderNo,
			"content":          m.Content,
			"status":           int8(m.Status),
			"copies":           m.Copies,
			"retry_count":      m.RetryCount,
			"error_msg":        m.ErrorMsg,
			"printed_at":       m.PrintedAt,
			"creator":          m.Creator,
			"created_at":       m.CreatedAt,
			"updater":          m.Updater,
			"updated_at":       m.UpdatedAt,
		})
	}

	_, err := tx.Table(storePrintTaskTable).Insert(rows)

	return err
}

// Paginate 按条件分页查询打印任务，按主键倒序。
// 入参：筛选条件与分页；返回任务列表与总数。
func (r *StorePrintTaskRepository) Paginate(ctx context.Context, filter StorePrintTaskFilter, pageNo, pageSize int) ([]model.StorePrintTaskModel, int64, error) {
	list := make([]model.StorePrintTaskModel, 0)
	var total int64

	query := r.applyFilter(facades.DB().WithContext(ctx).Table(storePrintTaskTable).WhereNull("deleted_at"), filter)
	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByID 按主键查询未删除的打印任务；未命中返回 (nil, nil)。
func (r *StorePrintTaskRepository) FindByID(ctx context.Context, id types.ID) (*model.StorePrintTaskModel, error) {
	var m model.StorePrintTaskModel
	if err := facades.DB().WithContext(ctx).
		Table(storePrintTaskTable).
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

// ResetForReprint 把已打印 / 失败 / 已取消的任务重置为待打印，重打次数加一。
// 入参：任务 ID；返回影响行数，0 表示任务已处于待打印状态。
// 说明：条件更新保证并发下同一张单只会被重置一次，避免重复出纸。
func (r *StorePrintTaskRepository) ResetForReprint(ctx context.Context, id types.ID) (int64, error) {
	res, err := facades.DB().WithContext(ctx).
		Table(storePrintTaskTable).
		Where("id", id).
		WhereIn("status", []any{
			int8(enums.PrintTaskStatusPrinted),
			int8(enums.PrintTaskStatusFailed),
			int8(enums.PrintTaskStatusCancelled),
		}).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":      int8(enums.PrintTaskStatusPending),
			"retry_count": frameworkDB.Raw("retry_count + 1"),
			"error_msg":   "",
			"printed_at":  nil,
			"updated_at":  types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// CancelByOrder 取消某订单下仍待打印的任务，用于订单取消后止住还没出纸的单。
// 入参：订单 ID；返回影响行数。
func (r *StorePrintTaskRepository) CancelByOrder(ctx context.Context, orderID types.ID) (int64, error) {
	res, err := facades.DB().WithContext(ctx).
		Table(storePrintTaskTable).
		Where("order_id", orderID).
		Where("status", int8(enums.PrintTaskStatusPending)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":     int8(enums.PrintTaskStatusCancelled),
			"updated_at": types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// applyFilter 拼接筛选条件，供列表与后续统计复用。
func (r *StorePrintTaskRepository) applyFilter(query db.Query, filter StorePrintTaskFilter) db.Query {
	if filter.StoreID > 0 {
		query = query.Where("store_id", filter.StoreID)
	} else if filter.RestrictStores {
		ids := printTaskStoreIDs(filter.StoreIDs)
		if len(ids) == 0 {
			return query.Where("id", 0)
		}
		query = query.WhereIn("store_id", ids)
	}
	if filter.PrinterID > 0 {
		query = query.Where("printer_id", filter.PrinterID)
	}
	if filter.PrintType != nil {
		query = query.Where("print_type", int8(*filter.PrintType))
	}
	if filter.Status != nil {
		query = query.Where("status", int8(*filter.Status))
	}
	if filter.OrderNo != "" {
		query = query.WhereLike("order_no", "%"+filter.OrderNo+"%")
	}

	return query
}
