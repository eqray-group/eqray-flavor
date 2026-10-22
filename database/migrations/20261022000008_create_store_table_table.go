package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000008CreateStoreTableTable 建 store_table 桌台表。
// 桌码全局唯一，桌号在同一门店内唯一；记录当前台位单与开台时间。
type M20261022000008CreateStoreTableTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000008CreateStoreTableTable) Signature() string {
	return "20261022000008_create_store_table_table"
}

// Up 建 store_table 表。
func (r *M20261022000008CreateStoreTableTable) Up() error {
	if facades.Schema().HasTable("store_table") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE store_table (
  id               bigint       NOT NULL COMMENT '主键',
  store_id         bigint       NOT NULL COMMENT '门店 ID，只能指向 type=3 的门店节点',
  area_id          bigint       NOT NULL COMMENT '所属区域 ID，必须与 store_id 属于同一门店',
  name             varchar(32)  NOT NULL COMMENT '桌号，如 A01、包间 1；同一门店内唯一',
  code             varchar(64)  NOT NULL COMMENT '桌码，全局唯一；默认按门店编码-桌号生成，创建后不可修改',
  seats            int          NOT NULL DEFAULT 4 COMMENT '座位数',
  sort             int          NOT NULL DEFAULT 0 COMMENT '区域内排序，越小越靠前',
  status           tinyint      NOT NULL DEFAULT 1 COMMENT '桌台状态：1 空闲 / 2 使用中 / 3 待清台',
  current_order_id bigint       NOT NULL DEFAULT 0 COMMENT '当前台位单 ID，0 表示没有进行中的台位单',
  opened_at        datetime     NULL COMMENT '开台时间，清台后置空',
  remark           varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator          varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at       datetime     NULL COMMENT '创建时间',
  updater          varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at       datetime     NULL COMMENT '更新时间',
  deleted_at       datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code),
  UNIQUE KEY uk_store_name (store_id, name),
  KEY idx_store_area_status (store_id, area_id, status),
  KEY idx_area (area_id),
  KEY idx_current_order (current_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='桌台表（门店内座位单元）'`)
}

// Down 删除 store_table 表。
func (r *M20261022000008CreateStoreTableTable) Down() error {
	return facades.Schema().DropIfExists("store_table")
}
