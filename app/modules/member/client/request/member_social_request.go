package request

import (
	"github.com/goravel/framework/contracts/http"
)

// MemberSocialLoginRequest 是会员第三方登录的入参。
type MemberSocialLoginRequest struct {
	// Type 渠道：wechat_mini / wechat_mp / wechat_open
	Type string `form:"type" json:"type"`
	// Code 微信授权码：小程序取 wx.login 返回值，公众号 / 开放平台取回调参数
	Code string `form:"code" json:"code"`
	// State 透传参数，仅用于排查问题
	State string `form:"state" json:"state"`
}

// Rules 返回校验规则。
func (r *MemberSocialLoginRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"type":  "required|in:wechat_mini,wechat_mp,wechat_open",
		"code":  "required|max:512",
		"state": "max:64",
	}
}

// Messages 覆盖默认提示。
func (r *MemberSocialLoginRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.required": "第三方登录渠道不能为空",
		"type.in":       "不支持的第三方登录渠道",
		"code.required": "第三方授权码不能为空",
		"code.max":      "第三方授权码不合法",
		"state.max":     "state 不能超过 64 个字符",
	}
}

// MemberSocialBindRequest 是会员绑定第三方账号的入参，字段与登录一致。
type MemberSocialBindRequest struct {
	Type  string `form:"type" json:"type"`
	Code  string `form:"code" json:"code"`
	State string `form:"state" json:"state"`
}

// Rules 返回校验规则。
func (r *MemberSocialBindRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"type":  "required|in:wechat_mini,wechat_mp,wechat_open",
		"code":  "required|max:512",
		"state": "max:64",
	}
}

// Messages 覆盖默认提示。
func (r *MemberSocialBindRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.required": "第三方登录渠道不能为空",
		"type.in":       "不支持的第三方登录渠道",
		"code.required": "第三方授权码不能为空",
		"code.max":      "第三方授权码不合法",
		"state.max":     "state 不能超过 64 个字符",
	}
}

// MemberRefreshRequest 是会员令牌续期的入参。
type MemberRefreshRequest struct {
	RefreshToken string `form:"refreshToken" json:"refreshToken"`
}

// Rules 返回校验规则。
func (r *MemberRefreshRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"refreshToken": "required",
	}
}

// Messages 覆盖默认提示。
func (r *MemberRefreshRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"refreshToken.required": "刷新令牌不能为空",
	}
}

// MemberLogoutRequest 是会员退出的入参，访问令牌从请求头读取。
type MemberLogoutRequest struct {
	RefreshToken string `form:"refreshToken" json:"refreshToken"`
}

// Rules 返回校验规则：刷新令牌可选，缺失时只注销访问令牌。
func (r *MemberLogoutRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{}
}
