package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"eqray-flavor/app/facades"
	baseAuth "eqray-flavor/app/modules/base/auth"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	infraService "eqray-flavor/app/modules/infra/service"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

const (
	smsSceneLogin          = "login"
	smsCodeTTL             = 5 * time.Minute
	smsResendInterval      = 60 * time.Second
	smsDailyLimit          = 10
	smsCodeKeyPrefix       = "member:sms:code:"
	smsRateKeyPrefix       = "member:sms:rate:"
	smsDailyKeyPrefix      = "member:sms:daily:"
	smsCodeParamName       = "code"
	smsDefaultTemplateCode = "member_sms_login"
)

var mobilePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// supportedSmsScenes 是允许的短信场景白名单：新增场景时在这里登记，
// 并在 infra_sms_template 里配一条同名模板，否则发送会因模板缺失失败。
var supportedSmsScenes = map[string]struct{}{
	smsSceneLogin: {},
}

// SmsLoginInput 是手机号 + 验证码登录（注册）入参。
type SmsLoginInput struct {
	Mobile    string
	Code      string
	IP        string
	UserAgent string
}

// SocialBindLoginInput 是手机号验证码登录后绑定第三方账号的入参，对齐 yudao 的 socialBindLogin。
type SocialBindLoginInput struct {
	Mobile      string
	Code        string
	SocialType  string
	SocialCode  string
	SocialState string
	IP          string
	UserAgent   string
}

// MemberAuthService 负责会员的手机号验证码登录 / 注册。
// 短信发送能力复用 infra 模块的 InfraSmsService（渠道与模板配置在 infra）。
type MemberAuthService struct {
	members *repository.MemberRepository
	socials *MemberSocialService
	sms     *infraService.InfraSmsService
}

// NewMemberAuthService 构造会员手机号登录服务，复用 infra 的短信发送能力。
func NewMemberAuthService() *MemberAuthService {
	return &MemberAuthService{
		members: repository.NewMemberRepository(),
		socials: NewMemberSocialService(),
		sms:     infraService.NewInfraSmsService(),
	}
}

// SendSmsCode 发送短信验证码：频率 60 秒、每日每手机号 10 次、验证码 5 分钟有效。
// 渠道与模板配置在 infra 模块（infra_sms_channel / infra_sms_template）。
func (s *MemberAuthService) SendSmsCode(ctx context.Context, mobile, scene string) error {
	mobile = strings.TrimSpace(mobile)
	if !mobilePattern.MatchString(mobile) {
		return baseErrs.New(memberErrs.MemberMobileInvalid)
	}

	// 场景来自前端，先归一化再参与比较和拼 Redis 键
	if scene == "" {
		scene = smsSceneLogin
	}
	// 白名单之外的场景直接拒绝，避免拿脏值去查模板、生成对不上的 Redis 键
	if _, ok := supportedSmsScenes[scene]; !ok {
		return baseErrs.New(memberErrs.MemberSmsSceneInvalid, scene)
	}

	// 发送频率限制
	rateKey := smsRateKey(scene, mobile)
	if facades.Cache().Has(rateKey) {
		return baseErrs.New(memberErrs.MemberSmsCodeTooFrequent)
	}

	// 每日发送上限
	dailyKey := smsDailyKey(scene, mobile)
	if !facades.Cache().Has(dailyKey) {
		if err := facades.Cache().Put(dailyKey, 0, ttlUntilMidnight()); err != nil {
			return err
		}
	}
	count, err := facades.Cache().Increment(dailyKey)
	if err != nil {
		return err
	}
	if count > smsDailyLimit {
		return baseErrs.New(memberErrs.MemberSmsCodeDailyLimit)
	}

	code, err := randomCode()
	if err != nil {
		return err
	}
	if err := facades.Cache().Put(smsCodeKey(scene, mobile), code, smsCodeTTL); err != nil {
		return err
	}

	if err := s.sms.Send(ctx, s.templateCode(scene), mobile, map[string]string{smsCodeParamName: code}); err != nil {
		return err
	}

	_ = facades.Cache().Put(rateKey, 1, smsResendInterval)

	return nil
}

// templateCode 返回场景对应的短信模板代码：登录用内置登录模板，其余场景直接用场景标识。
// 模板内容、绑定渠道与启停状态都配置在数据库（infra_sms_template / infra_sms_channel），
// 这里只负责给出模板代码，不再读配置文件。
func (s *MemberAuthService) templateCode(scene string) string {
	if scene == smsSceneLogin {
		return smsDefaultTemplateCode
	}

	return scene
}

// SmsLogin 手机号 + 验证码登录：手机号不存在时自动注册会员（对齐 yudao 的 createUserIfAbsent）。
func (s *MemberAuthService) SmsLogin(ctx context.Context, in SmsLoginInput) (*SocialLoginResult, error) {
	mobile := strings.TrimSpace(in.Mobile)
	if !mobilePattern.MatchString(mobile) {
		return nil, baseErrs.New(memberErrs.MemberMobileInvalid)
	}
	if err := s.verifySmsCode(smsSceneLogin, mobile, in.Code); err != nil {
		s.recordLogin(ctx, 0, mobile, baseEnums.LoginResultFail, in.IP, in.UserAgent)
		return nil, err
	}

	member, err := s.ensureMemberByMobile(ctx, mobile, in.IP)
	if err != nil {
		return nil, err
	}
	if member.Status != baseEnums.CommonStatusEnabled {
		s.recordLogin(ctx, member.ID, member.Nickname, baseEnums.LoginResultFail, in.IP, in.UserAgent)
		return nil, baseErrs.New(memberErrs.MemberDisabled)
	}

	return s.issueLogin(ctx, member, in.IP, in.UserAgent)
}

