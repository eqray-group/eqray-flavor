package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261020000001CreateInfraFileTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000001CreateInfraFileTable) Signature() string {
	return "20261020000001_create_infra_file_table"
}

// Up 建 infra_file 文件表（docs 6.5）。
// config_id 预留给多存储配置，当前统一使用 config/filesystems.go 里的 public 磁盘，落 0。
func (r *M20261020000001CreateInfraFileTable) Up() error {
	if facades.Schema().HasTable("infra_file") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE infra_file (
  id         bigint        NOT NULL COMMENT '主键',
  config_id  bigint        NOT NULL DEFAULT 0 COMMENT '文件配置 ID，0 表示默认配置',
  name       varchar(255)  NOT NULL COMMENT '原文件名',
  path       varchar(512)  NOT NULL COMMENT '存储路径',
  url        varchar(1024) NOT NULL DEFAULT '' COMMENT '访问地址',
  type       varchar(64)   NOT NULL DEFAULT '' COMMENT '文件 MIME 类型',
  size       bigint        NOT NULL DEFAULT 0 COMMENT '文件大小（字节）',
  creator    varchar(64)   NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime      NULL COMMENT '创建时间',
  updater    varchar(64)   NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime      NULL COMMENT '更新时间',
  deleted_at datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_config_id (config_id),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表'`)
}

// Down Reverse the migrations.
func (r *M20261020000001CreateInfraFileTable) Down() error {
	return facades.Schema().DropIfExists("infra_file")
}
