package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/member/model"
)

const addressTable = model.MemberAddressTable

// MemberAddressRepository 是 member_address 表的仓储。
// 地址由 C 端会员自己维护，所有查询都必须带 user_id，避免越权。
type MemberAddressRepository struct{}

func NewMemberAddressRepository() *MemberAddressRepository {
	return &MemberAddressRepository{}
}

// Create 插入一条收货地址。
func (r *MemberAddressRepository) Create(ctx context.Context, m *model.MemberAddressModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(addressTable).Insert(map[string]any{
		"id":             m.ID,
		"user_id":        m.UserID,
		"name":           m.Name,
		"mobile":         m.Mobile,
		"area_id":        m.AreaID,
		"detail_address": m.DetailAddress,
		"default_status": m.DefaultStatus,
		"creator":        m.Creator,
		"created_at":     m.CreatedAt,
		"updater":        m.Updater,
		"updated_at":     m.UpdatedAt,
	})

	return err
}

// Update 更新地址内容；带上 user_id 是为了防止改到别人的地址。
func (r *MemberAddressRepository) Update(ctx context.Context, m *model.MemberAddressModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(addressTable).
		Where("id", m.ID).
		Where("user_id", m.UserID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":           m.Name,
			"mobile":         m.Mobile,
			"area_id":        m.AreaID,
			"detail_address": m.DetailAddress,
			"default_status": m.DefaultStatus,
			"updater":        m.Updater,
			"updated_at":     m.UpdatedAt,
		})

	return err
}

// Delete 软删除地址；同样按 user_id 限定归属。
func (r *MemberAddressRepository) Delete(ctx context.Context, id, userID types.ID) error {
	now := types.Now()

	_, err := facades.DB().WithContext(ctx).Table(addressTable).
		Where("id", id).
		Where("user_id", userID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": now,
			"updated_at": now,
		})

	return err
}

// FindByID 查询当前会员名下的一条地址，不存在返回 nil。
func (r *MemberAddressRepository) FindByID(ctx context.Context, id, userID types.ID) (*model.MemberAddressModel, error) {
	var m model.MemberAddressModel

	if err := facades.DB().WithContext(ctx).Table(addressTable).
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

// ListByUser 查询当前会员的全部地址，默认地址排在最前。
func (r *MemberAddressRepository) ListByUser(ctx context.Context, userID types.ID) ([]model.MemberAddressModel, error) {
	list := make([]model.MemberAddressModel, 0)

	if err := facades.DB().WithContext(ctx).Table(addressTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		OrderByDesc("default_status").
		OrderByDesc("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ClearDefault 在事务内取消该会员的默认地址标记。
func (r *MemberAddressRepository) ClearDefault(ctx context.Context, tx db.Tx, userID types.ID) error {
	_, err := tx.Table(addressTable).
		Where("user_id", userID).
		Where("default_status", 1).
		WhereNull("deleted_at").
		Update(map[string]any{
			"default_status": 0,
			"updated_at":     types.Now(),
		})

	return err
}

// SetDefault 在事务内把指定地址设为默认。
func (r *MemberAddressRepository) SetDefault(ctx context.Context, tx db.Tx, id, userID types.ID) error {
	_, err := tx.Table(addressTable).
		Where("id", id).
		Where("user_id", userID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"default_status": 1,
			"updated_at":     types.Now(),
		})

	return err
}
