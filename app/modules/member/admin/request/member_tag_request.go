package request

import (
	"github.com/goravel/framework/contracts/http"
)

// MemberTagSaveRequest 是新增 / 编辑会员标签的入参。
type MemberTagSaveRequest struct {
	Name   string `form:"name" json:"name"`
	Status int8   `form:"status" json:"status"`
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *MemberTagSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   "required|max:30",
		"status": "required|integer|in:0,1",
		"remark": "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *MemberTagSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":   "标签名不能为空",
		"name.max":        "标签名不能超过 30 个字符",
		"status.required": "标签状态不能为空",
		"status.integer":  "标签状态取值不合法",
		"status.in":       "标签状态取值不合法",
		"remark.max":      "备注不能超过 255 个字符",
	}
}
