// Package result 提供 Controller 统一的响应输出助手。
package result

import (
	"errors"

	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/response"
)

// businessCodeKey 把本次请求的业务结果码写进上下文，供日志中间件在后置阶段读取。
// HTTP 状态码固定 200，只看状态码无法判断业务是否成功。
const businessCodeKey = "base_result_code"

// businessMessageKey 记录失败文案，API 错误日志需要它落 message 字段。
const businessMessageKey = "base_result_message"

// Success 输出成功响应，HTTP 状态码固定 200，data 无内容时传 nil（docs 8.1）。
func Success(ctx http.Context, data any) http.Response {
	ctx.WithValue(businessCodeKey, errs.Success.Code)

	return ctx.Response().Success().Json(response.Success(data))
}

// Page 输出分页成功响应，只含 list 与 total（docs 8.2）。
func Page(ctx http.Context, list any, total int64) http.Response {
	ctx.WithValue(businessCodeKey, errs.Success.Code)

	return ctx.Response().Success().Json(response.Page(list, total))
}

// Fail 统一输出失败响应：业务异常按自身错误码返回，非预期错误先记日志再按 500 返回。
// HTTP 状态码始终为 200，业务结果只由响应体的 code 表达（docs 8.1）。
func Fail(ctx http.Context, err error) http.Response {
	var be *errs.BusinessError
	code := errs.InternalServerError.Code
	message := errs.InternalServerError.Msg
	if errors.As(err, &be) {
		code = be.Code()
		message = be.Msg()
	} else {
		message = err.Error()
		facades.Log().WithContext(ctx.Context()).Error(err.Error())
	}
	ctx.WithValue(businessCodeKey, code)
	ctx.WithValue(businessMessageKey, message)

	return ctx.Response().Success().Json(response.Fail(err))
}

// MarkFail 只把失败码与文案写入上下文，供日志中间件记录，不输出响应。
// 适用于直接 Abort 的中间件（未登录、无权限等），保证 API 错误日志能拿到失败原因。
func MarkFail(ctx http.Context, code errs.ErrorCode) {
	ctx.WithValue(businessCodeKey, code.Code)
	ctx.WithValue(businessMessageKey, code.Msg)
}

// BusinessMessage 返回本次请求的失败文案，成功或未写入时为空串。
func BusinessMessage(ctx http.Context) string {
	if message, ok := ctx.Value(businessMessageKey).(string); ok {
		return message
	}

	return ""
}

// BusinessCode 返回本次请求的业务结果码：0 表示成功，其他为失败；未写入时返回 -1。
func BusinessCode(ctx http.Context) int {
	if code, ok := ctx.Value(businessCodeKey).(int); ok {
		return code
	}

	return -1
}
