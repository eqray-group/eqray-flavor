package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000023CreateMemberTagTable 建 member_tag 会员标签表（对齐 yudao）。
type M20261020000023CreateMemberTagTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000023CreateMemberTagTable) Signature() string {
	return "20261020000023_create_member_tag_table"
}

// Up 建 member_tag 表。
// 会员与标签是多对多，引用关系存在 member_user.tag_ids，不再单独建关联表。
func (r *M20261020000023CreateMemberTagTable) Up() error {
	if facades.Schema().HasTable("member_tag") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_tag (
  id         bigint       NOT NULL COMMENT '主键',
  name       varchar(30)  NOT NULL COMMENT '标签名',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员标签表'`)
}

// Down Reverse the migrations.
func (r *M20261020000023CreateMemberTagTable) Down() error {
	return facades.Schema().DropIfExists("member_tag")
}
