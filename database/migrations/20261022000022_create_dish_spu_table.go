package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000022CreateDishSpuTable 建 dish_spu 菜品（SPU）表。
// 一个 SPU 归属一个分类并挂多个 SKU；编码允许为空，非空时由业务层保证唯一（不建唯一键避免空串互斥）。
type M20261022000022CreateDishSpuTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000022CreateDishSpuTable) Signature() string {
	return "20261022000022_create_dish_spu_table"
}

// Up 建 dish_spu 表。
func (r *M20261022000022CreateDishSpuTable) Up() error {
	if facades.Schema().HasTable("dish_spu") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE dish_spu (
  id          bigint       NOT NULL COMMENT '主键',
  category_id bigint       NOT NULL COMMENT '所属分类 ID',
  name        varchar(64)  NOT NULL COMMENT '菜品名称，同一分类下唯一',
  code        varchar(64)  NOT NULL DEFAULT '' COMMENT '菜品编码，非空时全局唯一，导入匹配用',
  unit        varchar(16)  NOT NULL DEFAULT '份' COMMENT '销售单位，如份、杯、例',
  description varchar(500) NOT NULL DEFAULT '' COMMENT '菜品介绍',
  image       varchar(255) NOT NULL DEFAULT '' COMMENT '主图地址',
  tags        varchar(255) NOT NULL DEFAULT '' COMMENT '标签，逗号分隔，如招牌、微辣',
  allergens   varchar(255) NOT NULL DEFAULT '' COMMENT '过敏原，逗号分隔',
  status      tinyint      NOT NULL DEFAULT 1 COMMENT '上下架状态：1 上架 / 0 下架',
  sort        int          NOT NULL DEFAULT 0 COMMENT '分类内排序，越小越靠前',
  remark      varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator     varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at  datetime     NULL COMMENT '创建时间',
  updater     varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at  datetime     NULL COMMENT '更新时间',
  deleted_at  datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_category_name (category_id, name),
  KEY idx_category_status_sort (category_id, status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜品表（SPU，品牌级）'`)
}

// Down 删除 dish_spu 表。
func (r *M20261022000022CreateDishSpuTable) Down() error {
	return facades.Schema().DropIfExists("dish_spu")
}
