package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261023000023CreateCaterOrderLogTable 创建订单日志表 cater_order_log。
// 每次状态流转都写一条，含原状态、新状态、操作人与原因；只增不改。
type M20261023000023CreateCaterOrderLogTable struct{}

// Signature The unique signature for the migration.
func (r *M20261023000023CreateCaterOrderLogTable) Signature() string {
	return "20261023000023_create_cater_order_log_table"
}

// Up 创建 cater_order_log 表。
func (r *M20261023000023CreateCaterOrderLogTable) Up() error {
	if facades.Schema().HasTable("cater_order_log") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE cater_order_log (
  id            bigint       NOT NULL COMMENT '日志编号',
  order_id      bigint       NOT NULL COMMENT '订单 ID',
  from_status   smallint     NOT NULL DEFAULT 0 COMMENT '原状态，0 表示初始创建',
  to_status     smallint     NOT NULL COMMENT '新状态',
  operator_type tinyint      NOT NULL DEFAULT 1 COMMENT '操作方：1 顾客 / 2 店员 / 3 系统',
  operator_id   bigint       NOT NULL DEFAULT 0 COMMENT '操作人 ID，系统为 0',
  operator_name varchar(64)  NOT NULL DEFAULT '' COMMENT '操作人名称或系统动作描述',
  reason        varchar(255) NOT NULL DEFAULT '' COMMENT '变更原因，如拒单理由、超时关单',
  creator       varchar(64)  NOT NULL DEFAULT '' COMMENT '创建者',
  created_at    datetime     NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_order (order_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单状态日志'`)
}

// Down 删除 cater_order_log 表。
func (r *M20261023000023CreateCaterOrderLogTable) Down() error {
	return facades.Schema().DropIfExists("cater_order_log")
}
