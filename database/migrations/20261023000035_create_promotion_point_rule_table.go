package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000035CreatePromotionPointRuleTable 创建积分抵扣规则表 promotion_point_rule。
// 规则全品牌统一，只保留一条生效记录，修改即时生效。
type M20261023000035CreatePromotionPointRuleTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000035CreatePromotionPointRuleTable) Signature() string {
	return "20261023000035_create_promotion_point_rule_table"
}

// Up 创建 promotion_point_rule 表，并写入默认规则：100 积分抵 1 元，单笔最多抵 50 元。
func (r *M20261023000035CreatePromotionPointRuleTable) Up() error {
	if facades.Schema().HasTable("promotion_point_rule") {
		return nil
	}

	if err := facades.DB().Statement(`
CREATE TABLE promotion_point_rule (
  id                bigint      NOT NULL COMMENT '编号',
  exchange_rate     int         NOT NULL DEFAULT 100 COMMENT '兑换比例：多少积分抵 1 元（分）',
  max_deduct_amount bigint      NOT NULL DEFAULT 5000 COMMENT '单笔抵扣上限（分），0 表示不限制',
  min_order_amount  bigint      NOT NULL DEFAULT 0 COMMENT '使用积分抵扣的最低订单金额（分）',
  status            tinyint     NOT NULL DEFAULT 1 COMMENT '启用状态：0 关闭 / 1 开启',
  creator           varchar(64) NOT NULL DEFAULT '' COMMENT '创建者',
  created_at        datetime    NULL COMMENT '创建时间',
  updater           varchar(64) NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at        datetime    NULL COMMENT '更新时间',
  deleted_at        datetime    NULL COMMENT '删除时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分抵扣规则'`); err != nil {
		return err
	}

	_, err := facades.DB().Table("promotion_point_rule").Insert(map[string]any{
		"id":                int64(1001),
		"exchange_rate":     100,
		"max_deduct_amount": int64(5000),
		"min_order_amount":  int64(0),
		"status":            1,
		"creator":           "system",
		"created_at":        "2026-01-01 00:00:00",
		"updater":           "system",
		"updated_at":        "2026-01-01 00:00:00",
	})

	return err
}

// Down 删除 promotion_point_rule 表。
func (r *M20261023000035CreatePromotionPointRuleTable) Down() error {
	return facades.Schema().DropIfExists("promotion_point_rule")
}
