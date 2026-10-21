package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261019000002CreateSystemDictTypeTable struct{}

// Signature The unique signature for the migration.
func (r *M20261019000002CreateSystemDictTypeTable) Signature() string {
	return "20261019000002_create_system_dict_type_table"
}

// Up 建 system_dict_type 字典类型表（docs 6.3）。
func (r *M20261019000002CreateSystemDictTypeTable) Up() error {
	if facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE system_dict_type (
  id         bigint       NOT NULL COMMENT '主键',
  name       varchar(64)  NOT NULL COMMENT '字典名称',
  type       varchar(64)  NOT NULL COMMENT '字典类型编码，唯一',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  remark     varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典类型表'`)
}

// Down Reverse the migrations.
func (r *M20261019000002CreateSystemDictTypeTable) Down() error {
	return facades.Schema().DropIfExists("system_dict_type")
}
