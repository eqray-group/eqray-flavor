package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261020000004CreateInfraConfigTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000004CreateInfraConfigTable) Signature() string {
	return "20261020000004_create_infra_config_table"
}

// Up 建 infra_config 参数配置表（docs 6.4）。
// config_key 建唯一键：键由代码常量引用，改名等于换键。
func (r *M20261020000004CreateInfraConfigTable) Up() error {
	if facades.Schema().HasTable("infra_config") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE infra_config (
  id         bigint       NOT NULL COMMENT '主键',
  category   varchar(64)  NOT NULL DEFAULT '' COMMENT '参数分组',
  name       varchar(128) NOT NULL COMMENT '参数名称',
  config_key varchar(128) NOT NULL COMMENT '参数键，唯一',
  value      varchar(512) NOT NULL DEFAULT '' COMMENT '参数值',
  type       tinyint      NOT NULL DEFAULT 2 COMMENT '类型：1 系统内置 / 2 自定义',
  visible    tinyint      NOT NULL DEFAULT 0 COMMENT '是否对外公开：0 否 / 1 是',
  remark     varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_config_key (config_key),
  KEY idx_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='参数配置表'`)
}

// Down Reverse the migrations.
func (r *M20261020000004CreateInfraConfigTable) Down() error {
	return facades.Schema().DropIfExists("infra_config")
}
