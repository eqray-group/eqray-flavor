package bootstrap

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/schedule"

	"eqray-flavor/app/facades"
	infraService "eqray-flavor/app/modules/infra/service"
	payService "eqray-flavor/app/modules/pay/service"
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
		// 支付回调通知：每分钟扫描到点的任务，签名后推送给业务方，失败按间隔重试。
		facades.Schedule().Call(dispatchPayNotifyTasks).
			EveryMinute().
			Name("pay:dispatch-notify-task").
			OnOneServer(),
		// 超时关单：每 5 分钟关闭已过期的待支付单，关单前会先查渠道确认未收款。
		facades.Schedule().Call(closeExpiredPayOrders).
			EveryFiveMinutes().
			Name("pay:close-expired-order").
			OnOneServer(),
	}
}

// closeExpiredPayOrders 关闭一批已过期的待支付单，并把结果写入应用日志。
// 无入参与返回值：单笔关单失败已记录在任务错误里，这里只汇总调度层结果。
func closeExpiredPayOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	closed, err := payService.NewPayOrderService().CloseExpired(ctx, 100)
	if err != nil {
		facades.Log().Errorf("关闭超时支付单失败（本次关闭 %d 单）：%v", closed, err)
		return
	}

	if closed > 0 {
		facades.Log().Infof("关闭超时支付单完成，本次关闭 %d 单", closed)
	}
}

// dispatchPayNotifyTasks 派发一批支付 / 退款结果通知给业务方，并把结果写入应用日志。
// 无入参与返回值：单条通知失败已落在任务与日志表里，这里只记录调度层面的异常。
func dispatchPayNotifyTasks() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	handled, err := payService.NewPayNotifyTaskService().ExecuteDueTasks(ctx)
	if err != nil {
		facades.Log().Errorf("派发支付通知任务失败（本次处理 %d 条）：%v", handled, err)
		return
	}

	if handled > 0 {
		facades.Log().Infof("派发支付通知任务完成，本次处理 %d 条", handled)
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
