package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000018CreateInfraDistrictTable 建 infra_district 行政区划表。
// 省 / 市 / 区县 / 街道共用一张表：pid 自引用成树，area_code 与 parent_area_code 保存行政编码。
type M20261022000018CreateInfraDistrictTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000018CreateInfraDistrictTable) Signature() string {
	return "20261022000018_create_infra_district_table"
}

// Up 建 infra_district 表。
// 行政编码用于与第三方数据对齐，建普通索引不建唯一键：历史数据可能存在同码不同名的情况。
func (r *M20261022000018CreateInfraDistrictTable) Up() error {
	if facades.Schema().HasTable("infra_district") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE infra_district (
  id               bigint      NOT NULL COMMENT '主键',
  pid              bigint      NOT NULL DEFAULT 0 COMMENT '上级区划 ID，0 表示顶级',
  area_name        varchar(50) NOT NULL COMMENT '名称',
  short_name       varchar(50) NOT NULL COMMENT '简称',
  area_code        varchar(50) NOT NULL COMMENT '行政编码',
  parent_area_code varchar(50) NOT NULL COMMENT '父级行政编码',
  level            int         NOT NULL COMMENT '行政级别：1 省 / 2 市 / 3 区县 / 4 街道',
  sort             int         NOT NULL DEFAULT 1 COMMENT '排序',
  creator          varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at       datetime    NULL COMMENT '创建时间',
  updater          varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at       datetime    NULL COMMENT '更新时间',
  deleted_at       datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id) USING BTREE,
  KEY pid (pid) USING BTREE,
  KEY code (area_code) USING BTREE,
  KEY pcode (parent_area_code) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC COMMENT='行政区划表'`)
}

// Down 删除 infra_district 表。
func (r *M20261022000018CreateInfraDistrictTable) Down() error {
	return facades.Schema().DropIfExists("infra_district")
}
