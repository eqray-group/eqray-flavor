package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// EmployeeSaveRequest 是新增 / 编辑门店员工的入参。
// storeId、userId 是雪花 ID，前端必须按字符串提交，避免 JS Number 丢精度。
type EmployeeSaveRequest struct {
	// StoreID 门店 ID，只能指向门店类型节点
	StoreID types.ID `form:"storeId" json:"storeId"`
	// UserID 后台账号 ID
	UserID types.ID `form:"userId" json:"userId"`
	// Role 门店角色：1 店长 / 2 收银 / 3 后厨 / 4 服务员 / 5 骑手
	Role int8 `form:"role" json:"role"`
	// Status 启停状态：0 禁用 / 1 启用；只有启用记录计入可见门店
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回门店员工入参的校验规则。
func (r *EmployeeSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId": "integer|min:1",
		"userId":  "required|integer|min:1",
		"role":    "required|integer|in:1,2,3,4,5",
		"status":  "required|integer|in:0,1",
		"remark":  "max:255",
	}
}

// Messages 覆盖框架默认英文提示，保证接口文案统一为中文。
func (r *EmployeeSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.integer": "门店 ID 必须是整数",
		"storeId.min":     "门店 ID 不能小于 1",
		"userId.required": "后台账号不能为空",
		"userId.integer":  "后台账号 ID 必须是整数",
		"userId.min":      "后台账号 ID 不能小于 1",
		"role.required":   "门店角色不能为空",
		"role.integer":    "门店角色必须是整数",
		"role.in":         "门店角色取值不合法",
		"status.required": "员工状态不能为空",
		"status.integer":  "员工状态必须是整数",
		"status.in":       "员工状态取值不合法",
		"remark.max":      "备注不能超过 255 个字符",
	}
}
