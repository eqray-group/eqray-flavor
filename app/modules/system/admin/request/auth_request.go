// Package request 存放 system 模块的请求参数与校验规则。
package request

import (
	"github.com/goravel/framework/contracts/http"
)

// LoginRequest 是后台登录入参（docs 9.1）。
type LoginRequest struct {
	Username    string `form:"username" json:"username"`
	Password    string `form:"password" json:"password"`
	CaptchaID   string `form:"captchaId" json:"captchaId"`
	CaptchaCode string `form:"captchaCode" json:"captchaCode"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *LoginRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"username":    "required|max:64",
		"password":    "required|max:128",
		"captchaId":   "required|max:64",
		"captchaCode": "required|max:8",
	}
}

// RefreshRequest 是刷新令牌入参（docs 9.2）。
type RefreshRequest struct {
	RefreshToken string `form:"refreshToken" json:"refreshToken"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *RefreshRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"refreshToken": "required|max:128",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *RefreshRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"refreshToken.required": "刷新令牌不能为空",
		"refreshToken.max":      "刷新令牌格式不正确",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *LoginRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required":    "账号不能为空",
		"username.max":         "账号不能超过 64 个字符",
		"password.required":    "密码不能为空",
		"password.max":         "密码不能超过 128 个字符",
		"captchaId.required":   "验证码标识不能为空",
		"captchaCode.required": "验证码不能为空",
		"captchaCode.max":      "验证码格式不正确",
	}
}
