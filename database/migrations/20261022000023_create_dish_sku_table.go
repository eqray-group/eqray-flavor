package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000023CreateDishSkuTable 建 dish_sku 规格（SKU）表。
// 一个 SPU 下可有多个 SKU，同一 SPU 内规格名唯一；金额列一律存分。
type M20261022000023CreateDishSkuTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000023CreateDishSkuTable) Signature() string {
	return "20261022000023_create_dish_sku_table"
}

// Up 建 dish_sku 表。
func (r *M20261022000023CreateDishSkuTable) Up() error {
	if facades.Schema().HasTable("dish_sku") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE dish_sku (
  id               bigint      NOT NULL COMMENT '主键',
  spu_id           bigint      NOT NULL COMMENT '所属菜品 SPU ID',
  name             varchar(64) NOT NULL COMMENT '规格名称，如大份、中杯；同一 SPU 内唯一',
  code             varchar(64) NOT NULL DEFAULT '' COMMENT '规格编码，非空时全局唯一，导入匹配用',
  standard_price   bigint      NOT NULL DEFAULT 0 COMMENT '品牌标准价（分），必须大于 0',
  member_price     bigint      NOT NULL DEFAULT 0 COMMENT '品牌会员价（分），0 表示未设置',
  cost_price       bigint      NOT NULL DEFAULT 0 COMMENT '成本价（分），允许为 0',
  status           tinyint     NOT NULL DEFAULT 1 COMMENT '上下架状态：1 上架 / 0 下架',
  sort             int         NOT NULL DEFAULT 0 COMMENT 'SPU 内规格排序，越小越靠前',
  remark           varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator          varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at       datetime    NULL COMMENT '创建时间',
  updater          varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at       datetime    NULL COMMENT '更新时间',
  deleted_at       datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_spu_name (spu_id, name),
  KEY idx_spu_status_sort (spu_id, status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜品规格表（SKU，品牌级）'`)
}

// Down 删除 dish_sku 表。
func (r *M20261022000023CreateDishSkuTable) Down() error {
	return facades.Schema().DropIfExists("dish_sku")
}
