package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DateTimeLayout 是时间字段对外统一的 JSON 格式。
// 想改成毫秒时间戳，只需改这里与 MarshalJSON 的实现。
const DateTimeLayout = "2006-01-02 15:04:05"

// location 是读写时间字段使用的时区，默认跟随服务器本地时区。
// 部署到多时区时应在启动阶段调用 SetLocation，并与数据库连接的 loc 参数保持一致。
var location = time.Local

// SetLocation 设置时间字段的读写时区，传 nil 时忽略。
func SetLocation(loc *time.Location) {
	if loc != nil {
		location = loc
	}
}

// Location 返回当前使用的时区。
func Location() *time.Location { return location }

// DateTime 是统一的时间类型：数据库里是 datetime，JSON 里是 "2006-01-02 15:04:05"。
// Go 的 time.Time 默认序列化成 RFC3339（2026-10-17T09:49:56+08:00），与本项目约定不一致，
// 因此审计字段统一用它，避免每个 Model 各写一遍格式化。
type DateTime time.Time

// Now 返回当前时间。
func Now() DateTime { return DateTime(time.Now()) }

// Time 转回标准库时间。
func (t DateTime) Time() time.Time { return time.Time(t) }

// IsZero 判断是否为零值（未赋值）。
func (t DateTime) IsZero() bool { return time.Time(t).IsZero() }

// String 按 DateTimeLayout 输出字符串，零值返回空串。
func (t DateTime) String() string {
	if t.IsZero() {
		return ""
	}
	return t.Time().In(location).Format(DateTimeLayout)
}

// MarshalJSON 输出 "2006-01-02 15:04:05" 格式，零值输出空串而非 null。
func (t DateTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte(`""`), nil
	}
	return json.Marshal(t.String())
}

// UnmarshalJSON 兼容 "2006-01-02 15:04:05" 与 RFC3339 两种写法，空串视为零值。
func (t *DateTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*t = DateTime{}
		return nil
	}
	return t.parse(s)
}

// Value 实现 driver.Valuer：零值写 NULL，其余按配置时区转换后入库。
func (t DateTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time().In(location), nil
}

// Scan 实现 sql.Scanner：兼容 nil、time.Time、string、[]byte 四种来源。
func (t *DateTime) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*t = DateTime{}
	case time.Time:
		*t = DateTime(v)
	case string:
		return t.parse(v)
	case []byte:
		return t.parse(string(v))
	default:
		return fmt.Errorf("types.DateTime: 无法从 %T 扫描", src)
	}
	return nil
}

// parse 依次按项目布局与 RFC3339 解析，都失败时返回错误。
func (t *DateTime) parse(s string) error {
	for _, layout := range []string{DateTimeLayout, time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, s, location); err == nil {
			*t = DateTime(parsed)
			return nil
		}
	}
	return fmt.Errorf("types.DateTime: 无法解析 %q", s)
}
