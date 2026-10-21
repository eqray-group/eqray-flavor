package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
)

// MemberLevelTable 是表名常量，Model 与 Repository 共用同一来源。
const MemberLevelTable = "member_level"

// MemberLevelModel 对应 member_level 表，结构对齐 yudao 的会员等级。
// level 是等级的展示序号，member_user.level_id 引用本表主键。
type MemberLevelModel struct {
	baseModel.BaseModel
	// Level 等级序号，1-99，唯一
	Level int    `db:"level" json:"level"`
	Name  string `db:"name" json:"name"`
	// Icon 前端展示用的图标名，可为空
	Icon string `db:"icon" json:"icon"`
	// Experience 升级所需经验值
	Experience int `db:"experience" json:"experience"`
	// DiscountPercent 折扣百分比：100 表示无折扣，95 表示 95 折
	DiscountPercent int                        `db:"discount_percent" json:"discountPercent"`
	Description     string                     `db:"description" json:"description"`
	Status          baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Sort            int                        `db:"sort" json:"sort"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *MemberLevelModel) TableName() string {
	return MemberLevelTable
}
