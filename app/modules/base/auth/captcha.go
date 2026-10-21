package auth

import (
	"time"

	"github.com/mojocn/base64Captcha"

	"eqray-flavor/app/facades"
)

const (
	captchaHeight = 80
	captchaWidth  = 240
	captchaLength = 4
	captchaSkew   = 0.7
	captchaDots   = 80

	// 验证码答案有效期，与 base64Captcha 内存存储的默认值保持一致
	captchaTTL = 10 * time.Minute
	// Redis key 前缀，最终形如 captcha:<captchaId>
	captchaKeyPrefix = "captcha:"
)

var (
	captchaDriver                     = base64Captcha.NewDriverDigit(captchaHeight, captchaWidth, captchaLength, captchaSkew, captchaDots)
	captchaStore  base64Captcha.Store = redisStore{}
)

// redisStore 用 Redis 保存验证码答案：多实例部署时，生成与校验可能落在不同实例上，
// 进程内存储会直接校验失败。
type redisStore struct{}

// Set 写入验证码答案并设置过期时间。
func (redisStore) Set(id string, value string) error {
	return facades.Cache().Put(captchaKey(id), value, captchaTTL)
}

// Get 读取验证码答案；clear 为 true 时同时删除，保证一次校验后不可重放。
func (redisStore) Get(id string, clear bool) string {
	key := captchaKey(id)
	value := facades.Cache().GetString(key)
	if clear {
		facades.Cache().Forget(key)
	}

	return value
}

// Verify 比对答案，答案不存在（已过期或已用过）一律视为失败。
func (s redisStore) Verify(id, answer string, clear bool) bool {
	value := s.Get(id, clear)

	return value != "" && value == answer
}

// GenerateCaptcha 生成一张数字验证码，返回 captchaId 与可直接放进 img src 的 base64 图片。
// 答案存在 Redis，默认 10 分钟过期，校验一次即失效。
func GenerateCaptcha() (string, string, error) {
	c := base64Captcha.NewCaptcha(captchaDriver, captchaStore)
	id, image, _, err := c.Generate()
	if err != nil {
		return "", "", err
	}

	return id, image, nil
}

// VerifyCaptcha 校验验证码，无论成功与否都立即清除，避免重放。
func VerifyCaptcha(id, answer string) bool {
	if id == "" || answer == "" {
		return false
	}

	return captchaStore.Verify(id, answer, true)
}

// captchaKey 生成验证码在 Redis 中的键：前缀 + captchaId。
func captchaKey(id string) string {
	return captchaKeyPrefix + id
}
