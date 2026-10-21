package request

import (
	"github.com/goravel/framework/contracts/http"
)

// MemberSocialClientSaveRequest 是新增 / 编辑第三方登录配置的入参。
type MemberSocialClientSaveRequest struct {
	// Name 应用名
	Name string `form:"name" json:"name"`
	// SocialType 渠道：wechat_mini / wechat_mp / wechat_open
	SocialType string `form:"socialType" json:"socialType"`
	// ClientID 对应微信 appid
	ClientID string `form:"clientId" json:"clientId"`
	// ClientSecret 对应微信 appsecret；编辑留空表示不修改
	ClientSecret string `form:"clientSecret" json:"clientSecret"`
	// AgentID 预留：钉钉 / 企微等渠道
	AgentID string `form:"agentId" json:"agentId"`
	Status  int8   `form:"status" json:"status"`
}

// Rules 返回校验规则。
func (r *MemberSocialClientSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":         "required|max:64",
		"socialType":   "required|in:wechat_mini,wechat_mp,wechat_open",
		"clientId":     "required|max:128",
		"clientSecret": "max:256",
		"agentId":      "max:64",
		"status":       "required|integer|in:0,1",
	}
}

// Messages 覆盖默认提示。
func (r *MemberSocialClientSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":       "应用名称不能为空",
		"name.max":            "应用名称不能超过 64 个字符",
		"socialType.required": "第三方登录渠道不能为空",
		"socialType.in":       "不支持的第三方登录渠道",
		"clientId.required":   "ClientID 不能为空",
		"clientId.max":        "ClientID 不能超过 128 个字符",
		"clientSecret.max":    "ClientSecret 不能超过 256 个字符",
		"agentId.max":         "AgentID 不能超过 64 个字符",
		"status.required":     "状态不能为空",
		"status.integer":      "状态取值不合法",
		"status.in":           "状态取值不合法",
	}
}
