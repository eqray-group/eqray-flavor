package middleware

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
)

// OperateLogInput 是一次写操作留痕所需的数据。
type OperateLogInput struct {
	UserID        int64
	UserType      baseEnums.UserTypeEnum
	Type          string
	SubType       string
	Action        string
	Success       bool
	Extra         string
	RequestMethod string
	RequestURL    string
	RequestIP     string
	UserAgent     string
	Duration      int64
}

// OperateLogRecorder 由 system 模块注入，负责把操作日志写入 system_operate_log。
// 中间件放在 base，写库实现留在数据归属的 system，避免业务模块之间互相依赖。
type OperateLogRecorder func(ctx context.Context, in OperateLogInput) error

var operateLogRecorder OperateLogRecorder

// RegisterOperateLogRecorder 在应用启动时注入操作日志写入实现。
func RegisterOperateLogRecorder(recorder OperateLogRecorder) {
	operateLogRecorder = recorder
}

// OperateLog 采集写操作并在后置阶段交给 recorder 落库。
// 注册在 Permission 之内：只有通过鉴权并实际执行的请求才会留痕。
type OperateLog struct {
	// Type 是操作类型（模块），如「用户管理」
	Type string
	// SubType 是操作子类型（动作），如「新增用户」
	SubType string
}

// Signature 让每个操作语义对应独立的中间件签名。
func (m *OperateLog) Signature() string {
	return "operate_log_" + strings.NewReplacer(":", "_", " ", "_", "-", "_").Replace(m.Type+"_"+m.SubType)
}

// Handle 在后置阶段读取业务结果码并写日志；未登录或写库失败都不影响主流程。
func (m *OperateLog) Handle(ctx http.Context) {
	start := time.Now()

	ctx.Request().Next()

	recorder := operateLogRecorder
	if recorder == nil {
		return
	}

	userID := support.UserID(ctx.Context())
	if userID == 0 {
		return
	}

	// 只采集查询参数，不读请求体，避免干扰 Controller 的参数绑定
	extra := "{}"
	if queries := ctx.Request().Queries(); len(queries) > 0 {
		if encoded, err := json.Marshal(queries); err == nil {
			extra = string(encoded)
		}
	}

	_ = recorder(ctx.Context(), OperateLogInput{
		UserID:        userID,
		UserType:      baseEnums.UserTypeAdmin,
		Type:          m.Type,
		SubType:       m.SubType,
		Action:        m.Type + " - " + m.SubType,
		Success:       result.BusinessCode(ctx) == 0,
		Extra:         extra,
		RequestMethod: ctx.Request().Method(),
		RequestURL:    ctx.Request().Path(),
		RequestIP:     ctx.Request().Ip(),
		UserAgent:     ctx.Request().Header("User-Agent"),
		Duration:      time.Since(start).Milliseconds(),
	})
}
