package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/store/admin/request"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/service"
)

// EmployeeController 提供门店员工的分配、编辑、删除与账号选择接口。
type EmployeeController struct {
	service *service.StoreEmployeeService
}

// NewEmployeeController 构造门店员工控制器。
func NewEmployeeController() *EmployeeController {
	return &EmployeeController{service: service.NewStoreEmployeeService()}
}

// Index 分页查询门店员工，支持账号关键字、角色与状态筛选。
// 门店范围由请求头门店上下文或当前用户可见门店集合决定。
func (c *EmployeeController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.EmployeeListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if role := ctx.Request().QueryInt("role", 0); role > 0 {
		v := enums.StoreEmployeeRoleEnum(role)
		q.Role = &v
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

// UserOptions 搜索可分配的后台账号，供员工弹窗远程选择。
// 入参：keyword 账号 / 昵称 / 手机号；返回最多 20 个启用账号。
func (c *EmployeeController) UserOptions(ctx http.Context) http.Response {
	list, err := c.service.UserOptions(ctx.Context(), ctx.Request().Query("keyword"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Store 新增门店员工；员工只能分配到当前用户可见的门店类型节点。
// 入参：storeId、userId、role、status、remark；返回新记录 ID。
func (c *EmployeeController) Store(ctx http.Context) http.Response {
	var req request.EmployeeSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toEmployeeSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑门店员工；允许调整门店、账号、角色、状态与备注。
// 入参：路径 id + 员工字段；返回 null。
func (c *EmployeeController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.EmployeeSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toEmployeeSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除门店员工关联；关联表物理删除。
// 入参：路径 id；返回 null。
func (c *EmployeeController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toEmployeeSaveInput 把请求体转换为服务层入参。
func toEmployeeSaveInput(req request.EmployeeSaveRequest) service.EmployeeSaveInput {
	return service.EmployeeSaveInput{
		StoreID: req.StoreID,
		UserID:  req.UserID,
		Role:    enums.StoreEmployeeRoleEnum(req.Role),
		Status:  baseEnums.CommonStatusEnum(req.Status),
		Remark:  req.Remark,
	}
}
