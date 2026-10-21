// Package controller 存放 system 模块的控制器，只做取参、调 Service、出响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseAuth "eqray-flavor/app/modules/base/auth"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/admin/request"
	"eqray-flavor/app/modules/system/service"
)

// AuthController 提供后台认证接口。
type AuthController struct {
	service *service.AuthService
	users   *service.UserService
}

func NewAuthController() *AuthController {
	return &AuthController{
		service: service.NewAuthService(),
		users:   service.NewUserService(),
	}
}

// Captcha 生成登录图形验证码，验证码内容存 Redis，图片以 base64 返回。
// 无入参；返回 { captchaId, image }。
func (c *AuthController) Captcha(ctx http.Context) http.Response {
	id, image, err := baseAuth.GenerateCaptcha()
	if err != nil {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.InternalServerError, "验证码生成失败"))
	}

	return result.Success(ctx, map[string]any{
		"captchaId": id,
		"image":     image,
	})
}

// Profile 查询当前登录者的个人资料、角色与权限点，供前端初始化用户态。
// 无入参，用户身份取自令牌；返回个人资料，未登录返回未授权。
func (c *AuthController) Profile(ctx http.Context) http.Response {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	data, err := c.service.Profile(ctx.Context(), types.ID(userID))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, data)
}

// UpdateProfile 修改当前登录者的个人资料。
// 入参：nickname、email、mobile、sex、avatar；返回 null。
func (c *AuthController) UpdateProfile(ctx http.Context) http.Response {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	var req request.ProfileUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err := c.users.UpdateProfile(ctx.Context(), types.ID(userID), service.ProfileUpdateInput{
		Nickname: req.Nickname,
		Email:    req.Email,
		Mobile:   req.Mobile,
		Sex:      req.Sex,
		Avatar:   req.Avatar,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// UpdatePassword 修改当前登录者的登录密码，需要校验原密码。
// 入参：oldPassword、newPassword；返回 null。
func (c *AuthController) UpdatePassword(ctx http.Context) http.Response {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	var req request.ProfilePasswordRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.users.ChangePassword(ctx.Context(), types.ID(userID), req.OldPassword, req.NewPassword); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Login 后台账号密码登录，校验图形验证码后签发访问 / 刷新令牌。
// 入参：username、password、captchaId、captchaCode；返回用户信息、角色权限与令牌。
func (c *AuthController) Login(ctx http.Context) http.Response {
	var req request.LoginRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.service.Login(ctx.Context(), service.LoginInput{
		IP:          ctx.Request().Ip(),
		UserAgent:   ctx.Request().Header("User-Agent"),
		Username:    req.Username,
		Password:    req.Password,
		CaptchaID:   req.CaptchaID,
		CaptchaCode: req.CaptchaCode,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// Refresh 用刷新令牌换新的访问令牌。
// 入参：refreshToken；返回新的令牌对。
func (c *AuthController) Refresh(ctx http.Context) http.Response {
	var req request.RefreshRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	tokens, err := c.service.Refresh(req.RefreshToken)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tokens)
}

// Logout 注销后台登录态：清理令牌并写一条登录日志。
// 入参：Authorization 头 + refreshToken；返回 null。
func (c *AuthController) Logout(ctx http.Context) http.Response {
	c.service.Logout(
		ctx.Context(),
		baseAuth.BearerToken(ctx.Request().Header("Authorization")),
		ctx.Request().Input("refreshToken"),
		ctx.Request().Ip(),
		ctx.Request().Header("User-Agent"),
	)

	return result.Success(ctx, nil)
}
