// Package controller 提供 promotion 模块的后台管理接口。
package controller

import (
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/errs"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/param"
	"eqray-flavor/app/modules/base/result"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/promotion/admin/request"
	"eqray-flavor/app/modules/promotion/enums"
	"eqray-flavor/app/modules/promotion/service"
)

// dateTimeLayout 是时间入参格式，与接口返回的 DateTime 保持一致。
const dateTimeLayout = "2006-01-02 15:04:05"

// PromotionController 提供券模板、会员券、营销活动与积分规则的管理接口。
type PromotionController struct {
	service *service.PromotionService
}

// NewPromotionController 构造营销控制器。
func NewPromotionController() *PromotionController {
	return &PromotionController{service: service.NewPromotionService()}
}

// ---------- 优惠券模板 ----------

// Index 分页查询优惠券模板。
// 入参：keyword、type、status、pageNo、pageSize（查询串）；返回模板列表与总数。
func (c *PromotionController) Index(ctx http.Context) http.Response {
	var req request.CouponTemplatePageRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 15
	}

	query := service.CouponTemplateListQuery{
		Keyword: req.Keyword,
		Type:    enums.CouponType(req.Type),
		PageNo:  pageNo,
		PageSize: pageSize,
	}
	// 状态 0 是合法值，只有参数缺省才表示不限
	if ctx.Request().Query("status") != "" {
		status := req.Status
		query.Status = &status
	}

	rows, total, err := c.service.PageCouponTemplates(ctx.Context(), query)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, rows, total)
}

// Show 查询优惠券模板详情。
// 入参：路径 id；返回模板详情，不存在返回业务错误码。
func (c *PromotionController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	detail, err := c.service.DetailCouponTemplate(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, detail)
}

