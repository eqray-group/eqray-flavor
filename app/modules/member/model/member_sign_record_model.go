package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

const MemberSignRecordTable = "member_sign_record"

// MemberSignRecordModel 对应 member_sign_record 表，一天一条签到记录。
type MemberSignRecordModel struct {
	baseModel.BaseModel
	UserID   types.ID       `db:"user_id" json:"userId"`
	SignDate types.DateTime `db:"sign_date" json:"signDate"`
	// Day 本次签到是连续第几天，超过最大档位时仍记录真实天数
	Day int `db:"day" json:"day"`
	// Point 本次签到实际获得的积分
	Point int `db:"point" json:"point"`
	// Experience 本次签到实际获得的经验
	Experience int `db:"experience" json:"experience"`
}

// TableName 返回表名。
func (m *MemberSignRecordModel) TableName() string {
	return MemberSignRecordTable
}
