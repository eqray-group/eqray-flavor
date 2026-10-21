package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261017000002CreateSystemMenuTable struct{}

// Signature The unique signature for the migration.
func (r *M20261017000002CreateSystemMenuTable) Signature() string {
	return "20261017000002_create_system_menu_table"
}

// Up 建 system_menu 表（docs 6.3）。
// permission 不做唯一约束：只有按钮类节点填写，唯一性由业务层保证。
func (r *M20261017000002CreateSystemMenuTable) Up() error {
	if facades.Schema().HasTable("system_menu") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE system_menu (
  id             bigint       NOT NULL COMMENT '主键',
  name           varchar(64)  NOT NULL COMMENT '菜单名称',
  permission     varchar(128) NOT NULL DEFAULT '' COMMENT '权限标识，如 system:user:create',
  type           tinyint      NOT NULL DEFAULT 1 COMMENT '类型：1 目录 / 2 菜单 / 3 按钮',
  sort           int          NOT NULL DEFAULT 0 COMMENT '排序',
  parent_id      bigint       NOT NULL DEFAULT 0 COMMENT '父菜单 ID，0 为顶级',
  path           varchar(255) NOT NULL DEFAULT '' COMMENT '前端路由地址',
  icon           varchar(64)  NOT NULL DEFAULT '' COMMENT '图标',
  component      varchar(255) NOT NULL DEFAULT '' COMMENT '前端组件路径',
  component_name varchar(64)  NOT NULL DEFAULT '' COMMENT '组件名称',
  status         tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  visible        tinyint      NOT NULL DEFAULT 1 COMMENT '是否显示：0 隐藏 / 1 显示',
  keep_alive     tinyint      NOT NULL DEFAULT 1 COMMENT '是否缓存：0 否 / 1 是',
  always_show    tinyint      NOT NULL DEFAULT 0 COMMENT '是否始终显示：0 否 / 1 是',
  creator        varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at     datetime     NULL COMMENT '创建时间',
  updater        varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at     datetime     NULL COMMENT '更新时间',
  deleted_at     datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_permission (permission),
  KEY idx_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单与权限点表'`)
}

// Down Reverse the migrations.
func (r *M20261017000002CreateSystemMenuTable) Down() error {
	return facades.Schema().DropIfExists("system_menu")
}
