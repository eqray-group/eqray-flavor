package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
)

const MemberGroupTable = "member_group"

// MemberGroupModel 对应 member_group 表，会员通过 member_user.group_id 归属一个分组。
type MemberGroupModel struct {
	baseModel.BaseModel
	Name   string                     `db:"name" json:"name"`
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Remark string                     `db:"remark" json:"remark"`
}

// TableName 返回表名。
func (m *MemberGroupModel) TableName() string {
	return MemberGroupTable
}
