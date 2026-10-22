// Package repository 存放 system 模块的数据访问实现，只操作本模块的表。
package repository

import (
	"context"
	"strconv"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
)

const userTable = model.SystemUserTable

// UserFilter 是用户列表的筛选条件，零值表示不限制。
type UserFilter struct {
	Username string
	Nickname string
	Mobile   string
	Status   *baseEnums.CommonStatusEnum
}

// UserRepository 是 system_users 表的仓储。
type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Paginate 按条件分页查询，只查未软删的数据。
func (r *UserRepository) Paginate(ctx context.Context, filter UserFilter, pageNo, pageSize int) ([]model.SystemUserModel, int64, error) {
	list := make([]model.SystemUserModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(userTable).WhereNull("deleted_at")
	if filter.Username != "" {
		query = query.WhereLike("username", "%"+filter.Username+"%")
	}
	if filter.Nickname != "" {
		query = query.WhereLike("nickname", "%"+filter.Nickname+"%")
	}
	if filter.Mobile != "" {
		query = query.WhereLike("mobile", "%"+filter.Mobile+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ExistsBy 判断某一列是否已被其他用户占用；excludeID > 0 时排除自身。
func (r *UserRepository) ExistsBy(ctx context.Context, column, value string, excludeID types.ID) (bool, error) {
	if value == "" {
		return false, nil
	}

	var m model.SystemUserModel

	query := facades.DB().WithContext(ctx).Table(userTable).
		Where(column, value).
		WhereNull("deleted_at")
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	if err := query.First(&m); err != nil {
		return false, err
	}

	return m.ID != 0, nil
}

// FindByID 按主键查询未软删的用户；未命中返回 (nil, nil)。
func (r *UserRepository) FindByID(ctx context.Context, id types.ID) (*model.SystemUserModel, error) {
	var m model.SystemUserModel

	if err := facades.DB().WithContext(ctx).Table(userTable).
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

// FindByUsername 按登录名查询未软删的用户；未命中返回 (nil, nil)。
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.SystemUserModel, error) {
	var m model.SystemUserModel

	if err := facades.DB().WithContext(ctx).Table(userTable).
		Where("username", username).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}

	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// Create 插入用户；email / mobile 为 nil 时落库为 NULL（docs 6.8(1)）。
func (r *UserRepository) Create(ctx context.Context, m *model.SystemUserModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(userTable).Insert(map[string]any{
		"id":         m.ID,
		"username":   m.Username,
		"password":   m.Password,
		"nickname":   m.Nickname,
		"remark":     m.Remark,
		"email":      m.Email,
		"mobile":     m.Mobile,
		"sex":        m.Sex,
		"avatar":     m.Avatar,
		"status":     m.Status,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(systemErrs.SystemUserDataExists)
	}

	return err
}

// Update 更新用户资料，不改密码；唯一键冲突同样转成业务错误码。
func (r *UserRepository) Update(ctx context.Context, m *model.SystemUserModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(userTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"username":   m.Username,
			"nickname":   m.Nickname,
			"remark":     m.Remark,
			"email":      m.Email,
			"mobile":     m.Mobile,
			"sex":        m.Sex,
			"avatar":     m.Avatar,
			"status":     m.Status,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(systemErrs.SystemUserDataExists)
	}

	return err
}

// Delete 软删除；RowsAffected 为 0 说明记录已被删除或不存在。
func (r *UserRepository) Delete(ctx context.Context, m *model.SystemUserModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(userTable).
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
		return baseErrs.New(systemErrs.SystemUserNotExists, m.ID)
	}

	return nil
}

// UpdatePassword 只更新密码哈希，不触碰其它字段。
func (r *UserRepository) UpdatePassword(ctx context.Context, id types.ID, hash string) error {
	updater := ""
	if uid := support.UserID(ctx); uid != 0 {
		updater = strconv.FormatInt(uid, 10)
	}

	res, err := facades.DB().WithContext(ctx).Table(userTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"password":   hash,
			"updater":    updater,
			"updated_at": types.Now(),
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(systemErrs.SystemUserNotExists, id)
	}

	return nil
}

// UpdateLoginInfo 记录最后登录时间与 IP，失败不影响登录结果。
func (r *UserRepository) UpdateLoginInfo(ctx context.Context, id types.ID, ip string, at types.DateTime) error {
	_, err := facades.DB().WithContext(ctx).Table(userTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"login_ip":      ip,
			"last_login_at": at,
			"updated_at":    at,
		})

	return err
}

// FindByIDs 批量查询未软删的用户，供跨模块只读账号目录使用。
// 入参：ids 用户 ID 集合；返回用户列表，未命中的 ID 直接忽略。
func (r *UserRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.SystemUserModel, error) {
	list := make([]model.SystemUserModel, 0)
	if len(ids) == 0 {
		return list, nil
	}

	values := make([]any, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			values = append(values, id)
		}
	}
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(userTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// SearchEnabled 按账号 / 昵称 / 手机号搜索启用用户，供员工选择等下拉场景使用。
// 入参：keyword 关键字（空串不过滤）、limit 最大返回数；返回用户列表。
func (r *UserRepository) SearchEnabled(ctx context.Context, keyword string, limit int) ([]model.SystemUserModel, error) {
	list := make([]model.SystemUserModel, 0)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := facades.DB().WithContext(ctx).Table(userTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at")
	if keyword != "" {
		like := "%" + keyword + "%"
		// 关键字命中账号、昵称或手机号，分组包裹避免与状态条件串成 OR
		query = query.Where(func(q db.Query) db.Query {
			return q.WhereLike("username", like).
				OrWhereLike("nickname", like).
				OrWhereLike("mobile", like)
		})
	}

	if err := query.OrderByDesc("id").Limit(uint64(limit)).Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}
