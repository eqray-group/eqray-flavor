package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000028CreateDishStorePriceTable 建 dish_store_price 门店价格表。
// 门店价覆盖品牌价，(store_id, sku_id) 唯一；start_at / end_at 构成生效区间，为空表示长期有效。
type M20261022000028CreateDishStorePriceTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000028CreateDishStorePriceTable) Signature() string {
	return "20261022000028_create_dish_store_price_table"
}

// Up 建 dish_store_price 表。
func (r *M20261022000028CreateDishStorePriceTable) Up() error {
	if facades.Schema().HasTable("dish_store_price") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE dish_store_price (
  id                  bigint       NOT NULL COMMENT '主键',
  store_id            bigint       NOT NULL COMMENT '门店 ID，只能指向 type=3 的门店节点',
  sku_id              bigint       NOT NULL COMMENT '菜品规格 SKU ID',
  store_price         bigint       NOT NULL DEFAULT 0 COMMENT '门店价（分），0 表示不覆盖品牌标准价',
  store_member_price  bigint       NOT NULL DEFAULT 0 COMMENT '门店会员价（分），0 表示不覆盖品牌会员价',
  start_at            datetime     NULL COMMENT '生效开始时间，空表示立即生效',
  end_at              datetime     NULL COMMENT '生效结束时间，空表示长期有效',
  remark              varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator             varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at          datetime     NULL COMMENT '创建时间',
  updater             varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at          datetime     NULL COMMENT '更新时间',
  deleted_at          datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_sku (store_id, sku_id),
  KEY idx_sku (sku_id),
  KEY idx_store_valid (store_id, start_at, end_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='门店价格表（门店价覆盖）'`)
}

// Down 删除 dish_store_price 表。
func (r *M20261022000028CreateDishStorePriceTable) Down() error {
	return facades.Schema().DropIfExists("dish_store_price")
}
