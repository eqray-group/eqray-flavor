package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261017000003CreateSystemRoleTable struct{}

// Signature The unique signature for the migration.
func (r *M20261017000003CreateSystemRoleTable) Signature() string {
	return "20261017000003_create_system_role_table"
}

// Up 建 system_role 角色表（docs 6.3）。
// 角色编码唯一；角色软删除后编码仍被占用，避免与历史数据混淆。
// 内置角色数据由 20261017000008_seed_system_role 负责，建表与数据填充分离。
func (r *M20261017000003CreateSystemRoleTable) Up() error {
	if facades.Schema().HasTable("system_role") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE system_role (
  id           bigint       NOT NULL COMMENT '主键',
  name         varchar(64)  NOT NULL COMMENT '角色名称',
  code         varchar(64)  NOT NULL COMMENT '角色标识，唯一，如 SUPER_ADMIN',
  sort         int          NOT NULL DEFAULT 0 COMMENT '排序',
  status       tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  type         tinyint      NOT NULL DEFAULT 2 COMMENT '类型：1 内置角色 / 2 自定义角色',
  remark       varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  creator      varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime     NULL COMMENT '创建时间',
  updater      varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime     NULL COMMENT '更新时间',
  deleted_at   datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表'`)
}

// Down Reverse the migrations.
func (r *M20261017000003CreateSystemRoleTable) Down() error {
	return facades.Schema().DropIfExists("system_role")
}
