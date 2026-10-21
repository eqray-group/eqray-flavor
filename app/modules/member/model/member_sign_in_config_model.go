package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
)

const MemberSignInConfigTable = "member_sign_in_config"

// MemberSignInConfigModel 对应 member_sign_in_config 表：连续签到第 N 天的奖励档位（对齐 yudao）。
type MemberSignInConfigModel struct {
	baseModel.BaseModel
	// Day 连续签到第几天，从 1 开始
	Day int `db:"day" json:"day"`
	// Point 该档位奖励积分
	Point int `db:"point" json:"point"`
	// Experience 该档位奖励经验
	Experience int                        `db:"experience" json:"experience"`
	Status     baseEnums.CommonStatusEnum `db:"status" json:"status"`
}

// TableName 返回表名。
func (m *MemberSignInConfigModel) TableName() string {
	return MemberSignInConfigTable
}
