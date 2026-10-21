package request

import (
	"github.com/goravel/framework/contracts/http"
)

// MemberLevelSaveRequest 是新增 / 编辑会员等级的入参。
// level 仅新增时生效，编辑时后端会忽略该字段。
type MemberLevelSaveRequest struct {
	Level           int    `form:"level" json:"level"`
	Name            string `form:"name" json:"name"`
	Icon            string `form:"icon" json:"icon"`
	Experience      int    `form:"experience" json:"experience"`
	DiscountPercent int    `form:"discountPercent" json:"discountPercent"`
	Description     string `form:"description" json:"description"`
	Status          int8   `form:"status" json:"status"`
	Sort            int    `form:"sort" json:"sort"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *MemberLevelSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"level":           "required|integer|min:1|max:99",
		"name":            "required|max:64",
		"icon":            "max:64",
		"experience":      "integer|min:0",
		"discountPercent": "required|integer|min:1|max:100",
		"description":     "max:500",
		"status":          "required|integer|in:0,1",
		"sort":            "integer|min:0",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *MemberLevelSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"level.required":           "等级序号不能为空",
		"level.integer":            "等级序号必须是整数",
		"level.min":                "等级序号不能小于 1",
		"level.max":                "等级序号不能超过 99",
		"name.required":            "等级名称不能为空",
		"name.max":                 "等级名称不能超过 64 个字符",
		"icon.max":                 "图标名不能超过 64 个字符",
		"experience.integer":       "升级经验必须是整数",
		"experience.min":           "升级经验不能小于 0",
		"discountPercent.required": "折扣百分比不能为空",
		"discountPercent.integer":  "折扣百分比必须是整数",
		"discountPercent.min":      "折扣百分比不能小于 1",
		"discountPercent.max":      "折扣百分比不能超过 100",
		"description.max":          "等级说明不能超过 500 个字符",
		"status.required":          "等级状态不能为空",
		"status.integer":           "等级状态取值不合法",
		"status.in":                "等级状态取值不合法",
		"sort.integer":             "排序必须是整数",
		"sort.min":                 "排序不能小于 0",
	}
}
