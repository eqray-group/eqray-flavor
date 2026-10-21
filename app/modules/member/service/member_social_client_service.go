package service

import (
	"context"
	"strings"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

// MemberSocialClientSaveInput 是新增 / 编辑第三方登录配置的入参。
type MemberSocialClientSaveInput struct {
	Name         string
	SocialType   string
	ClientID     string
	ClientSecret string
	AgentID      string
	Status       int8
}

// MemberSocialClientListQuery 是配置列表的查询条件。
type MemberSocialClientListQuery struct {
	Keyword  string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// MemberSocialClientService 负责第三方登录渠道配置。
type MemberSocialClientService struct {
	clients *repository.MemberSocialClientRepository
}

// NewMemberSocialClientService 构造第三方登录渠道配置服务。
func NewMemberSocialClientService() *MemberSocialClientService {
	return &MemberSocialClientService{clients: repository.NewMemberSocialClientRepository()}
}

// Page 分页查询渠道配置。
func (s *MemberSocialClientService) Page(ctx context.Context, q MemberSocialClientListQuery) ([]model.MemberSocialClientModel, int64, error) {
	return s.clients.Paginate(ctx, repository.MemberSocialClientFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Status:  q.Status,
	}, q.PageNo, q.PageSize)
}

// Detail 查询配置详情。
func (s *MemberSocialClientService) Detail(ctx context.Context, id types.ID) (*model.MemberSocialClientModel, error) {
	m, err := s.clients.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(memberErrs.MemberSocialClientNotExists, id)
	}

	return m, nil
}

// Create 新增渠道配置：渠道唯一。
func (s *MemberSocialClientService) Create(ctx context.Context, in MemberSocialClientSaveInput) (types.ID, error) {
	if err := normalizeClientInput(&in, true); err != nil {
		return 0, err
	}

	exists, err := s.clients.ExistsByType(ctx, in.SocialType, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(memberErrs.MemberSocialClientTypeExists, SocialTypeName(in.SocialType))
	}

	m := &model.MemberSocialClientModel{
		Name:         in.Name,
		SocialType:   in.SocialType,
		ClientID:     in.ClientID,
		ClientSecret: in.ClientSecret,
		AgentID:      in.AgentID,
		Status:       baseEnums.CommonStatusEnum(in.Status),
	}
	if err := s.clients.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑渠道配置；client_secret 留空表示不修改。
func (s *MemberSocialClientService) Update(ctx context.Context, id types.ID, in MemberSocialClientSaveInput) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	if err := normalizeClientInput(&in, false); err != nil {
		return err
	}

	m.Name = in.Name
	m.ClientID = in.ClientID
	m.AgentID = in.AgentID
	m.Status = baseEnums.CommonStatusEnum(in.Status)
	if in.ClientSecret != "" {
		m.ClientSecret = in.ClientSecret
	}

	return s.clients.Update(ctx, m)
}

// Delete 删除渠道配置。
func (s *MemberSocialClientService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	return s.clients.Delete(ctx, m)
}

// normalizeClientInput 清洗入参并校验取值范围。
func normalizeClientInput(in *MemberSocialClientSaveInput, needSecret bool) error {
	in.Name = strings.TrimSpace(in.Name)
	in.SocialType = strings.TrimSpace(in.SocialType)
	in.ClientID = strings.TrimSpace(in.ClientID)
	in.ClientSecret = strings.TrimSpace(in.ClientSecret)
	in.AgentID = strings.TrimSpace(in.AgentID)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "应用名称不能为空")
	}
	if !IsSupportedSocialType(in.SocialType) {
		return baseErrs.WithMsg(baseErrs.BadRequest, "不支持的第三方登录渠道")
	}
	if in.ClientID == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "ClientID 不能为空")
	}
	if needSecret && in.ClientSecret == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "ClientSecret 不能为空")
	}
	if !baseEnums.CommonStatusEnum(in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "状态取值不合法")
	}

	return nil
}
