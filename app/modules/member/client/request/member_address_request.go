package request

import (
	"github.com/goravel/framework/contracts/http"
)

// MemberAddressRequest 是 C 端新增 / 编辑收货地址的入参。
// 会员 ID 一律从令牌取，请求体里不接受 userId。
type MemberAddressRequest struct {
	Name          string `form:"name" json:"name"`
	Mobile        string `form:"mobile" json:"mobile"`
	AreaID        int    `form:"areaId" json:"areaId"`
	DetailAddress string `form:"detailAddress" json:"detailAddress"`
	// DefaultStatus 仅新增时生效：1 表示新增后设为默认
	DefaultStatus int8 `form:"defaultStatus" json:"defaultStatus"`
}

// Rules 返回校验规则。
func (r *MemberAddressRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":          "required|max:30",
		"mobile":        "required|max:20",
		"areaId":        "integer|min:0",
		"detailAddress": "required|max:255",
		"defaultStatus": "integer|in:0,1",
	}
}

// Messages 覆盖默认提示。
func (r *MemberAddressRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":          "收件人不能为空",
		"name.max":               "收件人不能超过 30 个字符",
		"mobile.required":        "手机号不能为空",
		"mobile.max":             "手机号格式不正确",
		"areaId.integer":         "地区编号不合法",
		"areaId.min":             "地区编号不合法",
		"detailAddress.required": "详细地址不能为空",
		"detailAddress.max":      "详细地址不能超过 255 个字符",
		"defaultStatus.in":       "默认地址标记不合法",
		"defaultStatus.integer":  "默认地址标记不合法",
	}
}
