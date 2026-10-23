package enums

// PrintTaskStatusEnum 是打印任务的状态。
// 0 待打印 → 1 已打印 / 2 打印失败；3 已取消用于门店撤单或设备停用后的兜底。
type PrintTaskStatusEnum int8

const (
	// PrintTaskStatusPending 待打印：已入队，等待设备取走
	PrintTaskStatusPending PrintTaskStatusEnum = 0
	// PrintTaskStatusPrinted 已打印
	PrintTaskStatusPrinted PrintTaskStatusEnum = 1
	// PrintTaskStatusFailed 打印失败
	PrintTaskStatusFailed PrintTaskStatusEnum = 2
	// PrintTaskStatusCancelled 已取消
	PrintTaskStatusCancelled PrintTaskStatusEnum = 3
)

// String 返回打印任务状态的中文描述。
func (s PrintTaskStatusEnum) String() string {
	switch s {
	case PrintTaskStatusPending:
		return "待打印"
	case PrintTaskStatusPrinted:
		return "已打印"
	case PrintTaskStatusFailed:
		return "打印失败"
	case PrintTaskStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

// IsValid 判断打印任务状态取值是否合法。
func (s PrintTaskStatusEnum) IsValid() bool {
	switch s {
	case PrintTaskStatusPending, PrintTaskStatusPrinted, PrintTaskStatusFailed, PrintTaskStatusCancelled:
		return true
	default:
		return false
	}
}

// PrinterTypeFor 返回业务打印类型对应的设备类型。
// 结账单、预结单、退款单、交班与日结走前台小票机，厨打与退菜走后厨机，标签走标签机。
func PrinterTypeFor(printType PrintTypeEnum) PrinterTypeEnum {
	switch printType {
	case PrintTypeKitchen, PrintTypeReturn:
		return PrinterTypeKitchen
	case PrintTypeLabel:
		return PrinterTypeLabel
	default:
		return PrinterTypeCashier
	}
}
