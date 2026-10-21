package request

import (
	"github.com/goravel/framework/contracts/http"
)

// MemberSignConfigRequest 是后台新增 / 编辑签到档位的入参。
type MemberSignConfigRequest struct {
	Day        int  `form:"day" json:"day"`
	Point      int  `form:"point" json:"point"`
	Experience int  `form:"experience" json:"experience"`
	Status     int8 `form:"status" json:"status"`
}

// Rules 返回校验规则。
func (r *MemberSignConfigRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"day":        "required|integer|min:1|max:31",
		"point":      "integer|min:0",
		"experience": "integer|min:0",
		"status":     "integer|in:0,1",
	}
}

// Messages 覆盖默认提示。
func (r *MemberSignConfigRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"day.required":       "签到天数不能为空",
		"day.integer":        "签到天数必须是整数",
		"day.min":            "签到天数取值范围 1 ~ 31",
		"day.max":            "签到天数取值范围 1 ~ 31",
		"point.min":          "奖励积分不能为负数",
		"point.integer":      "奖励积分必须是整数",
		"experience.min":     "奖励经验不能为负数",
		"experience.integer": "奖励经验必须是整数",
		"status.in":          "状态取值不合法",
		"status.integer":     "状态取值不合法",
	}
}
