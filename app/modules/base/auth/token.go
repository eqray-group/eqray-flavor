// Package auth 提供框架级认证能力：双令牌（访问 / 刷新）与图片验证码，见 docs 9.2。
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
)

const (
	defaultAccessTTL  = 1800
	defaultRefreshTTL = 604800
	defaultPrefix     = "auth"

	accessKind  = "access"
	refreshKind = "refresh"
)

// 令牌作用域：不同作用域的令牌互不通用（Redis 键会多一层前缀）。
// 新增接入端时在这里加一个作用域常量即可，不需要再包一层包装包。
const (
	// ScopeAdmin 后台管理端作用域。
	ScopeAdmin = ""
	// ScopeMember C 端会员作用域。
	ScopeMember = "member"
)

// TokenPair 是登录 / 续期后返回给前端的一对令牌，有效期为秒。
//
// Access Token 短期有效，随每次请求放在 Authorization 头里；
// Refresh Token 长期有效，只用来换新的一对令牌，不参与业务请求。
type TokenPair struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
}

// Issue 为用户签发一对令牌。
//
// 两个令牌都是 32 字节随机串，服务端只在 Redis 保存它的 SHA-256，
// 这样 Redis 被读走也拿不到可用令牌，同时天然支持「删除即失效」。
func Issue(userID types.ID) (*TokenPair, error) {
	return IssueScoped(ScopeAdmin, userID)
}

// IssueScoped 按作用域签发令牌。
//
// scope 取 ScopeAdmin（后台账号）或 ScopeMember（C 端会员），Redis 键会多一层前缀，
// 两套令牌因此互不通用——后台令牌调不了会员接口，反之亦然。
func IssueScoped(scope string, userID types.ID) (*TokenPair, error) {
	accessTTL, refreshTTL := ttl()
	value := userID.String()

	accessToken, err := randomToken()
	if err != nil {
		return nil, err
	}
	refreshToken, err := randomToken()
	if err != nil {
		return nil, err
	}

	if err := facades.Cache().Put(key(scope, accessKind, accessToken), value, accessTTL); err != nil {
		return nil, err
	}
	if err := facades.Cache().Put(key(scope, refreshKind, refreshToken), value, refreshTTL); err != nil {
		facades.Cache().Forget(key(scope, accessKind, accessToken))
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int64(accessTTL / time.Second),
		RefreshExpiresIn: int64(refreshTTL / time.Second),
	}, nil
}

// Resolve 用访问令牌换取用户 ID；令牌不存在或已过期时返回未登录错误。
func Resolve(accessToken string) (types.ID, error) {
	return ResolveScoped(ScopeAdmin, accessToken)
}

// ResolveScoped 按作用域解析访问令牌，scope 语义与 IssueScoped 一致。
func ResolveScoped(scope, accessToken string) (types.ID, error) {
	if accessToken == "" {
		return 0, unauthorized()
	}

	raw := facades.Cache().GetString(key(scope, accessKind, accessToken))
	if raw == "" {
		return 0, unauthorized()
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, unauthorized()
	}

	return types.ID(id), nil
}

// Refresh 用刷新令牌换一对新令牌，并把旧刷新令牌作废（轮换）。
// 旧的访问令牌按自己的短 TTL 自然过期，不额外维护黑名单。
func Refresh(refreshToken string) (*TokenPair, error) {
	return RefreshScoped(ScopeAdmin, refreshToken)
}

// RefreshScoped 按作用域续期，换出的新令牌仍属于同一作用域。
func RefreshScoped(scope, refreshToken string) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, unauthorized()
	}

	raw := facades.Cache().GetString(key(scope, refreshKind, refreshToken))
	if raw == "" {
		return nil, unauthorized()
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return nil, unauthorized()
	}

	facades.Cache().Forget(key(scope, refreshKind, refreshToken))

	return IssueScoped(scope, types.ID(id))
}

// Revoke 登出时删除两个令牌，立即失去访问权与续期能力。
func Revoke(accessToken, refreshToken string) {
	RevokeScoped(ScopeAdmin, accessToken, refreshToken)
}

// RevokeScoped 按作用域注销令牌。
func RevokeScoped(scope, accessToken, refreshToken string) {
	if accessToken != "" {
		facades.Cache().Forget(key(scope, accessKind, accessToken))
	}
	if refreshToken != "" {
		facades.Cache().Forget(key(scope, refreshKind, refreshToken))
	}
}

// BearerToken 去掉可选的 Bearer 前缀，兼容前端带与不带两种写法。
func BearerToken(header string) string {
	header = strings.TrimSpace(header)
	if len(header) > 7 && strings.EqualFold(header[:7], "Bearer ") {
		return strings.TrimSpace(header[7:])
	}

	return header
}

// key 生成 Redis 键：{prefix}[:{scope}]:{kind}:{sha256(token)}，不落明文令牌。
func key(scope, kind, token string) string {
	sum := sha256.Sum256([]byte(token))

	prefix := facades.Config().GetString("token.prefix", defaultPrefix)
	if scope != "" {
		prefix += ":" + scope
	}

	return prefix + ":" + kind + ":" + hex.EncodeToString(sum[:])
}

// ttl 读取令牌有效期，配置为秒，缓存接口用 time.Duration。
func ttl() (time.Duration, time.Duration) {
	access := facades.Config().GetInt("token.access_ttl", defaultAccessTTL)
	refresh := facades.Config().GetInt("token.refresh_ttl", defaultRefreshTTL)

	return time.Duration(access) * time.Second, time.Duration(refresh) * time.Second
}

// randomToken 生成 32 字节随机串，URL 安全编码后直接放进请求头与请求体。
func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// unauthorized 构造未授权业务异常，供鉴权失败时统一返回 401 对应错误码。
func unauthorized() error {
	return baseErrs.New(baseErrs.Unauthorized)
}
