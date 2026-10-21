package service

import (
	"context"
	"strings"

	baseAuth "eqray-flavor/app/modules/base/auth"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

// SocialLoginInput 是第三方登录入参。
type SocialLoginInput struct {
	Type      string
	Code      string
	IP        string
	UserAgent string
}

// SocialBindInput 是当前会员绑定第三方账号的入参。
type SocialBindInput struct {
	MemberID types.ID
	Type     string
	Code     string
}

// SocialLoginResult 是第三方登录成功后的返回：令牌 + 会员档案。
type SocialLoginResult struct {
	*baseAuth.TokenPair
	Member *model.MemberUserModel `json:"member"`
}

// SocialBindingItem 是会员绑定列表的展示项，合并了社交账号的昵称与头像。
type SocialBindingItem struct {
	ID         types.ID       `db:"id" json:"id"`
	SocialType string         `db:"social_type" json:"socialType"`
	Nickname   string         `json:"nickname"`
	Avatar     string         `json:"avatar"`
	CreatedAt  types.DateTime `json:"createdAt"`
}

// MemberSocialService 负责会员的第三方登录与账号绑定。
type MemberSocialService struct {
	socials *repository.MemberSocialRepository
	members *repository.MemberRepository
	clients *repository.MemberSocialClientRepository
}

// NewMemberSocialService 构造会员社交登录服务。
func NewMemberSocialService() *MemberSocialService {
	return &MemberSocialService{
		socials: repository.NewMemberSocialRepository(),
		members: repository.NewMemberRepository(),
		clients: repository.NewMemberSocialClientRepository(),
	}
}

// AuthorizeURL 返回渠道授权地址；小程序无需跳转，返回空串。
func (s *MemberSocialService) AuthorizeURL(ctx context.Context, socialType, redirectURI, state string) (string, error) {
	provider, err := s.provider(ctx, socialType)
	if err != nil {
		return "", err
	}
	if socialType != SocialTypeWechatMini && strings.TrimSpace(redirectURI) == "" {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "回调地址不能为空")
	}

	return provider.AuthorizeURL(strings.TrimSpace(redirectURI), strings.TrimSpace(state))
}

// Login 第三方快捷登录（对齐 yudao 的 socialQuickLogin）：
//  1. code 换取第三方身份；
//  2. 已绑定则签发会员令牌，未绑定返回「未绑定」错误，不自动注册会员。
func (s *MemberSocialService) Login(ctx context.Context, in SocialLoginInput) (*SocialLoginResult, error) {
	profile, err := s.fetchProfile(ctx, in.Type, in.Code)
	if err != nil {
		s.recordLogin(ctx, 0, SocialTypeName(in.Type), baseEnums.LoginResultFail, in)
		return nil, err
	}

	socialUser, err := s.ensureSocialUser(ctx, profile)
	if err != nil {
		s.recordLogin(ctx, 0, SocialTypeName(in.Type), baseEnums.LoginResultFail, in)
		return nil, err
	}

	member, err := s.findBoundMember(ctx, socialUser)
	if err != nil {
		s.recordLogin(ctx, 0, socialUser.Nickname, baseEnums.LoginResultFail, in)
		return nil, err
	}
	if member.Status != baseEnums.CommonStatusEnabled {
		s.recordLogin(ctx, member.ID, member.Nickname, baseEnums.LoginResultFail, in)
		return nil, baseErrs.New(memberErrs.MemberDisabled)
	}

	tokens, err := baseAuth.IssueScoped(baseAuth.ScopeMember, member.ID)
	if err != nil {
		s.recordLogin(ctx, member.ID, member.Nickname, baseEnums.LoginResultFail, in)
		return nil, err
	}
	s.recordLogin(ctx, member.ID, member.Nickname, baseEnums.LoginResultSuccess, in)

	// 登录信息只用于展示，回写失败不阻断登录。
	now := types.Now()
	_ = s.members.UpdateLoginInfo(ctx, member.ID, in.IP, now)
	member.LoginIP = in.IP
	member.LoginDate = &now

	return &SocialLoginResult{TokenPair: tokens, Member: member}, nil
}

