package service

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"

	"github.com/go-pay/gopay"
	wxv3 "github.com/go-pay/gopay/wechat/v3"

	baseErrs "eqray-flavor/app/modules/base/errs"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payErrs "eqray-flavor/app/modules/pay/errs"
	"eqray-flavor/app/modules/pay/model"
)

// wechatChannelConfig 是微信渠道的 config JSON 结构，字段名与后台配置页一致。
type wechatChannelConfig struct {
	// AppID 微信 appId（小程序 / 公众号 / 开放平台）
	AppID string `json:"appId"`
	// MchID 微信商户号
	MchID string `json:"mchId"`
	// APIv3Key APIv3 密钥，32 字节
	APIv3Key string `json:"apiV3Key"`
	// SerialNo 商户证书序列号
	SerialNo string `json:"serialNo"`
	// PrivateKey 商户私钥 PEM 内容
	PrivateKey string `json:"privateKey"`
	// PublicKey 微信支付公钥 PEM，配置后开启回调验签
	PublicKey string `json:"publicKey"`
	// PublicKeyID 微信支付公钥 ID
	PublicKeyID string `json:"publicKeyId"`
}

// wechatGateway 基于 gopay 的微信支付 v3 适配器。
type wechatGateway struct {
	// code 渠道编码，决定调用哪个下单接口
	code string
	// cfg 渠道配置
	cfg *wechatChannelConfig
	// client gopay 微信 v3 客户端
	client *wxv3.ClientV3
	// privateKey 已解析的商户私钥，用于 JSAPI / App 二次签名
	privateKey *rsa.PrivateKey
	// wxPublicKey 已解析的微信支付公钥，用于回调验签
	wxPublicKey *rsa.PublicKey
}

// newWechatGateway 解析微信渠道配置并构造适配器。
// 入参：渠道模型；返回适配器，配置缺项时报业务错误。
func newWechatGateway(channel *model.PayChannelModel) (ChannelGateway, error) {
	cfg := &wechatChannelConfig{}
	if err := parseChannelConfig(channel.Config, cfg); err != nil {
		return nil, err
	}

	label := payEnums.ChannelLabel(channel.Code)
	if cfg.AppID == "" || cfg.MchID == "" || cfg.APIv3Key == "" || cfg.SerialNo == "" || cfg.PrivateKey == "" {
		return nil, baseErrs.New(payErrs.PayChannelConfigMissing, label, "appId / mchId / apiV3Key / serialNo / privateKey")
	}

	privateKey, err := parsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	client, err := wxv3.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.APIv3Key, cfg.PrivateKey)
	if err != nil {
		return nil, baseErrs.New(payErrs.PayChannelConfigMissing, label, err.Error())
	}
	// 配置了微信支付公钥时开启下单响应验签，回调验签在 ParseNotify 内自行完成
	var wxPublicKey *rsa.PublicKey
	if cfg.PublicKey != "" {
		wxPublicKey, err = parseRSAPublicKey(cfg.PublicKey)
		if err != nil {
			return nil, baseErrs.New(payErrs.PayChannelConfigMissing, label, "微信支付公钥无效")
		}
		if err = client.AutoVerifySignByPublicKey([]byte(cfg.PublicKey), cfg.PublicKeyID); err != nil {
			return nil, baseErrs.New(payErrs.PayChannelConfigMissing, label, "微信支付公钥无效")
		}
	}

	return &wechatGateway{code: channel.Code, cfg: cfg, client: client, privateKey: privateKey, wxPublicKey: wxPublicKey}, nil
}

