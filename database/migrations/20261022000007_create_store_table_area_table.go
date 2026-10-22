package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000007CreateStoreTableAreaTable 建 store_table_area 桌台区域表。
// 区域按门店归属，同一门店内名称唯一；删除采用物理删除。
type M20261022000007CreateStoreTableAreaTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000007CreateStoreTableAreaTable) Signature() string {
	return "20261022000007_create_store_table_area_table"
}

// Up 建 store_table_area 表。
func (r *M20261022000007CreateStoreTableAreaTable) Up() error {
	if facades.Schema().HasTable("store_table_area") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE store_table_area (
  id         bigint       NOT NULL COMMENT '主键',
  store_id   bigint       NOT NULL COMMENT '门店 ID，只能指向 type=3 的门店节点',
  name       varchar(32)  NOT NULL COMMENT '区域名称，如大厅、包间、露台',
  sort       int          NOT NULL DEFAULT 0 COMMENT '区域排序，越小越靠前',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间；关联表实际采用物理删除，保留字段与基类一致',
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_name (store_id, name),
  KEY idx_store_sort (store_id, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='桌台区域表（门店内用餐分区）'`)
}

// Down 删除 store_table_area 表。
func (r *M20261022000007CreateStoreTableAreaTable) Down() error {
	return facades.Schema().DropIfExists("store_table_area")
}
