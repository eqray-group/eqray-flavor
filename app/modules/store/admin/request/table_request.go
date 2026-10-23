package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// TableAreaSaveRequest 是新增 / 编辑桌台区域的入参。
// storeId 是雪花 ID，前端必须按字符串提交，避免 JS Number 丢精度。
type TableAreaSaveRequest struct {
	// StoreID 门店 ID；新增时未传则回填请求头门店，编辑时不可跨店移动
	StoreID types.ID `form:"storeId" json:"storeId"`
	// Name 区域名称，同一门店内唯一
	Name string `form:"name" json:"name"`
	// Sort 区域排序，越小越靠前
	Sort int `form:"sort" json:"sort"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回桌台区域入参的校验规则。
func (r *TableAreaSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId": "integer|min:1",
		"name":    "required|max:32",
		"sort":    "integer|min:0",
		"remark":  "max:255",
	}
}

// Messages 覆盖框架默认英文提示，保证接口文案统一为中文。
func (r *TableAreaSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.integer": "门店 ID 必须是整数",
		"storeId.min":     "门店 ID 不能小于 1",
		"name.required":   "区域名称不能为空",
		"name.max":        "区域名称不能超过 32 个字符",
		"sort.integer":    "排序必须是整数",
		"sort.min":        "排序不能小于 0",
		"remark.max":      "备注不能超过 255 个字符",
	}
}

// TableSaveRequest 是新增 / 编辑桌台的入参。
// code 仅新增时生效；未传时后端按「门店编码-桌号」自动生成。
type TableSaveRequest struct {
	// StoreID 门店 ID；新增时未传则回填请求头门店
	StoreID types.ID `form:"storeId" json:"storeId"`
	// AreaID 所属区域 ID，必须与门店一致
	AreaID types.ID `form:"areaId" json:"areaId"`
	// Name 桌号，如 A01
	Name string `form:"name" json:"name"`
	// Code 桌码，全局唯一；传空时自动生成
	Code string `form:"code" json:"code"`
	// Seats 座位数
	Seats int `form:"seats" json:"seats"`
	// Sort 区域内排序
	Sort int `form:"sort" json:"sort"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回桌台入参的校验规则。
func (r *TableSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId": "integer|min:1",
		"areaId":  "required|integer|min:1",
		"name":    "required|max:32",
		"code":    "max:64",
		"seats":   "required|integer|between:1,99",
		"sort":    "integer|min:0",
		"remark":  "max:255",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *TableSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.integer": "门店 ID 必须是整数",
		"storeId.min":     "门店 ID 不能小于 1",
		"areaId.required": "所属区域不能为空",
		"areaId.integer":  "所属区域 ID 必须是整数",
		"areaId.min":      "所属区域 ID 不能小于 1",
		"name.required":   "桌号不能为空",
		"name.max":        "桌号不能超过 32 个字符",
		"code.max":        "桌码不能超过 64 个字符",
		"seats.required":  "座位数不能为空",
		"seats.integer":   "座位数必须是整数",
		"seats.between":   "座位数需在 1 ~ 99 之间",
		"sort.integer":    "排序必须是整数",
		"sort.min":        "排序不能小于 0",
		"remark.max":      "备注不能超过 255 个字符",
	}
}

// TableBatchRequest 是批量生成桌台的入参。
type TableBatchRequest struct {
	// StoreID 门店 ID；未传则回填请求头门店
	StoreID types.ID `form:"storeId" json:"storeId"`
	// AreaID 所属区域 ID
	AreaID types.ID `form:"areaId" json:"areaId"`
	// Prefix 桌号前缀，如 A
	Prefix string `form:"prefix" json:"prefix"`
	// Start 起始序号
	Start int `form:"start" json:"start"`
	// Count 生成数量，单次最多 100
	Count int `form:"count" json:"count"`
	// Seats 每张桌台的座位数
	Seats int `form:"seats" json:"seats"`
}

// Rules 返回批量生成入参的校验规则。
func (r *TableBatchRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId": "integer|min:1",
		"areaId":  "required|integer|min:1",
		"prefix":  "required|max:16",
		"start":   "required|integer|between:1,9999",
		"count":   "required|integer|between:1,100",
		"seats":   "required|integer|between:1,99",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *TableBatchRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.integer": "门店 ID 必须是整数",
		"storeId.min":     "门店 ID 不能小于 1",
		"areaId.required": "所属区域不能为空",
		"areaId.integer":  "所属区域 ID 必须是整数",
		"areaId.min":      "所属区域 ID 不能小于 1",
		"prefix.required": "桌号前缀不能为空",
		"prefix.max":      "桌号前缀不能超过 16 个字符",
		"start.required":  "起始序号不能为空",
		"start.integer":   "起始序号必须是整数",
		"start.between":   "起始序号需在 1 ~ 9999 之间",
		"count.required":  "生成数量不能为空",
		"count.integer":   "生成数量必须是整数",
		"count.between":   "单次生成数量需在 1 ~ 100 之间",
		"seats.required":  "座位数不能为空",
		"seats.integer":   "座位数必须是整数",
		"seats.between":   "座位数需在 1 ~ 99 之间",
	}
}

// TableTransferRequest 是桌台换台的入参。
type TableTransferRequest struct {
	// ToTableID 目标桌台 ID，必须是同一门店的空闲桌台
	ToTableID types.ID `form:"toTableId" json:"toTableId"`
}

// Rules 返回换台入参的校验规则。
func (r *TableTransferRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"toTableId": "required|integer|min:1",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *TableTransferRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"toTableId.required": "请选择目标桌台",
		"toTableId.integer":  "目标桌台 ID 必须是整数",
		"toTableId.min":      "目标桌台 ID 不合法",
	}
}

// TableOpenRequest 是开台入参。
// 订单模块尚未接入时 currentOrderId 传 0，仅占台；后续可由订单服务回填台位单。
type TableOpenRequest struct {
	// CurrentOrderID 当前台位单 ID，0 表示尚未建单
	CurrentOrderID types.ID `form:"currentOrderId" json:"currentOrderId"`
}

// Rules 返回开台入参的校验规则。
func (r *TableOpenRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"currentOrderId": "integer|min:0",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *TableOpenRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"currentOrderId.integer": "台位单 ID 必须是整数",
		"currentOrderId.min":     "台位单 ID 不能小于 0",
	}
}
