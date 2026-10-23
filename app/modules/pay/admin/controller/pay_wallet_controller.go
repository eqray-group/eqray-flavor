package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/pay/admin/request"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/repository"
	"eqray-flavor/app/modules/pay/service"
)

// PayWalletController 提供会员钱包的查询、流水与余额调整接口。
type PayWalletController struct {
	service *service.PayWalletService
}

// NewPayWalletController 构造钱包控制器。
func NewPayWalletController() *PayWalletController {
	return &PayWalletController{service: service.NewPayWalletService()}
}

// Index 分页查询钱包，支持按用户与用户类型筛选。
// 入参：userId、userType、pageNo、pageSize（查询串）；返回钱包列表与总数。
func (c *PayWalletController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	userID, err := queryID(ctx, "userId")
	if err != nil {
		return result.Fail(ctx, err)
	}

	filter := repository.PayWalletFilter{UserID: userID}
	if raw := ctx.Request().Query("userType"); raw != "" {
		userType := int8(ctx.Request().QueryInt("userType", 0))
		filter.UserType = &userType
	}

	list, total, err := c.service.Page(ctx.Context(), filter, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 查询钱包详情。
// 入参：路径 id；返回钱包，不存在返回业务错误码。
func (c *PayWalletController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	wallet, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, wallet)
}

// Transactions 分页查询钱包流水，支持按业务类型与收支方向筛选。
// 入参：路径 id + bizType、type、pageNo、pageSize；返回流水列表与总数。
func (c *PayWalletController) Transactions(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	filter := repository.PayWalletTransactionFilter{WalletID: id}
	if raw := ctx.Request().QueryInt("bizType", 0); raw > 0 {
		bizType := payEnums.PayWalletBizTypeEnum(int8(raw))
		filter.BizType = &bizType
	}
	if raw := ctx.Request().Query("type"); raw != "" {
		direction := int8(ctx.Request().QueryInt("type", 0))
		filter.Type = &direction
	}

	list, total, err := c.service.Transactions(ctx.Context(), filter, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// ChangeBalance 调整钱包余额：price 为正表示加款、为负表示扣款，均写一条流水。
// 入参：路径 id + price、remark；返回 null，余额不足时返回业务错误码。
func (c *PayWalletController) ChangeBalance(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PayWalletChangeRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	in := service.PayWalletChangeInput{WalletID: id, Price: req.Price, Remark: req.Remark}
	if req.Price >= 0 {
		err = c.service.Income(ctx.Context(), in)
	} else {
		in.Price = -req.Price
		err = c.service.Expense(ctx.Context(), in)
	}
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
