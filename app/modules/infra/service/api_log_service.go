package service

import (
	"context"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/infra/repository"
)

// ApiLogService 负责 API 日志的后台维护动作（目前只有按保留期清理），
// 列表查询这类只读逻辑仍由控制器直接走仓储。
type ApiLogService struct {
	repo *repository.ApiLogRepository
}

// NewApiLogService 构造 API 日志服务，内部复用共享的日志仓储。
func NewApiLogService() *ApiLogService {
	return &ApiLogService{repo: repository.NewApiLogRepository()}
}

// PurgeExpiredAccess 按配置项 infra.api_log.access_retention_days（默认 7 天）
// 清理过期的 API 访问日志，供定时任务调用。
// 参数 ctx：控制单次清理的超时与取消；返回值：被删除的行数与错误。
func (s *ApiLogService) PurgeExpiredAccess(ctx context.Context) (int64, error) {
	days := facades.Config().GetInt("infra.api_log.access_retention_days", 7)

	return s.PurgeAccessBefore(ctx, days)
}

// PurgeAccessBefore 删除「当前时间往前 days 天」之前的访问日志。
// 参数 ctx：执行上下文；days：保留天数，<= 0 表示永久保留，直接返回 0。
// 返回值：被删除的行数与数据库错误。
func (s *ApiLogService) PurgeAccessBefore(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}

	cutoff := types.DateTime(time.Now().AddDate(0, 0, -days))

	return s.repo.DeleteAccessBefore(ctx, cutoff)
}
