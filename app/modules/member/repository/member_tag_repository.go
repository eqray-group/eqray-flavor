package repository

import (
	"context"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
)

const tagTable = model.MemberTagTable

// MemberTagFilter 是标签列表的筛选条件。
type MemberTagFilter struct {
	Keyword string
	Status  *baseEnums.CommonStatusEnum
}

// MemberTagRepository 是 member_tag 表的仓储。
type MemberTagRepository struct{}

func NewMemberTagRepository() *MemberTagRepository {
	return &MemberTagRepository{}
}

// Create 新增会员标签。
func (r *MemberTagRepository) Create(ctx context.Context, m *model.MemberTagModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(tagTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// Update 更新标签名称、状态与备注。
func (r *MemberTagRepository) Update(ctx context.Context, m *model.MemberTagModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(tagTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// Delete 软删除标签；记录不存在时返回业务错误码。
func (r *MemberTagRepository) Delete(ctx context.Context, m *model.MemberTagModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(tagTable).
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
		return baseErrs.New(memberErrs.MemberTagNotExists, m.ID)
	}

	return nil
}

// FindByID 按主键查询标签，不存在返回 nil。
func (r *MemberTagRepository) FindByID(ctx context.Context, id types.ID) (*model.MemberTagModel, error) {
	var m model.MemberTagModel

	if err := facades.DB().WithContext(ctx).Table(tagTable).
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

// FindByIDs 批量查询标签，供 C 端「我的资料」一次性带出标签名，避免 N+1。
func (r *MemberTagRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.MemberTagModel, error) {
	list := make([]model.MemberTagModel, 0)
	if len(ids) == 0 {
		return list, nil
	}

	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}

	if err := facades.DB().WithContext(ctx).Table(tagTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Enabled 返回启用中的标签，供会员表单多选使用。
func (r *MemberTagRepository) Enabled(ctx context.Context) ([]model.MemberTagModel, error) {
	list := make([]model.MemberTagModel, 0)

	if err := facades.DB().WithContext(ctx).Table(tagTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 分页查询标签，支持名称模糊匹配与状态筛选。
func (r *MemberTagRepository) Paginate(ctx context.Context, filter MemberTagFilter, pageNo, pageSize int) ([]model.MemberTagModel, int64, error) {
	list := make([]model.MemberTagModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(tagTable).WhereNull("deleted_at")
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
