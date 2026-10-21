package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	baseErrs "eqray-flavor/app/modules/base/errs"
	memberErrs "eqray-flavor/app/modules/member/errs"
)

// 第三方登录渠道标识，与 member_social_user.type 对应。
const (
	SocialTypeWechatMini = "wechat_mini"
	SocialTypeWechatMP   = "wechat_mp"
	SocialTypeWechatOpen = "wechat_open"
)

// 微信接口地址，抽成常量便于测试替换。
const (
	wechatCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"
	wechatAccessTokenURL  = "https://api.weixin.qq.com/sns/oauth2/access_token"
	wechatUserInfoURL     = "https://api.weixin.qq.com/sns/userinfo"
	wechatAuthorizeURL    = "https://open.weixin.qq.com/connect/oauth2/authorize"
	wechatQRConnectURL    = "https://open.weixin.qq.com/connect/qrconnect"
)

// socialProfile 是第三方平台返回的身份信息，屏蔽各渠道的差异。
type socialProfile struct {
	Type     string
	OpenID   string
	UnionID  string
	Nickname string
	Avatar   string
	RawData  string
}

// socialProvider 抽象第三方渠道：生成授权地址 + 用 code 换身份信息。
type socialProvider interface {
	AuthorizeURL(redirectURI, state string) (string, error)
	User(ctx context.Context, code string) (*socialProfile, error)
}

// SocialTypeName 返回渠道中文名，用于错误提示。
func SocialTypeName(socialType string) string {
	switch socialType {
	case SocialTypeWechatMini:
		return "微信小程序"
	case SocialTypeWechatMP:
		return "微信公众号"
	case SocialTypeWechatOpen:
		return "微信开放平台"
	default:
		return socialType
	}
}

// IsSupportedSocialType 判断渠道是否受支持。
func IsSupportedSocialType(socialType string) bool {
	switch socialType {
	case SocialTypeWechatMini, SocialTypeWechatMP, SocialTypeWechatOpen:
		return true
	default:
		return false
	}
}

