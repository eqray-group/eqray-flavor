package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261020000034CreateMemberSignInConfigTable 建 member_sign_in_config 签到奖励档位表（对齐 yudao）。
// 连续签到第 N 天对应一条档位，签到成功时按档位发放积分与经验。
type M20261020000034CreateMemberSignInConfigTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000034CreateMemberSignInConfigTable) Signature() string {
	return "20261020000034_create_member_sign_in_config_table"
}

// Up 建 member_sign_in_config 表。
// day 不加唯一键：档位允许软删后按同一天重新创建，唯一性由 Service 校验。
func (r *M20261020000034CreateMemberSignInConfigTable) Up() error {
	if facades.Schema().HasTable("member_sign_in_config") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE member_sign_in_config (
  id         bigint      NOT NULL COMMENT '主键',
  day        int         NOT NULL COMMENT '连续签到第几天',
  point      int         NOT NULL DEFAULT 0 COMMENT '该档位奖励积分',
  experience int         NOT NULL DEFAULT 0 COMMENT '该档位奖励经验',
  status     tinyint     NOT NULL DEFAULT 1 COMMENT '状态：0 关闭 / 1 开启',
  creator    varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime    NULL COMMENT '创建时间',
  updater    varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime    NULL COMMENT '更新时间',
  deleted_at datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_day (day)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员签到奖励档位表'`)
}

// Down 删除签到档位表。
func (r *M20261020000034CreateMemberSignInConfigTable) Down() error {
	return facades.Schema().DropIfExists("member_sign_in_config")
}
