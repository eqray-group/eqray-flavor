package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261020000009CreateInfraApiErrorLogTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000009CreateInfraApiErrorLogTable) Signature() string {
	return "20261020000009_create_infra_api_error_log_table"
}

// Up 建 infra_api_error_log API 错误日志表（docs 6.4），只在业务失败时写入。
func (r *M20261020000009CreateInfraApiErrorLogTable) Up() error {
	if facades.Schema().HasTable("infra_api_error_log") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE infra_api_error_log (
  id               bigint       NOT NULL COMMENT '主键',
  trace_id         varchar(64)  NOT NULL DEFAULT '' COMMENT '链路追踪 ID',
  user_id          bigint       NOT NULL DEFAULT 0 COMMENT '用户 ID',
  user_type        tinyint      NOT NULL DEFAULT 1 COMMENT '用户类型：1 后台用户 / 2 会员',
  application_name varchar(64)  NOT NULL DEFAULT 'backend' COMMENT '应用名',
  request_method   varchar(16)  NOT NULL DEFAULT '' COMMENT 'HTTP 方法',
  request_url      varchar(512) NOT NULL DEFAULT '' COMMENT '请求地址',
  request_params   text         NULL COMMENT '请求参数（脱敏）',
  user_ip          varchar(50)  NOT NULL DEFAULT '' COMMENT '来源 IP',
  user_agent       varchar(512) NOT NULL DEFAULT '' COMMENT '客户端信息',
  name             varchar(128) NOT NULL DEFAULT '' COMMENT '异常类名',
  message          text         NULL COMMENT '异常信息',
  exception_stack  text         NULL COMMENT '异常堆栈',
  process_status   tinyint      NOT NULL DEFAULT 0 COMMENT '处理状态：0 未处理 / 1 已处理',
  process_time     datetime     NULL COMMENT '处理时间',
  created_at       datetime     NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_trace_id (trace_id),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API 错误日志表'`)
}

// Down Reverse the migrations.
func (r *M20261020000009CreateInfraApiErrorLogTable) Down() error {
	return facades.Schema().DropIfExists("infra_api_error_log")
}
