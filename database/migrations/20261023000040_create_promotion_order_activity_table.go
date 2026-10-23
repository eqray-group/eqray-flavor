package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000040CreatePromotionOrderActivityTable 创建订单与活动的名额占用关联表 promotion_order_activity。
// 背景：活动名额按「活动 + 周期键」计数（docs/design/07 §5），下单时占用名额，
// 取消 / 超时关单 / 整单退款时必须把名额还回。但计数行本身只有 (activity_id, period_key)，
// 无法反查「某订单占了哪些名额」——尤其每日周期的活动跨天后再释放，
// 用当前日期算周期键会退错池子。因此下单时把订单、���动、周期键三者落库。
// 唯一键 (order_id, activity_id, period_key) 保证同一订单对同一活动同一周期只占用一次。
type M20261023000040CreatePromotionOrderActivityTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000040CreatePromotionOrderActivityTable) Signature() string {
	return "20261023000040_create_promotion_order_activity_table"
}

// Up 创建 promotion_order_activity 表。
func (r *M20261023000040CreatePromotionOrderActivityTable) Up() error {
	if facades.Schema().HasTable("promotion_order_activity") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE promotion_order_activity (
  id          bigint      NOT NULL COMMENT '编号',
  order_id    bigint      NOT NULL COMMENT '订单 ID',
  activity_id bigint      NOT NULL COMMENT '活动 ID',
  period_key  varchar(32) NOT NULL COMMENT '占用时的周期键：整期为 fixed，每日为 YYYY-MM-DD',
  status      tinyint     NOT NULL DEFAULT 1 COMMENT '状态：1 占用中 / 2 已释放',
  creator     varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
  created_at  datetime    NULL COMMENT '创建时间',
  updater     varchar(64) NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at  datetime    NULL COMMENT '更新时间',
  deleted_at  datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_activity_period (order_id, activity_id, period_key),
  KEY idx_order (order_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单占用活动名额关联'`)
}

// Down 删除 promotion_order_activity 表。
func (r *M20261023000040CreatePromotionOrderActivityTable) Down() error {
	return facades.Schema().DropIfExists("promotion_order_activity")
}
