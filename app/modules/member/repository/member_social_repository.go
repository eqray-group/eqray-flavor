package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/member/model"
)

// MemberSocialRepository 是第三方社交用户与绑定关系的仓储。
type MemberSocialRepository struct{}

func NewMemberSocialRepository() *MemberSocialRepository {
	return &MemberSocialRepository{}
}

// FindUser 按渠道 + openid 查询社交用户；未命中返回 (nil, nil)。
func (r *MemberSocialRepository) FindUser(ctx context.Context, socialType, openID string) (*model.MemberSocialUserModel, error) {
	var m model.MemberSocialUserModel

	if err := facades.DB().WithContext(ctx).Table(model.MemberSocialUserTable).
		Where("type", socialType).
		Where("openid", openID).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// CreateUser 插入社交用户，随调用方的连接（事务或普通连接）执行。
func (r *MemberSocialRepository) CreateUser(ctx context.Context, tx db.Tx, m *model.MemberSocialUserModel) error {
	m.FillCreate(ctx)

	_, err := executor(ctx, tx).Table(model.MemberSocialUserTable).Insert(map[string]any{
		"id":         m.ID,
		"type":       m.Type,
		"openid":     m.OpenID,
		"unionid":    m.UnionID,
		"nickname":   m.Nickname,
		"avatar":     m.Avatar,
		"raw_data":   m.RawData,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// UpdateUser 刷新社交用户资料（昵称、头像、unionid 可能随平台变更）。
func (r *MemberSocialRepository) UpdateUser(ctx context.Context, m *model.MemberSocialUserModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(model.MemberSocialUserTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"unionid":    m.UnionID,
			"nickname":   m.Nickname,
			"avatar":     m.Avatar,
			"raw_data":   m.RawData,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// FindBindingBySocialUser 按社交用户查绑定关系；未命中返回 (nil, nil)。
func (r *MemberSocialRepository) FindBindingBySocialUser(ctx context.Context, socialUserID types.ID) (*model.MemberSocialBindingModel, error) {
	var m model.MemberSocialBindingModel

	if err := facades.DB().WithContext(ctx).Table(model.MemberSocialBindingTable).
		Where("social_user_id", socialUserID).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindBindingByUserType 查某个会员在指定渠道上的绑定。
func (r *MemberSocialRepository) FindBindingByUserType(ctx context.Context, userID types.ID, socialType string) (*model.MemberSocialBindingModel, error) {
	var m model.MemberSocialBindingModel

	if err := facades.DB().WithContext(ctx).Table(model.MemberSocialBindingTable).
		Where("user_id", userID).
		Where("social_type", socialType).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindBindingByID 按主键查会员的绑定关系，防止越权解绑他人账号。
func (r *MemberSocialRepository) FindBindingByID(ctx context.Context, userID, id types.ID) (*model.MemberSocialBindingModel, error) {
	var m model.MemberSocialBindingModel

	if err := facades.DB().WithContext(ctx).Table(model.MemberSocialBindingTable).
		Where("id", id).
		Where("user_id", userID).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// CreateBinding 插入绑定关系。
func (r *MemberSocialRepository) CreateBinding(ctx context.Context, tx db.Tx, m *model.MemberSocialBindingModel) error {
	m.FillCreate(ctx)

	_, err := executor(ctx, tx).Table(model.MemberSocialBindingTable).Insert(map[string]any{
		"id":             m.ID,
		"user_id":        m.UserID,
		"social_type":    m.SocialType,
		"social_user_id": m.SocialUserID,
		"creator":        m.Creator,
		"created_at":     m.CreatedAt,
		"updater":        m.Updater,
		"updated_at":     m.UpdatedAt,
	})

	return err
}

// Bindings 查询会员的全部绑定关系。
func (r *MemberSocialRepository) Bindings(ctx context.Context, userID types.ID) ([]model.MemberSocialBindingModel, error) {
	list := make([]model.MemberSocialBindingModel, 0)

	if err := facades.DB().WithContext(ctx).Table(model.MemberSocialBindingTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		OrderByDesc("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// CountBindings 统计会员的绑定数量，用于解绑前的兜底保护。
func (r *MemberSocialRepository) CountBindings(ctx context.Context, userID types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(model.MemberSocialBindingTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		Count()
}

// DeleteBinding 硬删除绑定关系：解绑后同一渠道可重新绑定。
func (r *MemberSocialRepository) DeleteBinding(ctx context.Context, m *model.MemberSocialBindingModel) error {
	_, err := facades.DB().WithContext(ctx).Table(model.MemberSocialBindingTable).
		Where("id", m.ID).
		Delete()

	return err
}

// Users 批量查询社交用户，供绑定列表组装昵称与头像。
func (r *MemberSocialRepository) Users(ctx context.Context, ids []any) (map[types.ID]model.MemberSocialUserModel, error) {
	list := make([]model.MemberSocialUserModel, 0)
	if len(ids) == 0 {
		return map[types.ID]model.MemberSocialUserModel{}, nil
	}

	if err := facades.DB().WithContext(ctx).Table(model.MemberSocialUserTable).
		WhereIn("id", ids).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	result := make(map[types.ID]model.MemberSocialUserModel, len(list))
	for _, item := range list {
		result[item.ID] = item
	}

	return result, nil
}

// executor 统一事务与普通连接的写法，tx 为 nil 时走默认连接。
func executor(ctx context.Context, tx db.Tx) db.Tx {
	if tx != nil {
		return tx
	}

	return facades.DB().WithContext(ctx)
}
