package sms

import (
	"context"
	"encoding/json"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	baseErrs "eqray-flavor/app/modules/base/errs"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
)

// aliyunEndpoint 阿里云短信服务接入点。
const aliyunEndpoint = "dysmsapi.aliyuncs.com"

// aliyunSender 阿里云短信渠道，基于官方 SDK dysmsapi。
type aliyunSender struct {
	channel  *model.InfraSmsChannelModel
	template *model.InfraSmsTemplateModel
}

func (s *aliyunSender) Send(_ context.Context, payload *Payload) error {
	client, err := dysmsapi.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(s.channel.APIKey),
		AccessKeySecret: tea.String(s.channel.APISecret),
		Endpoint:        tea.String(aliyunEndpoint),
	})
	if err != nil {
		return baseErrs.New(infraErrs.InfraSmsSendFailed, err.Error())
	}

	paramJSON, err := json.Marshal(payload.Params)
	if err != nil {
		return baseErrs.New(infraErrs.InfraSmsSendFailed, "模板参数序列化失败")
	}

	resp, err := client.SendSms(&dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(payload.Mobile),
		SignName:      tea.String(s.channel.Signature),
		TemplateCode:  tea.String(s.template.APITemplateID),
		TemplateParam: tea.String(string(paramJSON)),
	})
	if err != nil {
		return baseErrs.New(infraErrs.InfraSmsSendFailed, err.Error())
	}
	if resp.Body == nil || resp.Body.Code == nil || tea.StringValue(resp.Body.Code) != "OK" {
		message := ""
		if resp.Body != nil && resp.Body.Message != nil {
			message = tea.StringValue(resp.Body.Message)
		}

		return baseErrs.New(infraErrs.InfraSmsSendFailed, message)
	}

	return nil
}
