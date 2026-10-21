package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/member/client/request"
	"eqray-flavor/app/modules/member/service"
)

// MemberAddressController 提供 C 端会员自助维护收货地址的接口。
// 所有操作的会员身份都取自令牌，地址归属由 Service 用 user_id 过滤。
type MemberAddressController struct {
	service *service.MemberAddressService
}

// NewMemberAddressController 构造 C 端会员地址控制器。
func NewMemberAddressController() *MemberAddressController {
	return &MemberAddressController{service: service.NewMemberAddressService()}
}

// Index 查询当前会员的全部收货地址，默认地址排在最前。
// 无入参，会员身份取自令牌；返回地址列表。
func (c *MemberAddressController) Index(ctx http.Context) http.Response {
	list, err := c.service.List(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Store 新增一条收货地址，会员的第一条地址自动置为默认。
// 入参见 MemberAddressRequest（收货人、手机号、地区、详细地址、是否默认）；返回新地址 ID。
func (c *MemberAddressController) Store(ctx http.Context) http.Response {
	var req request.MemberAddressRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), support.MemberID(ctx.Context()), service.MemberAddressInput{
		Name:          req.Name,
		Mobile:        req.Mobile,
		AreaID:        req.AreaID,
		DetailAddress: req.DetailAddress,
		DefaultStatus: req.DefaultStatus,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 修改指定地址的内容；默认标记只能通过 SetDefault 修改。
// 入参：路径 id + 地址内容；返回 null。
func (c *MemberAddressController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberAddressRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), support.MemberID(ctx.Context()), id, service.MemberAddressInput{
		Name:          req.Name,
		Mobile:        req.Mobile,
		AreaID:        req.AreaID,
		DetailAddress: req.DetailAddress,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除指定地址（软删），仅能删当前会员自己的地址。
// 入参：路径 id；返回 null。
func (c *MemberAddressController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), support.MemberID(ctx.Context()), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// SetDefault 把指定地址设为默认，事务内先清掉同会员的其它默认标记。
// 入参：路径 id；返回 null。
func (c *MemberAddressController) SetDefault(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.SetDefault(ctx.Context(), support.MemberID(ctx.Context()), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
