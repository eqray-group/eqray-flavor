package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
)

const MemberTagTable = "member_tag"

// MemberTagModel 对应 member_tag 表，会员通过 member_user.tag_ids 关联多个标签。
type MemberTagModel struct {
	baseModel.BaseModel
	Name   string                     `db:"name" json:"name"`
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Remark string                     `db:"remark" json:"remark"`
}

// TableName 返回表名。
func (m *MemberTagModel) TableName() string {
	return MemberTagTable
}
