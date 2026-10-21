// Package response 提供统一的 {code,msg,data} 响应封装，见 docs 8.1。
package response

import (
	"errors"

	"eqray-flavor/app/modules/base/errs"
)

// Body 是成功响应体：data 始终输出，避免前端按字段存在性判断。
type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// FailBody 是失败响应体：按 docs 8.1 省略 data 字段，判断成功一律用 code。
type FailBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// PageData 是分页响应，只含 list 与 total（docs 8.2）。
type PageData struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
}

// Success 构造成功响应，msg 固定为空字符串。
func Success(data any) Body {
	return Body{Code: errs.Success.Code, Msg: errs.Success.Msg, Data: data}
}

// Page 构造分页成功响应。
func Page(list any, total int64) Body {
	return Success(PageData{List: list, Total: total})
}

// Fail 把 error 转成失败响应：业务异常用其自身错误码，其余一律 500。
func Fail(err error) any {
	var be *errs.BusinessError
	if errors.As(err, &be) {
		return FailBody{Code: be.Code(), Msg: be.Msg()}
	}
	return FailBody{Code: errs.InternalServerError.Code, Msg: errs.InternalServerError.Msg}
}
