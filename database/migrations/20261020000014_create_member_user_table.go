package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000014CreateMemberUserTable 建 member_user 会员档案表（对齐 yudao）。
// 历史上先建 member_users 旧结构，再由 20261020000020 重命名并补列；
// 整理后一次建出最终结构，旧迁移与补迁移已删除。
type M20261020000014CreateMemberUserTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000014CreateMemberUserTable) Signature() string {
	return "20261020000014_create_member_user_table"
}

// Up 建 member_user 表。
// 手机号建唯一键：会员身份以手机号为准，软删除后号码不复用。
func (r *M20261020000014CreateMemberUserTable) Up() error {
	if facades.Schema().HasTable("member_user") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_user (
  id           bigint       NOT NULL COMMENT '主键',
  nickname     varchar(64)  NOT NULL COMMENT '昵称',
  mobile       varchar(20)  NOT NULL COMMENT '手机号，唯一',
  password     varchar(100) NOT NULL DEFAULT '' COMMENT '密码',
  avatar       varchar(512) NOT NULL DEFAULT '' COMMENT '头像地址',
  name         varchar(30)  NOT NULL DEFAULT '' COMMENT '真实姓名',
  sex          tinyint      NOT NULL DEFAULT 0 COMMENT '性别：0 未知 / 1 男 / 2 女',
  birthday     datetime     NULL COMMENT '出生日期',
  area_id      int          NOT NULL DEFAULT 0 COMMENT '所在地',
  level_id     bigint       NOT NULL DEFAULT 0 COMMENT '会员等级 ID',
  point        int          NOT NULL DEFAULT 0 COMMENT '积分',
  experience   int          NOT NULL DEFAULT 0 COMMENT '经验值',
  tag_ids      varchar(255) NOT NULL DEFAULT '' COMMENT '会员标签 ID，逗号分隔',
  mark         varchar(255) NOT NULL DEFAULT '' COMMENT '用户标记',
  register_ip  varchar(32)  NOT NULL DEFAULT '' COMMENT '注册 IP',
  login_ip     varchar(32)  NOT NULL DEFAULT '' COMMENT '最后登录 IP',
  login_date   datetime     NULL COMMENT '最后登录时间',
  group_id     bigint       NOT NULL DEFAULT 0 COMMENT '会员分组 ID',
  status       tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  creator      varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime     NULL COMMENT '创建时间',
  updater      varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime     NULL COMMENT '更新时间',
  deleted_at   datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mobile (mobile),
  KEY idx_status (status),
  KEY idx_created_at (created_at),
  KEY idx_level_id (level_id),
  KEY idx_group_id (group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员档案表'`)
}

// Down Reverse the migrations.
func (r *M20261020000014CreateMemberUserTable) Down() error {
	return facades.Schema().DropIfExists("member_user")
}
