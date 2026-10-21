package request

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
)

// UserCreateRequest 是新增用户的入参（docs 6.3）。
// email / mobile 可以不填，Service 会把空串转成 NULL（docs 6.8(1)）。
type UserCreateRequest struct {
	Username string                     `form:"username" json:"username"`
	Password string                     `form:"password" json:"password"`
	Nickname string                     `form:"nickname" json:"nickname"`
	Remark   string                     `form:"remark" json:"remark"`
	Email    string                     `form:"email" json:"email"`
	Mobile   string                     `form:"mobile" json:"mobile"`
	Sex      int8                       `form:"sex" json:"sex"`
	Avatar   string                     `form:"avatar" json:"avatar"`
	Status   baseEnums.CommonStatusEnum `form:"status" json:"status"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *UserCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"username": "required|max:64",
		"password": "required|min:6|max:72",
		"nickname": "max:64",
		"remark":   "max:500",
		"email":    "max:128",
		"mobile":   "max:20",
		"avatar":   "max:512",
		"sex":      "integer|in:0,1,2",
		"status":   "required|integer|in:0,1",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *UserCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required": "账号不能为空",
		"username.max":      "账号不能超过 64 个字符",
		"password.required": "密码不能为空",
		"password.min":      "密码不能少于 6 个字符",
		"password.max":      "密码不能超过 72 个字符",
		"nickname.max":      "昵称不能超过 64 个字符",
		"remark.max":        "备注不能超过 500 个字符",
		"email.max":         "邮箱不能超过 128 个字符",
		"mobile.max":        "手机号不能超过 20 个字符",
		"avatar.max":        "头像地址不能超过 512 个字符",
		"sex.integer":       "性别取值不合法",
		"sex.in":            "性别取值不合法",
		"status.required":   "状态不能为空",
		"status.integer":    "状态必须为整数",
		"status.in":         "状态取值不合法",
	}
}

// UserUpdateRequest 是编辑用户的入参，字段同样必填，保持 PUT 的幂等语义（docs 15.5）。
// 编辑不改密码，密码只能走重置接口。
type UserUpdateRequest struct {
	Username string                     `form:"username" json:"username"`
	Nickname string                     `form:"nickname" json:"nickname"`
	Remark   string                     `form:"remark" json:"remark"`
	Email    string                     `form:"email" json:"email"`
	Mobile   string                     `form:"mobile" json:"mobile"`
	Sex      int8                       `form:"sex" json:"sex"`
	Avatar   string                     `form:"avatar" json:"avatar"`
	Status   baseEnums.CommonStatusEnum `form:"status" json:"status"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *UserUpdateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"username": "required|max:64",
		"nickname": "max:64",
		"remark":   "max:500",
		"email":    "max:128",
		"mobile":   "max:20",
		"avatar":   "max:512",
		"sex":      "integer|in:0,1,2",
		"status":   "required|integer|in:0,1",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *UserUpdateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"username.required": "账号不能为空",
		"username.max":      "账号不能超过 64 个字符",
		"nickname.max":      "昵称不能超过 64 个字符",
		"remark.max":        "备注不能超过 500 个字符",
		"email.max":         "邮箱不能超过 128 个字符",
		"mobile.max":        "手机号不能超过 20 个字符",
		"avatar.max":        "头像地址不能超过 512 个字符",
		"sex.integer":       "性别取值不合法",
		"sex.in":            "性别取值不合法",
		"status.required":   "状态不能为空",
		"status.integer":    "状态必须为整数",
		"status.in":         "状态取值不合法",
	}
}

// UserAssignRoleRequest 是给用户分配角色的入参，roleIds 为空表示收回全部角色。
type UserAssignRoleRequest struct {
	RoleIDs []types.ID `form:"roleIds" json:"roleIds"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *UserAssignRoleRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{}
}

// UserResetPasswordRequest 是重置密码的入参，密码策略与新增保持一致。
type UserResetPasswordRequest struct {
	Password string `form:"password" json:"password"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *UserResetPasswordRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"password": "required|min:6|max:72",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *UserResetPasswordRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"password.required": "密码不能为空",
		"password.min":      "密码不能少于 6 个字符",
		"password.max":      "密码不能超过 72 个字符",
	}
}
