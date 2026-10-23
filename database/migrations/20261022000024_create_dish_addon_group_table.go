package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000024CreateDishAddonGroupTable 建 dish_addon_group 加料组表。
// 加料组是品牌级可复用配置，通过 dish_spu_addon 挂到具体菜品；名称全局唯一。
type M20261022000024CreateDishAddonGroupTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000024CreateDishAddonGroupTable) Signature() string {
	return "20261022000024_create_dish_addon_group_table"
}

// Up 建 dish_addon_group 表。
func (r *M20261022000024CreateDishAddonGroupTable) Up() error {
	if facades.Schema().HasTable("dish_addon_group") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE dish_addon_group (
  id         bigint       NOT NULL COMMENT '主键',
  name       varchar(64)  NOT NULL COMMENT '加料组名称，全局唯一，如加料、甜度、温度',
  required   tinyint      NOT NULL DEFAULT 0 COMMENT '是否必选：1 必选 / 0 可选',
  min_select int          NOT NULL DEFAULT 0 COMMENT '最少选择数量，必选时至少为 1',
  max_select int          NOT NULL DEFAULT 1 COMMENT '最多选择数量，0 表示不限',
  sort       int          NOT NULL DEFAULT 0 COMMENT '排序，越小越靠前',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：1 启用 / 0 禁用',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_name (name),
  KEY idx_status_sort (status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜品加料组表（品牌级）'`)
}

// Down 删除 dish_addon_group 表。
func (r *M20261022000024CreateDishAddonGroupTable) Down() error {
	return facades.Schema().DropIfExists("dish_addon_group")
}
