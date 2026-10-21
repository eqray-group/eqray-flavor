package request

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
)

// RoleCreateRequest 是新增角色的入参（docs 6.3）。
type RoleCreateRequest struct {
	Name   string                     `form:"name" json:"name"`
	Code   string                     `form:"code" json:"code"`
	Sort   int                        `form:"sort" json:"sort"`
	Status baseEnums.CommonStatusEnum `form:"status" json:"status"`
	Remark string                     `form:"remark" json:"remark"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *RoleCreateRequest) Rules(ctx http.Context) map[string]any {
	return roleRules()
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *RoleCreateRequest) Messages(ctx http.Context) map[string]string {
	return roleMessages()
}

// RoleUpdateRequest 是编辑角色的入参，字段同样必填（docs 15.5）。
type RoleUpdateRequest struct {
	Name   string                     `form:"name" json:"name"`
	Code   string                     `form:"code" json:"code"`
	Sort   int                        `form:"sort" json:"sort"`
	Status baseEnums.CommonStatusEnum `form:"status" json:"status"`
	Remark string                     `form:"remark" json:"remark"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *RoleUpdateRequest) Rules(ctx http.Context) map[string]any {
	return roleRules()
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *RoleUpdateRequest) Messages(ctx http.Context) map[string]string {
	return roleMessages()
}

// RoleAssignMenuRequest 是角色授权的入参，menuIds 为空表示收回全部授权。
type RoleAssignMenuRequest struct {
	MenuIDs []types.ID `form:"menuIds" json:"menuIds"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *RoleAssignMenuRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{}
}

// roleRules 是新增与编辑共用的校验规则。
func roleRules() map[string]any {
	return map[string]any{
		"name":   "required|max:64",
		"code":   "required|max:64|regex:^[A-Za-z][A-Za-z0-9_-]*$",
		"sort":   "integer",
		"status": "required|integer|in:0,1",
		"remark": "max:500",
	}
}

// roleMessages 是新增与编辑共用的中文文案。
func roleMessages() map[string]string {
	return map[string]string{
		"name.required":   "角色名称不能为空",
		"name.max":        "角色名称不能超过 64 个字符",
		"code.required":   "角色标识不能为空",
		"code.max":        "角色标识不能超过 64 个字符",
		"code.regex":      "角色标识只能包含字母、数字、下划线与中划线，且以字母开头",
		"sort.integer":    "排序必须为整数",
		"status.required": "状态不能为空",
		"status.integer":  "状态必须为整数",
		"status.in":       "状态取值不合法",
		"remark.max":      "备注不能超过 500 个字符",
	}
}
