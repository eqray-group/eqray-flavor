package enums

import "strconv"

// UserTypeEnum 是登录日志里的用户类型，后台与会员共用 auth_login_log。
type UserTypeEnum int8

const (
	UserTypeAdmin  UserTypeEnum = 1 // 后台用户
	UserTypeMember UserTypeEnum = 2 // 会员
)

// String 返回用户类型的文本描述。
func (t UserTypeEnum) String() string {
	switch t {
	case UserTypeAdmin:
		return "后台用户"
	case UserTypeMember:
		return "会员"
	default:
		return "未知(" + strconv.FormatInt(int64(t), 10) + ")"
	}
}

// IsValid 判断取值是否在枚举范围内。
func (t UserTypeEnum) IsValid() bool {
	return t == UserTypeAdmin || t == UserTypeMember
}

// LoginTypeEnum 是登录日志的动作类型。
type LoginTypeEnum int8

const (
	LoginTypeLogin  LoginTypeEnum = 1 // 登录
	LoginTypeLogout LoginTypeEnum = 2 // 登出
)

// String 返回动作类型的文本描述。
func (t LoginTypeEnum) String() string {
	switch t {
	case LoginTypeLogin:
		return "登录"
	case LoginTypeLogout:
		return "登出"
	default:
		return "未知(" + strconv.FormatInt(int64(t), 10) + ")"
	}
}

// IsValid 判断取值是否在枚举范围内。
func (t LoginTypeEnum) IsValid() bool {
	return t == LoginTypeLogin || t == LoginTypeLogout
}

// LoginResultEnum 是登录日志的结果，0 失败 / 1 成功。
type LoginResultEnum int8

const (
	LoginResultFail    LoginResultEnum = 0 // 失败
	LoginResultSuccess LoginResultEnum = 1 // 成功
)

// String 返回结果的文本描述。
func (r LoginResultEnum) String() string {
	switch r {
	case LoginResultFail:
		return "失败"
	case LoginResultSuccess:
		return "成功"
	default:
		return "未知(" + strconv.FormatInt(int64(r), 10) + ")"
	}
}

// IsValid 判断取值是否在枚举范围内。
func (r LoginResultEnum) IsValid() bool {
	return r == LoginResultFail || r == LoginResultSuccess
}
