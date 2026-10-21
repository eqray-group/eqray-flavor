package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/system/admin/request"
	"eqray-flavor/app/modules/system/service"
)

// UserController 提供 system_users 的 CRUD 接口。
type UserController struct {
	service *service.UserService
}

func NewUserController() *UserController {
	return &UserController{service: service.NewUserService()}
}

// Index 分页查询后台用户，支持账号、昵称、手机号与状态筛选。
// 入参：username、nickname、mobile、status、pageNo、pageSize（查询串）；返回用户列表与总数。
func (c *UserController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.UserListQuery{
		Username: ctx.Request().Query("username"),
		Nickname: ctx.Request().Query("nickname"),
		Mobile:   ctx.Request().Query("mobile"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := baseEnums.CommonStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 查询单个后台用户的详情。
// 入参：路径 id；返回用户详情，不存在返回业务错误码。
func (c *UserController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	user, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, user)
}

// Store 新增后台用户；账号唯一，密码由服务层加密后入库。
// 入参见 UserCreateRequest（账号、密码、昵称、邮箱、手机号等）；返回新用户 ID。
func (c *UserController) Store(ctx http.Context) http.Response {
	var req request.UserCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.UserSaveInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Remark:   req.Remark,
		Email:    req.Email,
		Mobile:   req.Mobile,
		Sex:      req.Sex,
		Avatar:   req.Avatar,
		Status:   req.Status,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑后台用户资料，密码不在本接口修改。
// 入参：路径 id + 用户字段；返回 null。
func (c *UserController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.UserUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.UserSaveInput{
		Username: req.Username,
		Nickname: req.Nickname,
		Remark:   req.Remark,
		Email:    req.Email,
		Mobile:   req.Mobile,
		Sex:      req.Sex,
		Avatar:   req.Avatar,
		Status:   req.Status,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除后台用户（软删），并清理其角色关联。
// 入参：路径 id；返回 null。
func (c *UserController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// ResetPassword 管理员重置指定用户的登录密码。
// 入参：路径 id + password（明文，服务层加密）；返回 null。
func (c *UserController) ResetPassword(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.UserResetPasswordRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.ResetPassword(ctx.Context(), id, req.Password); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Roles 查询指定用户已分配的角色 ID 列表，供角色分配弹窗回显。
// 入参：路径 id；返回角色 ID 数组。
func (c *UserController) Roles(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	roleIDs, err := c.service.RoleIDs(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, roleIDs)
}

// AssignRoles 覆盖式设置用户角色，传入的 ID 列表会替换原有角色，并刷新权限缓存版本。
// 入参：路径 id + roleIds；返回 null。
func (c *UserController) AssignRoles(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.UserAssignRoleRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.AssignRoles(ctx.Context(), id, req.RoleIDs); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
