package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

const MemberLevelRecordTable = "member_level_record"

// MemberLevelRecordModel 对应 member_level_record 表，记录会员等级变更。
// 后台调整等级、经验值变动都写入本表，用于追溯等级变化。
type MemberLevelRecordModel struct {
	baseModel.BaseModel
	UserID types.ID `db:"user_id" json:"userId"`
	// LevelID 变更后的等级 ID
	LevelID     types.ID `db:"level_id" json:"levelId"`
	Experience  int      `db:"experience" json:"experience"`
	Title       string   `db:"title" json:"title"`
	Description string   `db:"description" json:"description"`
}

// TableName 返回表名。
func (m *MemberLevelRecordModel) TableName() string {
	return MemberLevelRecordTable
}
