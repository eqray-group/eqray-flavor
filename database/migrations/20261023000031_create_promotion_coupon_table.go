package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000031CreatePromotionCouponTable 创建会员券表 promotion_coupon。
// 状态流转必须条件更新防并发重复使用，因此索引覆盖 (member_id, status) 与 (order_id, status)。
type M20261023000031CreatePromotionCouponTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000031CreatePromotionCouponTable) Signature() string {
	return "20261023000031_create_promotion_coupon_table"
}

// Up 创建 promotion_coupon 表。
func (r *M20261023000031CreatePromotionCouponTable) Up() error {
	if facades.Schema().HasTable("promotion_coupon") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE promotion_coupon (
  id          bigint        NOT NULL COMMENT '券编号',
  template_id bigint        NOT NULL COMMENT '券模板 ID',
  member_id   bigint        NOT NULL COMMENT '持有会员 ID',
  status      tinyint       NOT NULL DEFAULT 1 COMMENT '券状态：1 未使用 / 2 锁定 / 3 已使用 / 4 已过期',
  order_id    bigint        NOT NULL DEFAULT 0 COMMENT '占用该券的订单 ID',
  use_time    datetime      NULL COMMENT '核销时间',
  expire_time datetime      NOT NULL COMMENT '券过期时间',
  get_time    datetime      NOT NULL COMMENT '领取时间',
  creator     varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at  datetime      NULL COMMENT '创建时间',
  updater     varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at  datetime      NULL COMMENT '更新时间',
  deleted_at  datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_member_status (member_id, status),
  KEY idx_template_member (template_id, member_id),
  KEY idx_order (order_id, status),
  KEY idx_expire (status, expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员优惠券'`)
}

// Down 删除 promotion_coupon 表。
func (r *M20261023000031CreatePromotionCouponTable) Down() error {
	return facades.Schema().DropIfExists("promotion_coupon")
}
