package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261020000007CreateInfraFileConfigTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000007CreateInfraFileConfigTable) Signature() string {
	return "20261020000007_create_infra_file_config_table"
}

// Up 建 infra_file_config 文件配置表（docs 6.4）。
// master 由业务层保证至多一个，数据库不加唯一键（docs 6.8(6)）。
func (r *M20261020000007CreateInfraFileConfigTable) Up() error {
	if facades.Schema().HasTable("infra_file_config") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE infra_file_config (
  id         bigint      NOT NULL COMMENT '主键',
  name       varchar(64) NOT NULL COMMENT '配置名称',
  storage    varchar(32) NOT NULL DEFAULT 'local' COMMENT '存储类型：local 本地磁盘',
  master     tinyint     NOT NULL DEFAULT 0 COMMENT '是否主配置：0 否 / 1 是',
  config     text        NULL COMMENT '存储配置 JSON',
  creator    varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime    NULL COMMENT '创建时间',
  updater    varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime    NULL COMMENT '更新时间',
  deleted_at datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_master (master)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件配置表'`)
}

// Down Reverse the migrations.
func (r *M20261020000007CreateInfraFileConfigTable) Down() error {
	return facades.Schema().DropIfExists("infra_file_config")
}
