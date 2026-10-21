package service

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/database/db"
	"golang.org/x/crypto/bcrypt"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
	"eqray-flavor/app/modules/system/repository"
)

// UserListQuery 是用户列表的查询条件，零值表示不限制。
type UserListQuery struct {
	Username string
	Nickname string
	Mobile   string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// UserSaveInput 是新增 / 编辑用户的入参，由 Controller 从请求体转换而来。
// 密码只在新增与重置密码时使用，编辑资料不会改动密码。
type UserSaveInput struct {
	Username string
	Password string
	Nickname string
	Remark   string
	Email    string
	Mobile   string
	Sex      int8
	Avatar   string
	Status   baseEnums.CommonStatusEnum
}

// UserService 负责 system_users 的业务规则（docs 6.3 / 6.8）。
type UserService struct {
	users      *repository.UserRepository
	userRoles  *repository.UserRoleRepository
	roles      *repository.RoleRepository
	permission *PermissionService
}

func NewUserService() *UserService {
	return &UserService{
		users:      repository.NewUserRepository(),
		userRoles:  repository.NewUserRoleRepository(),
		roles:      repository.NewRoleRepository(),
		permission: NewPermissionService(),
	}
}

// Page 分页查询用户列表。
func (s *UserService) Page(ctx context.Context, q UserListQuery) ([]model.SystemUserModel, int64, error) {
	return s.users.Paginate(ctx, repository.UserFilter{
		Username: strings.TrimSpace(q.Username),
		Nickname: strings.TrimSpace(q.Nickname),
		Mobile:   strings.TrimSpace(q.Mobile),
		Status:   q.Status,
	}, q.PageNo, q.PageSize)
}

// Detail 查询用户详情，不存在时返回业务错误码。
func (s *UserService) Detail(ctx context.Context, id types.ID) (*model.SystemUserModel, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, baseErrs.New(systemErrs.SystemUserNotExists, id)
	}

	return user, nil
}

// Create 新增用户：先查重再落库，密码统一用 bcrypt 哈希。
// 约束冲突在 Repository 已转成 SystemUserDataExists，此处的前置校验只为给出更精确的提示。
func (s *UserService) Create(ctx context.Context, in UserSaveInput) (types.ID, error) {
	in.Username = strings.ToLower(strings.TrimSpace(in.Username))
	if in.Username == "" {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "账号不能为空")
	}
	if in.Password == "" {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "密码不能为空")
	}
	if !in.Status.IsValid() {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "用户状态不正确")
	}
	if err := s.checkUnique(ctx, in, 0); err != nil {
		return 0, err
	}

	hash, err := hashPassword(in.Password)
	if err != nil {
		return 0, err
	}

	user := &model.SystemUserModel{
		Username: in.Username,
		Password: hash,
		Nickname: defaultString(strings.TrimSpace(in.Nickname), in.Username),
		Remark:   strings.TrimSpace(in.Remark),
		Email:    optionalString(in.Email),
		Mobile:   optionalString(in.Mobile),
		Sex:      in.Sex,
		Avatar:   strings.TrimSpace(in.Avatar),
		Status:   in.Status,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return 0, err
	}

	return user.ID, nil
}

// Update 更新用户资料，不涉及密码；停用当前登录账号会被拒绝。
func (s *UserService) Update(ctx context.Context, id types.ID, in UserSaveInput) error {
	in.Username = strings.ToLower(strings.TrimSpace(in.Username))
	if in.Username == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "账号不能为空")
	}
	if !in.Status.IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "用户状态不正确")
	}

	user, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}
	if support.UserID(ctx) == int64(id) && in.Status == baseEnums.CommonStatusDisabled {
		return baseErrs.New(systemErrs.SystemUserUpdateSelf)
	}
	if err := s.checkUnique(ctx, in, id); err != nil {
		return err
	}

	user.Username = in.Username
	user.Nickname = defaultString(strings.TrimSpace(in.Nickname), in.Username)
	user.Remark = strings.TrimSpace(in.Remark)
	user.Email = optionalString(in.Email)
	user.Mobile = optionalString(in.Mobile)
	user.Sex = in.Sex
	user.Avatar = strings.TrimSpace(in.Avatar)
	user.Status = in.Status

	return s.users.Update(ctx, user)
}

// Delete 软删除用户，禁止删除当前登录账号。
func (s *UserService) Delete(ctx context.Context, id types.ID) error {
	if support.UserID(ctx) == int64(id) {
		return baseErrs.New(systemErrs.SystemUserDeleteSelf)
	}

	user, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	return s.users.Delete(ctx, user)
}

// ResetPassword 重置指定用户的密码。
func (s *UserService) ResetPassword(ctx context.Context, id types.ID, password string) error {
	if password == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "密码不能为空")
	}

	if _, err := s.Detail(ctx, id); err != nil {
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	return s.users.UpdatePassword(ctx, id, hash)
}