// Bind 把第三方账号绑定到当前会员。
func (s *MemberSocialService) Bind(ctx context.Context, in SocialBindInput) error {
	member, err := s.members.FindByID(ctx, in.MemberID)
	if err != nil {
		return err
	}
	if member == nil {
		return baseErrs.New(memberErrs.MemberNotExists, in.MemberID)
	}
	if member.Status != baseEnums.CommonStatusEnabled {
		return baseErrs.New(memberErrs.MemberDisabled)
	}

	profile, err := s.fetchProfile(ctx, in.Type, in.Code)
	if err != nil {
		return err
	}

	socialUser, err := s.ensureSocialUser(ctx, profile)
	if err != nil {
		return err
	}

	if binding, err := s.socials.FindBindingByUserType(ctx, in.MemberID, socialUser.Type); err != nil {
		return err
	} else if binding != nil {
		return baseErrs.New(memberErrs.MemberSocialTypeDuplicated, SocialTypeName(socialUser.Type))
	}

	if binding, err := s.socials.FindBindingBySocialUser(ctx, socialUser.ID); err != nil {
		return err
	} else if binding != nil {
		if binding.UserID == in.MemberID {
			return nil
		}
		return baseErrs.New(memberErrs.MemberSocialAlreadyBound)
	}

	err = s.socials.CreateBinding(ctx, nil, &model.MemberSocialBindingModel{
		UserID:       in.MemberID,
		SocialType:   socialUser.Type,
		SocialUserID: socialUser.ID,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(memberErrs.MemberSocialAlreadyBound)
	}

	return err
}

// Bindings 返回会员已绑定的第三方账号，供会员端与管理端查看。
func (s *MemberSocialService) Bindings(ctx context.Context, memberID types.ID) ([]SocialBindingItem, error) {
	bindings, err := s.socials.Bindings(ctx, memberID)
	if err != nil {
		return nil, err
	}

	ids := make([]any, 0, len(bindings))
	for _, item := range bindings {
		ids = append(ids, item.SocialUserID)
	}
	users, err := s.socials.Users(ctx, ids)
	if err != nil {
		return nil, err
	}

	items := make([]SocialBindingItem, 0, len(bindings))
	for _, item := range bindings {
		socialUser := users[item.SocialUserID]
		items = append(items, SocialBindingItem{
			ID:         item.ID,
			SocialType: item.SocialType,
			Nickname:   socialUser.Nickname,
			Avatar:     socialUser.Avatar,
			CreatedAt:  item.CreatedAt,
		})
	}

	return items, nil
}

// Unbind 解绑第三方账号；至少保留一个绑定，避免会员彻底失去登录方式。
func (s *MemberSocialService) Unbind(ctx context.Context, memberID, bindingID types.ID) error {
	binding, err := s.socials.FindBindingByID(ctx, memberID, bindingID)
	if err != nil {
		return err
	}
	if binding == nil {
		return baseErrs.New(memberErrs.MemberSocialBindingNotExists, bindingID)
	}

	total, err := s.socials.CountBindings(ctx, memberID)
	if err != nil {
		return err
	}
	if total <= 1 {
		return baseErrs.New(memberErrs.MemberSocialLastBinding)
	}

	return s.socials.DeleteBinding(ctx, binding)
}

// recordLogin 写会员登录日志，user_type 标记为会员，与后台登录日志区分。
func (s *MemberSocialService) recordLogin(
	ctx context.Context,
	memberID types.ID,
	username string,
	result baseEnums.LoginResultEnum,
	in SocialLoginInput,
) {
	baseAuth.RecordLoginLog(ctx, baseAuth.LoginLogInput{
		UserID:    memberID,
		UserType:  baseEnums.UserTypeMember,
		Username:  username,
		Type:      baseEnums.LoginTypeLogin,
		Result:    result,
		IP:        in.IP,
		UserAgent: in.UserAgent,
	})
}

// Me 返回当前登录会员的档案。
func (s *MemberSocialService) Me(ctx context.Context, memberID types.ID) (*model.MemberUserModel, error) {
	member, err := s.members.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, baseErrs.New(memberErrs.MemberNotExists, memberID)
	}
	if member.Status != baseEnums.CommonStatusEnabled {
		return nil, baseErrs.New(memberErrs.MemberDisabled)
	}

	return member, nil
}

