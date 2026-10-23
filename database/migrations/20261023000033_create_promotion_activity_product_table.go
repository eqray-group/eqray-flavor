package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000033CreatePromotionActivityProductTable 创建活动参与商品表 promotion_activity_product。
// 商品级活动必须配置参与商品；唯一键保证同一活动同一 SPU 不重复配置。
type M20261023000033CreatePromotionActivityProductTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000033CreatePromotionActivityProductTable) Signature() string {
	return "20261023000033_create_promotion_activity_product_table"
}

// Up 创建 promotion_activity_product 表。
func (r *M20261023000033CreatePromotionActivityProductTable) Up() error {
	if facades.Schema().HasTable("promotion_activity_product") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE promotion_activity_product (
  id          bigint      NOT NULL COMMENT '编号',
  activity_id bigint      NOT NULL COMMENT '活动 ID',
  spu_id      bigint      NOT NULL COMMENT '参与活动的菜品 SPU ID',
  creator     varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
  created_at  datetime    NULL COMMENT '创建时间',
  updater     varchar(64) NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at  datetime    NULL COMMENT '更新时间',
  deleted_at  datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_activity_spu (activity_id, spu_id),
  KEY idx_spu (spu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='营销活动参与商品'`)
}

// Down 删除 promotion_activity_product 表。
func (r *M20261023000033CreatePromotionActivityProductTable) Down() error {
	return facades.Schema().DropIfExists("promotion_activity_product")
}
