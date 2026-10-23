package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/model"
)

const (
	// payNotifyTaskTable 是回调通知任务表名，对应 yudao pay_notify_task
	payNotifyTaskTable = model.PayNotifyTaskTable
	// payNotifyLogTable 是回调通知日志表名，对应 yudao pay_notify_log
	payNotifyLogTable = model.PayNotifyLogTable
)

// PayNotifyRepository 负责回调通知任务与通知日志的读写。
// 任务按 next_notify_time 重试，每次执行写一条日志，逻辑对齐 yudao PayNotifyService。
type PayNotifyRepository struct{}

// NewPayNotifyRepository 构造回调通知仓储。
func NewPayNotifyRepository() *PayNotifyRepository {
	return &PayNotifyRepository{}
}

// PayNotifyTaskFilter 是通知任务列表的筛选条件。
type PayNotifyTaskFilter struct {
	// AppID 应用编号，0 表示不限
	AppID types.ID
	// Type 通知类型（1 支付 / 2 退款 / 3 转账），nil 表示不限
	Type *int8
	// Status 通知状态，nil 表示不限
	Status *int8
	// Keyword 商户订单号关键字
	Keyword string
}

// PaginateTasks 按条件分页查询通知任务，按主键倒序。
// 入参：筛选条件与分页；返回任务列表与总数。
func (r *PayNotifyRepository) PaginateTasks(ctx context.Context, filter PayNotifyTaskFilter, pageNo, pageSize int) ([]model.PayNotifyTaskModel, int64, error) {
	list := make([]model.PayNotifyTaskModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payNotifyTaskTable).WhereNull("deleted_at")
	if filter.AppID > 0 {
		query = query.Where("app_id", filter.AppID)
	}
	if filter.Type != nil {
		query = query.Where("type", *filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}
	if filter.Keyword != "" {
		query = query.WhereLike("merchant_order_id", "%"+filter.Keyword+"%")
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindTaskByID 按主键查询通知任务；未命中返回 (nil, nil)。
func (r *PayNotifyRepository) FindTaskByID(ctx context.Context, id types.ID) (*model.PayNotifyTaskModel, error) {
	var m model.PayNotifyTaskModel
	if err := facades.DB().WithContext(ctx).Table(payNotifyTaskTable).
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

// ListDueTasks 查询到点待通知的任务，供通知重试任务使用。
// 入参：当前时间与单次最大条数；返回可重试状态（等待 / 请求成功但结果失败 / 请求失败）
// 且 next_notify_time 已到的任务。最终失败(3) 与成功(2) 都是终态，不会被捞出。
func (r *PayNotifyRepository) ListDueTasks(ctx context.Context, now types.DateTime, limit uint64) ([]model.PayNotifyTaskModel, error) {
	list := make([]model.PayNotifyTaskModel, 0)
	if err := facades.DB().WithContext(ctx).Table(payNotifyTaskTable).
		WhereIn("status", []any{
			int8(payEnums.PayNotifyStatusWaiting),
			int8(payEnums.PayNotifyStatusRequestSuccess),
			int8(payEnums.PayNotifyStatusRequestFail),
		}).
		Where("next_notify_time <= ?", now).
		// 达到最大次数的任务不再重试，避免每轮调度都把它捞出来
		WhereRaw("notify_times < max_notify_times", []any{}).
		WhereNull("deleted_at").
		OrderBy("id").
		Limit(limit).
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// CreateTask 在事务内插入通知任务。
// 入参：事务与任务模型；返回错误表示落库失败。
func (r *PayNotifyRepository) CreateTask(ctx context.Context, tx db.Tx, m *model.PayNotifyTaskModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payNotifyTaskTable).Insert(map[string]any{
		"id":                m.ID,
		"app_id":            m.AppID,
		"type":              int8(m.Type),
		"data_id":           m.DataID,
		"merchant_order_id": m.MerchantOrderID,
		"notify_url":        m.NotifyURL,
		"status":            int8(m.Status),
		"next_notify_time":  m.NextNotifyTime,
		"last_execute_time": m.LastExecuteTime,
		"notify_times":      m.NotifyTimes,
		"max_notify_times":  m.MaxNotifyTimes,
		"creator":           m.Creator,
		"created_at":        m.CreatedAt,
		"updater":           m.Updater,
		"updated_at":        m.UpdatedAt,
	})

	return err
}

// UpdateTaskResult 回写通知结果与下次通知时间。
// 入参：事务与任务模型；返回错误表示更新失败。
func (r *PayNotifyRepository) UpdateTaskResult(ctx context.Context, tx db.Tx, m *model.PayNotifyTaskModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(payNotifyTaskTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":            int8(m.Status),
			"notify_times":      m.NotifyTimes,
			"next_notify_time":  m.NextNotifyTime,
			"last_execute_time": m.LastExecuteTime,
			"updater":           m.Updater,
			"updated_at":        m.UpdatedAt,
		})

	return err
}

// PaginateLogs 分页查询某个任务的通知日志，按主键倒序。
// 入参：任务 ID 与分页；返回日志列表与总数。
func (r *PayNotifyRepository) PaginateLogs(ctx context.Context, taskID types.ID, pageNo, pageSize int) ([]model.PayNotifyLogModel, int64, error) {
	list := make([]model.PayNotifyLogModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(payNotifyLogTable).
		Where("task_id", taskID).
		WhereNull("deleted_at").
		OrderByDesc("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CreateLog 在事务内写入一条通知日志。
// 入参：事务与日志模型；返回错误表示落库失败。
func (r *PayNotifyRepository) CreateLog(ctx context.Context, tx db.Tx, m *model.PayNotifyLogModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payNotifyLogTable).Insert(map[string]any{
		"id":           m.ID,
		"task_id":      m.TaskID,
		"notify_times": m.NotifyTimes,
		"response":     m.Response,
		"status":       int8(m.Status),
		"creator":      m.Creator,
		"created_at":   m.CreatedAt,
		"updater":      m.Updater,
		"updated_at":   m.UpdatedAt,
	})

	return err
}
