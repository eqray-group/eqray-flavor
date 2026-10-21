package request

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/enums"
)

// MenuCreateRequest 是新增菜单的入参（docs 6.3）。
// 目录 / 菜单 / 按钮共用一张表，type 决定其余字段的语义。
type MenuCreateRequest struct {
	Name          string                     `form:"name" json:"name"`
	Permission    string                     `form:"permission" json:"permission"`
	Type          enums.MenuTypeEnum         `form:"type" json:"type"`
	Sort          int                        `form:"sort" json:"sort"`
	ParentID      types.ID                   `form:"parentId" json:"parentId"`
	Path          string                     `form:"path" json:"path"`
	Icon          string                     `form:"icon" json:"icon"`
	Component     string                     `form:"component" json:"component"`
	ComponentName string                     `form:"componentName" json:"componentName"`
	Status        baseEnums.CommonStatusEnum `form:"status" json:"status"`
	Visible       int8                       `form:"visible" json:"visible"`
	KeepAlive     int8                       `form:"keepAlive" json:"keepAlive"`
	AlwaysShow    int8                       `form:"alwaysShow" json:"alwaysShow"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *MenuCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":          "required|max:64",
		"permission":    "max:128",
		"type":          "required|integer|in:1,2,3",
		"sort":          "integer",
		"path":          "max:255",
		"icon":          "max:64",
		"component":     "max:255",
		"componentName": "max:64",
		"status":        "required|integer|in:0,1",
		"visible":       "integer|in:0,1",
		"keepAlive":     "integer|in:0,1",
		"alwaysShow":    "integer|in:0,1",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *MenuCreateRequest) Messages(ctx http.Context) map[string]string {
	return menuMessages()
}

// MenuUpdateRequest 是编辑菜单的入参，字段同样必填，保持 PUT 的幂等语义（docs 15.5）。
type MenuUpdateRequest struct {
	Name          string                     `form:"name" json:"name"`
	Permission    string                     `form:"permission" json:"permission"`
	Type          enums.MenuTypeEnum         `form:"type" json:"type"`
	Sort          int                        `form:"sort" json:"sort"`
	ParentID      types.ID                   `form:"parentId" json:"parentId"`
	Path          string                     `form:"path" json:"path"`
	Icon          string                     `form:"icon" json:"icon"`
	Component     string                     `form:"component" json:"component"`
	ComponentName string                     `form:"componentName" json:"componentName"`
	Status        baseEnums.CommonStatusEnum `form:"status" json:"status"`
	Visible       int8                       `form:"visible" json:"visible"`
	KeepAlive     int8                       `form:"keepAlive" json:"keepAlive"`
	AlwaysShow    int8                       `form:"alwaysShow" json:"alwaysShow"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *MenuUpdateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":          "required|max:64",
		"permission":    "max:128",
		"type":          "required|integer|in:1,2,3",
		"sort":          "integer",
		"path":          "max:255",
		"icon":          "max:64",
		"component":     "max:255",
		"componentName": "max:64",
		"status":        "required|integer|in:0,1",
		"visible":       "integer|in:0,1",
		"keepAlive":     "integer|in:0,1",
		"alwaysShow":    "integer|in:0,1",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *MenuUpdateRequest) Messages(ctx http.Context) map[string]string {
	return menuMessages()
}

// menuMessages 是新增与编辑共用的中文文案。
func menuMessages() map[string]string {
	return map[string]string{
		"name.required":      "菜单名称不能为空",
		"name.max":           "菜单名称不能超过 64 个字符",
		"permission.max":     "权限标识不能超过 128 个字符",
		"type.required":      "菜单类型不能为空",
		"type.integer":       "菜单类型必须为整数",
		"type.in":            "菜单类型取值不合法",
		"sort.integer":       "排序必须为整数",
		"path.max":           "路由地址不能超过 255 个字符",
		"icon.max":           "图标不能超过 64 个字符",
		"component.max":      "组件路径不能超过 255 个字符",
		"componentName.max":  "组件名称不能超过 64 个字符",
		"status.required":    "状态不能为空",
		"status.integer":     "状态必须为整数",
		"status.in":          "状态取值不合法",
		"visible.integer":    "是否显示必须为整数",
		"visible.in":         "是否显示只能为 0 或 1",
		"keepAlive.integer":  "是否缓存必须为整数",
		"keepAlive.in":       "是否缓存只能为 0 或 1",
		"alwaysShow.integer": "是否始终显示必须为整数",
		"alwaysShow.in":      "是否始终显示只能为 0 或 1",
	}
}
