package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
)

// StorePrintTemplateTable 是打印模板表名常量。
const StorePrintTemplateTable = "store_print_template"

// StorePrintTemplateModel 对应 store_print_template 表，表示门店某类机型的小票模板。
// 同一门店下「设备类型 + 打印类型」只保留一个当前模板，内容变化时版本号自动递增；
// 历史打印任务在生成时保存模板版本与渲染快照，不在本表保留旧版本正文。
type StorePrintTemplateModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID，只能指向 type=3 的门店节点
	StoreID types.ID `db:"store_id" json:"storeId"`
	// Name 模板名称，便于运营区分
	Name string `db:"name" json:"name"`
	// PrinterType 适用设备类型：1 前台小票 / 2 后厨 / 3 标签 / 4 叫号
	PrinterType enums.PrinterTypeEnum `db:"printer_type" json:"printerType"`
	// PrintType 业务打印类型：1 结账单 / 2 预结单 / 3 厨打 / 4 退菜 /
	// 5 标签 / 6 退款单 / 7 交班 / 8 日结
	PrintType enums.PrintTypeEnum `db:"print_type" json:"printType"`
	// PaperWidth 纸宽，单位毫米，常用 58 / 76 / 80
	PaperWidth int `db:"paper_width" json:"paperWidth"`
	// Content 模板正文；由打印渲染器解释，内容变化时版本号递增
	Content string `db:"content" json:"content"`
	// Version 当前模板版本号，新建为 1，正文每次变化加 1
	Version int `db:"version" json:"version"`
	// Status 启停状态：0 禁用 / 1 启用
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *StorePrintTemplateModel) TableName() string {
	return StorePrintTemplateTable
}