// ProfileUpdateInput 是当前登录者编辑个人资料的入参。
type ProfileUpdateInput struct {
	Nickname string
	Email    string
	Mobile   string
	Sex      int8
	Avatar   string
}

// UpdateProfile 更新当前登录者自己的资料：账号、状态、密码都不在此接口调整（docs 9.3）。
// 头像留空表示不修改，避免前端未传时把已有头像清掉。
func (s *UserService) UpdateProfile(ctx context.Context, id types.ID, in ProfileUpdateInput) error {
	if in.Sex < 0 || in.Sex > 2 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "性别取值不合法")
	}

	user, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	email := strings.TrimSpace(in.Email)
	if exists, err := s.users.ExistsBy(ctx, "email", email, id); err != nil {
		return err
	} else if exists {
		return baseErrs.New(systemErrs.SystemEmailExists, email)
	}

	mobile := strings.TrimSpace(in.Mobile)
	if exists, err := s.users.ExistsBy(ctx, "mobile", mobile, id); err != nil {
		return err
	} else if exists {
		return baseErrs.New(systemErrs.SystemMobileExists, mobile)
	}

	user.Nickname = defaultString(strings.TrimSpace(in.Nickname), user.Username)
	user.Email = optionalString(email)
	user.Mobile = optionalString(mobile)
	user.Sex = in.Sex
	if avatar := strings.TrimSpace(in.Avatar); avatar != "" {
		user.Avatar = avatar
	}

	return s.users.Update(ctx, user)
}

// ChangePassword 先校验旧密码，再更新为新密码的哈希。
func (s *UserService) ChangePassword(ctx context.Context, id types.ID, oldPassword, newPassword string) error {
	if oldPassword == "" || newPassword == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "密码不能为空")
	}

	user, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)) != nil {
		return baseErrs.New(systemErrs.SystemOldPasswordWrong)
	}
	if oldPassword == newPassword {
		return baseErrs.WithMsg(baseErrs.BadRequest, "新密码不能与当前密码相同")
	}

	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.users.UpdatePassword(ctx, id, hash)
}

// RoleIDs 返回用户已绑定的角色 ID，分配角色弹窗回显用。
func (s *UserService) RoleIDs(ctx context.Context, id types.ID) ([]types.ID, error) {
	if _, err := s.Detail(ctx, id); err != nil {
		return nil, err
	}

	return s.userRoles.ListRoleIDsByUser(ctx, id)
}

// AssignRoles 覆盖式分配角色：校验角色存在且启用，事务内同步关联表，随后让权限缓存失效。
func (s *UserService) AssignRoles(ctx context.Context, id types.ID, roleIDs []types.ID) error {
	if _, err := s.Detail(ctx, id); err != nil {
		return err
	}

	roleIDs = uniqueIDs(roleIDs)
	if len(roleIDs) > 0 {
		roles, err := s.roles.FindByIDs(ctx, roleIDs)
		if err != nil {
			return err
		}

		enabled := 0
		for _, role := range roles {
			if role.Status == baseEnums.CommonStatusEnabled {
				enabled++
			}
		}
		if enabled != len(roleIDs) {
			return baseErrs.WithMsg(baseErrs.BadRequest, "存在无效或已停用的角色")
		}
	}

	err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.userRoles.Sync(ctx, tx, id, roleIDs)
	})
	if err != nil {
		return err
	}

	s.permission.Invalidate()

	return nil
}

// uniqueIDs 去重并过滤非法 ID。
func uniqueIDs(ids []types.ID) []types.ID {
	result := make([]types.ID, 0, len(ids))
	seen := make(map[types.ID]bool, len(ids))

	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}

	return result
}

// checkUnique 校验账号、邮箱、手机号是否被占用；excludeID > 0 时排除自身。
// 空邮箱 / 空手机号不参与查重，因为落库为 NULL（docs 6.8(1)）。
func (s *UserService) checkUnique(ctx context.Context, in UserSaveInput, excludeID types.ID) error {
	exists, err := s.users.ExistsBy(ctx, "username", in.Username, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(systemErrs.SystemUsernameExists, in.Username)
	}

	email := strings.TrimSpace(in.Email)
	if exists, err = s.users.ExistsBy(ctx, "email", email, excludeID); err != nil {
		return err
	}
	if exists {
		return baseErrs.New(systemErrs.SystemEmailExists, email)
	}

	mobile := strings.TrimSpace(in.Mobile)
	if exists, err = s.users.ExistsBy(ctx, "mobile", mobile, excludeID); err != nil {
		return err
	}
	if exists {
		return baseErrs.New(systemErrs.SystemMobileExists, mobile)
	}

	return nil
}

// hashPassword 生成 bcrypt 哈希，与登录时的比对保持一致。
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "密码长度不能超过 72 个字节")
	}

	return string(hash), nil
}

// optionalString 把空串转成 nil，保证唯一索引不比较 NULL（docs 6.8(1)）。
func optionalString(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}

	return &v
}

// defaultString 在 value 为空时返回 fallback。
func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
