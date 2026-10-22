package enums

// PrintTypeEnum 是打印模板对应的业务打印类型。
// 取值必须与数据字典 cater_print_type 保持一致，改动时两边同步（design/12）。
type PrintTypeEnum int8

const (
	PrintTypeCheckout  PrintTypeEnum = 1 // 结账单
	PrintTypePreBill   PrintTypeEnum = 2 // 预结单
	PrintTypeKitchen   PrintTypeEnum = 3 // 厨打单
	PrintTypeReturn    PrintTypeEnum = 4 // 退菜单
	PrintTypeLabel     PrintTypeEnum = 5 // 标签
	PrintTypeRefund    PrintTypeEnum = 6 // 退款单
	PrintTypeShift     PrintTypeEnum = 7 // 交班单
	PrintTypeDailyBill PrintTypeEnum = 8 // 日结单
)

// String 返回打印类型的中文描述。
func (t PrintTypeEnum) String() string {
	switch t {
	case PrintTypeCheckout:
		return "结账单"
	case PrintTypePreBill:
		return "预结单"
	case PrintTypeKitchen:
		return "厨打"
	case PrintTypeReturn:
		return "退菜"
	case PrintTypeLabel:
		return "标签"
	case PrintTypeRefund:
		return "退款单"
	case PrintTypeShift:
		return "交班"
	case PrintTypeDailyBill:
		return "日结"
	default:
		return "未知类型"
	}
}

// IsValid 判断打印类型取值是否合法。
func (t PrintTypeEnum) IsValid() bool {
	switch t {
	case PrintTypeCheckout, PrintTypePreBill, PrintTypeKitchen, PrintTypeReturn,
		PrintTypeLabel, PrintTypeRefund, PrintTypeShift, PrintTypeDailyBill:
		return true
	default:
		return false
	}
}
