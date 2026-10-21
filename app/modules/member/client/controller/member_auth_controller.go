package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseAuth "eqray-flavor/app/modules/base/auth"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/member/client/request"
	"eqray-flavor/app/modules/member/service"
)

// MemberAuthController 提供 C 端会员的第三方登录、手机号验证码登录与登录态接口。
type MemberAuthController struct {
	service *service.MemberSocialService
	auth    *service.MemberAuthService
}

// NewMemberAuthController 构造 C 端会员认证控制器，并装配社交与手机号登录服务。
func NewMemberAuthController() *MemberAuthController {
	return &MemberAuthController{
		service: service.NewMemberSocialService(),
		auth:    service.NewMemberAuthService(),
	}
}

// SmsCode 发送登录短信验证码，60 秒内同一手机号只能发一次。
// 入参：mobile、scene（默认 login）；返回 null。
func (c *MemberAuthController) SmsCode(ctx http.Context) http.Response {
	var req request.MemberSmsCodeRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.auth.SendSmsCode(ctx.Context(), req.Mobile, req.Scene); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// SmsLogin 手机号验证码登录；手机号不存在时自动注册会员。
// 入参：mobile、code；返回会员信息与访问 / 刷新令牌。
func (c *MemberAuthController) SmsLogin(ctx http.Context) http.Response {
	var req request.MemberSmsLoginRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.auth.SmsLogin(ctx.Context(), service.SmsLoginInput{
		Mobile:    req.Mobile,
		Code:      req.Code,
		IP:        ctx.Request().Ip(),
		UserAgent: ctx.Request().Header("User-Agent"),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// SocialBindLogin 先手机号登录取令牌，再绑定第三方账号；绑定失败不签发令牌。
// 入参：mobile、code、socialType、socialCode、socialState；返回会员信息与令牌。
func (c *MemberAuthController) SocialBindLogin(ctx http.Context) http.Response {
	var req request.MemberSocialBindLoginRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.auth.SocialBindLogin(ctx.Context(), service.SocialBindLoginInput{
		Mobile:      req.Mobile,
		Code:        req.Code,
		SocialType:  req.SocialType,
		SocialCode:  req.SocialCode,
		SocialState: req.SocialState,
		IP:          ctx.Request().Ip(),
		UserAgent:   ctx.Request().Header("User-Agent"),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// AuthorizeURL 生成第三方授权的跳转地址，凭据取自后台 member_social_client 配置。
// 入参：路径 type（渠道）+ redirectUri、state（查询串）；公众号 / 网站扫码返回跳转地址，小程序返回空串。
func (c *MemberAuthController) AuthorizeURL(ctx http.Context) http.Response {
	url, err := c.service.AuthorizeURL(
		ctx.Context(),
		ctx.Request().Route("type"),
		ctx.Request().Query("redirectUri"),
		ctx.Request().Query("state"),
	)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"url": url})
}

// SocialLogin 第三方快捷登录：已绑定才签发令牌，未绑定返回业务错误码，不自动注册。
// 入参：type（渠道）、code（授权码）；返回会员信息与令牌。
func (c *MemberAuthController) SocialLogin(ctx http.Context) http.Response {
	var req request.MemberSocialLoginRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.service.Login(ctx.Context(), service.SocialLoginInput{
		Type:      req.Type,
		Code:      req.Code,
		IP:        ctx.Request().Ip(),
		UserAgent: ctx.Request().Header("User-Agent"),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// Refresh 用刷新令牌换新的访问令牌，并重置刷新令牌有效期。
// 入参：refreshToken；返回新的令牌对。
func (c *MemberAuthController) Refresh(ctx http.Context) http.Response {
	var req request.MemberRefreshRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	tokens, err := c.service.Refresh(req.RefreshToken)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tokens)
}

// Logout 注销会员登录态：删除 Redis 中的访问令牌与刷新令牌。
// 入参：Authorization 头 + refreshToken；返回 null。
func (c *MemberAuthController) Logout(ctx http.Context) http.Response {
	var req request.MemberLogoutRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	c.service.Logout(
		baseAuth.BearerToken(ctx.Request().Header("Authorization")),
		req.RefreshToken,
	)

	return result.Success(ctx, nil)
}

// Me 查询当前登录会员的基本信息，供前端刷新登录态。
// 无入参，会员身份取自令牌；返回会员档案，令牌失效返回未授权。
func (c *MemberAuthController) Me(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if memberID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	member, err := c.service.Me(ctx.Context(), types.ID(memberID))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, member)
}

// Bind 把第三方账号绑定到当前会员，同一渠道只能绑定一个账号。
// 入参：type、code；会员身份取自令牌；返回 null。
func (c *MemberAuthController) Bind(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if memberID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	var req request.MemberSocialBindRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err := c.service.Bind(ctx.Context(), service.SocialBindInput{
		MemberID: types.ID(memberID),
		Type:     req.Type,
		Code:     req.Code,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Bindings 查询当前会员已绑定的第三方账号列表。
// 无入参，会员身份取自令牌；返回绑定记录列表。
func (c *MemberAuthController) Bindings(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if memberID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	list, err := c.service.Bindings(ctx.Context(), types.ID(memberID))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Unbind 解绑指定的第三方账号；至少保留一个绑定，避免会员失去登录方式。
// 入参：路径 id（绑定记录 ID）；返回 null。
func (c *MemberAuthController) Unbind(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if memberID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Unbind(ctx.Context(), types.ID(memberID), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
