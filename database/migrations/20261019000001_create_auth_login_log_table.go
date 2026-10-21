package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261019000001CreateAuthLoginLogTable struct{}

// Signature The unique signature for the migration.
func (r *M20261019000001CreateAuthLoginLogTable) Signature() string {
	return "20261019000001_create_auth_login_log_table"
}

// Up 建 auth_login_log 登录日志表（docs 6.2）。
// 表归 base 认证组件所有，只追加不删除，不建唯一键。
func (r *M20261019000001CreateAuthLoginLogTable) Up() error {
	if facades.Schema().HasTable("auth_login_log") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE auth_login_log (
  id         bigint       NOT NULL COMMENT '主键',
  user_id    bigint       NOT NULL DEFAULT 0 COMMENT '用户 ID，0 表示账号不存在',
  user_type  tinyint      NOT NULL DEFAULT 1 COMMENT '用户类型：1 后台用户 / 2 会员',
  username   varchar(64)  NOT NULL DEFAULT '' COMMENT '登录账号',
  type       tinyint      NOT NULL DEFAULT 1 COMMENT '类型：1 登录 / 2 登出',
  result     tinyint      NOT NULL DEFAULT 1 COMMENT '结果：0 失败 / 1 成功',
  user_ip    varchar(50)  NOT NULL DEFAULT '' COMMENT '来源 IP',
  user_agent varchar(512) NOT NULL DEFAULT '' COMMENT '客户端信息',
  created_at datetime     NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_user_id (user_id),
  KEY idx_username (username),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表'`)
}

// Down Reverse the migrations.
func (r *M20261019000001CreateAuthLoginLogTable) Down() error {
	return facades.Schema().DropIfExists("auth_login_log")
}
