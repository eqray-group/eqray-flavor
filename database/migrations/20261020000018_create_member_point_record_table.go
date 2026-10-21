package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000018CreateMemberPointRecordTable 建 member_point_record 积分流水表（对齐 yudao）。
// 历史上先建 member_point_logs 旧结构（member_id / type / points / balance / source / remark），
// 再由 20261020000020 重命名并改列；整理后一次建出最终结构。
type M20261020000018CreateMemberPointRecordTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000018CreateMemberPointRecordTable) Signature() string {
	return "20261020000018_create_member_point_record_table"
}

// Up 建 member_point_record 表。
// 只增不改，用于对账；point 用正负表示增减，业务类型放在 biz_type。
func (r *M20261020000018CreateMemberPointRecordTable) Up() error {
	if facades.Schema().HasTable("member_point_record") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_point_record (
  id          bigint       NOT NULL COMMENT '主键',
  user_id     bigint       NOT NULL COMMENT '会员 ID',
  biz_type    tinyint      NOT NULL DEFAULT 0 COMMENT '业务类型：0 其它 / 10 后台调整',
  biz_id      bigint       NOT NULL DEFAULT 0 COMMENT '业务编号',
  point       int          NOT NULL DEFAULT 0 COMMENT '积分，正数增加 / 负数扣减',
  total_point int          NOT NULL DEFAULT 0 COMMENT '变动后的积分',
  title       varchar(64)  NOT NULL DEFAULT '' COMMENT '流水标题',
  description varchar(255) NOT NULL DEFAULT '' COMMENT '流水描述',
  creator     varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at  datetime     NULL COMMENT '创建时间',
  updater     varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at  datetime     NULL COMMENT '更新时间',
  deleted_at  datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_member (user_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员积分流水表'`)
}

// Down Reverse the migrations.
func (r *M20261020000018CreateMemberPointRecordTable) Down() error {
	return facades.Schema().DropIfExists("member_point_record")
}
