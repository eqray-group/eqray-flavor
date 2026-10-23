package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000034CreatePromotionActivityQuotaTable 创建活动名额计数表 promotion_activity_quota。
// 名额按「活动 + 周期键」计数：整期键为 fixed，每日键为 YYYY-MM-DD，跨天自动重置。
// 唯一键保证并发下单不会各占一份名额。
type M20261023000034CreatePromotionActivityQuotaTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000034CreatePromotionActivityQuotaTable) Signature() string {
	return "20261023000034_create_promotion_activity_quota_table"
}

// Up 创建 promotion_activity_quota 表。
func (r *M20261023000034CreatePromotionActivityQuotaTable) Up() error {
	if facades.Schema().HasTable("promotion_activity_quota") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE promotion_activity_quota (
  id          bigint      NOT NULL COMMENT '编号',
  activity_id bigint      NOT NULL COMMENT '活动 ID',
  period_key  varchar(32) NOT NULL COMMENT '周期键：整期为 fixed，每日为 YYYY-MM-DD',
  used_count  int         NOT NULL DEFAULT 0 COMMENT '已占用名额',
  creator     varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
  created_at  datetime    NULL COMMENT '创建时间',
  updater     varchar(64) NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at  datetime    NULL COMMENT '更新时间',
  deleted_at  datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_activity_period (activity_id, period_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='营销活动名额计数'`)
}

// Down 删除 promotion_activity_quota 表。
func (r *M20261023000034CreatePromotionActivityQuotaTable) Down() error {
	return facades.Schema().DropIfExists("promotion_activity_quota")
}
