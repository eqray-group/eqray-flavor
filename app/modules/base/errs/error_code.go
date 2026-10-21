// Package errs 定义系统错误码与业务异常类型，见 docs 8.3。
package errs

import "fmt"

// ErrorCode 是错误码常量：code 决定前端处理，msg 是默认提示文案。
type ErrorCode struct {
	Code int
	Msg  string
}

// BusinessError 对应 yudao 的 ServiceException，由 Service 抛出、全局异常处理统一转换。
type BusinessError struct {
	code ErrorCode
	msg  string
}

// Error 实现 error 接口，输出 "[错误码] 提示文案"，便于日志定位问题。
func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code.Code, e.msg)
}

// Code 返回业务错误码。
func (e *BusinessError) Code() int { return e.code.Code }

// Msg 返回最终提示文案（可能已按占位符格式化）。
func (e *BusinessError) Msg() string { return e.msg }

// ErrorCode 返回原始错误码常量，便于日志与调试。
func (e *BusinessError) ErrorCode() ErrorCode { return e.code }

// 系统错误码固定段（0 ~ 999），全项目共用，不带模块前缀（docs 11.2）。
var (
	Success             = ErrorCode{Code: 0, Msg: ""}
	BadRequest          = ErrorCode{Code: 400, Msg: "请求参数不正确"}
	Unauthorized        = ErrorCode{Code: 401, Msg: "账号未登录"}
	Forbidden           = ErrorCode{Code: 403, Msg: "没有该操作权限"}
	NotFound            = ErrorCode{Code: 404, Msg: "请求未找到"}
	MethodNotAllowed    = ErrorCode{Code: 405, Msg: "请求方法不正确"}
	Locked              = ErrorCode{Code: 423, Msg: "请求失败，请稍后重试"}
	TooManyRequests     = ErrorCode{Code: 429, Msg: "请求过于频繁，请稍后重试"}
	InternalServerError = ErrorCode{Code: 500, Msg: "系统异常"}
	Unknown             = ErrorCode{Code: 999, Msg: "未知错误"}
)

// New 用常量文案构造业务异常；仅在传入 args 时按占位符格式化，避免文案里的 % 被误判。
func New(code ErrorCode, args ...any) *BusinessError {
	msg := code.Msg
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &BusinessError{code: code, msg: msg}
}

// WithMsg 用调用方的文案覆盖常量文案，仅用于参数校验这类文案来自校验器的场景。
func WithMsg(code ErrorCode, msg string) *BusinessError {
	return &BusinessError{code: code, msg: msg}
}
