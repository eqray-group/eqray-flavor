// Package controller 提供门店后台接口，只做取参、调用服务与统一响应。
package controller

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/admin/request"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/service"
)

// StoreController 提供门店的增删改查与组织树接口。
type StoreController struct {
	service *service.StoreService
}

// NewStoreController 构造门店控制器。
func NewStoreController() *StoreController {
	return &StoreController{service: service.NewStoreService()}
}

// Index 分页查询门店，支持名称 / 编码关键字与类型、状态筛选。
// 入参：keyword、type、status、entityId、pageNo、pageSize（查询串）；返回门店列表与总数。
// entityId 用于「组织树选中某节点，只看该节点下的门店」。
func (c *StoreController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.StoreListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if typ := ctx.Request().QueryInt("type", 0); typ > 0 {
		v := enums.StoreTypeEnum(typ)
		q.Type = &v
	}
	if status := ctx.Request().QueryInt("status", 0); status > 0 {
		v := enums.StoreStatusEnum(status)
		q.Status = &v
	}
	// 雪花 ID 走字符串解析，避免 QueryInt 在 32 位环境截断
	if raw := ctx.Request().Query("entityId"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "组织节点 ID 不合法"))
		}
		value := types.ID(parsed)
		q.EntityID = &value
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Tree 返回门店组织树，用于门店管理页的树形展示与上级节点选择。
// 无入参；返回以总部为根的嵌套节点数组。
func (c *StoreController) Tree(ctx http.Context) http.Response {
	tree, err := c.service.Tree(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tree)
}

// Visible 返回当前登录用户可见的门店节点（仅 type=3 的门店）。
// 供顶部门店切换器使用；超管与总部账号返回全部门店，门店员工按启用关联返回。
func (c *StoreController) Visible(ctx http.Context) http.Response {
	list, err := c.service.VisibleStores(ctx.Context(), types.ID(support.UserID(ctx.Context())))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show 查询单个门店详情。
// 入参：路径 id；返回门店详情，不存在返回业务错误码。
func (c *StoreController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// Store 新增门店；总部为根节点、大区挂总部、门店挂总部或大区。
// 入参：门店配置字段；返回新门店 ID。
func (c *StoreController) Store(ctx http.Context) http.Response {
	var req request.StoreSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑门店；编码创建后不可修改，入参中的 code 会被忽略。
// 入参：路径 id + 门店配置字段；返回 null。
func (c *StoreController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.StoreSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除门店；仍有下级节点时拒绝。
// 入参：路径 id；返回 null。
func (c *StoreController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toSaveInput 把请求体转换为服务层入参，避免 Controller 直接拼装模型。
// 入参：已通过校验的请求体；返回服务层可直接使用的门店字段。
func toSaveInput(req request.StoreSaveRequest) service.StoreSaveInput {
	return service.StoreSaveInput{
		ParentID:        req.ParentID,
		Type:            enums.StoreTypeEnum(req.Type),
		Code:            req.Code,
		Name:            req.Name,
		ShortName:       req.ShortName,
		Status:          enums.StoreStatusEnum(req.Status),
		ContactName:     req.ContactName,
		ContactMobile:   req.ContactMobile,
		DistrictID:      req.DistrictID,
		Address:         req.Address,
		Longitude:       req.Longitude,
		Latitude:        req.Latitude,
		OpenTime:        req.OpenTime,
		CloseTime:       req.CloseTime,
		DeliveryRule:    req.DeliveryRule,
		AutoAccept:      req.AutoAccept,
		OnlineEnabled:   req.OnlineEnabled,
		DeliveryEnabled: req.DeliveryEnabled,
		Sort:            req.Sort,
		Remark:          req.Remark,
	}
}
