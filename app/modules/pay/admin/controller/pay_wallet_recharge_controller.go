package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/pay/repository"
	"eqray-flavor/app/modules/pay/service"
)

// PayWalletRechargeController 提供充值记录的查询接口。
type PayWalletRechargeController struct {
	service *service.PayWalletService
}

// NewPayWalletRechargeController 构造充值记录控制器。
func NewPayWalletRechargeController() *PayWalletRechargeController {
	return &PayWalletRechargeController{service: service.NewPayWalletService()}
}

// Index 分页查询充值记录，支持按钱包与充值状态筛选。
// 入参：walletId、status、pageNo、pageSize（查询串）；返回充值记录与总数。
func (c *PayWalletRechargeController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	walletID, err := queryID(ctx, "walletId")
	if err != nil {
		return result.Fail(ctx, err)
	}

	filter := repository.PayWalletRechargeFilter{WalletID: walletID}
	// 状态 0（未支付）是合法值，只有参数缺省才表示不限
	if raw := ctx.Request().Query("status"); raw != "" {
		status := int8(ctx.Request().QueryInt("status", 0))
		filter.Status = &status
	}

	list, total, err := c.service.Recharges(ctx.Context(), filter, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 查询充值记录详情。
// 入参：路径 id；返回充值记录，不存在返回业务错误码。
func (c *PayWalletRechargeController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	row, err := c.service.RechargeDetail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}
