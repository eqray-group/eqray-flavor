package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000020CreateMemberLevelRecordTable 建 member_level_record 会员等级记录表（对齐 yudao）。
type M20261020000020CreateMemberLevelRecordTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000020CreateMemberLevelRecordTable) Signature() string {
	return "20261020000020_create_member_level_record_table"
}

// Up 建 member_level_record 表。
// 等级调整时由 Service 在事务内写入，用于追溯会员的等级变更历史。
func (r *M20261020000020CreateMemberLevelRecordTable) Up() error {
	if facades.Schema().HasTable("member_level_record") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_level_record (
  id          bigint       NOT NULL COMMENT '主键',
  user_id     bigint       NOT NULL COMMENT '会员 ID',
  level_id    bigint       NOT NULL DEFAULT 0 COMMENT '变更后的等级 ID',
  experience  int          NOT NULL DEFAULT 0 COMMENT '变更后的经验值',
  title       varchar(64)  NOT NULL DEFAULT '' COMMENT '记录标题',
  description varchar(255) NOT NULL DEFAULT '' COMMENT '记录描述',
  creator     varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at  datetime     NULL COMMENT '创建时间',
  updater     varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at  datetime     NULL COMMENT '更新时间',
  deleted_at  datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_user (user_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员等级记录表'`)
}

// Down Reverse the migrations.
func (r *M20261020000020CreateMemberLevelRecordTable) Down() error {
	return facades.Schema().DropIfExists("member_level_record")
}
