package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000025CreateDishAddonItemTable 建 dish_addon_item 加料项表。
// 加料项归属于加料组，组内名称唯一；加价存分，允许为 0（如免费加冰）。
type M20261022000025CreateDishAddonItemTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000025CreateDishAddonItemTable) Signature() string {
	return "20261022000025_create_dish_addon_item_table"
}

// Up 建 dish_addon_item 表。
func (r *M20261022000025CreateDishAddonItemTable) Up() error {
	if facades.Schema().HasTable("dish_addon_item") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE dish_addon_item (
  id         bigint       NOT NULL COMMENT '主键',
  group_id   bigint       NOT NULL COMMENT '所属加料组 ID',
  name       varchar(64)  NOT NULL COMMENT '加料项名称，组内唯一，如加珍珠、少冰',
  price      bigint       NOT NULL DEFAULT 0 COMMENT '加价（分），0 表示免费',
  sort       int          NOT NULL DEFAULT 0 COMMENT '组内排序，越小越靠前',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：1 启用 / 0 禁用',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_group_name (group_id, name),
  KEY idx_group_status_sort (group_id, status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜品加料项表（品牌级）'`)
}

// Down 删除 dish_addon_item 表。
func (r *M20261022000025CreateDishAddonItemTable) Down() error {
	return facades.Schema().DropIfExists("dish_addon_item")
}