// newSocialProvider 按渠道 + 客户端凭据构造渠道实现。
// 凭据来自后台配置的 member_social_client（对齐 yudao 的 system_social_client），
// 不再从 config / 环境变量读取 appid。
func newSocialProvider(socialType, clientID, clientSecret string) (socialProvider, error) {
	if !IsSupportedSocialType(socialType) {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "不支持的第三方登录渠道")
	}
	if clientID == "" || clientSecret == "" {
		return nil, baseErrs.New(memberErrs.MemberSocialClientUnavailable, SocialTypeName(socialType))
	}

	return &wechatProvider{
		socialType: socialType,
		appID:      clientID,
		appSecret:  clientSecret,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// wechatProvider 覆盖小程序、公众号与开放平台三种微信渠道。
type wechatProvider struct {
	socialType string
	appID      string
	appSecret  string
	client     *http.Client
}

// AuthorizeURL 生成跳转授权地址；小程序由前端 wx.login 直接取 code，返回空串。
func (p *wechatProvider) AuthorizeURL(redirectURI, state string) (string, error) {
	if p.socialType == SocialTypeWechatMini {
		return "", nil
	}

	query := url.Values{}
	query.Set("appid", p.appID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("state", state)

	endpoint := wechatAuthorizeURL
	if p.socialType == SocialTypeWechatOpen {
		endpoint = wechatQRConnectURL
		query.Set("scope", "snsapi_login")
	} else {
		query.Set("scope", "snsapi_userinfo")
	}

	return endpoint + "?" + query.Encode() + "#wechat_redirect", nil
}

// User 用 code 换取用户身份：小程序走 jscode2session，公众号 / 开放平台走 OAuth2。
func (p *wechatProvider) User(ctx context.Context, code string) (*socialProfile, error) {
	if strings.TrimSpace(code) == "" {
		return nil, baseErrs.New(memberErrs.MemberSocialCodeInvalid)
	}

	if p.socialType == SocialTypeWechatMini {
		return p.miniUser(ctx, code)
	}

	return p.oauthUser(ctx, code)
}

// miniUser 调用 jscode2session，只返回 openid / unionid，昵称头像由前端自行维护。
func (p *wechatProvider) miniUser(ctx context.Context, code string) (*socialProfile, error) {
	query := url.Values{}
	query.Set("appid", p.appID)
	query.Set("secret", p.appSecret)
	query.Set("js_code", code)
	query.Set("grant_type", "authorization_code")

	var resp struct {
		OpenID  string `json:"openid"`
		UnionID string `json:"unionid"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	raw, err := p.getJSON(ctx, wechatCode2SessionURL, query, &resp)
	if err != nil {
		return nil, err
	}
	if resp.ErrCode != 0 || resp.OpenID == "" {
		return nil, wechatError(resp.ErrCode, resp.ErrMsg)
	}

	return &socialProfile{
		Type:    p.socialType,
		OpenID:  resp.OpenID,
		UnionID: resp.UnionID,
		RawData: raw,
	}, nil
}

// oauthUser 先用 code 换 access_token，再拉取用户资料。
func (p *wechatProvider) oauthUser(ctx context.Context, code string) (*socialProfile, error) {
	tokenQuery := url.Values{}
	tokenQuery.Set("appid", p.appID)
	tokenQuery.Set("secret", p.appSecret)
	tokenQuery.Set("code", code)
	tokenQuery.Set("grant_type", "authorization_code")

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		OpenID      string `json:"openid"`
		UnionID     string `json:"unionid"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if _, err := p.getJSON(ctx, wechatAccessTokenURL, tokenQuery, &tokenResp); err != nil {
		return nil, err
	}
	if tokenResp.ErrCode != 0 || tokenResp.AccessToken == "" {
		return nil, wechatError(tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	infoQuery := url.Values{}
	infoQuery.Set("access_token", tokenResp.AccessToken)
	infoQuery.Set("openid", tokenResp.OpenID)
	infoQuery.Set("lang", "zh_CN")

	var infoResp struct {
		OpenID     string `json:"openid"`
		UnionID    string `json:"unionid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	raw, err := p.getJSON(ctx, wechatUserInfoURL, infoQuery, &infoResp)
	if err != nil {
		return nil, err
	}
	if infoResp.ErrCode != 0 {
		return nil, wechatError(infoResp.ErrCode, infoResp.ErrMsg)
	}

	if infoResp.OpenID == "" {
		infoResp.OpenID = tokenResp.OpenID
	}
	if infoResp.UnionID == "" {
		infoResp.UnionID = tokenResp.UnionID
	}

	return &socialProfile{
		Type:     p.socialType,
		OpenID:   infoResp.OpenID,
		UnionID:  infoResp.UnionID,
		Nickname: infoResp.Nickname,
		Avatar:   infoResp.HeadImgURL,
		RawData:  raw,
	}, nil
}

// getJSON 发起 GET 请求并把响应解密到 dest，同时返回原始报文用于留档。
func (p *wechatProvider) getJSON(ctx context.Context, endpoint string, query url.Values, dest any) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return "", baseErrs.New(memberErrs.MemberSocialRequestFailed, err.Error())
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", baseErrs.New(memberErrs.MemberSocialRequestFailed, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", baseErrs.New(memberErrs.MemberSocialRequestFailed, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return string(body), baseErrs.New(memberErrs.MemberSocialRequestFailed, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return string(body), baseErrs.New(memberErrs.MemberSocialRequestFailed, "响应格式不正确")
	}

	return string(body), nil
}

// wechatError 把微信错误码转换为业务错误：
// 授权码类错误统一提示重新授权，其余按原始 errmsg 返回，便于排查配置问题。
func wechatError(errCode int, errMsg string) error {
	switch errCode {
	case 40029, 40163, 41008, 42002:
		return baseErrs.New(memberErrs.MemberSocialCodeInvalid)
	}

	if errMsg == "" {
		errMsg = fmt.Sprintf("errcode=%d", errCode)
	}

	return baseErrs.New(memberErrs.MemberSocialRequestFailed, errMsg)
}
