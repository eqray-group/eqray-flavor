package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// MemberSaveRequest 是新增 / 编辑会员的入参，字段对齐 yudao 的 member_user。
type MemberSaveRequest struct {
	Nickname string `form:"nickname" json:"nickname"`
	Mobile   string `form:"mobile" json:"mobile"`
	Avatar   string `form:"avatar" json:"avatar"`
	// Name 真实姓名
	Name     string          `form:"name" json:"name"`
	Sex      int8            `form:"sex" json:"sex"`
	Birthday *types.DateTime `form:"birthday" json:"birthday"`
	AreaID   int             `form:"areaId" json:"areaId"`
	// LevelID 会员等级 ID
	LevelID    types.ID `form:"levelId" json:"levelId"`
	Point      int      `form:"point" json:"point"`
	Experience int      `form:"experience" json:"experience"`
	// TagIDs 标签 ID，逗号分隔
	TagIDs  string   `form:"tagIds" json:"tagIds"`
	Mark    string   `form:"mark" json:"mark"`
	GroupID types.ID `form:"groupId" json:"groupId"`
	// Status 会员状态：新增时不传默认启用；编辑时不传表示保持原状态
	Status *int8 `form:"status" json:"status"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *MemberSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"nickname":   "required|max:64",
		"mobile":     "required|max:20",
		"avatar":     "max:512",
		"name":       "max:30",
		"sex":        "integer|in:0,1,2",
		"areaId":     "integer|min:0",
		"levelId":    "required",
		"point":      "integer|min:0",
		"experience": "integer|min:0",
		"tagIds":     "max:255",
		"mark":       "max:255",
		"groupId":    "integer|min:0",
		// 不传 status 时跳过校验：新增默认启用，编辑保持原状态
		"status": "integer|in:0,1",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *MemberSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"nickname.required":  "昵称不能为空",
		"nickname.max":       "昵称不能超过 64 个字符",
		"mobile.required":    "手机号不能为空",
		"mobile.max":         "手机号不能超过 20 个字符",
		"avatar.max":         "头像地址不能超过 512 个字符",
		"name.max":           "真实姓名不能超过 30 个字符",
		"sex.integer":        "性别取值不合法",
		"sex.in":             "性别取值不合法",
		"areaId.integer":     "所在地取值不合法",
		"areaId.min":         "所在地取值不合法",
		"levelId.required":   "请选择会员等级",
		"point.integer":      "积分必须是整数",
		"point.min":          "积分不能小于 0",
		"experience.integer": "经验值必须是整数",
		"experience.min":     "经验值不能小于 0",
		"tagIds.max":         "标签数量过多",
		"mark.max":           "用户标记不能超过 255 个字符",
		"groupId.integer":    "会员分组取值不合法",
		"groupId.min":        "会员分组取值不合法",
		"status.integer":     "会员状态取值不合法",
		"status.in":          "会员状态取值不合法",
	}
}

// MemberStatusRequest 是启用 / 禁用会员的入参。
type MemberStatusRequest struct {
	Status int8 `form:"status" json:"status"`
}

// Rules 返回校验规则。
func (r *MemberStatusRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"status": "required|integer|in:0,1",
	}
}

// Messages 覆盖默认提示。
func (r *MemberStatusRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"status.required": "会员状态不能为空",
		"status.integer":  "会员状态取值不合法",
		"status.in":       "会员状态取值不合法",
	}
}

// MemberPointRequest 是调整会员积分的入参，point 用正负表示增减。
type MemberPointRequest struct {
	Point       int    `form:"point" json:"point"`
	Title       string `form:"title" json:"title"`
	Description string `form:"description" json:"description"`
}

// Rules 返回校验规则。
func (r *MemberPointRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"point":       "required|integer",
		"title":       "max:64",
		"description": "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *MemberPointRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"point.required":  "积分变动数量不能为空",
		"point.integer":   "积分变动数量必须是整数",
		"title.max":       "流水标题不能超过 64 个字符",
		"description.max": "流水描述不能超过 255 个字符",
	}
}

// MemberLevelAdjustRequest 是调整会员等级 / 经验值的入参。
type MemberLevelAdjustRequest struct {
	LevelID     types.ID `form:"levelId" json:"levelId"`
	Experience  int      `form:"experience" json:"experience"`
	Title       string   `form:"title" json:"title"`
	Description string   `form:"description" json:"description"`
}

// Rules 返回校验规则。
func (r *MemberLevelAdjustRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"levelId":     "integer|min:0",
		"experience":  "integer|min:0",
		"title":       "max:64",
		"description": "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *MemberLevelAdjustRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"levelId.integer":    "会员等级取值不合法",
		"levelId.min":        "会员等级取值不合法",
		"experience.integer": "经验值必须是整数",
		"experience.min":     "经验值不能小于 0",
		"title.max":          "记录标题不能超过 64 个字符",
		"description.max":    "记录描述不能超过 255 个字符",
	}
}
