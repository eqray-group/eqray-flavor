package controller

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/member/admin/request"
	"eqray-flavor/app/modules/member/service"
)

// MemberController 提供会员档案的增删改查与积分 / 等级调整接口。
type MemberController struct {
	service *service.MemberService
	socials *service.MemberSocialService
}

// NewMemberController 构造会员档案控制器，并装配会员与社交服务。
func NewMemberController() *MemberController {
	return &MemberController{
		service: service.NewMemberService(),
		socials: service.NewMemberSocialService(),
	}
}

// Index 分页查询会员档案，支持昵称、手机号、等级、分组与状态筛选。
// 入参：nickname、mobile、levelId、groupId、status、pageNo、pageSize（查询串）；返回会员列表与总数。
func (c *MemberController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.MemberListQuery{
		Nickname: ctx.Request().Query("nickname"),
		Mobile:   ctx.Request().Query("mobile"),
		LevelID:  queryID(ctx, "levelId"),
		GroupID:  queryID(ctx, "groupId"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
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

// Show 查询会员详情，含等级、标签、分组等展示字段。
// 入参：路径 id；返回会员详情，不存在返回业务错误码。
func (c *MemberController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// Store 后台新增会员档案；手机号唯一，重复时返回业务错误码。
// 入参见 MemberSaveRequest（昵称、手机号、等级、积分、标签、分组等）；返回新会员 ID。
func (c *MemberController) Store(ctx http.Context) http.Response {
	var req request.MemberSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), toMemberSaveInput(req))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑会员档案；积分与等级建议走专用的调整接口。
// 入参：路径 id + 会员字段；返回 null。
func (c *MemberController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, toMemberSaveInput(req)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// UpdateStatus 启用或禁用会员，禁用后 C 端接口会拒绝该会员。
// 入参：路径 id + status（0 禁用 / 1 启用）；返回 null。
func (c *MemberController) UpdateStatus(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberStatusRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.UpdateStatus(ctx.Context(), id, req.Status); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// AdjustPoint 后台调整会员积分，正数增加、负数扣减，并写一条积分流水。
// 入参：路径 id + point、title、description；返回 null。
func (c *MemberController) AdjustPoint(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberPointRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.AdjustPoint(ctx.Context(), id, req.Point, req.Title, req.Description); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// AdjustLevel 后台调整会员等级与经验，并写一条等级变更记录。
// 入参：路径 id + levelId、experience、title、description；返回 null。
func (c *MemberController) AdjustLevel(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberLevelAdjustRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.AdjustLevel(ctx.Context(), id, req.LevelID, req.Experience, req.Title, req.Description); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// PointRecords 分页查询指定会员的积分流水。
// 入参：路径 id + pageNo、pageSize；返回流水列表与总数。
func (c *MemberController) PointRecords(ctx http.Context) http.Response {
	id, page, err := c.pathIDAndPage(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.PointRecords(ctx.Context(), id, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// LevelRecords 分页查询指定会员的等级变更记录。
// 入参：路径 id + pageNo、pageSize；返回等级记录列表与总数。
func (c *MemberController) LevelRecords(ctx http.Context) http.Response {
	id, page, err := c.pathIDAndPage(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.LevelRecords(ctx.Context(), id, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// SignRecords 分页查询指定会员的签到记录。
// 入参：路径 id + pageNo、pageSize；返回签到记录列表与总数。
func (c *MemberController) SignRecords(ctx http.Context) http.Response {
	id, page, err := c.pathIDAndPage(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.SignRecords(ctx.Context(), id, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Addresses 查询指定会员的收货地址列表（只读），默认地址排最前。
// 入参：路径 id；返回地址列表。
func (c *MemberController) Addresses(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, err := c.service.Addresses(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// SocialBindings 管理端只读查看会员绑定的第三方账号，不支持在此解绑。
// 入参：路径 id；返回绑定记录列表。
func (c *MemberController) SocialBindings(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, err := c.socials.Bindings(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Destroy 删除会员档案（软删），已产生的积分 / 等级 / 签到记录保留。
// 入参：路径 id；返回 null。
func (c *MemberController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// toMemberSaveInput 把请求体转换为服务层入参。
func toMemberSaveInput(req request.MemberSaveRequest) service.MemberSaveInput {
	in := service.MemberSaveInput{
		Nickname:   req.Nickname,
		Mobile:     req.Mobile,
		Avatar:     req.Avatar,
		Name:       req.Name,
		Sex:        req.Sex,
		AreaID:     req.AreaID,
		LevelID:    req.LevelID,
		Point:      req.Point,
		Experience: req.Experience,
		TagIDs:     req.TagIDs,
		Mark:       req.Mark,
		GroupID:    req.GroupID,
		Status:     req.Status,
	}
	if req.Birthday != nil && !req.Birthday.IsZero() {
		birthday := *req.Birthday
		in.Birthday = &birthday
	}

	return in
}

// pathIDAndPage 统一解析路径 ID 与分页参数。
func (c *MemberController) pathIDAndPage(ctx http.Context) (types.ID, param.Page, error) {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return 0, page, err
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return 0, page, err
	}

	return id, page, nil
}

// queryID 解析雪花 ID 查询参数，非法值按「不限制」处理。
func queryID(ctx http.Context, key string) types.ID {
	raw := ctx.Request().Query(key)
	if raw == "" {
		return 0
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0
	}

	return types.ID(id)
}
