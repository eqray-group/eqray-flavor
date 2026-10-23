package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000026CreateDishSpuAddonTable 建 dish_spu_addon 菜品加料关联表。
// 只记录 SPU 挂了哪些加料组，(spu_id, addon_group_id) 唯一，避免重复挂载。
type M20261022000026CreateDishSpuAddonTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000026CreateDishSpuAddonTable) Signature() string {
	return "20261022000026_create_dish_spu_addon_table"
}

// Up 建 dish_spu_addon 表。
func (r *M20261022000026CreateDishSpuAddonTable) Up() error {
	if facades.Schema().HasTable("dish_spu_addon") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE dish_spu_addon (
  id             bigint       NOT NULL COMMENT '主键',
  spu_id         bigint       NOT NULL COMMENT '菜品 SPU ID',
  addon_group_id bigint       NOT NULL COMMENT '加料组 ID',
  sort           int          NOT NULL DEFAULT 0 COMMENT '菜品内加料组排序，越小越靠前',
  remark         varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator        varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at     datetime     NULL COMMENT '创建时间',
  updater        varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at     datetime     NULL COMMENT '更新时间',
  deleted_at     datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_spu_group (spu_id, addon_group_id),
  KEY idx_addon_group (addon_group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜品加料关联表（SPU 挂载加料组）'`)
}

// Down 删除 dish_spu_addon 表。
func (r *M20261022000026CreateDishSpuAddonTable) Down() error {
	return facades.Schema().DropIfExists("dish_spu_addon")
}
