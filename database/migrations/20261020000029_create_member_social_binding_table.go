package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000029CreateMemberSocialBindingTable 建 member_social_binding 会员第三方绑定表（对齐 yudao）。
type M20261020000029CreateMemberSocialBindingTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000029CreateMemberSocialBindingTable) Signature() string {
	return "20261020000029_create_member_social_binding_table"
}

// Up 建 member_social_binding 表。
// 一个会员在同一渠道只能绑定一个账号，一个社交账号也只能绑定一个会员；
// 解绑走物理删除，否则唯一键会挡住重新绑定。
func (r *M20261020000029CreateMemberSocialBindingTable) Up() error {
	if facades.Schema().HasTable("member_social_binding") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_social_binding (
  id             bigint      NOT NULL COMMENT '主键',
  user_id        bigint      NOT NULL COMMENT '会员 ID',
  social_type    varchar(32) NOT NULL COMMENT '渠道',
  social_user_id bigint      NOT NULL COMMENT '会员第三方账号 ID',
  creator        varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at     datetime    NULL COMMENT '创建时间',
  updater        varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at     datetime    NULL COMMENT '更新时间',
  deleted_at     datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_social_user (social_user_id),
  UNIQUE KEY uk_user_type (user_id, social_type),
  KEY idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员第三方账号绑定表'`)
}

// Down Reverse the migrations.
func (r *M20261020000029CreateMemberSocialBindingTable) Down() error {
	return facades.Schema().DropIfExists("member_social_binding")
}