// Store 新增或修改优惠券模板。
// 入参：路径 id（缺省为新增）与模板内容；返回模板 ID。
func (c *PromotionController) Store(ctx http.Context) http.Response {
	var req request.CouponTemplateSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	startTime, err := parseDateTime(req.StartTime)
	if err != nil {
		return result.Fail(ctx, err)
	}
	endTime, err := parseDateTime(req.EndTime)
	if err != nil {
		return result.Fail(ctx, err)
	}

	templateID, err := c.service.SaveCouponTemplate(ctx.Context(), req.ID, service.CouponTemplateSaveInput{
		Name: req.Name, Type: enums.CouponType(req.Type), Description: req.Description,
		ThresholdAmount: req.ThresholdAmount, DiscountAmount: req.DiscountAmount,
		DiscountRate: req.DiscountRate, MaxDiscountAmount: req.MaxDiscountAmount,
		Scope: enums.CouponScope(req.Scope), ScopeIDs: req.ScopeIDs,
		ChannelLimit: enums.CouponChannelLimit(req.ChannelLimit), IssueType: enums.CouponIssueType(req.IssueType),
		TotalCount: req.TotalCount, PerMemberLimit: req.PerMemberLimit,
		StartTime: startTime, EndTime: endTime, ValidDays: req.ValidDays,
		Status: baseStatus(req.Status),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, templateID)
}

// Destroy 删除优惠券模板。
// 入参：路径 id；返回空数据表示删除成功。
func (c *PromotionController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteCouponTemplate(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Issue 向指定会员发放优惠券。
// 入参：路径 id（券模板）+ memberId 与 count；返回实际发放张数。
func (c *PromotionController) Issue(ctx http.Context) http.Response {
	var req request.CouponTemplateIssueRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	count, err := c.service.Issue(ctx.Context(), service.IssueInput{
		TemplateID: id, MemberID: req.MemberID, Count: req.Count,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, count)
}

// ---------- 会员券 ----------

// CouponIndex 分页查询会员券。
// 入参：templateId、memberId、status、pageNo、pageSize（查询串）；返回券列表与总数。
func (c *PromotionController) CouponIndex(ctx http.Context) http.Response {
	var req request.CouponPageRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 15
	}

	query := service.CouponListQuery{
		TemplateID: req.TemplateID, MemberID: req.MemberID,
		PageNo: pageNo, PageSize: pageSize,
	}
	// 状态 0 不是合法值，缺省才表示不限
	if ctx.Request().Query("status") != "" {
		status := enums.CouponStatus(req.Status)
		query.Status = &status
	}

	rows, total, err := c.service.PageCoupons(ctx.Context(), query)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, rows, total)
}

// Revoke 作废会员券：未使用 → 已过期。
// 入参：路径 id；返回空数据表示作废成功。
func (c *PromotionController) Revoke(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.RevokeCoupon(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// ---------- 营销活动 ----------

// ActivityIndex 分页查询营销活动。
// 入参：keyword、type、status、pageNo、pageSize（查询串）；返回活动列表与总数。
func (c *PromotionController) ActivityIndex(ctx http.Context) http.Response {
	var req request.ActivityPageRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 15
	}

	query := service.ActivityListQuery{
		Keyword: req.Keyword,
		Type:    enums.ActivityType(req.Type),
		PageNo:  pageNo,
		PageSize: pageSize,
	}
	// 状态 0（未启用）是合法值，只有参数缺省才表示不限
	if ctx.Request().Query("status") != "" {
		status := enums.ActivityStatus(req.Status)
		query.Status = &status
	}

	rows, total, err := c.service.PageActivities(ctx.Context(), query)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, rows, total)
}

// ActivityShow 查询营销活动详情，含参与商品。
// 入参：路径 id；返回活动详情，不存在返回业务错误码。
func (c *PromotionController) ActivityShow(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	detail, err := c.service.DetailActivity(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, detail)
}

// ActivityStore 新增或修改营销活动。
// 入参：路径 id（缺省为新增）与活动内容；返回活动 ID。
func (c *PromotionController) ActivityStore(ctx http.Context) http.Response {
	var req request.ActivitySaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	startTime, err := parseDateTime(req.StartTime)
	if err != nil {
		return result.Fail(ctx, err)
	}
	endTime, err := parseDateTime(req.EndTime)
	if err != nil {
		return result.Fail(ctx, err)
	}

	activityID, err := c.service.SaveActivity(ctx.Context(), req.ID, service.ActivitySaveInput{
		Name: req.Name, Type: enums.ActivityType(req.Type), Level: enums.ActivityLevel(req.Level),
		Description: req.Description,
		Rule:        buildActivityRule(req),
		Scope:       enums.ActivityScope(req.Scope), ScopeIDs: req.ScopeIDs,
		Priority: req.Priority, MutexGroup: req.MutexGroup,
		Exclusive: req.Exclusive, Stackable: req.Stackable,
		QuotaLimit: req.QuotaLimit, QuotaCycle: enums.ActivityQuotaCycle(req.QuotaCycle),
		StartTime: startTime, EndTime: endTime, Status: enums.ActivityStatus(req.Status),
		ProductSPUIDs: req.ProductSPUIDs,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, activityID)
}

// ActivityDestroy 删除营销活动。
// 入参：路径 id；返回空数据表示删除成功。
func (c *PromotionController) ActivityDestroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteActivity(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// ---------- 积分规则 ----------

// PointRule 查询积分抵扣规则。
// 入参：无；返回当前规则。
func (c *PromotionController) PointRule(ctx http.Context) http.Response {
	rule, err := c.service.PointRule(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, rule)
}

// PointRuleStore 保存积分抵扣规则。
// 入参：兑换比例、单笔上限、最低订单金额与启用状态；返回空数据表示保存成功。
func (c *PromotionController) PointRuleStore(ctx http.Context) http.Response {
	var req request.PointRuleSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.UpdatePointRule(ctx.Context(), service.PointRuleOutput{
		ExchangeRate:     req.ExchangeRate,
		MaxDeductAmount: req.MaxDeductAmount,
		MinOrderAmount:  req.MinOrderAmount,
		Status:           req.Status,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// buildActivityRule 把请求里的规则字段装配成活动规则结构。
func buildActivityRule(req request.ActivitySaveRequest) service.ActivityRule {
	rule := service.ActivityRule{
		ThresholdAmount:   req.ThresholdAmount,
		DiscountAmount:    req.DiscountAmount,
		DiscountRate:      req.DiscountRate,
		MaxDiscountAmount: req.MaxDiscountAmount,
		SecondHalfRate:    req.SecondHalfRate,
		ComboPrice:        req.ComboPrice,
		AddonSpuID:        req.AddonSpuID,
		AddonPrice:        req.AddonPrice,
	}

	for _, tier := range req.Tiers {
		rule.Tiers = append(rule.Tiers, service.ActivityTier{
			ThresholdAmount: tier.ThresholdAmount,
			DiscountAmount:  tier.DiscountAmount,
		})
	}
	for _, combo := range req.ComboItems {
		rule.ComboItems = append(rule.ComboItems, service.ComboItem{
			SpuID:    combo.SpuID,
			Quantity: combo.Quantity,
		})
	}

	return rule
}

// parseDateTime 把 "yyyy-MM-dd HH:mm:ss" 解析成统一时间类型；空串返回零值表示不限。
func parseDateTime(raw string) (types.DateTime, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return types.DateTime{}, nil
	}

	parsed, err := time.ParseInLocation(dateTimeLayout, value, time.Local)
	if err != nil {
		return types.DateTime{}, errs.WithMsg(errs.BadRequest, "时间格式应为 yyyy-MM-dd HH:mm:ss")
	}

	return types.DateTime(parsed), nil
}

// baseStatus 把请求里的状态转成通用启用状态枚举。
func baseStatus(status int8) baseEnums.CommonStatusEnum {
	return baseEnums.CommonStatusEnum(status)
}
