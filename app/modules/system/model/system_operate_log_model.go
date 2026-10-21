package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
)

// SystemOperateLogTable 是表名常量，Model 与 Repository 共用同一来源。
const SystemOperateLogTable = "system_operate_log"

// SystemOperateLogModel 对应 system_operate_log 表（docs 6.3）。
// 日志只追加不删除，因此不嵌入 BaseModel（无 updater / deleted_at）。
type SystemOperateLogModel struct {
	ID            types.ID               `db:"id" json:"id"`
	TraceID       string                 `db:"trace_id" json:"traceId"`
	UserID        int64                  `db:"user_id" json:"userId"`
	UserType      baseEnums.UserTypeEnum `db:"user_type" json:"userType"`
	Type          string                 `db:"type" json:"type"`
	SubType       string                 `db:"sub_type" json:"subType"`
	BizID         string                 `db:"biz_id" json:"bizId"`
	Action        string                 `db:"action" json:"action"`
	Success       int8                   `db:"success" json:"success"`
	Extra         string                 `db:"extra" json:"extra"`
	RequestMethod string                 `db:"request_method" json:"requestMethod"`
	RequestURL    string                 `db:"request_url" json:"requestUrl"`
	RequestIP     string                 `db:"request_ip" json:"requestIp"`
	UserAgent     string                 `db:"user_agent" json:"userAgent"`
	Duration      int64                  `db:"duration" json:"duration"`
	CreatedAt     types.DateTime         `db:"created_at" json:"createdAt"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *SystemOperateLogModel) TableName() string {
	return SystemOperateLogTable
}
