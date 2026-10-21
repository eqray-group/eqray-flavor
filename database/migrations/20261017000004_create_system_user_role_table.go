package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261017000004CreateSystemUserRoleTable struct{}

// Signature The unique signature for the migration.
func (r *M20261017000004CreateSystemUserRoleTable) Signature() string {
	return "20261017000004_create_system_user_role_table"
}

// Up 建 system_user_role 用户角色关联表（docs 6.3）。
// 关联表直接硬删除，唯一键建在业务键上（docs 6.8）。
// 初始授权数据由 20261017000010_seed_system_user_role 负责，建表与数据填充分离。
func (r *M20261017000004CreateSystemUserRoleTable) Up() error {
	if facades.Schema().HasTable("system_user_role") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE system_user_role (
  id           bigint   NOT NULL COMMENT '主键',
  user_id      bigint   NOT NULL COMMENT '用户 ID',
  role_id      bigint   NOT NULL COMMENT '角色 ID',
  creator      varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime NULL COMMENT '创建时间',
  updater      varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime NULL COMMENT '更新时间',
  deleted_at   datetime NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_role (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表'`)
}

// Down Reverse the migrations.
func (r *M20261017000004CreateSystemUserRoleTable) Down() error {
	return facades.Schema().DropIfExists("system_user_role")
}
