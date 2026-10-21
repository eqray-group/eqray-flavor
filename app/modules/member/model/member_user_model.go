package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// MemberUserTable 是表名常量，Model 与 Repository 共用同一来源。
const MemberUserTable = "member_user"

// MemberUserModel 对应 member_user 表，表结构对齐 yudao 的会员用户设计。
// 登录账号体系仍归 system_users，本表只描述 C 端会员身份与消费属性。
type MemberUserModel struct {
	baseModel.BaseModel
	Mobile string `db:"mobile" json:"mobile"`
	// Password 预留：后续开放 C 端注册登录时存 bcrypt 哈希
	Password string `db:"password" json:"-"`
	Nickname string `db:"nickname" json:"nickname"`
	Avatar   string `db:"avatar" json:"avatar"`
	Name     string `db:"name" json:"name"`
	// Sex 0 未知 / 1 男 / 2 女
	Sex      int8            `db:"sex" json:"sex"`
	Birthday *types.DateTime `db:"birthday" json:"birthday"`
	AreaID   int             `db:"area_id" json:"areaId"`
	// LevelID 关联 member_level.id
	LevelID types.ID `db:"level_id" json:"levelId"`
	// Point 当前积分
	Point int `db:"point" json:"point"`
	// Experience 成长值
	Experience int `db:"experience" json:"experience"`
	// TagIDs 会员标签 ID 列表，逗号分隔
	TagIDs string `db:"tag_ids" json:"tagIds"`
	// Mark 后台标记，如「高价值客户」
	Mark       string                     `db:"mark" json:"mark"`
	RegisterIP string                     `db:"register_ip" json:"registerIp"`
	LoginIP    string                     `db:"login_ip" json:"loginIp"`
	LoginDate  *types.DateTime            `db:"login_date" json:"loginDate"`
	GroupID    types.ID                   `db:"group_id" json:"groupId"`
	Status     baseEnums.CommonStatusEnum `db:"status" json:"status"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *MemberUserModel) TableName() string {
	return MemberUserTable
}
