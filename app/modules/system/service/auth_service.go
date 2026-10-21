// Package service 存放 system 模块的业务逻辑，事务只在这一层开启。
package service

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	baseAuth "eqray-flavor/app/modules/base/auth"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/repository"
)

// LoginResult 是登录 / 续期成功后返回给前端的一对令牌。
type LoginResult = baseAuth.TokenPair

// LoginInput 是登录入参，用结构体承接避免 Login 的参数列表过长。
type LoginInput struct {
	IP          string
	UserAgent   string
	Username    string
	Password    string
	CaptchaID   string
	CaptchaCode string
}

// AuthService 负责后台登录认证（docs 9.1）。
type AuthService struct {
	users      *repository.UserRepository
	permission *PermissionService
}

// NewAuthService 构造 Service。
func NewAuthService() *AuthService {
	return &AuthService{
		users:      repository.NewUserRepository(),
		permission: NewPermissionService(),
	}
}

// Login 依次校验验证码、账号密码与账号状态，通过后签发令牌并刷新登录信息。
// 账号不存在与密码错误返回同一个错误码，避免被用来枚举账号。
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	if !baseAuth.VerifyCaptcha(in.CaptchaID, in.CaptchaCode) {
		s.recordLogin(ctx, in, 0, baseEnums.LoginResultFail)
		return nil, baseErrs.New(systemErrs.SystemCaptchaWrong)
	}

	user, err := s.users.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(in.Username)))
	if err != nil {
		return nil, err
	}
	if user == nil {
		s.recordLogin(ctx, in, 0, baseEnums.LoginResultFail)
		return nil, baseErrs.New(systemErrs.SystemLoginFailed)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)) != nil {
		s.recordLogin(ctx, in, user.ID, baseEnums.LoginResultFail)
		return nil, baseErrs.New(systemErrs.SystemLoginFailed)
	}
	if user.Status != baseEnums.CommonStatusEnabled {
		s.recordLogin(ctx, in, user.ID, baseEnums.LoginResultFail)
		return nil, baseErrs.New(systemErrs.SystemAccountDisabled)
	}

	tokens, err := baseAuth.Issue(user.ID)
	if err != nil {
		s.recordLogin(ctx, in, user.ID, baseEnums.LoginResultFail)
		return nil, err
	}

	s.recordLogin(ctx, in, user.ID, baseEnums.LoginResultSuccess)

	// 登录信息只用于展示，回写失败不阻断登录。
	_ = s.users.UpdateLoginInfo(ctx, user.ID, in.IP, types.Now())

	return tokens, nil
}

// recordLogin 写后台用户的登录日志，失败不影响主流程。
func (s *AuthService) recordLogin(ctx context.Context, in LoginInput, userID types.ID, result baseEnums.LoginResultEnum) {
	baseAuth.RecordLoginLog(ctx, baseAuth.LoginLogInput{
		UserID:    userID,
		UserType:  baseEnums.UserTypeAdmin,
		Username:  in.Username,
		Type:      baseEnums.LoginTypeLogin,
		Result:    result,
		IP:        in.IP,
		UserAgent: in.UserAgent,
	})
}

// Refresh 用刷新令牌换一对新令牌，旧刷新令牌随之作废（docs 9.2）。
func (s *AuthService) Refresh(refreshToken string) (*LoginResult, error) {
	return baseAuth.Refresh(refreshToken)
}

// Logout 删除服务端保存的访问令牌与刷新令牌，使两者立即失效，并写登出日志。
func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken, ip, userAgent string) {
	// 先解析用户，注销后访问令牌就从 Redis 里删掉了
	userID, _ := baseAuth.Resolve(accessToken)
	baseAuth.Revoke(accessToken, refreshToken)

	if userID == 0 {
		return
	}

	username := ""
	if user, err := s.users.FindByID(ctx, userID); err == nil && user != nil {
		username = user.Username
	}

	baseAuth.RecordLoginLog(ctx, baseAuth.LoginLogInput{
		UserID:    userID,
		UserType:  baseEnums.UserTypeAdmin,
		Username:  username,
		Type:      baseEnums.LoginTypeLogout,
		Result:    baseEnums.LoginResultSuccess,
		IP:        ip,
		UserAgent: userAgent,
	})
}

// Profile 返回当前登录者的基础信息、角色标识与权限标识（docs 9.3）。
func (s *AuthService) Profile(ctx context.Context, id types.ID) (map[string]any, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, baseErrs.New(systemErrs.SystemUserNotExists, id)
	}

	roles, err := s.permission.RoleCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	buttons, err := s.permission.Permissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"userId":   user.ID,
		"userName": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"email":    user.Email,
		"mobile":   user.Mobile,
		"sex":      user.Sex,
		"roles":    roles,
		"buttons":  buttons,
	}, nil
}