// Refresh 续期会员令牌。
func (s *MemberSocialService) Refresh(refreshToken string) (*baseAuth.TokenPair, error) {
	return baseAuth.RefreshScoped(baseAuth.ScopeMember, refreshToken)
}

// Logout 注销会员令牌。
func (s *MemberSocialService) Logout(accessToken, refreshToken string) {
	baseAuth.RevokeScoped(baseAuth.ScopeMember, accessToken, refreshToken)
}

// provider 按渠道从后台配置的 member_social_client 读取凭据并构造渠道实现。
func (s *MemberSocialService) provider(ctx context.Context, socialType string) (socialProvider, error) {
	client, err := s.clients.FindEnabledByType(ctx, socialType)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, baseErrs.New(memberErrs.MemberSocialClientUnavailable, SocialTypeName(socialType))
	}

	return newSocialProvider(socialType, client.ClientID, client.ClientSecret)
}

// fetchProfile 用 code 换取第三方身份信息。
func (s *MemberSocialService) fetchProfile(ctx context.Context, socialType, code string) (*socialProfile, error) {
	provider, err := s.provider(ctx, socialType)
	if err != nil {
		return nil, err
	}

	return provider.User(ctx, strings.TrimSpace(code))
}

// ensureSocialUser 查询或创建社交用户，并刷新可能变化的资料。
func (s *MemberSocialService) ensureSocialUser(ctx context.Context, profile *socialProfile) (*model.MemberSocialUserModel, error) {
	existing, err := s.socials.FindUser(ctx, profile.Type, profile.OpenID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		changed := existing.UnionID != profile.UnionID ||
			(existing.Nickname == "" && profile.Nickname != "") ||
			(existing.Avatar == "" && profile.Avatar != "")
		if changed {
			existing.UnionID = profile.UnionID
			if existing.Nickname == "" {
				existing.Nickname = profile.Nickname
			}
			if existing.Avatar == "" {
				existing.Avatar = profile.Avatar
			}
			existing.RawData = profile.RawData
			if err := s.socials.UpdateUser(ctx, existing); err != nil {
				return nil, err
			}
		}

		return existing, nil
	}

	m := &model.MemberSocialUserModel{
		Type:     profile.Type,
		OpenID:   profile.OpenID,
		UnionID:  profile.UnionID,
		Nickname: profile.Nickname,
		Avatar:   profile.Avatar,
		RawData:  profile.RawData,
	}
	if err := s.socials.CreateUser(ctx, nil, m); err != nil {
		// 并发首登时另一个请求已插入，回查即可。
		if support.IsDuplicate(err) {
			user, findErr := s.socials.FindUser(ctx, profile.Type, profile.OpenID)
			if findErr != nil {
				return nil, findErr
			}
			if user == nil {
				return nil, baseErrs.New(memberErrs.MemberSocialRequestFailed, "社交账号写入冲突，请重试")
			}

			return user, nil
		}

		return nil, err
	}

	return m, nil
}

// findBoundMember 按绑定关系找会员；未绑定直接返回业务错误，不自动注册。
func (s *MemberSocialService) findBoundMember(ctx context.Context, socialUser *model.MemberSocialUserModel) (*model.MemberUserModel, error) {
	binding, err := s.socials.FindBindingBySocialUser(ctx, socialUser.ID)
	if err != nil {
		return nil, err
	}
	if binding == nil {
		return nil, baseErrs.New(memberErrs.MemberSocialAccountNotBound)
	}

	member, err := s.members.FindByID(ctx, binding.UserID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, baseErrs.New(memberErrs.MemberNotExists, binding.UserID)
	}

	return member, nil
}
