// Package sms 抽象短信发送渠道：渠道与模板配置在数据库（对齐 yudao 的
// system_sms_channel / system_sms_template），这里只负责按渠道代码选择官方 SDK 实现。
package sms

import (
	"context"

	baseErrs "eqray-flavor/app/modules/base/errs"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
)

// Payload 是一次短信发送的内容。
type Payload struct {
	// Mobile 手机号（不带国家码）
	Mobile string
	// Content 渲染后的完整文案，日志渠道直接输出
	Content string
	// Params 模板变量，键为变量名，阿里云用 JSON 序列化后传给 TemplateParam
	Params map[string]string
	// ParamList 按模板 Params 顺序排列的变量值，腾讯云 TemplateParamSet 要求有序
	ParamList []string
}

// Sender 按渠道发送短信。
type Sender interface {
	Send(ctx context.Context, payload *Payload) error
}

// New 按渠道代码构造发送器；template 提供第三方平台上的模板 ID。
func New(channel *model.InfraSmsChannelModel, template *model.InfraSmsTemplateModel) (Sender, error) {
	switch channel.Code {
	case model.InfraSmsChannelDebugLog:
		return logSender{}, nil
	case model.InfraSmsChannelAliyun:
		return &aliyunSender{channel: channel, template: template}, nil
	case model.InfraSmsChannelTencent:
		return &tencentSender{channel: channel, template: template}, nil
	default:
		return nil, baseErrs.New(infraErrs.InfraSmsChannelNotExists, channel.Code)
	}
}
