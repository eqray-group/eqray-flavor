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

const levelTable = model.MemberLevelTable

// MemberLevelFilter 是会员等级列表的筛选条件，零值表示不限制。
type MemberLevelFilter struct {
	Keyword string
	Status  *baseEnums.CommonStatusEnum
}

// MemberLevelRepository 是 member_level 表的仓储。
type MemberLevelRepository struct{}

func NewMemberLevelRepository() *MemberLevelRepository {
	return &MemberLevelRepository{}
}

// Create 插入会员等级；等级序号冲突转换为业务错误码。
func (r *MemberLevelRepository) Create(ctx context.Context, m *model.MemberLevelModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(levelTable).Insert(map[string]any{
		"id":               m.ID,
		"level":            m.Level,
		"name":             m.Name,
		"icon":             m.Icon,
		"experience":       m.Experience,
		"discount_percent": m.DiscountPercent,
		"description":      m.Description,
		"status":           m.Status,
		"sort":             m.Sort,
		"creator":          m.Creator,
		"created_at":       m.CreatedAt,
		"updater":          m.Updater,
		"updated_at":       m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(memberErrs.MemberLevelExists, m.Level)
	}

	return err
}

// Update 更新等级可变字段；level 创建后不可修改，因此不在更新集合里。
func (r *MemberLevelRepository) Update(ctx context.Context, m *model.MemberLevelModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(levelTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":             m.Name,
			"icon":             m.Icon,
			"experience":       m.Experience,
			"discount_percent": m.DiscountPercent,
			"description":      m.Description,
			"status":           m.Status,
			"sort":             m.Sort,
			"updater":          m.Updater,
			"updated_at":       m.UpdatedAt,
		})

	return err
}

// Delete 软删除会员等级。
func (r *MemberLevelRepository) Delete(ctx context.Context, m *model.MemberLevelModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(levelTable).
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
		return baseErrs.New(memberErrs.MemberLevelNotExists, m.Level)
	}

	return nil
}

// FindByID 按主键查询未软删的等级；未命中返回 (nil, nil)。
func (r *MemberLevelRepository) FindByID(ctx context.Context, id types.ID) (*model.MemberLevelModel, error) {
	var m model.MemberLevelModel

	if err := facades.DB().WithContext(ctx).Table(levelTable).
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

// FindByLevel 按等级序号查询未软删的等级；未命中返回 (nil, nil)。
func (r *MemberLevelRepository) FindByLevel(ctx context.Context, level int) (*model.MemberLevelModel, error) {
	var m model.MemberLevelModel

	if err := facades.DB().WithContext(ctx).Table(levelTable).
		Where("level", level).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// Enabled 返回全部启用等级，按序号升序，用于会员表单下拉。
func (r *MemberLevelRepository) Enabled(ctx context.Context) ([]model.MemberLevelModel, error) {
	list := make([]model.MemberLevelModel, 0)

	if err := facades.DB().WithContext(ctx).Table(levelTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("level").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ExistsByLevel 判断等级序号是否已被占用，excludeID 用于编辑时排除自身。
func (r *MemberLevelRepository) ExistsByLevel(ctx context.Context, level int, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(levelTable).
		Where("level", level).
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

// CountMembersByLevelID 统计使用该等级的会员数量，用于删除前保护。
func (r *MemberLevelRepository) CountMembersByLevelID(ctx context.Context, levelID types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(model.MemberUserTable).
		Where("level_id", levelID).
		WhereNull("deleted_at").
		Count()
}

// Paginate 按条件分页查询会员等级。
func (r *MemberLevelRepository) Paginate(ctx context.Context, filter MemberLevelFilter, pageNo, pageSize int) ([]model.MemberLevelModel, int64, error) {
	list := make([]model.MemberLevelModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(levelTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderBy("sort").OrderBy("level").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
