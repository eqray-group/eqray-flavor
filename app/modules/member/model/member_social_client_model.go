package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
)

// MemberSocialClientTable 是第三方登录客户端配置表名，对齐 yudao 的 system_social_client。
const MemberSocialClientTable = "member_social_client"

// MemberSocialClientModel 对应 member_social_client 表：
// 每个渠道一条配置，登录时按 social_type 读取 client_id / client_secret。
type MemberSocialClientModel struct {
	baseModel.BaseModel
	// Name 应用名（仅后台展示用）
	Name string `db:"name" json:"name"`
	// SocialType 渠道：wechat_mini / wechat_mp / wechat_open
	SocialType string `db:"social_type" json:"socialType"`
	// ClientID 对应微信的 appid
	ClientID string `db:"client_id" json:"clientId"`
	// ClientSecret 对应微信的 appsecret
	ClientSecret string `db:"client_secret" json:"-"`
	// AgentID 预留：钉钉 / 企微等渠道的 agentId
	AgentID string                     `db:"agent_id" json:"agentId"`
	Status  baseEnums.CommonStatusEnum `db:"status" json:"status"`
}

// TableName 返回表名。
func (m *MemberSocialClientModel) TableName() string {
	return MemberSocialClientTable
}
