package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000028CreateMemberSocialUserTable 建 member_social_user 会员第三方账号表（对齐 yudao）。
type M20261020000028CreateMemberSocialUserTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000028CreateMemberSocialUserTable) Signature() string {
	return "20261020000028_create_member_social_user_table"
}

// Up 建 member_social_user 表。
// 保存第三方平台返回的身份信息，与会员档案通过 member_social_binding 关联。
func (r *M20261020000028CreateMemberSocialUserTable) Up() error {
	if facades.Schema().HasTable("member_social_user") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_social_user (
  id         bigint       NOT NULL COMMENT '主键',
  type       varchar(32)  NOT NULL COMMENT '渠道：wechat_mini / wechat_mp / wechat_open',
  openid     varchar(128) NOT NULL COMMENT '渠道内用户标识',
  unionid    varchar(128) NOT NULL DEFAULT '' COMMENT '开放平台用户标识',
  nickname   varchar(64)  NOT NULL DEFAULT '' COMMENT '昵称',
  avatar     varchar(512) NOT NULL DEFAULT '' COMMENT '头像地址',
  raw_data   text         NULL COMMENT '第三方原始报文',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_type_openid (type, openid),
  KEY idx_unionid (unionid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员第三方账号表'`)
}

// Down Reverse the migrations.
func (r *M20261020000028CreateMemberSocialUserTable) Down() error {
	return facades.Schema().DropIfExists("member_social_user")
}
