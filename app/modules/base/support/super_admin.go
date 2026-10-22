package support

import (
	"context"

	"eqray-flavor/app/modules/base/types"
)

// SuperAdminChecker 由持有角色数据的模块注入（当前为 system），
// 用于在基础能力里判断超管，避免 store 等业务模块直接依赖 system。
type SuperAdminChecker func(ctx context.Context, userID types.ID) (bool, error)

var superAdminChecker SuperAdminChecker

// RegisterSuperAdminChecker 在应用启动时注入超管判断实现。
func RegisterSuperAdminChecker(checker SuperAdminChecker) {
	superAdminChecker = checker
}

// IsSuperAdmin 判断当前用户是否为内置超管。
// 未注入实现时按非超管处理，保证缺装配时不会意外放行全部门店。
func IsSuperAdmin(ctx context.Context, userID types.ID) (bool, error) {
	if superAdminChecker == nil || userID <= 0 {
		return false, nil
	}

	return superAdminChecker(ctx, userID)
}
