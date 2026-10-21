package enums

// MenuTypeEnum 是 system_menu 的节点类型枚举，底层类型与表的 tinyint 对齐。
type MenuTypeEnum int8

const (
	MenuTypeDir    MenuTypeEnum = iota + 1 // 目录
	MenuTypeMenu                           // 菜单
	MenuTypeButton                         // 按钮（权限点）
)

// String 返回类型的中文描述，便于日志与接口输出。
func (t MenuTypeEnum) String() string {
	switch t {
	case MenuTypeDir:
		return "目录"
	case MenuTypeMenu:
		return "菜单"
	case MenuTypeButton:
		return "按钮"
	default:
		return "未知"
	}
}

// IsValid 判断类型取值是否合法。
func (t MenuTypeEnum) IsValid() bool {
	return t == MenuTypeDir || t == MenuTypeMenu || t == MenuTypeButton
}
