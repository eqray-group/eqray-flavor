package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000024CreateCaterCartTable 创建购物车表 cater_cart 与购物车明细表 cater_cart_item。
// 购物车是未下单的临时选购：一个会员在一个门店只有一张购物车，(store_id, member_id) 唯一。
type M20261023000024CreateCaterCartTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000024CreateCaterCartTable) Signature() string {
	return "20261023000024_create_cater_cart_table"
}

// Up 创建 cater_cart 与 cater_cart_item 表。
func (r *M20261023000024CreateCaterCartTable) Up() error {
	if !facades.Schema().HasTable("cater_cart") {
		if err := facades.DB().Statement(`
CREATE TABLE cater_cart (
  id         bigint   NOT NULL COMMENT '购物车编号',
  store_id   bigint   NOT NULL COMMENT '门店 ID',
  member_id  bigint   NOT NULL COMMENT '会员 ID',
  creator    varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
  created_at datetime NULL COMMENT '创建时间',
  updater    varchar(64) NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at datetime NULL COMMENT '更新时间',
  deleted_at datetime NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_store_member (store_id, member_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车'`); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("cater_cart_item") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE cater_cart_item (
  id         bigint        NOT NULL COMMENT '明细编号',
  cart_id    bigint        NOT NULL COMMENT '购物车 ID',
  sku_id     bigint        NOT NULL COMMENT '菜品规格 SKU ID',
  quantity   int           NOT NULL DEFAULT 1 COMMENT '数量',
  addons     varchar(1024) NOT NULL DEFAULT '' COMMENT '已选加料 JSON：[{addonItemId, quantity}]',
  remark     varchar(255)  NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at datetime      NULL COMMENT '创建时间',
  updater    varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at datetime      NULL COMMENT '更新时间',
  deleted_at datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_cart (cart_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车明细'`)
}

// Down 删除购物车两张表。
func (r *M20261023000024CreateCaterCartTable) Down() error {
	if err := facades.Schema().DropIfExists("cater_cart_item"); err != nil {
		return err
	}

	return facades.Schema().DropIfExists("cater_cart")
}
