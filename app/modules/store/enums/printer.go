package enums

// PrinterTypeEnum 是门店打印设备的用途类型。
// 取值必须与数据字典 cater_printer_type 保持一致，改动时两边同步（design/12）。
type PrinterTypeEnum int8

const (
	PrinterTypeCashier PrinterTypeEnum = 1 // 前台小票
	PrinterTypeKitchen PrinterTypeEnum = 2 // 后厨
	PrinterTypeLabel   PrinterTypeEnum = 3 // 标签
	PrinterTypeCall    PrinterTypeEnum = 4 // 叫号
)

// String 返回打印设备类型的中文描述，便于列表展示与日志定位。
func (t PrinterTypeEnum) String() string {
	switch t {
	case PrinterTypeCashier:
		return "前台小票"
	case PrinterTypeKitchen:
		return "后厨"
	case PrinterTypeLabel:
		return "标签"
	case PrinterTypeCall:
		return "叫号"
	default:
		return "未知类型"
	}
}

// IsValid 判断打印设备类型取值是否合法。
func (t PrinterTypeEnum) IsValid() bool {
	switch t {
	case PrinterTypeCashier, PrinterTypeKitchen, PrinterTypeLabel, PrinterTypeCall:
		return true
	default:
		return false
	}
}

// PrinterAccessModeEnum 是打印机的接入方式。
// 取值必须与数据字典 cater_printer_access 保持一致；一期以云打印机为主，
// 局域网与本地客户端作为兜底（design/09 §2）。
type PrinterAccessModeEnum int8

const (
	PrinterAccessCloud PrinterAccessModeEnum = 1 // 云打印机
	PrinterAccessLAN   PrinterAccessModeEnum = 2 // 局域网打印
	PrinterAccessLocal PrinterAccessModeEnum = 3 // 本地打印客户端
)

// String 返回接入方式的中文描述。
func (m PrinterAccessModeEnum) String() string {
	switch m {
	case PrinterAccessCloud:
		return "云打印机"
	case PrinterAccessLAN:
		return "局域网"
	case PrinterAccessLocal:
		return "本地客户端"
	default:
		return "未知方式"
	}
}

// IsValid 判断接入方式取值是否合法。
func (m PrinterAccessModeEnum) IsValid() bool {
	switch m {
	case PrinterAccessCloud, PrinterAccessLAN, PrinterAccessLocal:
		return true
	default:
		return false
	}
}
