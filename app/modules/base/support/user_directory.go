package support

import (
	"context"
	"errors"

	"eqray-flavor/app/modules/base/types"
)

// UserSummary 是跨模块只读账号摘要，供员工分配等场景展示与选择。
// 只包含展示所需字段，不暴露密码、角色等敏感信息。
type UserSummary struct {
	ID       types.ID
	Username string
	Nickname string
	Mobile   string
	Status   int8
}

// UserDirectory 由持有后台账号数据的模块注入（当前为 system）。
// store 等业务模块只读账号摘要，不直接依赖 system 的仓储或模型。
type UserDirectory interface {
	// FindUsers 按 ID 批量查询账号摘要，未命中的 ID 直接忽略。
	FindUsers(ctx context.Context, ids []types.ID) ([]UserSummary, error)
	// SearchUsers 按账号 / 昵称 / 手机号搜索启用账号，limit 为最大返回数。
	SearchUsers(ctx context.Context, keyword string, limit int) ([]UserSummary, error)
}

var userDirectory UserDirectory

// RegisterUserDirectory 在应用启动时注入账号目录实现。
func RegisterUserDirectory(directory UserDirectory) {
	userDirectory = directory
}

// FindUserSummaries 批量查询账号摘要，未注入实现时返回明确错误。
func FindUserSummaries(ctx context.Context, ids []types.ID) ([]UserSummary, error) {
	if userDirectory == nil {
		return nil, errors.New("账号目录未初始化")
	}

	return userDirectory.FindUsers(ctx, ids)
}

// SearchUserSummaries 搜索启用账号摘要，未注入实现时返回明确错误。
func SearchUserSummaries(ctx context.Context, keyword string, limit int) ([]UserSummary, error) {
	if userDirectory == nil {
		return nil, errors.New("账号目录未初始化")
	}

	return userDirectory.SearchUsers(ctx, keyword, limit)
}
