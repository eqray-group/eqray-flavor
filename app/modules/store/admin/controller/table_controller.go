package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/admin/request"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/service"
)

// TableController 提供桌台的增删改查、批量生成与开台 / 清台接口。
type TableController struct {
	service *service.StoreTableService
}

// NewTableController 构造桌台控制器。
func NewTableController() *TableController {
	return &TableController{service: service.NewStoreTableService()}
}

// Index 分页查询桌台，支持桌号 / 桌码、区域与状态筛选。
// 门店范围由请求头门店上下文或当前用户可见门店集合决定。
func (c *TableController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.TableListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		AreaID:   types.ID(ctx.Request().QueryInt("areaId", 0)),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", 0); status > 0 {
		v := enums.TableStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Store 新增桌台；未传桌码时按「门店编码-桌号」自动生成。
// 入参：storeId、areaId、name、code、seats、sort、remark；返回新桌台 ID。
func (c *TableController) Store(ctx http.Context) http.Response {
	var req request.TableSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toTableSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑桌台；桌码创建后不可修改。
// 入参：路径 id + 桌台字段；返回 null。
func (c *TableController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.TableSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toTableSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除桌台；使用中的桌台必须清台后才能删除。
// 入参：路径 id；返回 null。
func (c *TableController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Batch 批量生成桌台；桌号按前缀补零，桌码按「门店编码-桌号」生成。
// 入参：storeId、areaId、prefix、start、count、seats；返回生成数量。
func (c *TableController) Batch(ctx http.Context) http.Response {
	var req request.TableBatchRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	count, err := c.service.BatchCreate(ctx.Context(), toTableBatchInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"count": count})
}

// Open 开台：仅空闲桌台可开，状态改为使用中。
// 入参：路径 id + currentOrderId（可为 0）；返回 null。
func (c *TableController) Open(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.TableOpenRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Open(ctx.Context(), id, req.CurrentOrderID); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Clear 清台：使用中或待清台的桌台恢复空闲。
// 入参：路径 id；返回 null。
func (c *TableController) Clear(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Clear(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Transfer 换台：把进行中的台位单换到另一张空闲桌台。
// 入参：路径 id（源台）+ toTableId；返回 null。
func (c *TableController) Transfer(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.TableTransferRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Transfer(ctx.Context(), id, req.ToTableID); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toTableSaveInput 把请求体转换为服务层入参。
func toTableSaveInput(req request.TableSaveRequest) service.TableSaveInput {
	return service.TableSaveInput{
		StoreID: req.StoreID,
		AreaID:  req.AreaID,
		Name:    req.Name,
		Code:    req.Code,
		Seats:   req.Seats,
		Sort:    req.Sort,
		Remark:  req.Remark,
	}
}

// toTableBatchInput 把批量生成请求转换为服务层入参。
func toTableBatchInput(req request.TableBatchRequest) service.TableBatchInput {
	return service.TableBatchInput{
		StoreID: req.StoreID,
		AreaID:  req.AreaID,
		Prefix:  req.Prefix,
		Start:   req.Start,
		Count:   req.Count,
		Seats:   req.Seats,
	}
}
