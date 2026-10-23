package controller

import (
	"strconv"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
)

// baseStatus 把请求里的 int8 状态转换为通用启停枚举，取值范围由请求校验保证。
func baseStatus(status int8) baseEnums.CommonStatusEnum {
	return baseEnums.CommonStatusEnum(status)
}

// parseID 解析雪花 ID 查询参数，空值或非法值按 0（表示不限）处理。
func parseID(raw string) types.ID {
	if raw == "" {
		return 0
	}

	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0
	}

	return types.ID(value)
}

// badRequest 构造带中文文案的参数错误，避免在控制器里拼接错误码。
func badRequest(message string) error {
	return baseErrs.WithMsg(baseErrs.BadRequest, message)
}