// Submit 按渠道编码调用微信下单接口。
// 入参：支付单号、标题、金额、回调地址、openid；返回前端支付参数（JSAPI 签名参数 / Native 二维码 / H5 链接）。
func (g *wechatGateway) Submit(ctx context.Context, req SubmitRequest) (*SubmitResult, error) {
	bm := make(gopay.BodyMap)
	bm.Set("appid", g.cfg.AppID).
		Set("mchid", g.cfg.MchID).
		Set("description", req.Subject).
		Set("out_trade_no", req.No).
		Set("notify_url", req.NotifyURL).
		Set("amount", gopay.BodyMap{"total": req.Price, "currency": "CNY"})
	if req.Attach != "" {
		bm.Set("attach", req.Attach)
	}

	switch g.code {
	case payEnums.ChannelWxLite, payEnums.ChannelWxPub:
		if req.OpenID == "" {
			return nil, baseErrs.New(payErrs.PayChannelConfigMissing, payEnums.ChannelLabel(g.code), "payerOpenId")
		}
		bm.Set("payer", gopay.BodyMap{"openid": req.OpenID})

		rsp, err := g.client.V3TransactionJsapi(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}
		if rsp.Response == nil {
			return nil, errGateway(errorFromResponse(rsp.Error, rsp.ErrResponse))
		}

		params, err := g.jsapiParams(rsp.Response.PrepayId)
		if err != nil {
			return nil, err
		}

		return &SubmitResult{
			Extras:    marshalExtras(map[string]any{"prepayId": rsp.Response.PrepayId}),
			PayParams: params,
		}, nil

	case payEnums.ChannelWxNative:
		rsp, err := g.client.V3TransactionNative(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}
		if rsp.Response == nil {
			return nil, errGateway(errorFromResponse(rsp.Error, rsp.ErrResponse))
		}

		return &SubmitResult{
			Extras:    marshalExtras(map[string]any{"codeUrl": rsp.Response.CodeUrl}),
			PayParams: map[string]any{"codeUrl": rsp.Response.CodeUrl},
		}, nil

	case payEnums.ChannelWxWap:
		rsp, err := g.client.V3TransactionH5(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}
		if rsp.Response == nil {
			return nil, errGateway(errorFromResponse(rsp.Error, rsp.ErrResponse))
		}

		return &SubmitResult{
			Extras:    marshalExtras(map[string]any{"h5Url": rsp.Response.H5Url}),
			PayParams: map[string]any{"h5Url": rsp.Response.H5Url},
		}, nil

	default:
		rsp, err := g.client.V3TransactionApp(ctx, bm)
		if err != nil {
			return nil, errGateway(err)
		}
		if rsp.Response == nil {
			return nil, errGateway(errorFromResponse(rsp.Error, rsp.ErrResponse))
		}

		return &SubmitResult{
			Extras:    marshalExtras(map[string]any{"prepayId": rsp.Response.PrepayId}),
			PayParams: map[string]any{"prepayId": rsp.Response.PrepayId},
		}, nil
	}
}

// Refund 调用微信 v3 退款接口。
// 入参：支付单号、退款单号、原支付金额与退款金额（分）、退款原因与退款通知地址；
// 返回渠道退款单号与渠道是否已同步确认退款完成（微信常为 PROCESSING，需等退款回调）。
func (g *wechatGateway) Refund(ctx context.Context, req RefundRequest) (string, bool, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.No).
		Set("out_refund_no", req.RefundNo).
		Set("amount", gopay.BodyMap{
			"refund":   req.RefundPrice,
			"total":    req.PayPrice,
			"currency": "CNY",
		})
	if req.Reason != "" {
		bm.Set("reason", req.Reason)
	}
	if req.NotifyURL != "" {
		bm.Set("notify_url", req.NotifyURL)
	}

	rsp, err := g.client.V3Refund(ctx, bm)
	if err != nil {
		return "", false, baseErrs.New(payErrs.PayRefundSubmitFail, trimMessage(err.Error()))
	}
	if rsp.Response == nil {
		return "", false, baseErrs.New(payErrs.PayRefundSubmitFail,
			trimMessage(errorFromResponse(rsp.Error, rsp.ErrResponse).Error()))
	}

	return rsp.Response.RefundId, rsp.Response.Status == payEnums.WechatRefundStatusSuccess, nil
}

// Query 查询微信订单状态，用于补单与关单前的二次确认。
// 入参：支付单号；返回是否已支付、渠道订单号、付款人与支付金额（分）。
func (g *wechatGateway) Query(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	rsp, err := g.client.V3TransactionQueryOrder(ctx, wxv3.OutTradeNo, req.No)
	if err != nil {
		return nil, baseErrs.New(payErrs.PayOrderQueryFail, trimMessage(err.Error()))
	}
	if rsp.Response == nil {
		return nil, baseErrs.New(payErrs.PayOrderQueryFail,
			trimMessage(errorFromResponse(rsp.Error, rsp.ErrResponse).Error()))
	}

	result := &QueryResult{
		Paid:           rsp.Response.TradeState == payEnums.WechatTradeStateSuccess,
		ChannelOrderNo: rsp.Response.TransactionId,
	}
	if rsp.Response.Payer != nil {
		result.ChannelUserID = rsp.Response.Payer.Openid
	}
	if rsp.Response.Amount != nil {
		result.Price = int64(rsp.Response.Amount.Total)
	}

	return result, nil
}

// Close 关闭微信订单，关闭后用户无法再对该单付款。
// 入参：支付单号；返回错误表示渠道拒绝关单（如订单已支付）。
func (g *wechatGateway) Close(ctx context.Context, req QueryRequest) error {
	rsp, err := g.client.V3TransactionCloseOrder(ctx, req.No)
	if err != nil {
		return baseErrs.New(payErrs.PayOrderCloseFail, trimMessage(err.Error()))
	}
	if rsp != nil && rsp.Code != 0 {
		return baseErrs.New(payErrs.PayOrderCloseFail, trimMessage(errorFromResponse(rsp.Error, rsp.ErrResponse).Error()))
	}

	return nil
}

