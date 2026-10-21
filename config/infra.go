package config

import (
	"eqray-flavor/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("infra", map[string]any{
		"api_log": map[string]any{
			// API 访问日志保留天数：bootstrap/schedule.go 的定时任务每天按此值清理
			// infra_api_access_log 中更早的记录；<= 0 表示永久保留、不清理。
			"access_retention_days": config.Env("INFRA_API_LOG_RETENTION_DAYS", 7),
		},
	})
}
