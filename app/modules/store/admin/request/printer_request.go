package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// PrinterSaveRequest 是新增 / 编辑打印设备的入参。
// secret 只在新增或修改密钥时填写；编辑留空表示沿用原密钥，接口永远不回显明文。
type PrinterSaveRequest struct {
	// StoreID 门店 ID；新增时未传则回填请求头门店，编辑时不可跨店移动
	StoreID types.ID `form:"storeId" json:"storeId"`
	// Name 设备名称，如前台小票机
	Name string `form:"name" json:"name"`
	// DeviceNo 设备号，同一门店内唯一
	DeviceNo string `form:"deviceNo" json:"deviceNo"`
	// Type 设备类型：1 前台小票 / 2 后厨 / 3 标签 / 4 叫号
	Type int8 `form:"type" json:"type"`
	// AccessMode 接入方式：1 云打印机 / 2 局域网 / 3 本地客户端
	AccessMode int8 `form:"accessMode" json:"accessMode"`
	// Config 非敏感配置 JSON，禁止写密钥
	Config string `form:"config" json:"config"`
	// Secret 设备密钥明文；编辑留空表示不修改
	Secret string `form:"secret" json:"secret"`
	// Copies 默认打印份数
	Copies int `form:"copies" json:"copies"`
	// Status 启停状态：0 禁用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回打印设备入参的校验规则。
func (r *PrinterSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId":    "integer|min:1",
		"name":       "required|max:64",
		"deviceNo":   "required|max:64",
		"type":       "required|integer|in:1,2,3,4",
		"accessMode": "required|integer|in:1,2,3",
		"config":     "max:2000",
		"secret":     "max:512",
		"copies":     "required|integer|between:1,9",
		"status":     "required|integer|in:0,1",
		"remark":     "max:255",
	}
}

// Messages 覆盖框架默认英文提示，保证接口文案统一为中文。
func (r *PrinterSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.integer":     "门店 ID 必须是整数",
		"storeId.min":         "门店 ID 不能小于 1",
		"name.required":       "设备名称不能为空",
		"name.max":            "设备名称不能超过 64 个字符",
		"deviceNo.required":   "设备号不能为空",
		"deviceNo.max":        "设备号不能超过 64 个字符",
		"type.required":       "设备类型不能为空",
		"type.integer":        "设备类型必须是整数",
		"type.in":             "设备类型取值不合法",
		"accessMode.required": "接入方式不能为空",
		"accessMode.integer":  "接入方式必须是整数",
		"accessMode.in":       "接入方式取值不合法",
		"config.max":          "打印配置不能超过 2000 个字符",
		"secret.max":          "设备密钥不能超过 512 字节",
		"copies.required":     "默认打印份数不能为空",
		"copies.integer":      "默认打印份数必须是整数",
		"copies.between":      "默认打印份数需在 1 ~ 9 之间",
		"status.required":     "设备状态不能为空",
		"status.integer":      "设备状态必须是整数",
		"status.in":           "设备状态取值不合法",
		"remark.max":          "备注不能超过 255 个字符",
	}
}
