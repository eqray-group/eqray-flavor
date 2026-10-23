package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000021CreateCaterOrderItemTable 创建订单明细表 cater_order_item。
// 菜品快照：名称、单价、数量、加料费、分摊优惠与小计全部落库，订单完成后不受菜品改价影响。
type M20261023000021CreateCaterOrderItemTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000021CreateCaterOrderItemTable) Signature() string {
	return "20261023000021_create_cater_order_item_table"
}

// Up 创建 cater_order_item 表。
func (r *M20261023000021CreateCaterOrderItemTable) Up() error {
	if facades.Schema().HasTable("cater_order_item") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE cater_order_item (
  id                  bigint       NOT NULL COMMENT '明细编号',
  order_id            bigint       NOT NULL COMMENT '订单 ID',
  spu_id              bigint       NOT NULL COMMENT '菜品 SPU ID',
  sku_id              bigint       NOT NULL COMMENT '菜品规格 SKU ID',
  spu_name            varchar(128) NOT NULL COMMENT '菜品名称快照',
  sku_name            varchar(64)  NOT NULL DEFAULT '' COMMENT '规格名称快照',
  original_price      bigint       NOT NULL COMMENT '下单时标准价 / 门店价（分），用于对账',
  unit_price          bigint       NOT NULL COMMENT '成交单价（分），含会员价与单品级优惠后的价格',
  quantity            int          NOT NULL COMMENT '数量',
  addon_amount        bigint       NOT NULL DEFAULT 0 COMMENT '加料费合计（分）',
  item_discount_amount bigint      NOT NULL DEFAULT 0 COMMENT '分摊到本明细的整单优惠（分）',
  subtotal            bigint       NOT NULL COMMENT '小计（分）= (unit_price * quantity + addon_amount) - item_discount_amount',
  status              tinyint      NOT NULL DEFAULT 1 COMMENT '明细状态：1 正常 / 2 已退',
  orderer             varchar(64)  NOT NULL DEFAULT '' COMMENT '点菜人，堂食多人点菜时使用',
  remark              varchar(255) NOT NULL DEFAULT '' COMMENT '明细备注，如少辣',
  creator             varchar(64)  NOT NULL DEFAULT '' COMMENT '创建者',
  created_at          datetime     NULL COMMENT '创建时间',
  updater             varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at          datetime     NULL COMMENT '更新时间',
  deleted_at          datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_order (order_id),
  KEY idx_sku (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单明细'`)
}

// Down 删除 cater_order_item 表。
func (r *M20261023000021CreateCaterOrderItemTable) Down() error {
	return facades.Schema().DropIfExists("cater_order_item")
}
