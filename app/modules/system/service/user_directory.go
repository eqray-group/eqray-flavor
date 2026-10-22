package service

import (
	"context"

	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
	"eqray-flavor/app/modules/system/repository"
)

// UserDirectory 实现 base 的只读账号目录接口，供 store 等业务模块查询账号摘要。
// 只暴露展示字段，不返回密码、角色等敏感信息。
type UserDirectory struct {
	users *repository.UserRepository
}

// NewUserDirectory 构造账号目录。
func NewUserDirectory() *UserDirectory {
	return &UserDirectory{users: repository.NewUserRepository()}
}

// FindUsers 按 ID 批量查询账号摘要，未命中的 ID 直接忽略。
// 入参：ids 用户 ID 集合；返回账号摘要列表。
func (d *UserDirectory) FindUsers(ctx context.Context, ids []types.ID) ([]support.UserSummary, error) {
	rows, err := d.users.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return toUserSummaries(rows), nil
}

// SearchUsers 按账号 / 昵称 / 手机号搜索启用账号，供员工选择等场景使用。
// 入参：keyword 关键字、limit 最大返回数；返回账号摘要列表。
func (d *UserDirectory) SearchUsers(ctx context.Context, keyword string, limit int) ([]support.UserSummary, error) {
	rows, err := d.users.SearchEnabled(ctx, keyword, limit)
	if err != nil {
		return nil, err
	}

	return toUserSummaries(rows), nil
}

// toUserSummaries 把 system_users 模型转换为跨模块账号摘要。
func toUserSummaries(rows []model.SystemUserModel) []support.UserSummary {
	list := make([]support.UserSummary, 0, len(rows))
	for _, row := range rows {
		mobile := ""
		if row.Mobile != nil {
			mobile = *row.Mobile
		}
		list = append(list, support.UserSummary{
			ID:       row.ID,
			Username: row.Username,
			Nickname: row.Nickname,
			Mobile:   mobile,
			Status:   int8(row.Status),
		})
	}

	return list
}
