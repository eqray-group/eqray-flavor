package service

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/enums"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
	"eqray-flavor/app/modules/system/repository"
)

// RoleListQuery 是角色列表的查询条件，零值表示不限制。
type RoleListQuery struct {
	Name     string
	Code     string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// RoleSaveInput 是新增 / 编辑角色的入参。
type RoleSaveInput struct {
	Name   string
	Code   string
	Sort   int
	Status baseEnums.CommonStatusEnum
	Remark string
}

// RoleService 负责 system_role 的业务规则与角色授权（docs 6.3、9.3）。
type RoleService struct {
	roles      *repository.RoleRepository
	userRoles  *repository.UserRoleRepository
	roleMenus  *repository.RoleMenuRepository
	menus      *repository.MenuRepository
	permission *PermissionService
}

func NewRoleService() *RoleService {
	return &RoleService{
		roles:      repository.NewRoleRepository(),
		userRoles:  repository.NewUserRoleRepository(),
		roleMenus:  repository.NewRoleMenuRepository(),
		menus:      repository.NewMenuRepository(),
		permission: NewPermissionService(),
	}
}

// Page 分页查询角色。
func (s *RoleService) Page(ctx context.Context, q RoleListQuery) ([]model.SystemRoleModel, int64, error) {
	return s.roles.Paginate(ctx, repository.RoleFilter{
		Name:   strings.TrimSpace(q.Name),
		Code:   strings.TrimSpace(q.Code),
		Status: q.Status,
	}, q.PageNo, q.PageSize)
}

// Options 返回全部启用角色，供用户分配角色时选择。
func (s *RoleService) Options(ctx context.Context) ([]model.SystemRoleModel, error) {
	return s.roles.ListEnabled(ctx)
}

// Detail 查询角色详情。
func (s *RoleService) Detail(ctx context.Context, id types.ID) (*model.SystemRoleModel, error) {
	role, err := s.roles.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, baseErrs.New(systemErrs.SystemRoleNotExists, id)
	}

	return role, nil
}

// Create 新增自定义角色，角色标识保留调用方传入的大小写。
func (s *RoleService) Create(ctx context.Context, in RoleSaveInput) (types.ID, error) {
	name, code, err := normalizeRole(in)
	if err != nil {
		return 0, err
	}

	exists, err := s.roles.ExistsByCode(ctx, code, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(systemErrs.SystemRoleCodeExists, code)
	}

	role := &model.SystemRoleModel{
		Name:   name,
		Code:   code,
		Sort:   in.Sort,
		Status: in.Status,
		Type:   enums.RoleTypeCustom,
		Remark: strings.TrimSpace(in.Remark),
	}
	if err := s.roles.Create(ctx, role); err != nil {
		return 0, err
	}

	s.permission.Invalidate()

	return role.ID, nil
}

// Update 编辑角色；内置角色不允许改标识或禁用，避免把自己锁在门外。
func (s *RoleService) Update(ctx context.Context, id types.ID, in RoleSaveInput) error {
	existing, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	name, code, err := normalizeRole(in)
	if err != nil {
		return err
	}

	if existing.Type == enums.RoleTypeBuiltin {
		if code != existing.Code || in.Status != baseEnums.CommonStatusEnabled {
			return baseErrs.New(systemErrs.SystemRoleBuiltin)
		}
	}

	exists, err := s.roles.ExistsByCode(ctx, code, existing.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(systemErrs.SystemRoleCodeExists, code)
	}

	existing.Name = name
	existing.Code = code
	existing.Sort = in.Sort
	existing.Status = in.Status
	existing.Remark = strings.TrimSpace(in.Remark)

	if err := s.roles.Update(ctx, existing); err != nil {
		return err
	}

	s.permission.Invalidate()

	return nil
}

// Delete 删除角色：内置角色与被引用的角色都不允许删除。
func (s *RoleService) Delete(ctx context.Context, id types.ID) error {
	role, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}
	if role.Type == enums.RoleTypeBuiltin {
		return baseErrs.New(systemErrs.SystemRoleBuiltin)
	}

	count, err := s.userRoles.CountUsersByRole(ctx, role.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(systemErrs.SystemRoleInUse)
	}

	// 角色与其授权关系一起删除，避免留下孤儿授权
	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.roleMenus.DeleteByRole(ctx, tx, role.ID); err != nil {
			return err
		}

		return s.roles.Delete(ctx, role)
	})
	if err != nil {
		return err
	}

	s.permission.Invalidate()

	return nil
}

// MenuIDs 返回角色已授权的菜单 ID，前端回显勾选状态。
func (s *RoleService) MenuIDs(ctx context.Context, id types.ID) ([]types.ID, error) {
	if _, err := s.Detail(ctx, id); err != nil {
		return nil, err
	}

	return s.roleMenus.ListMenuIDsByRole(ctx, id)
}

// AssignMenus 覆盖式授权：把角色可访问的菜单对齐到 menuIDs。
func (s *RoleService) AssignMenus(ctx context.Context, id types.ID, menuIDs []types.ID) error {
	role, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	// 校验菜单都存在，避免写入悬空授权
	if len(menuIDs) > 0 {
		list, err := s.menus.ListAll(ctx)
		if err != nil {
			return err
		}
		valid := make(map[types.ID]bool, len(list))
		for _, item := range list {
			valid[item.ID] = true
		}
		for _, menuID := range menuIDs {
			if menuID > 0 && !valid[menuID] {
				return baseErrs.WithMsg(baseErrs.BadRequest, "存在无效的菜单")
			}
		}
	}

	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.roleMenus.Sync(ctx, tx, role.ID, menuIDs)
	})
	if err != nil {
		return err
	}

	s.permission.Invalidate()

	return nil
}

// normalizeRole 归一化名称与角色标识，并校验必填与状态。
func normalizeRole(in RoleSaveInput) (string, string, error) {
	name := strings.TrimSpace(in.Name)
	code := strings.TrimSpace(in.Code)

	if name == "" {
		return "", "", baseErrs.WithMsg(baseErrs.BadRequest, "角色名称不能为空")
	}
	if code == "" {
		return "", "", baseErrs.WithMsg(baseErrs.BadRequest, "角色标识不能为空")
	}
	if !in.Status.IsValid() {
		return "", "", baseErrs.WithMsg(baseErrs.BadRequest, "角色状态不正确")
	}

	return name, code, nil
}
