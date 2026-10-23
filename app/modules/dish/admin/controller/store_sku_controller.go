package controller

import (
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/dish/admin/request"
	"eqray-flavor/app/modules/dish/service"
)

// dateTimeLayout 是门店价生效区间的入参格式，与接口返回的 DateTime 保持一致。
const dateTimeLayout = "2006-01-02 15:04:05"

// StoreSkuController 提供门店维度的菜品可售与门店价维护接口。
type StoreSkuController struct {
	service *service.StoreSkuService
}

// NewStoreSkuController 构造门店菜品控制器。
func NewStoreSkuController() *StoreSkuController {
	return &StoreSkuController{service: service.NewStoreSkuService()}
}

// Index 分页查询当前门店的规格配置，未配置的规格也会列出并标记 configured=false。
// 入参：keyword、categoryId、pageNo、pageSize（查询串）+ 请求头门店；返回规格配置列表与总数。
func (c *StoreSkuController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), service.StoreSkuListQuery{
		Keyword:    ctx.Request().Query("keyword"),
		CategoryID: parseID(ctx.Request().Query("categoryId")),
		PageNo:     page.PageNo,
		PageSize:   page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Save 保存单规格的门店配置与门店价覆盖；门店来自请求头，不接受请求体指定。
// 入参：路径 skuId + 可售开关、排序、门店价、区间；返回 null。
func (c *StoreSkuController) Save(ctx http.Context) http.Response {
	skuID, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.StoreSkuSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	startAt, err := parseDateTime(req.StartAt)
	if err != nil {
		return result.Fail(ctx, err)
	}
	endAt, err := parseDateTime(req.EndAt)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Save(ctx.Context(), skuID, service.StoreSkuSaveInput{
		Available:        req.Available,
		SoldOut:          req.SoldOut,
		OnlineOnly:       req.OnlineOnly,
		Sort:             req.Sort,
		Remark:           req.Remark,
		StorePrice:       req.StorePrice,
		StoreMemberPrice: req.StoreMemberPrice,
		StartAt:          startAt,
		EndAt:            endAt,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Batch 批量调整门店可售开关，用于列表批量上架 / 下架 / 沽清。
// 入参：skuIds + 需要变更的开关；返回 null。
func (c *StoreSkuController) Batch(ctx http.Context) http.Response {
	var req request.StoreSkuBatchRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.BatchUpdate(ctx.Context(), service.StoreSkuBatchInput{
		SkuIDs:     req.SkuIDs,
		Available:  req.Available,
		SoldOut:    req.SoldOut,
		OnlineOnly: req.OnlineOnly,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// parseDateTime 把 "yyyy-MM-dd HH:mm:ss" 解析成统一时间类型；空串返回 nil 表示不限。
func parseDateTime(raw string) (*types.DateTime, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.ParseInLocation(dateTimeLayout, value, time.Local)
	if err != nil {
		return nil, badRequest("生效时间格式应为 yyyy-MM-dd HH:mm:ss")
	}

	result := types.DateTime(parsed)

	return &result, nil
}
