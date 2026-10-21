package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261020000002CreateSystemOperateLogTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000002CreateSystemOperateLogTable) Signature() string {
	return "20261020000002_create_system_operate_log_table"
}

// Up 建 system_operate_log 操作日志表（docs 6.3）。
// 日志只追加不删除，因此不设 deleted_at，也不建唯一键。
func (r *M20261020000002CreateSystemOperateLogTable) Up() error {
	if facades.Schema().HasTable("system_operate_log") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE system_operate_log (
  id             bigint       NOT NULL COMMENT '主键',
  trace_id       varchar(64)  NOT NULL DEFAULT '' COMMENT '链路追踪 ID',
  user_id        bigint       NOT NULL DEFAULT 0 COMMENT '操作人 ID',
  user_type      tinyint      NOT NULL DEFAULT 1 COMMENT '用户类型：1 后台用户 / 2 会员',
  type           varchar(64)  NOT NULL DEFAULT '' COMMENT '操作类型（模块）',
  sub_type       varchar(64)  NOT NULL DEFAULT '' COMMENT '操作子类型（动作）',
  biz_id         varchar(64)  NOT NULL DEFAULT '' COMMENT '业务主键',
  action         varchar(255) NOT NULL DEFAULT '' COMMENT '操作描述',
  success        tinyint      NOT NULL DEFAULT 1 COMMENT '是否成功：0 失败 / 1 成功',
  extra          text         NULL COMMENT '扩展信息（查询参数等，已脱敏）',
  request_method varchar(16)  NOT NULL DEFAULT '' COMMENT 'HTTP 方法',
  request_url    varchar(512) NOT NULL DEFAULT '' COMMENT '请求地址',
  request_ip     varchar(50)  NOT NULL DEFAULT '' COMMENT '请求 IP',
  user_agent     varchar(512) NOT NULL DEFAULT '' COMMENT '客户端标识',
  duration       bigint       NOT NULL DEFAULT 0 COMMENT '耗时（毫秒）',
  created_at     datetime     NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_user_id (user_id),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表'`)
}

// Down Reverse the migrations.
func (r *M20261020000002CreateSystemOperateLogTable) Down() error {
	return facades.Schema().DropIfExists("system_operate_log")
}
