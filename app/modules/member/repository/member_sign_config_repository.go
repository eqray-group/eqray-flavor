package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/member/model"
)

const signInConfigTable = model.MemberSignInConfigTable

// MemberSignConfigFilter 是签到档位列表的筛选条件，零值表示不限制。
type MemberSignConfigFilter struct {
	Day    *int
	Status *baseEnums.CommonStatusEnum
}

// MemberSignConfigRepository 是 member_sign_in_config 表的仓储：签到奖励档位。
type MemberSignConfigRepository struct{}

func NewMemberSignConfigRepository() *MemberSignConfigRepository {
	return &MemberSignConfigRepository{}
}

// Enabled 返回全部启用的档位，按 day 升序；签到按连续天数取档位时使用。
func (r *MemberSignConfigRepository) Enabled(ctx context.Context) ([]model.MemberSignInConfigModel, error) {
	list := make([]model.MemberSignInConfigModel, 0)

	if err := facades.DB().WithContext(ctx).Table(signInConfigTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderBy("day").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 分页查询档位，按 day 升序；支持按天数、状态筛选。
func (r *MemberSignConfigRepository) Paginate(ctx context.Context, filter MemberSignConfigFilter, pageNo, pageSize int) ([]model.MemberSignInConfigModel, int64, error) {
	list := make([]model.MemberSignInConfigModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(signInConfigTable).WhereNull("deleted_at")
	if filter.Day != nil {
		query = query.Where("day", *filter.Day)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderBy("day").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByID 按主键查询档位，不存在返回 nil。
func (r *MemberSignConfigRepository) FindByID(ctx context.Context, id types.ID) (*model.MemberSignInConfigModel, error) {
	var m model.MemberSignInConfigModel

	if err := facades.DB().WithContext(ctx).Table(signInConfigTable).
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

// ExistsByDay 判断第几天是否已被占用，excludeID 用于编辑时排除自身。
func (r *MemberSignConfigRepository) ExistsByDay(ctx context.Context, day int, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(signInConfigTable).
		Where("day", day).
		WhereNull("deleted_at")
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var m model.MemberSignInConfigModel
	if err := query.First(&m); err != nil {
		return false, err
	}

	return m.ID != 0, nil
}

// Create 新增档位。
func (r *MemberSignConfigRepository) Create(ctx context.Context, m *model.MemberSignInConfigModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(signInConfigTable).Insert(map[string]any{
		"id":         m.ID,
		"day":        m.Day,
		"point":      m.Point,
		"experience": m.Experience,
		"status":     m.Status,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// Update 更新档位。
func (r *MemberSignConfigRepository) Update(ctx context.Context, m *model.MemberSignInConfigModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(signInConfigTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"day":        m.Day,
			"point":      m.Point,
			"experience": m.Experience,
			"status":     m.Status,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// Delete 软删除档位。
func (r *MemberSignConfigRepository) Delete(ctx context.Context, m *model.MemberSignInConfigModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(signInConfigTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": m.UpdatedAt,
			"updated_at": m.UpdatedAt,
		})

	return err
}
