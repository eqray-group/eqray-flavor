package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// PrintTemplateSaveRequest 是新增 / 编辑打印模板的入参。
// 版本号由服务层维护，客户端不能直接提交，避免绕过正文变化校验。
type PrintTemplateSaveRequest struct {
	// StoreID 门店 ID；新增时未传则回填请求头门店，编辑时不可跨店移动
	StoreID types.ID `form:"storeId" json:"storeId"`
	// Name 模板名称，便于运营区分
	Name string `form:"name" json:"name"`
	// PrinterType 适用设备类型：1 前台小票 / 2 后厨 / 3 标签 / 4 叫号
	PrinterType int8 `form:"printerType" json:"printerType"`
	// PrintType 业务打印类型：1 结账单 / 2 预结单 / 3 厨打 / 4 退菜 /
	// 5 标签 / 6 退款单 / 7 交班 / 8 日结
	PrintType int8 `form:"printType" json:"printType"`
	// PaperWidth 纸宽，单位毫米
	PaperWidth int `form:"paperWidth" json:"paperWidth"`
	// Content 模板正文
	Content string `form:"content" json:"content"`
	// Status 启停状态：0 禁用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回打印模板入参的校验规则。
func (r *PrintTemplateSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId":     "integer|min:1",
		"name":        "required|max:64",
		"printerType": "required|integer|in:1,2,3,4",
		"printType":   "required|integer|in:1,2,3,4,5,6,7,8",
		"paperWidth":  "required|integer|between:40,200",
		"content":     "required|max:65535",
		"status":      "required|integer|in:0,1",
		"remark":      "max:255",
	}
}

// Messages 覆盖框架默认英文提示，保证接口文案统一为中文。
func (r *PrintTemplateSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.integer":      "门店 ID 必须是整数",
		"storeId.min":          "门店 ID 不能小于 1",
		"name.required":        "模板名称不能为空",
		"name.max":             "模板名称不能超过 64 个字符",
		"printerType.required": "设备类型不能为空",
		"printerType.integer":  "设备类型必须是整数",
		"printerType.in":       "设备类型取值不合法",
		"printType.required":   "打印类型不能为空",
		"printType.integer":    "打印类型必须是整数",
		"printType.in":         "打印类型取值不合法",
		"paperWidth.required":  "纸宽不能为空",
		"paperWidth.integer":   "纸宽必须是整数",
		"paperWidth.between":   "纸宽需在 40 ~ 200 mm 之间",
		"content.required":     "模板内容不能为空",
		"content.max":          "模板内容不能超过 65535 个字符",
		"status.required":      "模板状态不能为空",
		"status.integer":       "模板状态必须是整数",
		"status.in":            "模板状态取值不合法",
		"remark.max":           "备注不能超过 255 个字符",
	}
}