// ParseNotify 验签并解密微信回调报文，解析支付结果。
// 入参：原始请求体与请求头（需带 Wechatpay-* 四个签名头）；返回统一通知结果。
func (g *wechatGateway) ParseNotify(ctx context.Context, body []byte, headers map[string]string, form map[string]string) (*NotifyResult, error) {
	// 未配置微信支付公钥时无法验签，一律拒绝，避免伪造通知改单
	if g.wxPublicKey == nil {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}
	if err := wxv3.V3VerifySignByPK(
		headerValue(headers, wxv3.HeaderTimestamp),
		headerValue(headers, wxv3.HeaderNonce),
		string(body),
		headerValue(headers, wxv3.HeaderSignature),
		g.wxPublicKey,
	); err != nil {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}

	var notify struct {
		EventType string `json:"event_type"`
		Resource  struct {
			Ciphertext     string `json:"ciphertext"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &notify); err != nil {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}
	if notify.Resource.Ciphertext == "" {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}

	plain, err := wxv3.V3DecryptNotifyCipherTextToBytes(
		notify.Resource.Ciphertext,
		notify.Resource.Nonce,
		notify.Resource.AssociatedData,
		g.cfg.APIv3Key,
	)
	if err != nil {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}

	// 支付回调与退款回调解密后字段不同，用同一结构体兜住，缺的字段为零值
	var data struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		OutRefundNo   string `json:"out_refund_no"`
		RefundStatus  string `json:"refund_status"`
		Payer         struct {
			OpenID string `json:"openid"`
		} `json:"payer"`
		Amount struct {
			Total  int64 `json:"total"`
			Refund int64 `json:"refund"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plain, &data); err != nil {
		return nil, baseErrs.New(payErrs.PayOrderNotifySignInvalid)
	}

	return &NotifyResult{
		EventType:      notify.EventType,
		OutTradeNo:     data.OutTradeNo,
		ChannelOrderNo: data.TransactionID,
		ChannelUserID:  data.Payer.OpenID,
		Price:          data.Amount.Total,
		Success:        data.TradeState == "SUCCESS",
		RefundNo:       data.OutRefundNo,
		RefundPrice:    data.Amount.Refund,
		RefundSuccess:  data.RefundStatus == payEnums.WechatRefundStatusSuccess,
		RawData:        string(plain),
	}, nil
}

// jsapiParams 生成小程序 / 公众号调起支付参数（含 paySign）。
// 入参：预支付会话 ID；返回前端 wx.requestPayment 所需字段。
func (g *wechatGateway) jsapiParams(prepayID string) (map[string]any, error) {
	timestamp := nowTimestamp()
	nonce, err := randomNonce()
	if err != nil {
		return nil, baseErrs.New(payErrs.PayOrderSubmitFail, "随机串生成失败")
	}
	pkg := "prepay_id=" + prepayID

	signature, err := rsaSignSHA256(g.privateKey, g.cfg.AppID+"\n"+timestamp+"\n"+nonce+"\n"+pkg+"\n")
	if err != nil {
		return nil, baseErrs.New(payErrs.PayOrderSubmitFail, "微信支付参数签名失败")
	}

	return map[string]any{
		"appId":     g.cfg.AppID,
		"timeStamp": timestamp,
		"nonceStr":  nonce,
		"package":   pkg,
		"signType":  "RSA",
		"paySign":   signature,
	}, nil
}

// errorFromResponse 把 gopay 的错误码与错误响应合并成 error。
// 入参：SDK 返回的 Error 文本与错误响应结构；返回可用于提示的 error。
func errorFromResponse(message string, response wxv3.ErrResponse) error {
	if response.Code != "" {
		return &gatewayError{message: response.Code + " " + response.Message}
	}
	if message == "" {
		message = "渠道未返回业务数据"
	}

	return &gatewayError{message: message}
}

// headerValue 按忽略大小写的方式取请求头值，兼容不同框架对头名大小写的处理。
// 入参：回调请求头映射与目标头名；返回头值，缺失时为空串。
func headerValue(headers map[string]string, key string) string {
	if value, ok := headers[key]; ok {
		return value
	}
	for name, value := range headers {
		if strings.EqualFold(name, key) {
			return value
		}
	}

	return ""
}

// parseRSAPublicKey 解析 RSA 公钥 PEM，兼容 PKIX（PUBLIC KEY）与 PKCS#1（RSA PUBLIC KEY）。
// 入参：公钥 PEM 文本；返回公钥对象，格式不合法时报错。
func parseRSAPublicKey(content string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(content)))
	if block == nil {
		return nil, gatewayErr("微信支付公钥 PEM 解析失败")
	}

	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, gatewayErr("微信支付公钥仅支持 PKIX / PKCS#1")
}

// marshalExtras 把渠道额外参数序列化成 JSON，序列化失败时返回空对象。
func marshalExtras(data map[string]any) string {
	raw, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}

	return string(raw)
}
