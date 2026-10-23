package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strconv"
	"strings"
	"time"

	baseErrs "eqray-flavor/app/modules/base/errs"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payErrs "eqray-flavor/app/modules/pay/errs"
	"eqray-flavor/app/modules/pay/model"
)

// SubmitRequest 是渠道下单入参，金额单位为分。
type SubmitRequest struct {
	// No 支付单号，作为渠道的 out_trade_no
	No string
	// Subject 商品标题
	Subject string
	// Body 商品描述
	Body string
	// Price 支付金额（分）
	Price int64
	// NotifyURL 渠道异步通知地址
	NotifyURL string
	// OpenID 微信 JSAPI / 小程序支付必填
	OpenID string
	// AuthCode 用户付款码，支付宝条码支付（alipay_bar）必填
	AuthCode string
	// ClientIP 用户 IP，支付宝部分产品需要
	ClientIP string
	// Attach 附加数据，回调原样带回
	Attach string
}

// SubmitResult 是渠道下单结果。
type SubmitResult struct {
	// Extras 渠道额外参数 JSON，写入 pay_order_extension.channel_extras
	Extras string
	// PayParams 前端调起支付所需参数，按渠道不同（JSAPI 返回签名参数、Native 返回二维码链接）
	PayParams map[string]any
}

// RefundRequest 是渠道退款入参，金额单位为分。
type RefundRequest struct {
	// No 支付单号（渠道侧的 out_trade_no）
	No string
	// RefundNo 退款单号（渠道侧的 out_refund_no）
	RefundNo string
	// PayPrice 原支付金额（分）
	PayPrice int64
	// RefundPrice 本次退款金额（分）
	RefundPrice int64
	// Reason 退款原因
	Reason string
	// NotifyURL 渠道退款结果通知地址，微信需要，支付宝同步返回无需
	NotifyURL string
}

// NotifyResult 是渠道回调解析结果，支付回调与退款回调共用。
// 支付回调只有 OutTradeNo 等字段；退款回调额外带 RefundNo 与 RefundPrice。
type NotifyResult struct {
	// EventType 渠道事件类型，如微信的 TRANSACTION.SUCCESS / REFUND.SUCCESS
	EventType string
	// OutTradeNo 商户支付单号
	OutTradeNo string
	// ChannelOrderNo 渠道订单号
	ChannelOrderNo string
	// ChannelUserID 渠道用户标识，如微信 openid
	ChannelUserID string
	// Price 渠道回传的支付金额（分）
	Price int64
	// Success 渠道是否判定支付成功
	Success bool
	// RefundNo 商户退款单号，仅退款回调有值
	RefundNo string
	// RefundPrice 渠道回传的退款金额（分），仅退款回调有值
	RefundPrice int64
	// RefundSuccess 渠道是否判定退款成功，仅退款回调有意义
	RefundSuccess bool
	// RawData 回调原文，写入拓展单便于排查
	RawData string
}

// QueryRequest 是渠道查单 / 关单入参。
type QueryRequest struct {
	// No 支付单号（渠道侧的 out_trade_no）
	No string
}

// QueryResult 是渠道查单结果，用于补单与关单前的二次确认。
type QueryResult struct {
	// Paid 渠道是否已确认收款
	Paid bool
	// ChannelOrderNo 渠道订单号
	ChannelOrderNo string
	// ChannelUserID 渠道用户标识，如微信 openid
	ChannelUserID string
	// Price 渠道回传的支付金额（分）
	Price int64
}

// ChannelGateway 是渠道适配接口，微信 / 支付宝 / 模拟支付各自实现。
// 所有实现都不处理业务状态，只负责与渠道交互，状态推进由订单服务负责。
type ChannelGateway interface {
	// Submit 调用渠道下单，返回前端支付参数或错误
	Submit(ctx context.Context, req SubmitRequest) (*SubmitResult, error)
	// Refund 调用渠道退款，返回渠道退款单号与渠道是否已同步确认退款完成
	Refund(ctx context.Context, req RefundRequest) (string, bool, error)
	// Query 查询渠道订单状态，用于补单与关单前的二次确认
	Query(ctx context.Context, req QueryRequest) (*QueryResult, error)
	// Close 关闭渠道订单，关闭后用户无法再付款
	Close(ctx context.Context, req QueryRequest) error
	// ParseNotify 验签并解析渠道回调，返回统一结果
	ParseNotify(ctx context.Context, body []byte, headers map[string]string, form map[string]string) (*NotifyResult, error)
}

// channelGatewayFactory 按渠道编码创建适配器，便于按渠道注册扩展。
type channelGatewayFactory func(channel *model.PayChannelModel) (ChannelGateway, error)

// channelGatewayFactories 保存渠道编码到适配器的映射。
var channelGatewayFactories = map[string]channelGatewayFactory{
	payEnums.ChannelWxPub:     newWechatGateway,
	payEnums.ChannelWxLite:    newWechatGateway,
	payEnums.ChannelWxApp:     newWechatGateway,
	payEnums.ChannelWxNative:  newWechatGateway,
	payEnums.ChannelWxWap:     newWechatGateway,
	payEnums.ChannelAlipayPc:  newAlipayGateway,
	payEnums.ChannelAlipayWap: newAlipayGateway,
	payEnums.ChannelAlipayApp: newAlipayGateway,
	payEnums.ChannelAlipayQr:  newAlipayGateway,
	payEnums.ChannelAlipayBar: newAlipayGateway,
	payEnums.ChannelMock:      newMockGateway,
}

