package sms

import (
	"context"

	"eqray-flavor/app/facades"
)

// logSender 本地日志渠道（DebugLog）：不真实发送短信，只把文案写入日志，
// 供开发 / 测试环境使用，与 yudao 的 Debug 渠道一致。
type logSender struct{}

func (logSender) Send(_ context.Context, payload *Payload) error {
	facades.Log().Infof("[短信验证码] mobile=%s content=%s", payload.Mobile, payload.Content)

	return nil
}
