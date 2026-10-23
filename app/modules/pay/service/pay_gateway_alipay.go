package service

import (
	"context"
	"encoding/json"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-pay/gopay"
	alipay "github.com/go-pay/gopay/alipay"

	baseErrs "eqray-flavor/app/modules/base/errs"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payErrs "eqray-flavor/app/modules/pay/errs"
	"eqray-flavor/app/modules/pay/model"
)

// alipayChannelConfig 是支付宝渠道的 config JSON 结构，字段名与后台配置页保持一致。
type alipayChannelConfig struct {
	// AppID 支付宝应用 appId
	AppID string `json:"appId"`
	// PrivateKey 应用私钥 PEM，支持 PKCS#1 与 PKCS#8
	PrivateKey string `json:"privateKey"`
	// AlipayPublicKey 支付宝公钥 PEM，用于异步通知验签；留空则拒绝一切回调
	AlipayPublicKey string `json:"alipayPublicKey"`
	// SellerID 收款支付宝账号 PID，多商户场景填写，单商户可留空
	SellerID string `json:"sellerId"`
	// IsProd 是否生产环境，false 时走新版沙箱网关
	IsProd bool `json:"isProd"`
}

// alipayGateway 基于 gopay 的支付宝开放平台适配器。
type alipayGateway struct {
	// code 渠道编码，决定调用哪个下单接口
	code string
	// cfg 渠道配置
	cfg *alipayChannelConfig
	// client gopay 支付宝客户端
	client *alipay.Client
}

// newAlipayGateway 解析支付宝渠道配置并构造适配器。
// 入参：渠道模型；返回适配器，配置缺项或私钥非法时报业务错误。
func newAlipayGateway(channel *model.PayChannelModel) (ChannelGateway, error) {
	cfg := &alipayChannelConfig{}
	if err := parseChannelConfig(channel.Config, cfg); err != nil {
		return nil, err
	}

	label := payEnums.ChannelLabel(channel.Code)
	if cfg.AppID == "" || cfg.PrivateKey == "" {
		return nil, baseErrs.New(payErrs.PayChannelConfigMissing, label, "appId / privateKey")
	}

	client, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return nil, baseErrs.New(payErrs.PayChannelConfigMissing, label, trimMessage(err.Error()))
	}

	return &alipayGateway{code: channel.Code, cfg: cfg, client: client}, nil
}

// Submit 按渠道编码调用支付宝下单接口。
// 入参：支付单号、标题、金额（分）、回调地址、付款码；返回前端支付参数（支付链接 / orderStr / 二维码 / 交易号）。
func (g *alipayGateway) Submit(ctx context.Context, req SubmitRequest) (*SubmitResult, error) {
	bm := g.baseBody(req)

	switch g.code {
	case payEnums.ChannelAlipayPc:
		payURL, err := g.client.TradePagePay(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}

		return &SubmitResult{
			Extras:    marshalExtras(map[string]any{"payUrl": payURL}),
			PayParams: map[string]any{"payUrl": payURL},
		}, nil

	case payEnums.ChannelAlipayWap:
		payURL, err := g.client.TradeWapPay(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}

		return &SubmitResult{
			Extras:    marshalExtras(map[string]any{"payUrl": payURL}),
			PayParams: map[string]any{"payUrl": payURL},
		}, nil

	case payEnums.ChannelAlipayApp:
		orderStr, err := g.client.TradeAppPay(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}

		return &SubmitResult{
			Extras:    marshalExtras(map[string]any{"orderStr": orderStr}),
			PayParams: map[string]any{"orderStr": orderStr},
		}, nil

	case payEnums.ChannelAlipayQr:
		rsp, err := g.client.TradePrecreate(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}
		if rsp.Response == nil {
			return nil, errGateway(gatewayErr("渠道未返回业务数据"))
		}

		return &SubmitResult{
			Extras:    marshalExtras(map[string]any{"qrCode": rsp.Response.QrCode}),
			PayParams: map[string]any{"qrCode": rsp.Response.QrCode},
		}, nil

	default:
		// 条码支付需要用户在收银台出示的付款码，缺失时无法发起交易
		if req.AuthCode == "" {
			return nil, baseErrs.New(payErrs.PayChannelConfigMissing, payEnums.ChannelLabel(g.code), "authCode")
		}
		bm.Set("scene", "bar_code").Set("auth_code", req.AuthCode)

		rsp, err := g.client.TradePay(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}
		if rsp.Response == nil {
			return nil, errGateway(gatewayErr("渠道未返回业务数据"))
		}

		return &SubmitResult{
			Extras: marshalExtras(map[string]any{
				"tradeNo":    rsp.Response.TradeNo,
				"buyerLogon": rsp.Response.BuyerLogonId,
			}),
			PayParams: map[string]any{"tradeNo": rsp.Response.TradeNo},
		}, nil
	}
}

