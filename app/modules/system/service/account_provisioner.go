package service

import (
	"context"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
)

// AccountProvisioner 实现 base 的账号开通接口，供 store 等业务模块自动创建骑手账号。
// 账号与手机号唯一性、bcrypt 密码哈希全部复用 system 的新增用户逻辑。
type AccountProvisioner struct {
	users *UserService
}

// NewAccountProvisioner 构造账号开通实现。
func NewAccountProvisioner() *AccountProvisioner {
	return &AccountProvisioner{users: NewUserService()}
}

// CreateAccount 创建后台账号并返回账号 ID。
// 入参：in 账号、明文初始密码与资料；返回新账号 ID，重复时返回 system 模块的业务错误。
func (p *AccountProvisioner) CreateAccount(ctx context.Context, in support.AccountProvisionInput) (types.ID, error) {
	return p.users.Create(ctx, UserSaveInput{
		Username: in.Username,
		Password: in.Password,
		Nickname: in.Nickname,
		Remark:   in.Remark,
		Mobile:   in.Mobile,
		Status:   baseEnums.CommonStatusEnum(in.Status),
	})
}
