// Package enums 存放跨模块共用的枚举（docs 11.1）。
package enums

// CommonStatusEnum 是通用的启停状态枚举，用户、角色、菜单等表的 status 字段共用。
// 底层类型与各表的 tinyint 对齐。
type CommonStatusEnum int8

const (
	CommonStatusDisabled CommonStatusEnum = 0 // 禁用
	CommonStatusEnabled  CommonStatusEnum = 1 // 启用
)

// String 返回状态的中文描述，便于日志与接口输出。
func (s CommonStatusEnum) String() string {
	switch s {
	case CommonStatusDisabled:
		return "禁用"
	case CommonStatusEnabled:
		return "启用"
	default:
		return "未知"
	}
}

// IsValid 判断状态取值是否合法。
// 注意：0 是合法值，不能用零值判断「未传」，未传由请求校验的 required 保证（docs 11.1）。
func (s CommonStatusEnum) IsValid() bool {
	return s == CommonStatusDisabled || s == CommonStatusEnabled
}
