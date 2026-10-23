package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/pay/admin/request"
	"eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/model"
	"eqray-flavor/app/modules/pay/service"
)

// PayWalletPackageController 提供充值套餐的增删改查与下拉接口。
type PayWalletPackageController struct {
	service *service.PayWalletService
}

// NewPayWalletPackageController 构造充值套餐控制器。
func NewPayWalletPackageController() *PayWalletPackageController {
	return &PayWalletPackageController{service: service.NewPayWalletService()}
}

// Index 分页查询充值套餐（含已关闭）。
// 入参：pageNo、pageSize（查询串）；返回套餐列表与总数。
func (c *PayWalletPackageController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Packages(ctx.Context(), page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Options 返回开启状态的充值套餐，供充值页面的档位下拉使用。
// 无入参；返回按实付金额升序的套餐列表。
func (c *PayWalletPackageController) Options(ctx http.Context) http.Response {
	list, err := c.service.EnabledPackages(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Store 新增充值套餐。
// 入参：套餐名、实付金额、赠送金额、状态与描述；返回新套餐主键。
func (c *PayWalletPackageController) Store(ctx http.Context) http.Response {
	var req request.PayWalletPackageSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.CreatePackage(ctx.Context(), toPayWalletPackageModel(0, req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑充值套餐，已产生的充值记录不受影响。
// 入参：路径 id + 套餐字段；返回 null。
func (c *PayWalletPackageController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PayWalletPackageSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.UpdatePackage(ctx.Context(), toPayWalletPackageModel(id, req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除充值套餐（软删除），历史充值记录保留。
// 入参：路径 id；返回 null。
func (c *PayWalletPackageController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeletePackage(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toPayWalletPackageModel 把请求体转成套餐模型，避免 Controller 直接拼装字段。
// 入参：套餐主键（新增传 0）与请求体；返回服务层可直接写入的模型。
func toPayWalletPackageModel(id types.ID, req request.PayWalletPackageSaveRequest) model.PayWalletRechargePackageModel {
	pkg := model.PayWalletRechargePackageModel{
		Name:        req.Name,
		PayPrice:    req.PayPrice,
		BonusPrice:  req.BonusPrice,
		Status:      enums.PayEnableStatusEnum(req.Status),
		Description: req.Description,
	}
	pkg.ID = id

	return pkg
}
