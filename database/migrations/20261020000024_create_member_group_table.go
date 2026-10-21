package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000024CreateMemberGroupTable 建 member_group 会员分组表（对齐 yudao）。
type M20261020000024CreateMemberGroupTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000024CreateMemberGroupTable) Signature() string {
	return "20261020000024_create_member_group_table"
}

// Up 建 member_group 表。
// 分组用于会员分群运营，会员通过 member_user.group_id 引用。
func (r *M20261020000024CreateMemberGroupTable) Up() error {
	if facades.Schema().HasTable("member_group") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_group (
  id         bigint       NOT NULL COMMENT '主键',
  name       varchar(30)  NOT NULL COMMENT '分组名',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员分组表'`)
}

// Down Reverse the migrations.
func (r *M20261020000024CreateMemberGroupTable) Down() error {
	return facades.Schema().DropIfExists("member_group")
}
