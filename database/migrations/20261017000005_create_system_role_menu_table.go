package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261017000005CreateSystemRoleMenuTable struct{}

// Signature The unique signature for the migration.
func (r *M20261017000005CreateSystemRoleMenuTable) Signature() string {
	return "20261017000005_create_system_role_menu_table"
}

// Up 建 system_role_menu 关联表（docs 6.3）。
// 关联表直接硬删除，取消授权后可以再次授予同一菜单（docs 6.8）。
func (r *M20261017000005CreateSystemRoleMenuTable) Up() error {
	if facades.Schema().HasTable("system_role_menu") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE system_role_menu (
  id           bigint   NOT NULL COMMENT '主键',
  role_id      bigint   NOT NULL COMMENT '角色 ID',
  menu_id      bigint   NOT NULL COMMENT '菜单 ID',
  creator      varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime NULL COMMENT '创建时间',
  updater      varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime NULL COMMENT '更新时间',
  deleted_at   datetime NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_role_menu (role_id, menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表'`)
}

// Down Reverse the migrations.
func (r *M20261017000005CreateSystemRoleMenuTable) Down() error {
	return facades.Schema().DropIfExists("system_role_menu")
}
