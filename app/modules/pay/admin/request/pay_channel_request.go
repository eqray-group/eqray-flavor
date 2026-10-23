package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// PayChannelSaveRequest 是新增 / 编辑支付渠道的入参。
// appId 是雪花 ID，前端必须按字符串提交；编辑时应用与渠道编码不可变更，后端会忽略。
// config 是渠道参数 JSON 字符串，敏感字段回传掩码值表示沿用原密钥。
type PayChannelSaveRequest struct {
	// AppID 所属应用编号
	AppID types.ID `form:"appId" json:"appId"`
	// Code 渠道编码，取值见 enums.PayChannelItems
	Code string `form:"code" json:"code"`
	// Status 开启状态：0 开启 / 1 关闭
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
	// FeeRate 渠道费率，单位百分比
	FeeRate float64 `form:"feeRate" json:"feeRate"`
	// Config 渠道配置 JSON，如 {"appId":"","mchId":""}
	Config string `form:"config" json:"config"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *PayChannelSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		// appId 是雪花 ID，按字符串提交，合法性由 types.ID 反序列化兜底
		"code":    "required|max:32",
		"status":  "required|integer|in:0,1",
		"remark":  "max:255",
		"feeRate": "numeric|between:0,100",
		"config":  "required",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *PayChannelSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"code.required":   "渠道编码不能为空",
		"code.max":        "渠道编码不能超过 32 个字符",
		"status.required": "渠道状态不能为空",
		"status.integer":  "渠道状态必须是整数",
		"status.in":       "渠道状态只能是 0（开启）或 1（关闭）",
		"remark.max":      "备注不能超过 255 个字符",
		"feeRate.numeric": "渠道费率必须是数字",
		"feeRate.between": "渠道费率必须在 0 ~ 100 之间",
		"config.required": "渠道配置不能为空",
	}
}
