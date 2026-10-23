package controller

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/dish/admin/request"
	"eqray-flavor/app/modules/dish/service"
)

// SpuController 提供菜品的增删改查与上下架接口。
type SpuController struct {
	service *service.SpuService
}

// NewSpuController 构造菜品控制器。
func NewSpuController() *SpuController {
	return &SpuController{service: service.NewSpuService()}
}

// Index 分页查询菜品，支持名称/编码关键字、分类（含下级）与上下架状态筛选。
// 入参：keyword、categoryId、status、pageNo、pageSize（查询串）；返回菜品列表与总数。
func (c *SpuController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	categoryID := parseID(ctx.Request().Query("categoryId"))
	list, total, err := c.service.Page(ctx.Context(), service.SpuListQuery{
		Keyword:    ctx.Request().Query("keyword"),
		CategoryID: categoryID,
		Status:     parseStatusFilter(ctx.Request().Query("status")),
		PageNo:     page.PageNo,
		PageSize:   page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 返回菜品详情，包含规格列表与已挂载的加料组。
// 入参：路径 id；返回菜品、规格与加料组 ID 列表。
func (c *SpuController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	detail, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, detail)
}

// Store 新增菜品；规格与加料组随菜品一起提交，至少需要一条规格。
// 入参：菜品字段 + skus + addonGroupIds；返回新菜品 ID。
func (c *SpuController) Store(ctx http.Context) http.Response {
	var req request.SpuSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toSpuSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑菜品；规格整表覆盖，未提交的存量规格会被删除（至少保留一条）。
// 入参：路径 id + 菜品字段 + skus + addonGroupIds；返回 null。
func (c *SpuController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.SpuSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toSpuSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// UpdateStatus 切换菜品上下架状态，供列表开关直接调用。
// 入参：路径 id + status；返回 null。
func (c *SpuController) UpdateStatus(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.SpuStatusRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.UpdateStatus(ctx.Context(), id, baseStatus(req.Status)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除菜品；规格、门店可售与价格配置、加料组关联一并清理。
// 入参：路径 id；返回 null。
func (c *SpuController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toSpuSaveInput 把请求体转换为服务层入参，规格逐条转换。
func toSpuSaveInput(req request.SpuSaveRequest) service.SpuSaveInput {
	skus := make([]service.SpuSkuInput, 0, len(req.Skus))
	for _, sku := range req.Skus {
		skus = append(skus, service.SpuSkuInput{
			ID:            sku.ID,
			Name:          sku.Name,
			Code:          sku.Code,
			StandardPrice: sku.StandardPrice,
			MemberPrice:   sku.MemberPrice,
			CostPrice:     sku.CostPrice,
			Status:        baseStatus(sku.Status),
			Sort:          sku.Sort,
			Remark:        sku.Remark,
		})
	}

	return service.SpuSaveInput{
		CategoryID:    req.CategoryID,
		Name:          req.Name,
		Code:          req.Code,
		Unit:          req.Unit,
		Description:   req.Description,
		Image:         req.Image,
		Tags:          req.Tags,
		Allergens:     req.Allergens,
		Status:        baseStatus(req.Status),
		Sort:          req.Sort,
		Remark:        req.Remark,
		Skus:          skus,
		AddonGroupIDs: req.AddonGroupIDs,
	}
}

// parseStatusFilter 解析列表状态筛选：空字符串表示不限，其余按 0/1 解析。
func parseStatusFilter(raw string) *int8 {
	if raw == "" {
		return nil
	}

	value, err := strconv.ParseInt(raw, 10, 8)
	if err != nil {
		return nil
	}
	status := int8(value)

	return &status
}
