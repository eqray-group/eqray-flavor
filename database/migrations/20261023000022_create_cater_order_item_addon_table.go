package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000022CreateCaterOrderItemAddonTable 创建明细加料快照表 cater_order_item_addon。
// 只存名称与加价快照，加料项后续改名改价不影响历史订单。
type M20261023000022CreateCaterOrderItemAddonTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000022CreateCaterOrderItemAddonTable) Signature() string {
	return "20261023000022_create_cater_order_item_addon_table"
}

// Up 创建 cater_order_item_addon 表。
func (r *M20261023000022CreateCaterOrderItemAddonTable) Up() error {
	if facades.Schema().HasTable("cater_order_item_addon") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE cater_order_item_addon (
  id            bigint      NOT NULL COMMENT '主键',
  order_id      bigint      NOT NULL COMMENT '订单 ID',
  order_item_id bigint      NOT NULL COMMENT '订单明细 ID',
  addon_item_id bigint      NOT NULL COMMENT '加料项 ID',
  addon_name    varchar(64) NOT NULL COMMENT '加料名称快照',
  price         bigint      NOT NULL DEFAULT 0 COMMENT '单份加价（分）',
  quantity      int         NOT NULL DEFAULT 1 COMMENT '份数',
  creator       varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
  created_at    datetime    NULL COMMENT '创建时间',
  updater       varchar(64) NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at    datetime    NULL COMMENT '更新时间',
  deleted_at    datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_order_item (order_item_id),
  KEY idx_order (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单明细加料快照'`)
}

// Down 删除 cater_order_item_addon 表。
func (r *M20261023000022CreateCaterOrderItemAddonTable) Down() error {
	return facades.Schema().DropIfExists("cater_order_item_addon")
}
