package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/member/service"
)

// MemberSignController 提供 C 端每日签到与签到状态查询。
type MemberSignController struct {
	service *service.MemberSignService
}

// NewMemberSignController 构造 C 端会员签到控制器。
func NewMemberSignController() *MemberSignController {
	return &MemberSignController{service: service.NewMemberSignService()}
}

// Sign 执行每日签到：写签到记录、加积分与经验、写积分流水，必要时升等级。
// 无请求体，会员身份取自令牌；同一会员同一天只能签一次，重复签到返回业务错误码；返回本次签到结果。
func (c *MemberSignController) Sign(ctx http.Context) http.Response {
	res, err := c.service.Sign(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// Status 查询今日签到状态：是否已签、连续与累计天数、本月已签日期、今日与次日可得奖励。
// 无入参，会员身份取自令牌；返回签到状态。
func (c *MemberSignController) Status(ctx http.Context) http.Response {
	res, err := c.service.Status(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}
