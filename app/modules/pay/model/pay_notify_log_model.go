package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayNotifyLogTable 是回调通知日志表名常量，对应 yudao pay_notify_log。
const PayNotifyLogTable = "pay_notify_log"

// PayNotifyLogModel 对应 yudao pay_notify_log 表，记录每次回调通知的请求与结果。
// 与 pay_notify_task 一对多，排查商户回调失败时看这张表。
type PayNotifyLogModel struct {
	baseModel.BaseModel
	// TaskID 通知任务编号
	TaskID types.ID `db:"task_id" json:"taskId"`
	// NotifyTimes 第几次通知
	NotifyTimes int8 `db:"notify_times" json:"notifyTimes"`
	// Response 通知响应内容
	Response string `db:"response" json:"response"`
	// Status 通知状态：0 等待 / 10 成功 / 20 失败
	Status enums.PayNotifyStatusEnum `db:"status" json:"status"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayNotifyLogModel) TableName() string {
	return PayNotifyLogTable
}
