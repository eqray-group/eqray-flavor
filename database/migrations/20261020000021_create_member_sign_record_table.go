package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000021CreateMemberSignRecordTable 建 member_sign_record 会员签到记录表（对齐 yudao）。
type M20261020000021CreateMemberSignRecordTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000021CreateMemberSignRecordTable) Signature() string {
	return "20261020000021_create_member_sign_record_table"
}

// Up 建 member_sign_record 表。
// 同一会员同一天只能签到一次，用唯一键保证；删除记录不应等于可以再签一次。
func (r *M20261020000021CreateMemberSignRecordTable) Up() error {
	if facades.Schema().HasTable("member_sign_record") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_sign_record (
  id         bigint      NOT NULL COMMENT '主键',
  user_id    bigint      NOT NULL COMMENT '会员 ID',
  sign_date  date        NOT NULL COMMENT '签到日期',
  day        int         NOT NULL DEFAULT 0 COMMENT '连续签到第几天',
  point      int         NOT NULL DEFAULT 0 COMMENT '签到获得积分',
  experience int         NOT NULL DEFAULT 0 COMMENT '本次获得经验',
  creator    varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime    NULL COMMENT '创建时间',
  updater    varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime    NULL COMMENT '更新时间',
  deleted_at datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_date (user_id, sign_date),
  KEY idx_user (user_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员签到记录表'`)
}

// Down Reverse the migrations.
func (r *M20261020000021CreateMemberSignRecordTable) Down() error {
	return facades.Schema().DropIfExists("member_sign_record")
}