// baseBody 组装支付宝公共下单参数：单号、金额（转元）、标题、回调地址与附加数据。
// 入参：渠道下单请求；返回可继续追加渠道专有字段的 BodyMap。
func (g *alipayGateway) baseBody(req SubmitRequest) gopay.BodyMap {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.No).
		Set("total_amount", centToYuan(req.Price)).
		Set("subject", req.Subject)
	if req.NotifyURL != "" {
		bm.Set("notify_url", req.NotifyURL)
	}
	if req.Body != "" {
		bm.Set("body", req.Body)
	}
	if req.Attach != "" {
		// 支付宝把 passback_params 原样回传，与微信 attach 语义一致
		bm.Set("passback_params", req.Attach)
	}
	if g.cfg.SellerID != "" {
		bm.Set("seller_id", g.cfg.SellerID)
	}

	return bm
}

// Refund 调用支付宝统一收单退款接口。
// 入参：支付单号、退款单号、原支付金额与退款金额（分）、退款原因；
// 返回渠道交易号（缺失时回退退款单号）与 fund_change 是否为 Y（支付宝同步确认退款完成）。
func (g *alipayGateway) Refund(ctx context.Context, req RefundRequest) (string, bool, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.No).
		Set("out_request_no", req.RefundNo).
		Set("refund_amount", centToYuan(req.RefundPrice))
	if req.Reason != "" {
		bm.Set("refund_reason", req.Reason)
	}

	rsp, err := g.client.TradeRefund(ctx, bm)
	if err != nil {
		return "", false, baseErrs.New(payErrs.PayRefundSubmitFail, trimMessage(err.Error()))
	}
	if rsp.Response == nil {
		return "", false, baseErrs.New(payErrs.PayRefundSubmitFail, "渠道未返回退款结果")
	}

	refundNo := rsp.Response.TradeNo
	if refundNo == "" {
		refundNo = req.RefundNo
	}

	// 支付宝退款是同步接口，fund_change=Y 才表示资金真的发生变动
	return refundNo, rsp.Response.FundChange == "Y", nil
}

// Query 查询支付宝订单状态，用于补单与关单前的二次确认。
// 入参：支付单号；返回是否已支付、渠道交易号、付款账号与支付金额（分）。
func (g *alipayGateway) Query(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.No)

	rsp, err := g.client.TradeQuery(ctx, bm)
	if err != nil {
		return nil, baseErrs.New(payErrs.PayOrderQueryFail, trimMessage(err.Error()))
	}
	if rsp.Response == nil {
		return nil, baseErrs.New(payErrs.PayOrderQueryFail, "渠道未返回查单结果")
	}

	status := rsp.Response.TradeStatus

	return &QueryResult{
		Paid:           status == "TRADE_SUCCESS" || status == "TRADE_FINISHED",
		ChannelOrderNo: rsp.Response.TradeNo,
		ChannelUserID:  rsp.Response.BuyerLogonId,
		Price:          yuanToCent(rsp.Response.TotalAmount),
	}, nil
}

// Close 关闭支付宝订单，关闭后用户无法再对该单付款。
// 入参：支付单号；返回错误表示渠道拒绝关单（如订单已支付）。
func (g *alipayGateway) Close(ctx context.Context, req QueryRequest) error {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.No)

	if _, err := g.client.TradeClose(ctx, bm); err != nil {
		return baseErrs.New(payErrs.PayOrderCloseFail, trimMessage(err.Error()))
	}

	return nil
}

// ParseNotify 校验支付宝异步通知签名并解析支付结果。
// 入参：原始请求体与已解析的表单参数（表单为空时回退解析请求体）；返回统一通知结果。
func (g *alipayGateway) ParseNotify(ctx context.Context, body []byte, headers map[string]string, form map[string]string) (*NotifyResult, error) {
	params := form
	if len(params) == 0 && len(body) > 0 {
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
		}
		params = make(map[string]string, len(values))
		for key, value := range values {
			if len(value) > 0 {
				params[key] = value[0]
			}
		}
	}
	if len(params) == 0 || params["sign"] == "" {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}
	// 未配置支付宝公钥时无法验签，一律拒绝，避免伪造通知改单
	if g.cfg.AlipayPublicKey == "" {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}

	bm := make(gopay.BodyMap, len(params))
	for key, value := range params {
		bm.Set(key, value)
	}
	// VerifySign 会从入参里摘掉 sign / sign_type，因此不能直接复用 params
	ok, err := alipay.VerifySign(g.cfg.AlipayPublicKey, bm)
	if err != nil || !ok {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}

	status := params["trade_status"]
	userID := params["buyer_id"]
	if userID == "" {
		userID = params["buyer_logon_id"]
	}

	return &NotifyResult{
		OutTradeNo:     params["out_trade_no"],
		ChannelOrderNo: params["trade_no"],
		ChannelUserID:  userID,
		Price:          yuanToCent(params["total_amount"]),
		Success:        status == "TRADE_SUCCESS" || status == "TRADE_FINISHED",
		RawData:        rawNotify(body, params),
	}, nil
}

// rawNotify 生成可落库的回调原文：优先用请求体，请求体为空时用表单参数序列化。
// 入参：原始请求体与已解析表单；返回 JSON 或表单原文。
func rawNotify(body []byte, params map[string]string) string {
	if len(body) > 0 {
		return string(body)
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return ""
	}

	return string(raw)
}

// yuanToCent 把支付宝回传的元金额（如 "10.00"）换算成分，非法金额按 0 处理。
// 入参：元金额字符串；返回金额（分）。
func yuanToCent(amount string) int64 {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return 0
	}
	value, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0
	}

	return int64(math.Round(value * 100))
}
