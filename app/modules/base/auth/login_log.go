package auth

import (
	"context"
	"strings"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
)

const loginLogTable = "auth_login_log"

// LoginLog 对应 auth_login_log，由 base 认证组件独占读写（docs 6.2）。
type LoginLog struct {
	ID        types.ID                  `db:"id" json:"id"`
	UserID    types.ID                  `db:"user_id" json:"userId"`
	UserType  baseEnums.UserTypeEnum    `db:"user_type" json:"userType"`
	Username  string                    `db:"username" json:"username"`
	Type      baseEnums.LoginTypeEnum   `db:"type" json:"type"`
	Result    baseEnums.LoginResultEnum `db:"result" json:"result"`
	UserIP    string                    `db:"user_ip" json:"userIp"`
	UserAgent string                    `db:"user_agent" json:"userAgent"`
	CreatedAt types.DateTime            `db:"created_at" json:"createdAt"`
}

// LoginLogInput 是写入登录日志的入参。
type LoginLogInput struct {
	UserID    types.ID
	UserType  baseEnums.UserTypeEnum
	Username  string
	Type      baseEnums.LoginTypeEnum
	Result    baseEnums.LoginResultEnum
	IP        string
	UserAgent string
}

// LoginLogFilter 是登录日志的筛选条件，零值表示不限制。
type LoginLogFilter struct {
	Username string
	Type     *baseEnums.LoginTypeEnum
	Result   *baseEnums.LoginResultEnum
}

// RecordLoginLog 写一条登录 / 登出日志。
// 日志写入失败只记框架日志，不影响登录主流程。
func RecordLoginLog(ctx context.Context, in LoginLogInput) {
	if !in.Type.IsValid() {
		in.Type = baseEnums.LoginTypeLogin
	}
	if !in.Result.IsValid() {
		in.Result = baseEnums.LoginResultFail
	}
	if !in.UserType.IsValid() {
		in.UserType = baseEnums.UserTypeAdmin
	}

	_, err := facades.DB().WithContext(ctx).Table(loginLogTable).Insert(map[string]any{
		"id":         support.NewID(),
		"user_id":    in.UserID,
		"user_type":  in.UserType,
		"username":   truncate(strings.TrimSpace(in.Username), 64),
		"type":       in.Type,
		"result":     in.Result,
		"user_ip":    truncate(strings.TrimSpace(in.IP), 50),
		"user_agent": truncate(strings.TrimSpace(in.UserAgent), 512),
		"created_at": types.Now(),
	})
	if err != nil {
		facades.Log().WithContext(ctx).Warning("auth_login_log write failed: " + err.Error())
	}
}

// PageLoginLogs 分页查询登录日志，按时间倒序。
func PageLoginLogs(ctx context.Context, filter LoginLogFilter, pageNo, pageSize int) ([]LoginLog, int64, error) {
	list := make([]LoginLog, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(loginLogTable)
	if filter.Username != "" {
		query = query.WhereLike("username", "%"+filter.Username+"%")
	}
	if filter.Type != nil {
		query = query.Where("type", *filter.Type)
	}
	if filter.Result != nil {
		query = query.Where("result", *filter.Result)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// truncate 按字符截断，避免超长 UA 写库报错。
func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}

	return string(runes[:limit])
}
