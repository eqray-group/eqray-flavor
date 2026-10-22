package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000011CreateStoreRiderTable 建 store_rider 骑手档案表。
// 骑手与门店员工一对一；停用只改状态，删除采用软删除以保留历史。
type M20261022000011CreateStoreRiderTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000011CreateStoreRiderTable) Signature() string {
	return "20261022000011_create_store_rider_table"
}

// Up 建 store_rider 表。
// employee_id 唯一保证一名门店员工只有一份骑手档案（design/24 §5）。
func (r *M20261022000011CreateStoreRiderTable) Up() error {
	if facades.Schema().HasTable("store_rider") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE store_rider (
  id             bigint       NOT NULL COMMENT '主键',
  store_id       bigint       NOT NULL COMMENT '门店 ID，与关联员工的门店一致',
  employee_id    bigint       NOT NULL COMMENT '门店员工关联 ID，一名员工只能有一份骑手档案',
  status         tinyint      NOT NULL DEFAULT 1 COMMENT '骑手状态：1 空闲 / 2 接单中 / 3 配送中 / 4 休息 / 5 停用',
  max_orders     int          NOT NULL DEFAULT 3 COMMENT '接单并发上限',
  current_orders int          NOT NULL DEFAULT 0 COMMENT '当前进行中的配送单数量',
  today_orders   int          NOT NULL DEFAULT 0 COMMENT '今日累计接单量，跨日由接单条件更新重置',
  stat_date      date         NULL COMMENT '今日单量统计日期',
  remark         varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator        varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at     datetime     NULL COMMENT '创建时间',
  updater        varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at     datetime     NULL COMMENT '更新时间',
  deleted_at     datetime     NULL COMMENT '删除时间；停用不影响历史，删除为软删除',
  PRIMARY KEY (id),
  UNIQUE KEY uk_employee (employee_id),
  KEY idx_store_status (store_id, status),
  KEY idx_store_employee (store_id, employee_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='骑手档案表（骑手与门店员工一对一）'`)
}

// Down 删除 store_rider 表。
func (r *M20261022000011CreateStoreRiderTable) Down() error {
	return facades.Schema().DropIfExists("store_rider")
}
