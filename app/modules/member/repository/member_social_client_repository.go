package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
)

const socialClientTable = model.MemberSocialClientTable

// MemberSocialClientRepository 是 member_social_client 表的仓储。
type MemberSocialClientRepository struct{}

func NewMemberSocialClientRepository() *MemberSocialClientRepository {
	return &MemberSocialClientRepository{}
}

// Create 新增渠道配置；渠道冲突转换为业务错误码。
func (r *MemberSocialClientRepository) Create(ctx context.Context, m *model.MemberSocialClientModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(socialClientTable).Insert(map[string]any{
		"id":            m.ID,
		"name":          m.Name,
		"social_type":   m.SocialType,
		"client_id":     m.ClientID,
		"client_secret": m.ClientSecret,
		"agent_id":      m.AgentID,
		"status":        m.Status,
		"creator":       m.Creator,
		"created_at":    m.CreatedAt,
		"updater":       m.Updater,
		"updated_at":    m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(memberErrs.MemberSocialClientTypeExists, m.SocialType)
	}

	return err
}

// Update 更新渠道配置；client_secret 传空时沿用原值，避免列表页脱敏后回写覆盖。
func (r *MemberSocialClientRepository) Update(ctx context.Context, m *model.MemberSocialClientModel) error {
	m.FillUpdate(ctx)

	data := map[string]any{
		"name":       m.Name,
		"client_id":  m.ClientID,
		"agent_id":   m.AgentID,
		"status":     m.Status,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	}
	if m.ClientSecret != "" {
		data["client_secret"] = m.ClientSecret
	}

	_, err := facades.DB().WithContext(ctx).Table(socialClientTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(data)

	return err
}

// Delete 软删除渠道配置。
func (r *MemberSocialClientRepository) Delete(ctx context.Context, m *model.MemberSocialClientModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(socialClientTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(memberErrs.MemberSocialClientNotExists, m.ID)
	}

	return nil
}

// FindByID 按主键查询未软删的配置；未命中返回 (nil, nil)。
func (r *MemberSocialClientRepository) FindByID(ctx context.Context, id types.ID) (*model.MemberSocialClientModel, error) {
	var m model.MemberSocialClientModel

	if err := facades.DB().WithContext(ctx).Table(socialClientTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindEnabledByType 按渠道查询启用中的配置，供登录时读取凭据。
func (r *MemberSocialClientRepository) FindEnabledByType(ctx context.Context, socialType string) (*model.MemberSocialClientModel, error) {
	var m model.MemberSocialClientModel

	if err := facades.DB().WithContext(ctx).Table(socialClientTable).
		Where("social_type", socialType).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ExistsByType 判断渠道是否已被占用，excludeID 用于编辑时排除自身。
func (r *MemberSocialClientRepository) ExistsByType(ctx context.Context, socialType string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(socialClientTable).
		Where("social_type", socialType).
		WhereNull("deleted_at")
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	total, err := query.Count()
	if err != nil {
		return false, err
	}

	return total > 0, nil
}

// Paginate 分页查询渠道配置。
func (r *MemberSocialClientRepository) Paginate(ctx context.Context, filter MemberSocialClientFilter, pageNo, pageSize int) ([]model.MemberSocialClientModel, int64, error) {
	list := make([]model.MemberSocialClientModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(socialClientTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// MemberSocialClientFilter 是渠道配置列表的筛选条件。
type MemberSocialClientFilter struct {
	Keyword string
	Status  *baseEnums.CommonStatusEnum
}
