// Package controller 提供支付模块的 C 端接口，只处理会员本人的数据。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/pay/client/request"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/repository"
	"eqray-flavor/app/modules/pay/service"
)

// payUserTypeMember 是 C 端固定使用的用户类型（会员）。
const payUserTypeMember int8 = 1

// PayWalletController 提供会员钱包余额、流水、充值档位与发起充值的接口。
type PayWalletController struct {
	service *service.PayWalletService
}

// NewPayWalletController 构造 C 端钱包控制器。
func NewPayWalletController() *PayWalletController {
	return &PayWalletController{service: service.NewPayWalletService()}
}

// Wallet 查询我的钱包，首次访问自动创建零余额钱包。
// 入参：会员令牌；返回余额、冻结金额与累计收支。
func (c *PayWalletController) Wallet(ctx http.Context) http.Response {
	wallet, err := c.service.EnsureWallet(ctx.Context(), support.MemberID(ctx.Context()), payUserTypeMember)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, wallet)
}

// Transactions 分页查询我的钱包流水。
// 入参：bizType、type、pageNo、pageSize；返回流水列表与总数。
func (c *PayWalletController) Transactions(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	wallet, err := c.service.EnsureWallet(ctx.Context(), support.MemberID(ctx.Context()), payUserTypeMember)
	if err != nil {
		return result.Fail(ctx, err)
	}

	filter := repository.PayWalletTransactionFilter{WalletID: wallet.ID}
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

// Packages 返回可选的充值档位。
// 无入参；返回开启状态的充值套餐，按实付金额升序。
func (c *PayWalletController) Packages(ctx http.Context) http.Response {
	list, err := c.service.EnabledPackages(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// CreateRecharge 发起钱包充值：创建充值记录并返回支付调起参数。
// 入参：应用、渠道、套餐与渠道需要的用户标识；返回充值记录编号与调起参数。
func (c *PayWalletController) CreateRecharge(ctx http.Context) http.Response {
	var req request.PayWalletRechargeRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	memberID := support.MemberID(ctx.Context())

	row, err := c.service.CreateRecharge(ctx.Context(), service.PayWalletRechargeCreateInput{
		UserID:      memberID,
		UserType:    payUserTypeMember,
		PackageID:   req.PackageID,
		AppID:       req.AppID,
		ChannelCode: req.ChannelCode,
		OpenID:      req.OpenID,
		AuthCode:    req.AuthCode,
		UserIP:      req.UserIP,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}

// PayByBalance 用钱包余额支付业务订单：余额足够时立即扣款并返回已支付支付单。
// 入参：应用、商户订单号、商品标题与金额（分）；返回支付单信息与支付后的钱包余额。
// 说明：同一应用的商户订单号只能支付一次，重复提交会返回重复下单错误。
func (c *PayWalletController) PayByBalance(ctx http.Context) http.Response {
	var req request.PayWalletPayRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	row, err := c.service.PayByBalance(ctx.Context(), service.PayWalletPayInput{
		AppID:           req.AppID,
		UserID:          support.MemberID(ctx.Context()),
		UserType:        payUserTypeMember,
		MerchantOrderID: req.MerchantOrderID,
		Subject:         req.Subject,
		Body:            req.Body,
		Price:           req.Price,
		UserIP:          req.UserIP,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, row)
}
