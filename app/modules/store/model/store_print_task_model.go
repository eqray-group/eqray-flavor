package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
)

// StorePrintTaskTable 是门店打印任务表名常量。
const StorePrintTaskTable = "store_print_task"

// StorePrintTaskModel 对应 store_print_task 表，一条记录表示「某台设备要打一张单」。
// 正文在入队时就渲染好并存快照，模板后续被改动也不会影响历史单据的重打结果。
type StorePrintTaskModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID
	StoreID types.ID `db:"store_id" json:"storeId"`
	// PrinterID 打印设备 ID
	PrinterID types.ID `db:"printer_id" json:"printerId"`
	// PrinterName 设备名称快照，设备改名或删除后仍可追溯
	PrinterName string `db:"printer_name" json:"printerName"`
	// TemplateID 打印模板 ID
	TemplateID types.ID `db:"template_id" json:"templateId"`
	// TemplateVersion 生成时的模板版本号
	TemplateVersion int `db:"template_version" json:"templateVersion"`
	// PrintType 打印类型，取值见 enums.PrintTypeEnum
	PrintType enums.PrintTypeEnum `db:"print_type" json:"printType"`
	// OrderID 订单 ID，非订单类打印为 0
	OrderID types.ID `db:"order_id" json:"orderId"`
	// OrderNo 订单号快照
	OrderNo string `db:"order_no" json:"orderNo"`
	// Content 渲染后的打印正文
	Content string `db:"content" json:"content"`
	// Status 任务状态
	Status enums.PrintTaskStatusEnum `db:"status" json:"status"`
	// Copies 打印份数
	Copies int `db:"copies" json:"copies"`
	// RetryCount 重打次数
	RetryCount int `db:"retry_count" json:"retryCount"`
	// ErrorMsg 失败原因
	ErrorMsg string `db:"error_msg" json:"errorMsg"`
	// PrintedAt 打印成功时间
	PrintedAt *types.DateTime `db:"printed_at" json:"printedAt"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *StorePrintTaskModel) TableName() string {
	return StorePrintTaskTable
}
