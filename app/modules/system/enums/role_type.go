package enums

// RoleTypeEnum 是 system_role 的角色类型枚举。
type RoleTypeEnum int8

const (
	RoleTypeBuiltin RoleTypeEnum = iota + 1 // 内置角色
	RoleTypeCustom                          // 自定义角色
)

// String 返回类型的中文描述。
func (t RoleTypeEnum) String() string {
	switch t {
	case RoleTypeBuiltin:
		return "内置角色"
	case RoleTypeCustom:
		return "自定义角色"
	default:
		return "未知"
	}
}

// IsValid 判断类型取值是否合法。
func (t RoleTypeEnum) IsValid() bool {
	return t == RoleTypeBuiltin || t == RoleTypeCustom
}
