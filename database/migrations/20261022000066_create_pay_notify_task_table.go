package migrations

import (
	"eqray-flavor/app/facades"
)

// M20261022000066CreatePayNotifyTaskTable 创建回调通知任务表，结构对齐 yudao pay_notify_task。
// type：1 支付 / 2 退款 / 3 转账；status：0 等待 / 10 成功 / 20 失败，失败按 next_notify_time 重试。
type M20261022000066CreatePayNotifyTaskTable struct{}

// Signature The unique signature for the migration.
func (r *M20261022000066CreatePayNotifyTaskTable) Signature() string {
	return "20261022000066_create_pay_notify_task_table"
}

// Up 创建 pay_notify_task 表。
func (r *M20261022000066CreatePayNotifyTaskTable) Up() error {
	if facades.Schema().HasTable("pay_notify_task") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_notify_task (
  id                bigint        NOT NULL COMMENT '任务编号',
  app_id            bigint        NOT NULL COMMENT '应用编号',
  type              tinyint       NOT NULL COMMENT '通知类型：1 支付 / 2 退款 / 3 转账',
  data_id           bigint        NOT NULL COMMENT '业务数据编号',
  merchant_order_id varchar(64)   NOT NULL DEFAULT '' COMMENT '商户订单号',
  notify_url        varchar(1024) NOT NULL DEFAULT '' COMMENT '异步通知地址',
  status            tinyint       NOT NULL DEFAULT 0 COMMENT '通知状态：0 等待 / 10 成功 / 20 失败',
  next_notify_time  datetime      NOT NULL COMMENT '下一次通知时间',
  last_execute_time datetime      NOT NULL COMMENT '最后一次执行时间',
  notify_times      tinyint       NOT NULL DEFAULT 0 COMMENT '当前通知次数',
  max_notify_times  tinyint       NOT NULL DEFAULT 0 COMMENT '最大通知次数',
  creator           varchar(64)   NOT NULL DEFAULT '' COMMENT '创建者',
  created_at        datetime      NULL COMMENT '创建时间',
  updater           varchar(64)   NOT NULL DEFAULT '' COMMENT '更新者',
  updated_at        datetime      NULL COMMENT '更新时间',
  deleted_at        datetime      NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  KEY idx_status_next (status, next_notify_time),
  KEY idx_type_data (type, data_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付回调通知任务'`)
}

// Down 删除 pay_notify_task 表。
func (r *M20261022000066CreatePayNotifyTaskTable) Down() error {
	return facades.Schema().DropIfExists("pay_notify_task")
}
