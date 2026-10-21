package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/member/admin/request"
	"eqray-flavor/app/modules/member/service"
)

// MemberSignConfigController 提供后台签到奖励档位的维护接口。
type MemberSignConfigController struct {
	service *service.MemberSignConfigService
}

// NewMemberSignConfigController 构造签到奖励档位控制器。
func NewMemberSignConfigController() *MemberSignConfigController {
	return &MemberSignConfigController{service: service.NewMemberSignConfigService()}
}

// Index 分页查询签到奖励档位，按连续签到天数升序。
// 入参：day（签到天数）、status（状态）、pageNo、pageSize（查询串，均可选）；返回档位列表与总数。
func (c *MemberSignConfigController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.MemberSignConfigListQuery{
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if day := ctx.Request().QueryInt("day", 0); day > 0 {
		q.Day = &day
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := baseEnums.CommonStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Store 新增签到奖励档位；同一天只能有一条档位，重复时返回业务错误码。
// 入参：day（1~31）、point、experience、status；返回 null。
func (c *MemberSignConfigController) Store(ctx http.Context) http.Response {
	var req request.MemberSignConfigRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Create(ctx.Context(), service.MemberSignConfigInput{
		Day:        req.Day,
		Point:      req.Point,
		Experience: req.Experience,
		Status:     req.Status,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Update 编辑签到奖励档位，同样校验天数不与其他档位重复。
// 入参：路径 id + day、point、experience、status；返回 null。
func (c *MemberSignConfigController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberSignConfigRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, service.MemberSignConfigInput{
		Day:        req.Day,
		Point:      req.Point,
		Experience: req.Experience,
		Status:     req.Status,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除签到奖励档位；至少保留一条启用中的档位，否则返回业务错误码。
// 入参：路径 id；返回 null。
func (c *MemberSignConfigController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