// SocialBindLogin 手机号验证码登录（注册）并绑定第三方账号，返回登录令牌。
func (s *MemberAuthService) SocialBindLogin(ctx context.Context, in SocialBindLoginInput) (*SocialLoginResult, error) {
	mobile := strings.TrimSpace(in.Mobile)
	if !mobilePattern.MatchString(mobile) {
		return nil, baseErrs.New(memberErrs.MemberMobileInvalid)
	}
	if err := s.verifySmsCode(smsSceneLogin, mobile, in.Code); err != nil {
		s.recordLogin(ctx, 0, mobile, baseEnums.LoginResultFail, in.IP, in.UserAgent)
		return nil, err
	}

	member, err := s.ensureMemberByMobile(ctx, mobile, in.IP)
	if err != nil {
		return nil, err
	}
	if member.Status != baseEnums.CommonStatusEnabled {
		s.recordLogin(ctx, member.ID, member.Nickname, baseEnums.LoginResultFail, in.IP, in.UserAgent)
		return nil, baseErrs.New(memberErrs.MemberDisabled)
	}

	// 绑定第三方账号；失败不签发令牌
	if err := s.socials.Bind(ctx, SocialBindInput{
		MemberID: member.ID,
		Type:     in.SocialType,
		Code:     in.SocialCode,
	}); err != nil {
		return nil, err
	}

	return s.issueLogin(ctx, member, in.IP, in.UserAgent)
}

// verifySmsCode 校验并消费验证码，无论成功与否都删除，避免重放。
func (s *MemberAuthService) verifySmsCode(scene, mobile, code string) error {
	key := smsCodeKey(scene, mobile)
	actual := facades.Cache().GetString(key)
	facades.Cache().Forget(key)

	if actual == "" || actual != strings.TrimSpace(code) {
		return baseErrs.New(memberErrs.MemberSmsCodeWrong)
	}

	return nil
}

// ensureMemberByMobile 按手机号查会员，不存在则创建（默认昵称 + 启用状态）。
func (s *MemberAuthService) ensureMemberByMobile(ctx context.Context, mobile, ip string) (*model.MemberUserModel, error) {
	m, err := s.members.FindByMobile(ctx, mobile)
	if err != nil {
		return nil, err
	}
	if m != nil {
		return m, nil
	}

	m = &model.MemberUserModel{
		Mobile:     mobile,
		Nickname:   "用户" + mobile[len(mobile)-4:],
		RegisterIP: ip,
		Status:     baseEnums.CommonStatusEnabled,
	}
	if err := s.members.Create(ctx, m); err != nil {
		// 并发注册：另一个请求已创建，回查即可
		if support.IsDuplicate(err) {
			return s.members.FindByMobile(ctx, mobile)
		}

		return nil, err
	}

	return m, nil
}

// issueLogin 签发令牌、回写登录信息并记登录日志。
func (s *MemberAuthService) issueLogin(ctx context.Context, member *model.MemberUserModel, ip, userAgent string) (*SocialLoginResult, error) {
	tokens, err := baseAuth.IssueScoped(baseAuth.ScopeMember, member.ID)
	if err != nil {
		s.recordLogin(ctx, member.ID, member.Nickname, baseEnums.LoginResultFail, ip, userAgent)
		return nil, err
	}
	s.recordLogin(ctx, member.ID, member.Nickname, baseEnums.LoginResultSuccess, ip, userAgent)

	now := types.Now()
	_ = s.members.UpdateLoginInfo(ctx, member.ID, ip, now)
	member.LoginIP = ip
	member.LoginDate = &now

	return &SocialLoginResult{TokenPair: tokens, Member: member}, nil
}

// recordLogin 写会员登录日志，user_type 标记为会员。
func (s *MemberAuthService) recordLogin(ctx context.Context, memberID types.ID, username string, result baseEnums.LoginResultEnum, ip, userAgent string) {
	baseAuth.RecordLoginLog(ctx, baseAuth.LoginLogInput{
		UserID:    memberID,
		UserType:  baseEnums.UserTypeMember,
		Username:  username,
		Type:      baseEnums.LoginTypeLogin,
		Result:    result,
		IP:        ip,
		UserAgent: userAgent,
	})
}

// smsCodeKey 返回验证码缓存键：按场景 + 手机号隔离。
func smsCodeKey(scene, mobile string) string {
	return smsCodeKeyPrefix + scene + ":" + mobile
}

// smsRateKey 返回发送频率限制键：同一手机号同一场景 60 秒内只能发一次。
func smsRateKey(scene, mobile string) string {
	return smsRateKeyPrefix + scene + ":" + mobile
}

// smsDailyKey 返回每日发送计数键：带日期后缀，跨天自动失效。
func smsDailyKey(scene, mobile string) string {
	return smsDailyKeyPrefix + scene + ":" + mobile + ":" + time.Now().Format("20060102")
}

// ttlUntilMidnight 计算到当天 24 点的剩余时长，用于每日计数键的过期时间。
func ttlUntilMidnight() time.Duration {
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	return tomorrow.Sub(now)
}

// randomCode 生成 6 位数字验证码。
func randomCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}
