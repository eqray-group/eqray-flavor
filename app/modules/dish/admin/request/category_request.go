package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// CategorySaveRequest 是新增 / 编辑菜品分类的入参。
// parentId 为雪花 ID，前端必须按字符串提交，避免 JS Number 丢精度；0 表示根分类。
type CategorySaveRequest struct {
	// ParentID 上级分类 ID，0 表示根分类
	ParentID types.ID `form:"parentId" json:"parentId"`
	// Name 分类名称，同一上级下唯一
	Name string `form:"name" json:"name"`
	// Image 分类图标地址
	Image string `form:"image" json:"image"`
	// Sort 排序，越小越靠前
	Sort int `form:"sort" json:"sort"`
	// Status 启停状态：0 禁用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回菜品分类入参的校验规则。
func (r *CategorySaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"parentId": "integer|min:0",
		"name":     "required|max:64",
		"image":    "max:255",
		"sort":     "integer|min:0",
		"status":   "required|integer|in:0,1",
		"remark":   "max:255",
	}
}

// Messages 覆盖框架默认英文提示，保证接口文案统一为中文。
func (r *CategorySaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"parentId.integer": "上级分类 ID 必须是整数",
		"parentId.min":     "上级分类 ID 不能小于 0",
		"name.required":    "分类名称不能为空",
		"name.max":         "分类名称不能超过 64 个字符",
		"image.max":        "分类图标地址不能超过 255 个字符",
		"sort.integer":     "排序必须是整数",
		"sort.min":         "排序不能小于 0",
		"status.required":  "分类状态不能为空",
		"status.integer":   "分类状态必须是整数",
		"status.in":        "分类状态只能是 0 或 1",
		"remark.max":       "备注不能超过 255 个字符",
	}
}
