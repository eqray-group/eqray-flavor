package config

import (
	"github.com/goravel/framework/contracts/cache"
	redisFacades "github.com/goravel/redis/facades"

	"eqray-flavor/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("cache", map[string]any{
		// 默认缓存驱动：访问令牌 / 刷新令牌与后续的权限缓存都落在 Redis，
		// 多实例部署时才能共享同一份登录态（docs 9.2）。
		"default": config.Env("CACHE_DRIVER", "redis"),

		// Cache Stores
		//
		// 保留 memory 便于单测；生产走 redis。
		"stores": map[string]any{
			"memory": map[string]any{
				"driver": "memory",
			},
			"redis": map[string]any{
				"driver": "custom",
				// 对应 database.redis.<connection>，见 config/database.go
				"connection": config.Env("REDIS_CACHE_CONNECTION", "default"),
				"via": func() (cache.Driver, error) {
					return redisFacades.Cache("redis")
				},
			},
		},

		// Cache Key Prefix
		//
		// Must: a-zA-Z0-9_-
		"prefix": config.Env("CACHE_PREFIX", "eqray_cache"),
	})
}
