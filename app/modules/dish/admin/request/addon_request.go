package request

import (
	"github.com/goravel/framework/contracts/http"
)

// AddonGroupSaveRequest 是新增 / 编辑加料组的入参。
type AddonGroupSaveRequest struct {
	// Name 加料组名称，全局唯一
	Name string `form:"name" json:"name"`
	// Required 是否必选：0 可选 / 1 必选
	Required int8 `form:"required" json:"required"`
	// MinSelect 最少选择数量
	MinSelect int `form:"minSelect" json:"minSelect"`
	// MaxSelect 最多选择数量，0 表示不限
	MaxSelect int `form:"maxSelect" json:"maxSelect"`
	// Sort 排序，越小越靠前
	Sort int `form:"sort" json:"sort"`
	// Status 启停状态：0 禁用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回加料组入参的校验规则。
func (r *AddonGroupSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":      "required|max:64",
		"required":  "required|integer|in:0,1",
		"minSelect": "integer|min:0",
		"maxSelect": "integer|min:0",
		"sort":      "integer|min:0",
		"status":    "required|integer|in:0,1",
		"remark":    "max:255",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *AddonGroupSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":     "加料组名称不能为空",
		"name.max":          "加料组名称不能超过 64 个字符",
		"required.required": "必选标记不能为空",
		"required.integer":  "必选标记必须是整数",
		"required.in":       "必选标记只能是 0 或 1",
		"minSelect.integer": "最少选择数量必须是整数",
		"minSelect.min":     "最少选择数量不能小于 0",
		"maxSelect.integer": "最多选择数量必须是整数",
		"maxSelect.min":     "最多选择数量不能小于 0",
		"sort.integer":      "排序必须是整数",
		"sort.min":          "排序不能小于 0",
		"status.required":   "加料组状态不能为空",
		"status.integer":    "加料组状态必须是整数",
		"status.in":         "加料组状态只能是 0 或 1",
		"remark.max":        "备注不能超过 255 个字符",
	}
}

// AddonItemSaveRequest 是新增 / 编辑加料项的入参。
type AddonItemSaveRequest struct {
	// Name 加料项名称，组内唯一
	Name string `form:"name" json:"name"`
	// Price 加价（分），0 表示免费
	Price int64 `form:"price" json:"price"`
	// Sort 组内排序
	Sort int `form:"sort" json:"sort"`
	// Status 启停状态：0 禁用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回加料项入参的校验规则。
func (r *AddonItemSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   "required|max:64",
		"price":  "required|integer|min:0",
		"sort":   "integer|min:0",
		"status": "required|integer|in:0,1",
		"remark": "max:255",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *AddonItemSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":   "加料项名称不能为空",
		"name.max":        "加料项名称不能超过 64 个字符",
		"price.required":  "加价不能为空",
		"price.integer":   "加价必须是整数（单位分）",
		"price.min":       "加价不能为负数",
		"sort.integer":    "排序必须是整数",
		"sort.min":        "排序不能小于 0",
		"status.required": "加料项状态不能为空",
		"status.integer":  "加料项状态必须是整数",
		"status.in":       "加料项状态只能是 0 或 1",
		"remark.max":      "备注不能超过 255 个字符",
	}
}
