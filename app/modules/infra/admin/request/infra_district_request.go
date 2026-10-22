package request

import (
	"github.com/goravel/framework/contracts/http"
)

// InfraDistrictSaveRequest 是新增 / 编辑行政区划的入参。
// 上级与行政编码只在新增时生效：编辑时后端忽略，避免编码层级被破坏。
type InfraDistrictSaveRequest struct {
	// Pid 上级区划 ID，0 表示顶级
	Pid int64 `form:"pid" json:"pid"`
	// AreaName 区划名称
	AreaName string `form:"areaName" json:"areaName"`
	// ShortName 简称
	ShortName string `form:"shortName" json:"shortName"`
	// AreaCode 行政编码
	AreaCode string `form:"areaCode" json:"areaCode"`
	// Sort 同级排序
	Sort int `form:"sort" json:"sort"`
}

// Rules 返回校验规则。
func (r *InfraDistrictSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"pid":       "integer|min:0",
		"areaName":  "required|max:50",
		"shortName": "required|max:50",
		"areaCode":  "required|max:50",
		"sort":      "integer|min:0",
	}
}

// Messages 覆盖默认提示，保证接口文案统一为中文。
func (r *InfraDistrictSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"pid.integer":        "上级区划必须是整数",
		"pid.min":            "上级区划不能为负数",
		"areaName.required":  "区划名称不能为空",
		"areaName.max":       "区划名称不能超过 50 个字符",
		"shortName.required": "区划简称不能为空",
		"shortName.max":      "区划简称不能超过 50 个字符",
		"areaCode.required":  "行政编码不能为空",
		"areaCode.max":       "行政编码不能超过 50 个字符",
		"sort.integer":       "排序必须是整数",
		"sort.min":           "排序不能小于 0",
	}
}
