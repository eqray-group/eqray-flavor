package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/model"
	"eqray-flavor/app/modules/system/repository"
)

const (
	// 权限数据变更时前移版本号，旧缓存由其 TTL 自然过期（docs 9.3）
	permissionVersionKey = "rbac:version"
	permissionCacheTTL   = 30 * time.Minute
	// 内置超管角色标识，命中后跳过接口级权限校验（docs 9.3）
	superAdminCode = "SUPER_ADMIN"
)

// PermissionService 负责 RBAC 权限集合的计算与缓存。
type PermissionService struct {
	userRoles *repository.UserRoleRepository
	roles     *repository.RoleRepository
	roleMenus *repository.RoleMenuRepository
	menus     *repository.MenuRepository
}

func NewPermissionService() *PermissionService {
	return &PermissionService{
		userRoles: repository.NewUserRoleRepository(),
		roles:     repository.NewRoleRepository(),
		roleMenus: repository.NewRoleMenuRepository(),
		menus:     repository.NewMenuRepository(),
	}
}

// IsSuperAdmin 判断用户是否拥有内置超管角色。
func (s *PermissionService) IsSuperAdmin(ctx context.Context, userID types.ID) (bool, error) {
	roles, err := s.enabledRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Code == superAdminCode {
			return true, nil
		}
	}

	return false, nil
}

// Allows 判断用户是否拥有指定权限标识，供 base 的权限中间件注入调用。
// 超级管理员直接放行，与中间件原有的放行规则一致（docs 9.3）。
func (s *PermissionService) Allows(ctx context.Context, userID types.ID, code string) (bool, error) {
	super, err := s.IsSuperAdmin(ctx, userID)
	if err != nil {
		return false, err
	}
	if super {
		return true, nil
	}

	codes, err := s.Permissions(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, item := range codes {
		if item == code {
			return true, nil
		}
	}

	return false, nil
}

// RoleCodes 返回用户已启用角色的标识集合，登录信息展示用。
func (s *PermissionService) RoleCodes(ctx context.Context, userID types.ID) ([]string, error) {
	roles, err := s.enabledRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
	}

	return codes, nil
}

// Menus 返回用户可访问的启用菜单（含按钮），用于生成前端动态路由（后端权限模式）。
// 超级管理员直接返回全部启用菜单，与 Permissions 的放行规则一致。
func (s *PermissionService) Menus(ctx context.Context, userID types.ID) ([]model.SystemMenuModel, error) {
	all, err := s.menus.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	enabled := make([]model.SystemMenuModel, 0, len(all))
	for _, menu := range all {
		if menu.Status == baseEnums.CommonStatusEnabled {
			enabled = append(enabled, menu)
		}
	}

	roles, err := s.enabledRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.Code == superAdminCode {
			return enabled, nil
		}
	}
	if len(roles) == 0 {
		return []model.SystemMenuModel{}, nil
	}

	roleIDs := make([]types.ID, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}

	menuIDs, err := s.roleMenus.ListMenuIDsByRoles(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	granted := make(map[types.ID]bool, len(menuIDs))
	for _, id := range menuIDs {
		granted[id] = true
	}

	menus := make([]model.SystemMenuModel, 0, len(enabled))
	for _, menu := range enabled {
		if granted[menu.ID] {
			menus = append(menus, menu)
		}
	}

	return menus, nil
}

// Permissions 返回用户拥有的权限标识集合，命中 Redis 缓存时直接返回。
// 超级管理员返回全部权限（用于前端按钮展示），中间件另有直接放行通道。
func (s *PermissionService) Permissions(ctx context.Context, userID types.ID) ([]string, error) {
	cacheKey := s.permissionCacheKey(userID)
	if cached := facades.Cache().GetString(cacheKey); cached != "" {
		return strings.Split(cached, ","), nil
	}

	permissions, err := s.loadPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(permissions) > 0 {
		_ = facades.Cache().Put(cacheKey, strings.Join(permissions, ","), permissionCacheTTL)
	}

	return permissions, nil
}

// Invalidate 角色、角色授权、用户角色或菜单权限变化后调用，让所有用户的权限缓存失效。
func (s *PermissionService) Invalidate() {
	_ = facades.Cache().Put(
		permissionVersionKey,
		strconv.FormatInt(time.Now().UnixNano(), 10),
		permissionCacheTTL,
	)
}

// loadPermissions 按 用户 → 角色 → 菜单权限 计算权限集合。
func (s *PermissionService) loadPermissions(ctx context.Context, userID types.ID) ([]string, error) {
	roles, err := s.enabledRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return []string{}, nil
	}

	for _, role := range roles {
		if role.Code == superAdminCode {
			return s.menus.ListPermissions(ctx)
		}
	}

	roleIDs := make([]types.ID, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}

	menuIDs, err := s.roleMenus.ListMenuIDsByRoles(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	return s.menus.ListPermissionsByIDs(ctx, menuIDs)
}

// enabledRoles 返回用户绑定的启用角色。
func (s *PermissionService) enabledRoles(ctx context.Context, userID types.ID) ([]model.SystemRoleModel, error) {
	roleIDs, err := s.userRoles.ListRoleIDsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []model.SystemRoleModel{}, nil
	}

	roles, err := s.roles.FindByIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	enabled := make([]model.SystemRoleModel, 0, len(roles))
	for _, role := range roles {
		if role.Status == baseEnums.CommonStatusEnabled {
			enabled = append(enabled, role)
		}
	}

	return enabled, nil
}

// permissionCacheKey 带版本号，版本变化后旧键自然失效。
func (s *PermissionService) permissionCacheKey(userID types.ID) string {
	version := facades.Cache().GetString(permissionVersionKey)
	if version == "" {
		version = "0"
	}

	return "rbac:perms:" + userID.String() + ":" + version
}
