package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// MemberSocialBindingTable 是会员与第三方账号的绑定表名。
const MemberSocialBindingTable = "member_social_binding"

// MemberSocialBindingModel 对应 member_social_binding 表：
// 一个会员在同一渠道只能绑定一个账号，一个社交账号也只能绑定一个会员。
type MemberSocialBindingModel struct {
	baseModel.BaseModel
	// UserID 会员 ID，关联 member_user.id
	UserID types.ID `db:"user_id" json:"userId"`
	// SocialType 第三方类型，与 member_social_user.type 一致
	SocialType string `db:"social_type" json:"socialType"`
	// SocialUserID 社交用户 ID，关联 member_social_user.id
	SocialUserID types.ID `db:"social_user_id" json:"socialUserId"`
}

// TableName 返回表名。
func (m *MemberSocialBindingModel) TableName() string {
	return MemberSocialBindingTable
}
