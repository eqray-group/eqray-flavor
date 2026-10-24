package migrations

import (
	"eqray-flavor/app/facades"
)

type M20261020000025CreateMemberSocialClientTable struct{}

// Signature The unique signature for the migration.
func (r *M20261020000025CreateMemberSocialClientTable) Signature() string {
	return "20261020000025_create_member_social_client_table"
}

// Up 建第三方登录渠道配置表（对齐 yudao 的 system_social_client）。
// 早期版本曾在这里补回 member_user.mobile 的唯一键（为自动注册放宽过约束），
// 该唯一键已并入 20261020000014 建表语句，这里不再重复处理。
func (r *M20261020000025CreateMemberSocialClientTable) Up() error {
	if !facades.Schema().HasTable("member_social_client") {
		if err := facades.DB().Statement(`
CREATE TABLE member_social_client (
  id            bigint       NOT NULL COMMENT '主键',
  name          varchar(64)  NOT NULL COMMENT '应用名',
  social_type   varchar(32)  NOT NULL COMMENT '渠道：wechat_mini / wechat_mp / wechat_open',
  client_id     varchar(128) NOT NULL COMMENT '客户端 ID（微信 appid）',
  client_secret varchar(256) NOT NULL COMMENT '客户端密钥（微信 appsecret）',
  agent_id      varchar(64)  NOT NULL DEFAULT '' COMMENT '代理编号（钉钉/企微等渠道）',
  status        tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  creator       varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at    datetime     NULL COMMENT '创建时间',
  updater       varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at    datetime     NULL COMMENT '更新时间',
  deleted_at    datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_social_type (social_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员第三方登录渠道配置表'`); err != nil {
			return err
		}
	}

	return nil
}

// Down 回滚配置表。
func (r *M20261020000025CreateMemberSocialClientTable) Down() error {
	return facades.Schema().DropIfExists("member_social_client")
}
