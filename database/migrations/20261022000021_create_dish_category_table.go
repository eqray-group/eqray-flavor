package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000021CreateDishCategoryTable 建 dish_category 菜品分类表。
// 分类是品牌级树结构，同一上级下名称唯一；删除采用物理删除，删除前由业务层校验下级分类与菜品引用。
type M20261022000021CreateDishCategoryTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000021CreateDishCategoryTable) Signature() string {
	return "20261022000021_create_dish_category_table"
}

// Up 建 dish_category 表。
func (r *M20261022000021CreateDishCategoryTable) Up() error {
	if facades.Schema().HasTable("dish_category") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE dish_category (
  id         bigint       NOT NULL COMMENT '主键',
  parent_id  bigint       NOT NULL DEFAULT 0 COMMENT '上级分类 ID，0 表示根分类',
  name       varchar(64)  NOT NULL COMMENT '分类名称，同一上级下唯一',
  image      varchar(255) NOT NULL DEFAULT '' COMMENT '分类图标地址',
  sort       int          NOT NULL DEFAULT 0 COMMENT '排序，越小越靠前',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：1 启用 / 0 禁用',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  deleted_at datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_parent_name (parent_id, name),
  KEY idx_parent_status_sort (parent_id, status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜品分类表（品牌级树结构）'`)
}

// Down 删除 dish_category 表。
func (r *M20261022000021CreateDishCategoryTable) Down() error {
	return facades.Schema().DropIfExists("dish_category")
}
