package request

import (
	"github.com/goravel/framework/contracts/http"
)

// MemberSmsCodeRequest 是发送短信验证码的入参。
type MemberSmsCodeRequest struct {
	Mobile string `form:"mobile" json:"mobile"`
	// Scene 场景标识，默认 login
	Scene string `form:"scene" json:"scene"`
}

// Rules 返回校验规则。
func (r *MemberSmsCodeRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"mobile": "required|max:20",
		"scene":  "max:32",
	}
}

// Messages 覆盖默认提示。
func (r *MemberSmsCodeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"mobile.required": "手机号不能为空",
		"mobile.max":      "手机号格式不正确",
		"scene.max":       "场景标识不合法",
	}
}

// MemberSmsLoginRequest 是手机号 + 验证码登录的入参。
type MemberSmsLoginRequest struct {
	Mobile string `form:"mobile" json:"mobile"`
	Code   string `form:"code" json:"code"`
}

// Rules 返回校验规则。
func (r *MemberSmsLoginRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"mobile": "required|max:20",
		"code":   "required|max:6",
	}
}

// Messages 覆盖默认提示。
func (r *MemberSmsLoginRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"mobile.required": "手机号不能为空",
		"mobile.max":      "手机号格式不正确",
		"code.required":   "验证码不能为空",
		"code.max":        "验证码不合法",
	}
}

// MemberSocialBindLoginRequest 是手机号验证码登录后绑定第三方账号的入参。
type MemberSocialBindLoginRequest struct {
	Mobile      string `form:"mobile" json:"mobile"`
	Code        string `form:"code" json:"code"`
	SocialType  string `form:"socialType" json:"socialType"`
	SocialCode  string `form:"socialCode" json:"socialCode"`
	SocialState string `form:"socialState" json:"socialState"`
}

// Rules 返回校验规则。
func (r *MemberSocialBindLoginRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"mobile":      "required|max:20",
		"code":        "required|max:6",
		"socialType":  "required|in:wechat_mini,wechat_mp,wechat_open",
		"socialCode":  "required|max:512",
		"socialState": "max:64",
	}
}

// Messages 覆盖默认提示。
func (r *MemberSocialBindLoginRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"mobile.required":     "手机号不能为空",
		"mobile.max":          "手机号格式不正确",
		"code.required":       "验证码不能为空",
		"code.max":            "验证码不合法",
		"socialType.required": "第三方登录渠道不能为空",
		"socialType.in":       "不支持的第三方登录渠道",
		"socialCode.required": "第三方授权码不能为空",
		"socialCode.max":      "第三方授权码不合法",
		"socialState.max":     "state 不能超过 64 个字符",
	}
}
