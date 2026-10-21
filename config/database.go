package config

import (
	"eqray-flavor/app/facades"

	"github.com/goravel/framework/contracts/database/driver"
	mysqlFacades "github.com/goravel/mysql/facades"
)

func init() {
	config := facades.Config()
	config.Add("database", map[string]any{
		// Default database connection name
		"default": config.Env("DB_CONNECTION", "mysql"),
		// Database connections
		"connections": map[string]any{
			// MySQL 8.0：字符集 utf8mb4、排序规则 utf8mb4_unicode_ci，见 docs 第 2 节
			"mysql": map[string]any{
				"host":     config.Env("DB_HOST", "127.0.0.1"),
				"port":     config.Env("DB_PORT", "3306"),
				"database": config.Env("DB_DATABASE"),
				"username": config.Env("DB_USERNAME"),
				"password": config.Env("DB_PASSWORD"),
				"charset":  "utf8mb4",
				"loc":      "Local",
				"singular": false,
				"prefix":   "",
				"via": func() (driver.Driver, error) {
					return mysqlFacades.Mysql("mysql")
				},
			},
		},
		// Redis 连接：缓存默认走 redis store，令牌就存在这里（docs 9.2）。
		// 配置项名固定为 database.redis.<connection>.*，由 goravel/redis 读取。
		"redis": map[string]any{
			"default": map[string]any{
				"host":     config.Env("REDIS_HOST", "127.0.0.1"),
				"port":     config.Env("REDIS_PORT", "6379"),
				"password": config.Env("REDIS_PASSWORD", ""),
				"database": config.Env("REDIS_DB", 0),
			},
		},
		// Pool configuration
		"pool": map[string]any{
			// Sets the maximum number of connections in the idle
			// connection pool.
			//
			// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
			// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
			//
			// If n <= 0, no idle connections are retained.
			"max_idle_conns": 10,
			// Sets the maximum number of open connections to the database.
			//
			// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
			// MaxIdleConns, then MaxIdleConns will be reduced to match the new
			// MaxOpenConns limit.
			//
			// If n <= 0, then there is no limit on the number of open connections.
			"max_open_conns": 100,
			// Sets the maximum amount of time a connection may be idle.
			//
			// Expired connections may be closed lazily before reuse.
			//
			// If d <= 0, connections are not closed due to a connection's idle time.
			// Unit: Second
			"conn_max_idletime": 3600,
			// Sets the maximum amount of time a connection may be reused.
			//
			// Expired connections may be closed lazily before reuse.
			//
			// If d <= 0, connections are not closed due to a connection's age.
			// Unit: Second
			"conn_max_lifetime": 3600,
		},
		// Sets the threshold for slow queries in milliseconds, the slow query will be logged.
		// Unit: Millisecond
		"slow_threshold": 200,
		// Migration Repository Table
		//
		// This table keeps track of all the migrations that have already run for
		// your application. Using this information, we can determine which of
		// the migrations on disk haven't actually been run in the database.
		"migrations": map[string]any{
			"table": "migrations",
		},
	})
}
