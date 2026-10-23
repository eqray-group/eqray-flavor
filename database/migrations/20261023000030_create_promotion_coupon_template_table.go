package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000030CreatePromotionCouponTemplateTable 创建优惠券模板表 promotion_coupon_template。
// 券模板只描述规则，金额一律整型分；折扣率用万分比（7500 = 75 折）。
type M20261023000030CreatePromotionCouponTemplateTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000030CreatePromotionCouponTemplateTable) Signature() string {
	return "20261023000030_create_promotion_coupon_template_table"
}

// Up 创建 promotion_coupon_template 表。
func (r *M20261023000030CreatePromotionCouponTemplateTable) Up() error {
	if facades.Schema().HasTable("promotion_coupon_template") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE promotion_coupon_template (
  id                  bigint        NOT NULL COMMENT '模板编号',
  name                varchar(64)   NOT NULL COMMENT '券名称',
  type                tinyint       NOT NULL COMMENT '券类型：1 满减 / 2 折扣 / 3 单品 / 4 兑换 / 5 代金',
  description         varchar(255)  NOT NULL DEFAULT '' COMMENT '券说明',
  threshold_amount    bigint        NOT NULL DEFAULT 0 COMMENT '使用门槛（分），0 表示无门槛',
  discount_amount     bigint        NOT NULL DEFAULT 0 COMMENT '抵扣金额（分）',
  discount_rate       int           NOT NULL DEFAULT 0 COMMENT '折扣率（万分比）：7500 表示 75 折',
  max_discount_amount bigint        NOT NULL DEFAULT 0 COMMENT '折扣封顶（分），0 表示不封顶',
  scope               tinyint       NOT NULL DEFAULT 1 COMMENT '适用范围：1 全品牌 / 2 指定门店 / 3 指定商品',
  scope_ids           varchar(1000) NOT NULL DEFAULT '' COMMENT '适用对象 ID 列表，逗号分隔',
  channel_limit       tinyint       NOT NULL DEFAULT 0 COMMENT '渠道限制：0 不限 / 1 仅线上 / 2 仅线下',
  issue_type          tinyint       NOT NULL DEFAULT 1 COMMENT '发放方式：1 主动领取 / 2 消费返券 / 3 新人礼 / 4 定向发券',
  total_count         int           NOT NULL DEFAULT 0 COMMENT '发行总量，0 表示不限量',
  issued_count        int           NOT NULL DEFAULT 0 COMMENT '已发放数量',
  per_member_limit    int           NOT NULL DEFAULT 0 COMMENT '每人限领张数，0 表示不限',
  start_time          datetime      NOT NULL COMMENT '生效开始时间',
  end_time            datetime      NOT NULL COMMENT '生效结束时间',
  valid_days          int           NOT NULL DEFAULT 0 COMMENT '领取后有效天数，0 表示长期有效',
  status              tinyint       NOT NULL DEFAULT 1 COMMENT '启用状态：0 关闭 / 1 开启',
  creator             varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at          datetime      NULL COMMENT '创建时间',
  updater             varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at          datetime      NULL COMMENT '更新时间',
  deleted_at          datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_type_status (type, status),
  KEY idx_time (start_time, end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券模板'`)
}

// Down 删除 promotion_coupon_template 表。
func (r *M20261023000030CreatePromotionCouponTemplateTable) Down() error {
	return facades.Schema().DropIfExists("promotion_coupon_template")
}
