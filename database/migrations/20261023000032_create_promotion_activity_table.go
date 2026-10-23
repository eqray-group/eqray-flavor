package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000032CreatePromotionActivityTable 创建营销活动表 promotion_activity。
// 规则细节存 rule JSON；优先级、互斥组、独占与可叠加标记支撑 07 §5 的判定顺序。
type M20261023000032CreatePromotionActivityTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000032CreatePromotionActivityTable) Signature() string {
	return "20261023000032_create_promotion_activity_table"
}

// Up 创建 promotion_activity 表。
func (r *M20261023000032CreatePromotionActivityTable) Up() error {
	if facades.Schema().HasTable("promotion_activity") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE promotion_activity (
  id          bigint        NOT NULL COMMENT '活动编号',
  name        varchar(64)   NOT NULL COMMENT '活动名称',
  type        tinyint       NOT NULL COMMENT '活动玩法：1 满减 / 2 第二份半价 / 3 限时折扣 / 4 套餐特惠 / 5 加价购',
  level       tinyint       NOT NULL DEFAULT 2 COMMENT '作用层级：1 商品级 / 2 订单级',
  description varchar(255)  NOT NULL DEFAULT '' COMMENT '活动说明',
  rule        text          NULL COMMENT '玩法规则 JSON，结构见 promotion service 层',
  scope       tinyint       NOT NULL DEFAULT 1 COMMENT '适用门店范围：1 全品牌 / 2 指定门店',
  scope_ids   varchar(1000) NOT NULL DEFAULT '' COMMENT '适用门店 ID 列表，逗号分隔',
  priority    int           NOT NULL DEFAULT 0 COMMENT '优先级，数值大的先判定',
  mutex_group varchar(64)   NOT NULL DEFAULT '' COMMENT '互斥组，组内只生效一个',
  exclusive   tinyint       NOT NULL DEFAULT 0 COMMENT '是否独占：1 表示生效后其他活动终止',
  stackable   tinyint       NOT NULL DEFAULT 1 COMMENT '是否可叠加：1 表示优惠金额累加',
  quota_limit int           NOT NULL DEFAULT 0 COMMENT '名额上限，0 表示不限',
  quota_cycle tinyint       NOT NULL DEFAULT 1 COMMENT '名额周期：1 整个活动期 / 2 每日重置',
  start_time  datetime      NOT NULL COMMENT '活动开始时间',
  end_time    datetime      NOT NULL COMMENT '活动结束时间',
  status      tinyint       NOT NULL DEFAULT 0 COMMENT '启用状态：0 未启用 / 1 已启用',
  creator     varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at  datetime      NULL COMMENT '创建时间',
  updater     varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at  datetime      NULL COMMENT '更新时间',
  deleted_at  datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_status_time (status, start_time, end_time),
  KEY idx_mutex (mutex_group),
  KEY idx_type_level (type, level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='营销活动'`)
}

// Down 删除 promotion_activity 表。
func (r *M20261023000032CreatePromotionActivityTable) Down() error {
	return facades.Schema().DropIfExists("promotion_activity")
}
