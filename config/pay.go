package config

import (
	"eqray-flavor/app/facades"
)

// init 注册支付模块配置。
// 商户号、密钥与证书已改为数据库配置：在后台「支付渠道」页维护 pay_channel.config，
// 下单与回调验签都按渠道主键取对应配置，因此这里不再放 wechat / alipay 参数。
// 唯一保留的是渠道回调的基础地址，用于拼装：
//
//	{notify_base_url}/api/v1/pay/callbacks/order/{channelId}
//	{notify_base_url}/api/v1/pay/callbacks/refund/{channelId}
//
// 改配置需要重启服务。
func init() {
	facades.Config().Add("pay", map[string]any{
		// 支付回调基础地址，如 https://api.example.com；为空时下单直接报错，不会静默成功
		"notify_base_url": facades.Config().Env("PAY_NOTIFY_BASE_URL", ""),
	})
}
