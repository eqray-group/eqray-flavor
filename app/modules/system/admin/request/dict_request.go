package request

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
)

// DictTypeCreateRequest 是新增字典类型的入参（docs 6.3）。
type DictTypeCreateRequest struct {
	Name   string                     `form:"name" json:"name"`
	Type   string                     `form:"type" json:"type"`
	Status baseEnums.CommonStatusEnum `form:"status" json:"status"`
	Remark string                     `form:"remark" json:"remark"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *DictTypeCreateRequest) Rules(ctx http.Context) map[string]any {
	return dictTypeRules()
}

// Messages 覆盖框架默认的英文提示。
func (r *DictTypeCreateRequest) Messages(ctx http.Context) map[string]string {
	return dictTypeMessages()
}

// DictTypeUpdateRequest 是编辑字典类型的入参，类型编码不可修改。
type DictTypeUpdateRequest struct {
	Name   string                     `form:"name" json:"name"`
	Type   string                     `form:"type" json:"type"`
	Status baseEnums.CommonStatusEnum `form:"status" json:"status"`
	Remark string                     `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *DictTypeUpdateRequest) Rules(ctx http.Context) map[string]any {
	return dictTypeRules()
}

// Messages 覆盖框架默认的英文提示。
func (r *DictTypeUpdateRequest) Messages(ctx http.Context) map[string]string {
	return dictTypeMessages()
}

func dictTypeRules() map[string]any {
	return map[string]any{
		"name":   "required|max:64",
		"type":   "required|max:64|regex:^[a-z][a-z0-9_]*$",
		"status": "required|integer|in:0,1",
		"remark": "max:500",
	}
}

func dictTypeMessages() map[string]string {
	return map[string]string{
		"name.required":   "字典名称不能为空",
		"name.max":        "字典名称不能超过 64 个字符",
		"type.required":   "字典类型不能为空",
		"type.max":        "字典类型不能超过 64 个字符",
		"type.regex":      "字典类型只能包含小写字母、数字与下划线，且以字母开头",
		"status.required": "状态不能为空",
		"status.integer":  "状态必须为整数",
		"status.in":       "状态取值不合法",
		"remark.max":      "备注不能超过 500 个字符",
	}
}

// DictDataCreateRequest 是新增字典数据的入参。
type DictDataCreateRequest struct {
	Sort      int                        `form:"sort" json:"sort"`
	Label     string                     `form:"label" json:"label"`
	Value     string                     `form:"value" json:"value"`
	DictType  string                     `form:"dictType" json:"dictType"`
	Status    baseEnums.CommonStatusEnum `form:"status" json:"status"`
	ColorType string                     `form:"colorType" json:"colorType"`
	CSSClass  string                     `form:"cssClass" json:"cssClass"`
	Remark    string                     `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *DictDataCreateRequest) Rules(ctx http.Context) map[string]any {
	return dictDataRules()
}

// Messages 覆盖框架默认的英文提示。
func (r *DictDataCreateRequest) Messages(ctx http.Context) map[string]string {
	return dictDataMessages()
}

// DictDataUpdateRequest 是编辑字典数据的入参。
type DictDataUpdateRequest struct {
	Sort      int                        `form:"sort" json:"sort"`
	Label     string                     `form:"label" json:"label"`
	Value     string                     `form:"value" json:"value"`
	DictType  string                     `form:"dictType" json:"dictType"`
	Status    baseEnums.CommonStatusEnum `form:"status" json:"status"`
	ColorType string                     `form:"colorType" json:"colorType"`
	CSSClass  string                     `form:"cssClass" json:"cssClass"`
	Remark    string                     `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *DictDataUpdateRequest) Rules(ctx http.Context) map[string]any {
	return dictDataRules()
}

// Messages 覆盖框架默认的英文提示。
func (r *DictDataUpdateRequest) Messages(ctx http.Context) map[string]string {
	return dictDataMessages()
}

func dictDataRules() map[string]any {
	return map[string]any{
		"dictType":  "required|max:64",
		"label":     "required|max:64",
		"value":     "required|max:64",
		"sort":      "integer",
		"status":    "required|integer|in:0,1",
		"colorType": "in:default,primary,success,info,warning,danger",
		"cssClass":  "max:64",
		"remark":    "max:500",
	}
}

func dictDataMessages() map[string]string {
	return map[string]string{
		"dictType.required": "字典类型不能为空",
		"dictType.max":      "字典类型不能超过 64 个字符",
		"label.required":    "显示文本不能为空",
		"label.max":         "显示文本不能超过 64 个字符",
		"value.required":    "字典键值不能为空",
		"value.max":         "字典键值不能超过 64 个字符",
		"sort.integer":      "排序必须为整数",
		"status.required":   "状态不能为空",
		"status.integer":    "状态必须为整数",
		"status.in":         "状态取值不合法",
		"colorType.in":      "标签样式取值不合法",
		"cssClass.max":      "样式类不能超过 64 个字符",
		"remark.max":        "备注不能超过 500 个字符",
	}
}
