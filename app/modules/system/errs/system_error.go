// Package errs 定义 system 模块的业务错误码，编码规则见 docs 8.3(2)。
package errs

import baseErrs "eqray-flavor/app/modules/base/errs"

// system 模块业务错误码：1-002-{子模块}-{错误码}，auth 占用 001 子模块。
var (
	SystemCaptchaWrong    = baseErrs.ErrorCode{Code: 1_002_001_001, Msg: "验证码错误或已过期"}
	SystemLoginFailed     = baseErrs.ErrorCode{Code: 1_002_001_002, Msg: "账号或密码错误"}
	SystemAccountDisabled = baseErrs.ErrorCode{Code: 1_002_001_003, Msg: "账号已被停用"}

	// user 子模块
	SystemUserNotExists    = baseErrs.ErrorCode{Code: 1_002_002_001, Msg: "用户【%v】不存在"}
	SystemUsernameExists   = baseErrs.ErrorCode{Code: 1_002_002_002, Msg: "账号【%s】已存在"}
	SystemEmailExists      = baseErrs.ErrorCode{Code: 1_002_002_003, Msg: "邮箱【%s】已被占用"}
	SystemMobileExists     = baseErrs.ErrorCode{Code: 1_002_002_004, Msg: "手机号【%s】已被占用"}
	SystemUserDeleteSelf   = baseErrs.ErrorCode{Code: 1_002_002_005, Msg: "不能删除当前登录账号"}
	SystemUserUpdateSelf   = baseErrs.ErrorCode{Code: 1_002_002_006, Msg: "不能停用当前登录账号"}
	SystemUserDataExists   = baseErrs.ErrorCode{Code: 1_002_002_007, Msg: "账号、邮箱或手机号已存在"}
	SystemOldPasswordWrong = baseErrs.ErrorCode{Code: 1_002_002_008, Msg: "当前密码不正确"}

	// menu 子模块
	SystemMenuNotExists     = baseErrs.ErrorCode{Code: 1_002_004_001, Msg: "菜单【%v】不存在"}
	SystemMenuHasChildren   = baseErrs.ErrorCode{Code: 1_002_004_002, Msg: "存在子菜单，不可删除"}
	SystemMenuParentInvalid = baseErrs.ErrorCode{Code: 1_002_004_003, Msg: "上级菜单不存在或不是目录 / 菜单"}
	SystemMenuParentSelf    = baseErrs.ErrorCode{Code: 1_002_004_004, Msg: "上级菜单不能是自己或自己的子级"}

	// role 子模块
	SystemRoleInUse      = baseErrs.ErrorCode{Code: 1_002_003_001, Msg: "角色已被用户引用，不可删除"}
	SystemRoleNotExists  = baseErrs.ErrorCode{Code: 1_002_003_002, Msg: "角色【%v】不存在"}
	SystemRoleCodeExists = baseErrs.ErrorCode{Code: 1_002_003_003, Msg: "角色标识【%s】已存在"}
	SystemRoleBuiltin    = baseErrs.ErrorCode{Code: 1_002_003_004, Msg: "内置角色不允许删除或修改标识"}

	// dict 子模块
	SystemDictTypeNotExists = baseErrs.ErrorCode{Code: 1_002_005_001, Msg: "字典类型【%v】不存在"}
	SystemDictTypeExists    = baseErrs.ErrorCode{Code: 1_002_005_002, Msg: "字典类型标识【%s】已存在"}
	SystemDictTypeImmutable = baseErrs.ErrorCode{Code: 1_002_005_003, Msg: "字典类型标识创建后不可修改"}
	SystemDictDataNotExists = baseErrs.ErrorCode{Code: 1_002_005_004, Msg: "字典数据【%v】不存在"}
	SystemDictValueExists   = baseErrs.ErrorCode{Code: 1_002_005_005, Msg: "字典键值【%s】已存在"}
)
