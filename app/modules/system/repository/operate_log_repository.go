package repository

import (
	"context"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

const operateLogTable = model.SystemOperateLogTable

// OperateLogFilter 是操作日志列表的筛选条件，零值表示不限制。
type OperateLogFilter struct {
	Type    string
	SubType string
	Success *int8
	Keyword string
}

// OperateLogRepository 是 system_operate_log 表的仓储。
type OperateLogRepository struct{}

func NewOperateLogRepository() *OperateLogRepository {
	return &OperateLogRepository{}
}

// Create 追加一条操作日志；日志表只有 id 与 created_at，没有审计更新字段。
func (r *OperateLogRepository) Create(ctx context.Context, m *model.SystemOperateLogModel) error {
	if m.ID == 0 {
		m.ID = support.NewID()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = types.Now()
	}

	_, err := facades.DB().WithContext(ctx).Table(operateLogTable).Insert(map[string]any{
		"id":             m.ID,
		"trace_id":       m.TraceID,
		"user_id":        m.UserID,
		"user_type":      m.UserType,
		"type":           m.Type,
		"sub_type":       m.SubType,
		"biz_id":         m.BizID,
		"action":         m.Action,
		"success":        m.Success,
		"extra":          m.Extra,
		"request_method": m.RequestMethod,
		"request_url":    m.RequestURL,
		"request_ip":     m.RequestIP,
		"user_agent":     m.UserAgent,
		"duration":       m.Duration,
		"created_at":     m.CreatedAt,
	})

	return err
}

// Paginate 按条件分页查询操作日志。
func (r *OperateLogRepository) Paginate(ctx context.Context, filter OperateLogFilter, pageNo, pageSize int) ([]model.SystemOperateLogModel, int64, error) {
	list := make([]model.SystemOperateLogModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(operateLogTable)
	if filter.Type != "" {
		query = query.Where("type", filter.Type)
	}
	if filter.SubType != "" {
		query = query.Where("sub_type", filter.SubType)
	}
	if filter.Success != nil {
		query = query.Where("success", *filter.Success)
	}
	if filter.Keyword != "" {
		query = query.WhereLike("action", "%"+filter.Keyword+"%")
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
