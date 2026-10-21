// Package middleware 存放 infra 模块的 API 日志中间件。
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
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/infra/model"
	"eqray-flavor/app/modules/infra/repository"
)

// maxResponseBody 是响应体的落库上限，避免大响应把日志表撑爆。
const maxResponseBody = 2048

// RequestLog 是全局中间件：每个请求落一条访问日志，业务失败时补一条错误日志（docs 6.4）。
type RequestLog struct{}

// Signature 是中间件唯一标识。
func (m *RequestLog) Signature() string {
	return "infra_request_log"
}

// Handle 在请求前后记录时间、用户与业务结果。
func (m *RequestLog) Handle(ctx http.Context) {
	beginAt := time.Now()
	ctx.Request().Next()

	// 静态资源与探活请求不落库
	path := ctx.Request().Path()
	if path == "/" || strings.HasPrefix(path, "/storage") {
		return
	}

	userID := support.UserID(ctx.Context())
	if userID == 0 && path == "/api/v1/system/auth/captcha" {
		return
	}

	method := ctx.Request().Method()
	url := path
	params := "{}"
	if queries := ctx.Request().Queries(); len(queries) > 0 {
		if encoded, err := json.Marshal(queries); err == nil {
			params = string(encoded)
		}
	}

	code := result.BusinessCode(ctx)
	// 请求上下文在响应结束后可能被复用，异步落库前先把需要的值取出来
	message := result.BusinessMessage(ctx)
	// 404 / 405 等由框架直接返回的失败没有业务文案，兜底避免错误日志出现空行
	if code != 0 && message == "" {
		message = "请求失败，未返回业务错误信息"
	}
	ip := ip(ctx)
	userAgent := ctx.Request().Header("User-Agent")
	endAt := time.Now()
	duration := endAt.Sub(beginAt).Milliseconds()

	ir := repository.NewApiLogRepository()
	repoCtx := context.Background()

	// 异步落库，避免日志写库拖慢接口响应
	go func() {
		access := &model.InfraApiAccessLogModel{
			TraceID:         traceID(beginAt),
			UserID:          userID,
			UserType:        int8(baseEnums.UserTypeAdmin),
			ApplicationName: "backend",
			RequestMethod:   method,
			RequestURL:      url,
			RequestParams:   params,
			ResponseBody:    "",
			UserIP:          ip,
			UserAgent:       userAgent,
			BeginAt:         types.DateTime(beginAt),
			EndAt:           types.DateTime(endAt),
			Duration:        int(duration),
			ResultCode:      code,
		}
		_ = ir.CreateAccess(repoCtx, access)

		if code == 0 {
			return
		}

		_ = ir.CreateError(repoCtx, &model.InfraApiErrorLogModel{
			TraceID:         access.TraceID,
			UserID:          userID,
			UserType:        int8(baseEnums.UserTypeAdmin),
			ApplicationName: "backend",
			RequestMethod:   method,
			RequestURL:      url,
			RequestParams:   params,
			UserIP:          access.UserIP,
			UserAgent:       access.UserAgent,
			Name:            "BusinessError",
			Message:         message,
			ExceptionStack:  "",
			ProcessStatus:   0,
		})
	}()
}

// traceID 用开始时间纳秒值充当轻量链路标识，不引入额外中间件。
func traceID(begin time.Time) string {
	return time.Unix(0, begin.UnixNano()).Format("20060102150405.000000")
}

// ip 读取来源 IP，异常时返回空串。
func ip(ctx http.Context) string {
	return ctx.Request().Ip()
}
