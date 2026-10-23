package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000067CreatePayNotifyLogTable 创建回调通知日志表，结构对齐 yudao pay_notify_log。
// 每条通知任务每次执行写一条日志，记录响应内容与结果。
type M20261022000067CreatePayNotifyLogTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000067CreatePayNotifyLogTable) Signature() string {
	return "20261022000067_create_pay_notify_log_table"
}

// Up 创建 pay_notify_log 表。
func (r *M20261022000067CreatePayNotifyLogTable) Up() error {
	if facades.Schema().HasTable("pay_notify_log") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_notify_log (
  id           bigint        NOT NULL COMMENT '日志编号',
  task_id      bigint        NOT NULL COMMENT '通知任务编号',
  notify_times tinyint       NOT NULL DEFAULT 0 COMMENT '第几次通知',
  response     varchar(2048) NOT NULL DEFAULT '' COMMENT '通知响应内容',
  status       tinyint       NOT NULL DEFAULT 0 COMMENT '通知状态：0 等待 / 10 成功 / 20 失败',
  creator      varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at   datetime      NULL COMMENT '创建时间',
  updater      varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at   datetime      NULL COMMENT '更新时间',
  deleted_at   datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_task (task_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付回调通知日志'`)
}

// Down 删除 pay_notify_log 表。
func (r *M20261022000067CreatePayNotifyLogTable) Down() error {
	return facades.Schema().DropIfExists("pay_notify_log")
}
