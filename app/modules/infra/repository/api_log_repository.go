package repository

import (
	"context"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/infra/model"
)

const (
	apiAccessLogTable = model.InfraApiAccessLogTable
	apiErrorLogTable  = model.InfraApiErrorLogTable
)

// ApiLogFilter 是 API 日志列表的筛选条件，零值表示不限制。
type ApiLogFilter struct {
	RequestURL string
	ResultCode *int
	Keyword    string
}

// ApiLogRepository 负责两张 API 日志表的写入与查询。
type ApiLogRepository struct{}

func NewApiLogRepository() *ApiLogRepository {
	return &ApiLogRepository{}
}

// CreateAccess 追加访问日志。
func (r *ApiLogRepository) CreateAccess(ctx context.Context, m *model.InfraApiAccessLogModel) error {
	if m.ID == 0 {
		m.ID = support.NewID()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = types.Now()
	}

	_, err := facades.DB().WithContext(ctx).Table(apiAccessLogTable).Insert(map[string]any{
		"id":               m.ID,
		"trace_id":         m.TraceID,
		"user_id":          m.UserID,
		"user_type":        m.UserType,
		"application_name": m.ApplicationName,
		"request_method":   m.RequestMethod,
		"request_url":      m.RequestURL,
		"request_params":   m.RequestParams,
		"response_body":    m.ResponseBody,
		"user_ip":          m.UserIP,
		"user_agent":       m.UserAgent,
		"operate_module":   m.OperateModule,
		"operate_name":     m.OperateName,
		"begin_at":         m.BeginAt,
		"end_at":           m.EndAt,
		"duration":         m.Duration,
		"result_code":      m.ResultCode,
		"created_at":       m.CreatedAt,
	})

	return err
}

// CreateError 追加错误日志。
func (r *ApiLogRepository) CreateError(ctx context.Context, m *model.InfraApiErrorLogModel) error {
	if m.ID == 0 {
		m.ID = support.NewID()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = types.Now()
	}

	_, err := facades.DB().WithContext(ctx).Table(apiErrorLogTable).Insert(map[string]any{
		"id":               m.ID,
		"trace_id":         m.TraceID,
		"user_id":          m.UserID,
		"user_type":        m.UserType,
		"application_name": m.ApplicationName,
		"request_method":   m.RequestMethod,
		"request_url":      m.RequestURL,
		"request_params":   m.RequestParams,
		"user_ip":          m.UserIP,
		"user_agent":       m.UserAgent,
		"name":             m.Name,
		"message":          m.Message,
		"exception_stack":  m.ExceptionStack,
		"process_status":   m.ProcessStatus,
		"process_time":     m.ProcessTime,
		"created_at":       m.CreatedAt,
	})

	return err
}

// PaginateAccess 分页查询访问日志。
func (r *ApiLogRepository) PaginateAccess(ctx context.Context, filter ApiLogFilter, pageNo, pageSize int) ([]model.InfraApiAccessLogModel, int64, error) {
	list := make([]model.InfraApiAccessLogModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(apiAccessLogTable)
	if filter.RequestURL != "" {
		query = query.WhereLike("request_url", "%"+filter.RequestURL+"%")
	}
	if filter.ResultCode != nil {
		query = query.Where("result_code", *filter.ResultCode)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// PaginateError 分页查询错误日志。
func (r *ApiLogRepository) PaginateError(ctx context.Context, filter ApiLogFilter, pageNo, pageSize int) ([]model.InfraApiErrorLogModel, int64, error) {
	list := make([]model.InfraApiErrorLogModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(apiErrorLogTable)
	if filter.RequestURL != "" {
		query = query.WhereLike("request_url", "%"+filter.RequestURL+"%")
	}
	if filter.Keyword != "" {
		query = query.WhereLike("message", "%"+filter.Keyword+"%")
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// DeleteAccessBefore 删除创建时间早于 before 的访问日志。
// 参数 before：截止时间，早于它的记录会被物理删除（命中 idx_created_at 索引）。
// 返回值：受影响的行数与数据库错误。
func (r *ApiLogRepository) DeleteAccessBefore(ctx context.Context, before types.DateTime) (int64, error) {
	result, err := facades.DB().WithContext(ctx).Table(apiAccessLogTable).
		Where("created_at < ?", before).
		Delete()
	if err != nil {
		return 0, err
	}
	if result == nil {
		return 0, nil
	}

	return result.RowsAffected, nil
}

// MarkErrorProcessed 把错误日志标记为已处理。
func (r *ApiLogRepository) MarkErrorProcessed(ctx context.Context, id types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(apiErrorLogTable).
		Where("id", id).
		Update(map[string]any{
			"process_status": 1,
			"process_time":   types.Now(),
		})

	return err
}
