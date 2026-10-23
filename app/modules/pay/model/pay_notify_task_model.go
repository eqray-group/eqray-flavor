package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/enums"
)

// PayNotifyTaskTable 是回调通知任务表名常量，对应 yudao pay_notify_task。
const PayNotifyTaskTable = "pay_notify_task"

// PayNotifyTaskModel 对应 yudao pay_notify_task 表，表示一条待通知商户的回调任务。
// 通知失败按 next_notify_time 重试，直到达到 max_notify_times。
type PayNotifyTaskModel struct {
	baseModel.BaseModel
	// AppID 应用编号
	AppID types.ID `db:"app_id" json:"appId"`
	// Type 通知类型：1 支付单 / 2 退款单
	Type enums.PayNotifyTypeEnum `db:"type" json:"type"`
	// DataID 业务数据编号（支付单 / 退款单 ID）
	DataID types.ID `db:"data_id" json:"dataId"`
	// MerchantOrderID 商户订单号
	MerchantOrderID string `db:"merchant_order_id" json:"merchantOrderId"`
	// NotifyURL 异步通知地址
	NotifyURL string `db:"notify_url" json:"notifyUrl"`
	// Status 通知状态：1 等待 / 2 成功 / 3 失败 / 4 请求成功但结果失败 / 5 请求失败
	Status enums.PayNotifyStatusEnum `db:"status" json:"status"`
	// NextNotifyTime 下一次通知时间
	NextNotifyTime types.DateTime `db:"next_notify_time" json:"nextNotifyTime"`
	// LastExecuteTime 最后一次执行时间
	LastExecuteTime types.DateTime `db:"last_execute_time" json:"lastExecuteTime"`
	// NotifyTimes 当前已通知次数
	NotifyTimes int8 `db:"notify_times" json:"notifyTimes"`
	// MaxNotifyTimes 最大通知次数
	MaxNotifyTimes int8 `db:"max_notify_times" json:"maxNotifyTimes"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PayNotifyTaskModel) TableName() string {
	return PayNotifyTaskTable
}
