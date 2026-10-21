package sms

import (
	"context"
	"encoding/json"

	tcCommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcProfile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcRegions "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	tcSms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	baseErrs "eqray-flavor/app/modules/base/errs"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
)

// tencentRegion 默认地域，可在渠道 extra 里覆盖。
const tencentRegion = tcRegions.Guangzhou

// tencentExtra 渠道 extra 的 JSON 结构。
type tencentExtra struct {
	Region   string `json:"region"`
	SdkAppID string `json:"sdkAppId"`
}

// tencentSender 腾讯云短信渠道，基于官方 SDK sms v20210111。
type tencentSender struct {
	channel  *model.InfraSmsChannelModel
	template *model.InfraSmsTemplateModel
}

func (s *tencentSender) Send(_ context.Context, payload *Payload) error {
	client, err := tcSms.NewClient(
		tcCommon.NewCredential(s.channel.APIKey, s.channel.APISecret),
		s.region(),
		tcProfile.NewClientProfile(),
	)
	if err != nil {
		return baseErrs.New(infraErrs.InfraSmsSendFailed, err.Error())
	}

	req := tcSms.NewSendSmsRequest()
	req.SmsSdkAppId = tcCommon.StringPtr(s.sdkAppID())
	req.SignName = tcCommon.StringPtr(s.channel.Signature)
	req.TemplateId = tcCommon.StringPtr(s.template.APITemplateID)
	req.TemplateParamSet = tcCommon.StringPtrs(payload.ParamList)
	req.PhoneNumberSet = tcCommon.StringPtrs([]string{"+86" + payload.Mobile})

	resp, err := client.SendSms(req)
	if err != nil {
		return baseErrs.New(infraErrs.InfraSmsSendFailed, err.Error())
	}
	if resp.Response == nil || len(resp.Response.SendStatusSet) == 0 {
		return baseErrs.New(infraErrs.InfraSmsSendFailed, "短信平台未返回发送状态")
	}

	status := resp.Response.SendStatusSet[0]
	if status.Code == nil || strPtr(status.Code) != "Ok" {
		message := ""
		if status.Message != nil {
			message = strPtr(status.Message)
		}

		return baseErrs.New(infraErrs.InfraSmsSendFailed, message)
	}

	return nil
}

// region 返回地域，优先取渠道 extra 配置。
func (s *tencentSender) region() string {
	extra := s.extra()
	if extra.Region != "" {
		return extra.Region
	}

	return tencentRegion
}

// sdkAppID 返回短信应用 ID，来自渠道 extra。
func (s *tencentSender) sdkAppID() string {
	return s.extra().SdkAppID
}

func (s *tencentSender) extra() tencentExtra {
	var extra tencentExtra
	if s.channel.Extra == "" {
		return extra
	}
	_ = json.Unmarshal([]byte(s.channel.Extra), &extra)

	return extra
}

// strPtr 安全解引用字符串指针，nil 时返回空串。
func strPtr(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}
