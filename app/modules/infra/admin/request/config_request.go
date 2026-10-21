package request

import (
	"github.com/goravel/framework/contracts/http"
)

// ConfigCreateRequest 是新增参数配置的入参。
type ConfigCreateRequest struct {
	Category  string `form:"category" json:"category"`
	Name      string `form:"name" json:"name"`
	ConfigKey string `form:"configKey" json:"configKey"`
	Value     string `form:"value" json:"value"`
	Type      int8   `form:"type" json:"type"`
	Visible   int8   `form:"visible" json:"visible"`
	Remark    string `form:"remark" json:"remark"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *ConfigCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"category":  "max:64",
		"name":      "required|max:128",
		"configKey": "required|max:128",
		"value":     "max:512",
		"type":      "integer|in:1,2",
		"visible":   "integer|in:0,1",
		"remark":    "max:500",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *ConfigCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"category.max":       "参数分组不能超过 64 个字符",
		"name.required":      "参数名称不能为空",
		"name.max":           "参数名称不能超过 128 个字符",
		"configKey.required": "参数键不能为空",
		"configKey.max":      "参数键不能超过 128 个字符",
		"value.max":          "参数值不能超过 512 个字符",
		"type.integer":       "参数类型不合法",
		"type.in":            "参数类型不合法",
		"visible.integer":    "是否公开取值不合法",
		"visible.in":         "是否公开取值不合法",
		"remark.max":         "备注不能超过 500 个字符",
	}
}

// ConfigUpdateRequest 是编辑参数配置的入参；参数键与类型创建后不可修改。
type ConfigUpdateRequest struct {
	Category string `form:"category" json:"category"`
	Name     string `form:"name" json:"name"`
	Value    string `form:"value" json:"value"`
	Visible  int8   `form:"visible" json:"visible"`
	Remark   string `form:"remark" json:"remark"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *ConfigUpdateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"category": "max:64",
		"name":     "required|max:128",
		"value":    "max:512",
		"visible":  "integer|in:0,1",
		"remark":   "max:500",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *ConfigUpdateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"category.max":    "参数分组不能超过 64 个字符",
		"name.required":   "参数名称不能为空",
		"name.max":        "参数名称不能超过 128 个字符",
		"value.max":       "参数值不能超过 512 个字符",
		"visible.integer": "是否公开取值不合法",
		"visible.in":      "是否公开取值不合法",
		"remark.max":      "备注不能超过 500 个字符",
	}
}
