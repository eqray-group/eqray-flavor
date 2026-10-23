package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000027CreateDishStoreSkuTable 建 dish_store_sku 门店可售表。
// 门店级配置，(store_id, sku_id) 唯一；可售、沽清与仅线上三个开关共同决定顾客端菜单。
type M20261022000027CreateDishStoreSkuTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000027CreateDishStoreSkuTable) Signature() string {
	return "20261022000027_create_dish_store_sku_table"
}

// Up 建 dish_store_sku 表。
func (r *M20261022000027CreateDishStoreSkuTable) Up() error {
	if facades.Schema().HasTable("dish_store_sku") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE dish_store_sku (
  id         bigint       NOT NULL COMMENT '主键',
  store_id   bigint       NOT NULL COMMENT '门店 ID，只能指向 type=3 的门店节点',
  sku_id     bigint       NOT NULL COMMENT '菜品规格 SKU ID',
  available  tinyint      NOT NULL DEFAULT 1 COMMENT '门店是否可售：1 可售 / 0 下架',
  sold_out   tinyint      NOT NULL DEFAULT 0 COMMENT '沽清标记：1 已沽清 / 0 正常',
  online_only tinyint     NOT NULL DEFAULT 0 COMMENT '是否仅线上可售：1 仅线上 / 0 不限',
  sort       int          NOT NULL DEFAULT 0 COMMENT '门店内排序，越小越靠前',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_sku (store_id, sku_id),
  KEY idx_sku (sku_id),
  KEY idx_store_available (store_id, available, sold_out)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='门店可售表（门店 + SKU）'`)
}

// Down 删除 dish_store_sku 表。
func (r *M20261022000027CreateDishStoreSkuTable) Down() error {
	return facades.Schema().DropIfExists("dish_store_sku")
}
