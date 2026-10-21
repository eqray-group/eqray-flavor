// Package repository 存放 system 模块的数据访问实现，只操作本模块的表。
package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
)

const userRoleTable = model.SystemUserRoleTable

// UserRoleRepository 是 system_user_role 关联表的仓储。
type UserRoleRepository struct{}

func NewUserRoleRepository() *UserRoleRepository {
	return &UserRoleRepository{}
}

// ListRoleIDsByUser 返回用户当前绑定的角色 ID（未软删）。
func (r *UserRoleRepository) ListRoleIDsByUser(ctx context.Context, userID types.ID) ([]types.ID, error) {
	rows := make([]model.SystemUserRoleModel, 0)

	if err := facades.DB().WithContext(ctx).Table(userRoleTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		Get(&rows); err != nil {
		return nil, err
	}

	ids := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.RoleID)
	}

	return ids, nil
}

// CountUsersByRole 统计仍引用该角色的用户数，删除角色前校验。
func (r *UserRoleRepository) CountUsersByRole(ctx context.Context, roleID types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(userRoleTable).
		Where("role_id", roleID).
		WhereNull("deleted_at").
		Count()
}

// Sync 把用户的角色集合对齐到 roleIDs：多余的软删除，缺少的补插入。
// 由 Service 开启事务后传入 tx，保证多写一致（docs 3）。
func (r *UserRoleRepository) Sync(ctx context.Context, tx db.Tx, userID types.ID, roleIDs []types.ID) error {
	rows := make([]model.SystemUserRoleModel, 0)
	if err := tx.Table(userRoleTable).
		Select("id", "role_id").
		Where("user_id", userID).
		WhereNull("deleted_at").
		Get(&rows); err != nil {
		return err
	}

	wanted := toIDSet(roleIDs)
	alive := make(map[types.ID]bool, len(rows))

	for _, row := range rows {
		alive[row.RoleID] = true
		if wanted[row.RoleID] {
			continue
		}
		// 关联表不再软删除，直接物理删除，避免占用 (user_id, role_id) 唯一键
		if _, err := tx.Table(userRoleTable).
			Where("id", row.ID).
			Delete(); err != nil {
			return err
		}
	}

	for _, roleID := range roleIDs {
		if roleID <= 0 || alive[roleID] {
			continue
		}

		m := &model.SystemUserRoleModel{UserID: userID, RoleID: roleID}
		m.FillCreate(ctx)

		if _, err := tx.Table(userRoleTable).Insert(map[string]any{
			"id":         m.ID,
			"user_id":    m.UserID,
			"role_id":    m.RoleID,
			"creator":    m.Creator,
			"created_at": m.CreatedAt,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		}); err != nil {
			return err
		}
	}

	return nil
}

// toIDSet 把 ID 集合转成去重 set。
func toIDSet(ids []types.ID) map[types.ID]bool {
	set := make(map[types.ID]bool, len(ids))
	for _, id := range ids {
		if id > 0 {
			set[id] = true
		}
	}

	return set
}
