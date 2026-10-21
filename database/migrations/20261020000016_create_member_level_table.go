package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000016CreateMemberLevelTable 建 member_level 会员等级表（对齐 yudao）。
// 历史上先建 member_levels 旧结构（growth / discount），再由 20261020000020 重命名并改列；
// 整理后一次建出最终结构。
type M20261020000016CreateMemberLevelTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000016CreateMemberLevelTable) Signature() string {
	return "20261020000016_create_member_level_table"
}

// Up 建 member_level 表。
// level 是会员档案引用等级的业务键，建唯一键保证不会出现重复等级。
func (r *M20261020000016CreateMemberLevelTable) Up() error {
	if facades.Schema().HasTable("member_level") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_level (
  id               bigint       NOT NULL COMMENT '主键',
  level            int          NOT NULL COMMENT '等级序号：1-99，唯一',
  name             varchar(64)  NOT NULL COMMENT '等级名称',
  icon             varchar(64)  NOT NULL DEFAULT '' COMMENT '图标名',
  experience       int          NOT NULL DEFAULT 0 COMMENT '升级所需经验值',
  discount_percent int          NOT NULL DEFAULT 100 COMMENT '折扣百分比：100 无折扣 / 95 表示 95 折',
  description      varchar(500) NOT NULL DEFAULT '' COMMENT '等级说明',
  status           tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  sort             int          NOT NULL DEFAULT 0 COMMENT '排序，越小越靠前',
  creator          varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at       datetime     NULL COMMENT '创建时间',
  updater          varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at       datetime     NULL COMMENT '更新时间',
  deleted_at       datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_level (level),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员等级表'`)
}

// Down Reverse the migrations.
func (r *M20261020000016CreateMemberLevelTable) Down() error {
	return facades.Schema().DropIfExists("member_level")
}
