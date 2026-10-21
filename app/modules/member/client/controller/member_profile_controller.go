package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/member/client/request"
	"eqray-flavor/app/modules/member/service"
)

// MemberProfileController 提供 C 端会员的「我的资料」查询与修改。
type MemberProfileController struct {
	service *service.MemberProfileService
}

// NewMemberProfileController 构造 C 端会员资料控制器。
func NewMemberProfileController() *MemberProfileController {
	return &MemberProfileController{service: service.NewMemberProfileService()}
}

// Show 查看「我的资料」：基础档案 + 等级名、下一等级所需经验、标签名与分组名。
// 无入参，会员身份取自令牌；返回资料详情。
func (c *MemberProfileController) Show(ctx http.Context) http.Response {
	profile, err := c.service.Profile(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, profile)
}

// Update 修改「我的资料」：昵称、头像、姓名、性别、生日、地区。
// 入参见 MemberProfileUpdateRequest，生日格式 2006-01-02；手机号、积分、等级不可自助修改；返回 null。
func (c *MemberProfileController) Update(ctx http.Context) http.Response {
	var req request.MemberProfileUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	birthday, err := req.BirthdayTime()
	if err != nil {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "生日格式应为 2006-01-02"))
	}

	if err := c.service.Update(ctx.Context(), support.MemberID(ctx.Context()), service.MemberProfileUpdateInput{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Name:     req.Name,
		Sex:      req.Sex,
		Birthday: birthday,
		AreaID:   req.AreaID,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
