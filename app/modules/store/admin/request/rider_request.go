package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// RiderCreateRequest 是新增骑手档案的入参。
// userId 为空时按 mobile 自动开通后台账号，姓名与手机号由服务层强校验。
type RiderCreateRequest struct {
	// StoreID 门店 ID；未传时回填请求头门店
	StoreID types.ID `form:"storeId" json:"storeId"`
	// UserID 已有后台账号 ID，0 表示自动开通
	UserID types.ID `form:"userId" json:"userId"`
	// Name 骑手姓名，自动开通账号时作为昵称
	Name string `form:"name" json:"name"`
	// Mobile 骑手手机号，自动开通账号时作为登录账号
	Mobile string `form:"mobile" json:"mobile"`
	// MaxOrders 接单并发上限
	MaxOrders int `form:"maxOrders" json:"maxOrders"`
	// Status 骑手状态：1 空闲 / 2 接单中 / 3 配送中 / 4 休息 / 5 停用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回新增骑手入参的校验规则。
func (r *RiderCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId":   "integer|min:1",
		"userId":    "integer|min:0",
		"name":      "max:64",
		"mobile":    "max:20",
		"maxOrders": "required|integer|between:1,20",
		"status":    "required|integer|in:1,2,3,4,5",
		"remark":    "max:255",
	}
}

// Messages 覆盖框架默认英文提示，保证接口文案统一为中文。
func (r *RiderCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.integer":    "门店 ID 必须是整数",
		"storeId.min":        "门店 ID 不能小于 1",
		"userId.integer":     "后台账号 ID 必须是整数",
		"userId.min":         "后台账号 ID 不能小于 0",
		"name.max":           "骑手姓名不能超过 64 个字符",
		"mobile.max":         "骑手手机号不能超过 20 个字符",
		"maxOrders.required": "接单并发上限不能为空",
		"maxOrders.integer":  "接单并发上限必须是整数",
		"maxOrders.between":  "接单并发上限需在 1 ~ 20 之间",
		"status.required":    "骑手状态不能为空",
		"status.integer":     "骑手状态必须是整数",
		"status.in":          "骑手状态取值不合法",
		"remark.max":         "备注不能超过 255 个字符",
	}
}

// RiderUpdateRequest 是编辑骑手档案的入参。
// 门店、员工与后台账号创建后不可变更，只调整状态、并发上限与备注。
type RiderUpdateRequest struct {
	// MaxOrders 接单并发上限
	MaxOrders int `form:"maxOrders" json:"maxOrders"`
	// Status 骑手状态：1 空闲 / 2 接单中 / 3 配送中 / 4 休息 / 5 停用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回编辑骑手档案入参的校验规则。
func (r *RiderUpdateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"maxOrders": "required|integer|between:1,20",
		"status":    "required|integer|in:1,2,3,4,5",
		"remark":    "max:255",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *RiderUpdateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"maxOrders.required": "接单并发上限不能为空",
		"maxOrders.integer":  "接单并发上限必须是整数",
		"maxOrders.between":  "接单并发上限需在 1 ~ 20 之间",
		"status.required":    "骑手状态不能为空",
		"status.integer":     "骑手状态必须是整数",
		"status.in":          "骑手状态取值不合法",
		"remark.max":         "备注不能超过 255 个字符",
	}
}
