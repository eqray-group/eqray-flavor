package service

import (
	"context"
	"strings"

	baseErrs "eqray-flavor/app/modules/base/errs"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	infraModel "eqray-flavor/app/modules/infra/model"
	"eqray-flavor/app/modules/infra/repository"
	"eqray-flavor/app/modules/infra/support/sms"
)

// InfraSmsService 提供按模板代码发送短信的能力，供 member 等业务模块复用。
type InfraSmsService struct {
	templates *repository.InfraSmsTemplateRepository
	channels  *repository.InfraSmsChannelRepository
}

func NewInfraSmsService() *InfraSmsService {
	return &InfraSmsService{
		templates: repository.NewInfraSmsTemplateRepository(),
		channels:  repository.NewInfraSmsChannelRepository(),
	}
}

// Send 按模板代码发送短信：模板 → 模板绑定的渠道 → 对应官方 SDK。
// values 的键需与模板 params 声明的变量名一致。
func (s *InfraSmsService) Send(ctx context.Context, templateCode, mobile string, values map[string]string) error {
	template, err := s.templates.FindEnabledByCode(ctx, templateCode)
	if err != nil {
		return err
	}
	if template == nil {
		return baseErrs.New(infraErrs.InfraSmsTemplateNotExists, templateCode)
	}

	channel, err := s.channels.FindEnabledByID(ctx, template.ChannelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return baseErrs.New(infraErrs.InfraSmsChannelNotExists, template.ChannelID)
	}

	sender, err := sms.New(channel, template)
	if err != nil {
		return err
	}

	return sender.Send(ctx, &sms.Payload{
		Mobile:    mobile,
		Content:   renderSmsContent(template, values),
		Params:    values,
		ParamList: buildParamList(template.Params, values),
	})
}

// renderSmsContent 用变量值替换模板里的 {name} 占位符，供日志渠道输出完整文案。
func renderSmsContent(template *infraModel.InfraSmsTemplateModel, values map[string]string) string {
	content := template.Content
	for name, value := range values {
		content = strings.ReplaceAll(content, "{"+name+"}", value)
	}

	return content
}

// buildParamList 按模板 params 声明顺序取值，腾讯云要求参数有序。
func buildParamList(params string, values map[string]string) []string {
	if params == "" {
		return nil
	}

	names := strings.Split(params, ",")
	list := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		list = append(list, values[name])
	}

	return list
}
