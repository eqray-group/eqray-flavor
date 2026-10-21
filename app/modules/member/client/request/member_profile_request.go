package request

import (
	"time"

	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// MemberProfileUpdateRequest 是 C 端修改个人资料的入参。
// 手机号、积分、等级、状态不在这里，避免会员自助篡改。
type MemberProfileUpdateRequest struct {
	Nickname string `form:"nickname" json:"nickname"`
	Avatar   string `form:"avatar" json:"avatar"`
	Name     string `form:"name" json:"name"`
	Sex      int8   `form:"sex" json:"sex"`
	// Birthday 只接收日期，格式 2006-01-02；传空串表示清空生日
	Birthday string `form:"birthday" json:"birthday"`
	AreaID   int    `form:"areaId" json:"areaId"`
}

// Rules 返回校验规则。
func (r *MemberProfileUpdateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"nickname": "required|max:32",
		"avatar":   "max:255",
		"name":     "max:32",
		"sex":      "integer|in:0,1,2",
		"birthday": "max:20",
		"areaId":   "integer|min:0",
	}
}

// Messages 覆盖默认提示。
func (r *MemberProfileUpdateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname.required": "昵称不能为空",
		"nickname.max":      "昵称不能超过 32 个字符",
		"avatar.max":        "头像地址过长",
		"name.max":          "真实姓名不能超过 32 个字符",
		"sex.in":            "性别取值不合法",
		"sex.integer":       "性别取值不合法",
		"birthday.max":      "生日格式不正确",
		"areaId.min":        "地区编号不合法",
		"areaId.integer":    "地区编号不合法",
	}
}

// BirthdayTime 把日期字符串解析成统一时间类型，空串返回 nil。
func (r *MemberProfileUpdateRequest) BirthdayTime() (*types.DateTime, error) {
	if r.Birthday == "" {
		return nil, nil
	}

	parsed, err := time.ParseInLocation("2006-01-02", r.Birthday, types.Location())
	if err != nil {
		parsed, err = time.ParseInLocation(types.DateTimeLayout, r.Birthday, types.Location())
	}
	if err != nil {
		return nil, err
	}

	value := types.DateTime(parsed)

	return &value, nil
}
