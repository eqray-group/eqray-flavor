package model

import (
	"eqray-flavor/app/modules/base/types"
)

// InfraApiAccessLogTable 是表名常量，Model 与 Repository 共用同一来源。
const InfraApiAccessLogTable = "infra_api_access_log"

// InfraApiAccessLogModel 对应 infra_api_access_log 表（docs 6.4）。
type InfraApiAccessLogModel struct {
	ID              types.ID       `db:"id" json:"id"`
	TraceID         string         `db:"trace_id" json:"traceId"`
	UserID          int64          `db:"user_id" json:"userId"`
	UserType        int8           `db:"user_type" json:"userType"`
	ApplicationName string         `db:"application_name" json:"applicationName"`
	RequestMethod   string         `db:"request_method" json:"requestMethod"`
	RequestURL      string         `db:"request_url" json:"requestUrl"`
	RequestParams   string         `db:"request_params" json:"requestParams"`
	ResponseBody    string         `db:"response_body" json:"responseBody"`
	UserIP          string         `db:"user_ip" json:"userIp"`
	UserAgent       string         `db:"user_agent" json:"userAgent"`
	OperateModule   string         `db:"operate_module" json:"operateModule"`
	OperateName     string         `db:"operate_name" json:"operateName"`
	BeginAt         types.DateTime `db:"begin_at" json:"beginAt"`
	EndAt           types.DateTime `db:"end_at" json:"endAt"`
	Duration        int            `db:"duration" json:"duration"`
	ResultCode      int            `db:"result_code" json:"resultCode"`
	CreatedAt       types.DateTime `db:"created_at" json:"createdAt"`
}

// TableName 返回表名。
func (m *InfraApiAccessLogModel) TableName() string {
	return InfraApiAccessLogTable
}
