package service

import "strings"

// maskSecret 生成密钥脱敏值：空值返回空串，长度不超过 4 全部隐藏，
// 否则隐藏前段只保留末四位，如 ****abcd。
// 入参：密钥明文；返回脱敏字符串，可直接回显给前端。
func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	runes := []rune(secret)
	if len(runes) <= 4 {
		return "****"
	}

	return "****" + string(runes[len(runes)-4:])
}

// isMaskedValue 判断配置里的值是否为脱敏占位值（以 **** 开头），
// 编辑渠道时这类值表示沿用原密钥。
// 入参：配置项文本；返回是否为脱敏值。
func isMaskedValue(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), "****")
}
