package model

import (
	"eqray-flavor/app/modules/base/types"
)

// InfraApiErrorLogTable 是表名常量，Model 与 Repository 共用同一来源。
const InfraApiErrorLogTable = "infra_api_error_log"

// InfraApiErrorLogModel 对应 infra_api_error_log 表（docs 6.4）。
type InfraApiErrorLogModel struct {
	ID              types.ID       `db:"id" json:"id"`
	TraceID         string         `db:"trace_id" json:"traceId"`
	UserID          int64          `db:"user_id" json:"userId"`
	UserType        int8           `db:"user_type" json:"userType"`
	ApplicationName string         `db:"application_name" json:"applicationName"`
	RequestMethod   string         `db:"request_method" json:"requestMethod"`
	RequestURL      string         `db:"request_url" json:"requestUrl"`
	RequestParams   string         `db:"request_params" json:"requestParams"`
	UserIP          string         `db:"user_ip" json:"userIp"`
	UserAgent       string         `db:"user_agent" json:"userAgent"`
	Name            string         `db:"name" json:"name"`
	Message         string         `db:"message" json:"message"`
	ExceptionStack  string         `db:"exception_stack" json:"exceptionStack"`
	ProcessStatus   int8           `db:"process_status" json:"processStatus"`
	ProcessTime     types.DateTime `db:"process_time" json:"processTime"`
	CreatedAt       types.DateTime `db:"created_at" json:"createdAt"`
}

// TableName 返回表名。
func (m *InfraApiErrorLogModel) TableName() string {
	return InfraApiErrorLogTable
}
