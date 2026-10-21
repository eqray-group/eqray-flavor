// Package types 存放跨模块共用的基础类型。
package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ID 是雪花 ID 的统一类型：数据库里是 int64，JSON 里一律是字符串。
// 见 docs 8.4「ID 字段的序列化」。
type ID int64

// String 把雪花 ID 转成十进制字符串，供 JSON、日志与 Redis 键使用。
func (i ID) String() string { return strconv.FormatInt(int64(i), 10) }

// MarshalJSON 把 ID 序列化成字符串，避免前端 JS 处理 19 位整数时丢精度。
func (i ID) MarshalJSON() ([]byte, error) { return json.Marshal(i.String()) }

// UnmarshalJSON 解析请求里的 ID：兼容字符串与数字，空串、null 视为 0。
func (i *ID) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*i = 0
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("types.ID: 无法解析 %q: %w", s, err)
	}
	*i = ID(v)
	return nil
}

// Value 让 GORM 把 ID 当作 int64 写入数据库。
func (i ID) Value() (driver.Value, error) { return int64(i), nil }

// Scan 实现 sql.Scanner：兼容 nil、int64、[]byte、string 四种来源。
func (i *ID) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*i = 0
	case int64:
		*i = ID(v)
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return fmt.Errorf("types.ID: 无法从 %s 扫描: %w", string(v), err)
		}
		*i = ID(n)
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("types.ID: 无法从 %s 扫描: %w", v, err)
		}
		*i = ID(n)
	default:
		return fmt.Errorf("types.ID: 无法从 %T 扫描", src)
	}
	return nil
}
