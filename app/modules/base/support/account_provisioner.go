package support

import (
	"context"
	"errors"

	"eqray-flavor/app/modules/base/types"
)

// AccountProvisionInput 是跨模块开通后台账号的最小入参。
// 密码由业务模块生成并以明文传入，账号模块负责哈希后落库，明文不保存。
type AccountProvisionInput struct {
	Username string
	Password string
	Nickname string
	Mobile   string
	Remark   string
	Status   int8
}

// AccountProvisioner 由持有后台账号数据的模块注入（当前为 system）。
// store 等业务模块通过该接口开通骑手账号，不直接依赖 system 的仓储或模型。
type AccountProvisioner interface {
	// CreateAccount 创建后台账号并返回账号 ID；账号或手机号重复时返回账号模块的业务错误。
	CreateAccount(ctx context.Context, in AccountProvisionInput) (types.ID, error)
}

var accountProvisioner AccountProvisioner

// RegisterAccountProvisioner 在应用启动时注入账号开通实现。
func RegisterAccountProvisioner(provisioner AccountProvisioner) {
	accountProvisioner = provisioner
}

// CreateAccount 开通后台账号，未注入实现时返回明确错误。
// 入参：in 账号、明文初始密码与资料；返回新账号 ID。
func CreateAccount(ctx context.Context, in AccountProvisionInput) (types.ID, error) {
	if accountProvisioner == nil {
		return 0, errors.New("账号开通能力未初始化")
	}

	return accountProvisioner.CreateAccount(ctx, in)
}
