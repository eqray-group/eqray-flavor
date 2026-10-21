package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// InfraSmsChannelSaveRequest 是新增 / 编辑短信渠道的入参。
type InfraSmsChannelSaveRequest struct {
	// Name 渠道名称
	Name string `form:"name" json:"name"`
	// Code 渠道代码：DebugLog / AliyunSms / TencentSms
	Code string `form:"code" json:"code"`
	// APIKey 阿里云 AccessKeyId / 腾讯云 SecretId
	APIKey string `form:"apiKey" json:"apiKey"`
	// APISecret 阿里云 AccessKeySecret / 腾讯云 SecretKey，编辑留空表示不修改
	APISecret string `form:"apiSecret" json:"apiSecret"`
	// Signature 短信签名
	Signature string `form:"signature" json:"signature"`
	// Extra 扩展配置 JSON：腾讯云 {"region","sdkAppId"}
	Extra  string `form:"extra" json:"extra"`
	Status int8   `form:"status" json:"status"`
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *InfraSmsChannelSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":      "required|max:64",
		"code":      "required|in:DebugLog,AliyunSms,TencentSms",
		"apiKey":    "max:128",
		"apiSecret": "max:256",
		"signature": "max:64",
		"extra":     "max:512",
		"status":    "required|integer|in:0,1",
		"remark":    "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *InfraSmsChannelSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":   "渠道名称不能为空",
		"name.max":        "渠道名称不能超过 64 个字符",
		"code.required":   "渠道代码不能为空",
		"code.in":         "不支持的短信渠道",
		"apiKey.max":      "APIKey 不能超过 128 个字符",
		"apiSecret.max":   "APISecret 不能超过 256 个字符",
		"signature.max":   "短信签名不能超过 64 个字符",
		"extra.max":       "扩展配置不能超过 512 个字符",
		"status.required": "状态不能为空",
		"status.integer":  "状态取值不合法",
		"status.in":       "状态取值不合法",
		"remark.max":      "备注不能超过 255 个字符",
	}
}

// InfraSmsTemplateSaveRequest 是新增 / 编辑短信模板的入参。
type InfraSmsTemplateSaveRequest struct {
	ChannelID     types.ID `form:"channelId" json:"channelId"`
	Code          string   `form:"code" json:"code"`
	Name          string   `form:"name" json:"name"`
	Type          int8     `form:"type" json:"type"`
	APITemplateID string   `form:"apiTemplateId" json:"apiTemplateId"`
	Content       string   `form:"content" json:"content"`
	Params        string   `form:"params" json:"params"`
	Status        int8     `form:"status" json:"status"`
	Remark        string   `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *InfraSmsTemplateSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"channelId":     "required",
		"code":          "required|max:64",
		"name":          "required|max:64",
		"type":          "required|integer|in:1,2",
		"apiTemplateId": "max:128",
		"content":       "required|max:512",
		"params":        "max:128",
		"status":        "required|integer|in:0,1",
		"remark":        "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *InfraSmsTemplateSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"channelId.required": "请选择短信渠道",
		"code.required":      "模板代码不能为空",
		"code.max":           "模板代码不能超过 64 个字符",
		"name.required":      "模板名称不能为空",
		"name.max":           "模板名称不能超过 64 个字符",
		"type.required":      "模板类型不能为空",
		"type.integer":       "模板类型取值不合法",
		"type.in":            "模板类型取值不合法",
		"apiTemplateId.max":  "第三方模板 ID 不能超过 128 个字符",
		"content.required":   "模板内容不能为空",
		"content.max":        "模板内容不能超过 512 个字符",
		"params.max":         "模板变量不能超过 128 个字符",
		"status.required":    "状态不能为空",
		"status.integer":     "状态取值不合法",
		"status.in":          "状态取值不合法",
		"remark.max":         "备注不能超过 255 个字符",
	}
}
