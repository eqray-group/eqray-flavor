package config

import (
	"eqray-flavor/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("token", map[string]any{
		// 访问令牌有效期，单位秒：短期有效，过期后由刷新令牌换新（docs 9.2）
		"access_ttl": config.GetInt("ACCESS_TOKEN_TTL", 1800),
		// 刷新令牌有效期，单位秒：长期有效，登录一次可维持较长时间
		"refresh_ttl": config.GetInt("REFRESH_TOKEN_TTL", 604800),
		// Redis key 前缀，最终形如 auth:access:<token> / auth:refresh:<token>
		"prefix": config.GetString("TOKEN_PREFIX", "auth"),
	})
}
