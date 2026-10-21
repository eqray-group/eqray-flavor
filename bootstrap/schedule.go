package bootstrap

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/schedule"

	"eqray-flavor/app/facades"
	infraService "eqray-flavor/app/modules/infra/service"
)

// Schedule 注册应用级定时任务，由 foundation.Setup().WithSchedule 装载，
// 随 app.Start() 的 schedule runner 一起启动。
func Schedule() []schedule.Event {
	return []schedule.Event{
		// 访问日志保留策略：每天 03:10 清理 infra_api_access_log 中超过保留天数的记录，
		// 保留天数见 config/infra.go（默认 7 天，可用 INFRA_API_LOG_RETENTION_DAYS 覆盖）。
		// OnOneServer 保证多实例部署时同一时刻只有一台机器执行。
		facades.Schedule().Call(purgeExpiredApiAccessLogs).
			DailyAt("03:10").
			Name("infra:purge-api-access-log").
			OnOneServer(),
	}
}

// purgeExpiredApiAccessLogs 执行一次访问日志清理，并把结果写入应用日志。
// 无入参与返回值：异常只记录日志，避免把错误抛给调度器打断后续任务。
func purgeExpiredApiAccessLogs() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	deleted, err := infraService.NewApiLogService().PurgeExpiredAccess(ctx)
	if err != nil {
		facades.Log().Errorf("清理 API 访问日志失败：%v", err)
		return
	}

	if deleted > 0 {
		facades.Log().Infof("清理 API 访问日志完成，共删除 %d 条", deleted)
	}
}
