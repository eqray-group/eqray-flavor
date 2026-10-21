package service

import (
	"context"

	baseAuth "eqray-flavor/app/modules/base/auth"
	baseEnums "eqray-flavor/app/modules/base/enums"
)

// LoginLogListQuery 是登录日志列表的查询条件。
type LoginLogListQuery struct {
	Username string
	Type     *baseEnums.LoginTypeEnum
	Result   *baseEnums.LoginResultEnum
	PageNo   int
	PageSize int
}

// LogService 读取登录日志；表归 base 的认证组件，system 只通过其接口获取。
type LogService struct{}

func NewLogService() *LogService {
	return &LogService{}
}

// LoginLogs 分页查询登录日志（docs 7.1）。
func (s *LogService) LoginLogs(ctx context.Context, q LoginLogListQuery) ([]baseAuth.LoginLog, int64, error) {
	return baseAuth.PageLoginLogs(ctx, baseAuth.LoginLogFilter{
		Username: q.Username,
		Type:     q.Type,
		Result:   q.Result,
	}, q.PageNo, q.PageSize)
}
