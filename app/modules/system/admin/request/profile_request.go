package request

import (
	"github.com/goravel/framework/contracts/http"
)

// ProfileUpdateRequest 是当前登录者编辑个人资料的入参。
// 账号与状态不允许自助修改，只能由管理员在用户管理中调整。
type ProfileUpdateRequest struct {
	Nickname string `form:"nickname" json:"nickname"`
	Email    string `form:"email" json:"email"`
	Mobile   string `form:"mobile" json:"mobile"`
	Sex      int8   `form:"sex" json:"sex"`
	Avatar   string `form:"avatar" json:"avatar"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *ProfileUpdateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"nickname": "required|max:64",
		"email":    "max:128",
		"mobile":   "max:20",
		"avatar":   "max:512",
		"sex":      "integer|in:0,1,2",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *ProfileUpdateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname.required": "昵称不能为空",
		"nickname.max":      "昵称不能超过 64 个字符",
		"email.max":         "邮箱不能超过 128 个字符",
		"mobile.max":        "手机号不能超过 20 个字符",
		"avatar.max":        "头像地址不能超过 512 个字符",
		"sex.integer":       "性别取值不合法",
		"sex.in":            "性别取值不合法",
	}
}

// ProfilePasswordRequest 是当前登录者修改密码的入参，需要先校验旧密码。
type ProfilePasswordRequest struct {
	OldPassword string `form:"oldPassword" json:"oldPassword"`
	NewPassword string `form:"newPassword" json:"newPassword"`
}

// Rules 返回校验规则，密码策略与新增用户保持一致。
func (r *ProfilePasswordRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"oldPassword": "required|min:6|max:72",
		"newPassword": "required|min:6|max:72",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *ProfilePasswordRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"oldPassword.required": "当前密码不能为空",
		"oldPassword.min":      "当前密码不能少于 6 个字符",
		"oldPassword.max":      "当前密码不能超过 72 个字符",
		"newPassword.required": "新密码不能为空",
		"newPassword.min":      "新密码不能少于 6 个字符",
		"newPassword.max":      "新密码不能超过 72 个字符",
	}
}
