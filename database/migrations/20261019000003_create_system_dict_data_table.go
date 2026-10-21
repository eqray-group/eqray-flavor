package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261019000003CreateSystemDictDataTable struct{}

// Signature The unique signature for the migration.
func (r *M20261019000003CreateSystemDictDataTable) Signature() string {
	return "20261019000003_create_system_dict_data_table"
}

// Up 建 system_dict_data 字典数据表（docs 6.3）。
func (r *M20261019000003CreateSystemDictDataTable) Up() error {
	if facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE system_dict_data (
  id         bigint       NOT NULL COMMENT '主键',
  sort       int          NOT NULL DEFAULT 0 COMMENT '排序',
  label      varchar(64)  NOT NULL COMMENT '显示文本',
  value      varchar(64)  NOT NULL COMMENT '实际值',
  dict_type  varchar(64)  NOT NULL COMMENT '所属字典类型编码',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  color_type varchar(32)  NOT NULL DEFAULT 'default' COMMENT '标签样式',
  css_class  varchar(64)  NOT NULL DEFAULT '' COMMENT '自定义样式类',
  remark     varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_type_value (dict_type, value),
  KEY idx_dict_type (dict_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典数据表'`)
}

// Down Reverse the migrations.
func (r *M20261019000003CreateSystemDictDataTable) Down() error {
	return facades.Schema().DropIfExists("system_dict_data")
}