// NewChannelGateway 按渠道编码创建适配器。
// 入参：渠道模型（config 为渠道参数 JSON）；返回适配器，编码不支持时报错。
func NewChannelGateway(channel *model.PayChannelModel) (ChannelGateway, error) {
	factory, ok := channelGatewayFactories[channel.Code]
	if !ok {
		return nil, baseErrs.New(payErrs.PayChannelCodeInvalid, channel.Code)
	}

	return factory(channel)
}

// parseChannelConfig 把渠道 config JSON 解析到目标结构。
// 入参：config 文本与目标指针；返回错误表示 JSON 不合法。
func parseChannelConfig(raw string, target any) error {
	if strings.TrimSpace(raw) == "" {
		return baseErrs.New(payErrs.PayChannelConfigInvalid)
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return baseErrs.New(payErrs.PayChannelConfigInvalid)
	}

	return nil
}

// mockGateway 是模拟支付适配器，下单直接成功，用于本地联调与演示。
type mockGateway struct{}

// newMockGateway 构造模拟支付适配器。
func newMockGateway(channel *model.PayChannelModel) (ChannelGateway, error) {
	return &mockGateway{}, nil
}

// Submit 模拟下单：不出网，直接返回空支付参数，由订单服务立即置为已支付。
func (g *mockGateway) Submit(ctx context.Context, req SubmitRequest) (*SubmitResult, error) {
	return &SubmitResult{
		Extras:    `{"mock":true}`,
		PayParams: map[string]any{"mock": true},
	}, nil
}

// Refund 模拟退款：直接返回以退款单号为后缀的渠道退款号，并视为退款完成。
func (g *mockGateway) Refund(ctx context.Context, req RefundRequest) (string, bool, error) {
	return "MOCK" + req.RefundNo, true, nil
}

// Query 模拟查单：无法远端核对，统一返回未支付，由调用方按本地状态处理。
func (g *mockGateway) Query(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	return &QueryResult{Paid: false}, nil
}

// Close 模拟关单：无远端订单，直接视为关闭成功。
func (g *mockGateway) Close(ctx context.Context, req QueryRequest) error {
	return nil
}

// ParseNotify 模拟回调：不参与验签，返回失败表示 mock 渠道不走回调。
func (g *mockGateway) ParseNotify(ctx context.Context, body []byte, headers map[string]string, form map[string]string) (*NotifyResult, error) {
	return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
}

// parsePrivateKey 解析 RSA 私钥 PEM，兼容 PKCS#8 与 PKCS#1。
// 入参：私钥 PEM 文本；返回私钥对象，格式不合法时报业务错误。
func parsePrivateKey(content string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(content)))
	if block == nil {
		return nil, baseErrs.New(payErrs.PayChannelConfigMissing, "RSA 私钥", "PEM 解析失败")
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, baseErrs.New(payErrs.PayChannelConfigMissing, "RSA 私钥", "仅支持 PKCS#8 / PKCS#1")
}

// rsaSignSHA256 用私钥对内容做 SHA256withRSA 签名并 base64 编码。
// 入参：私钥与待签名内容；返回签名串。
func rsaSignSHA256(privateKey *rsa.PrivateKey, message string) (string, error) {
	digest := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// randomNonce 生成 32 位十六进制随机串，用于渠道签名防重放。
func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return strings.ToLower(base64.RawURLEncoding.EncodeToString(buf))[:32], nil
}

// centToYuan 把金额（分）格式化成两位小数的元，支付宝接口要求元字符串。
// 入参：金额（分）；返回如 "10.00"。
func centToYuan(cent int64) string {
	sign := ""
	if cent < 0 {
		sign = "-"
		cent = -cent
	}

	return sign + strconv.FormatInt(cent/100, 10) + "." + leftPad2(strconv.FormatInt(cent%100, 10))
}

// leftPad2 把 0~99 的整数补足两位。
func leftPad2(value string) string {
	if len(value) == 1 {
		return "0" + value
	}

	return value
}

// nowTimestamp 返回当前秒级时间戳字符串。
func nowTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// errGateway 把底层 SDK 错误包装成业务错误，避免泄露内部实现。
// 入参：错误；返回带渠道失败提示的业务错误。
func errGateway(err error) error {
	if err == nil {
		return nil
	}

	return baseErrs.New(payErrs.PayOrderSubmitFail, trimMessage(err.Error()))
}

// trimMessage 截断过长的渠道错误文本，避免把 SDK 长报文直接回给前端。
// 入参：原始错误文本；返回最多 160 个字符的文本。
func trimMessage(message string) string {
	if len([]rune(message)) > 160 {
		return string([]rune(message)[:160])
	}

	return message
}

// gatewayError 是渠道错误的轻量实现，避免直接把 SDK 错误抛到上层。
type gatewayError struct {
	// message 错误描述
	message string
}

// Error 实现 error 接口，返回渠道错误描述。
func (e *gatewayError) Error() string {
	return e.message
}

// gatewayErr 构造渠道错误，用于渠道未返回业务数据等无底层错误的场景。
// 入参：错误描述；返回实现 error 的渠道错误。
func gatewayErr(message string) error {
	return &gatewayError{message: message}
}

// errorsAs 兼容包装 errors.As 的调用点，保持 service 内错误判断集中。
func errorsAs(err error, target any) bool {
	return errors.As(err, target)
}
