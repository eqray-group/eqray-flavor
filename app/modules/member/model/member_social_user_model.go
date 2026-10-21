package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
)

// MemberSocialUserTable 是第三方社交用户表名。
const MemberSocialUserTable = "member_social_user"

// MemberSocialUserModel 对应 member_social_user 表，保存第三方平台返回的身份信息，
// 与会员档案（member_user）通过 member_social_binding 关联。
type MemberSocialUserModel struct {
	baseModel.BaseModel
	// Type 第三方类型：wechat_mini / wechat_mp / wechat_open
	Type string `db:"type" json:"type"`
	// OpenID 同一应用内的用户唯一标识
	OpenID string `db:"openid" json:"openid"`
	// UnionID 同一开放平台账号下的用户唯一标识，小程序 / 公众号可据此合并身份
	UnionID  string `db:"unionid" json:"unionid"`
	Nickname string `db:"nickname" json:"nickname"`
	Avatar   string `db:"avatar" json:"avatar"`
	// RawData 第三方返回的原始报文，便于排查问题
	RawData string `db:"raw_data" json:"-"`
}

// TableName 返回表名。
func (m *MemberSocialUserModel) TableName() string {
	return MemberSocialUserTable
}
