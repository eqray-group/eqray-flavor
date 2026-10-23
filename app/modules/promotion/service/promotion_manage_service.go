// promotion_manage_service.go 实现 promotion 模块的后台管理能力：券模板、会员券与活动的增删改查。
// 管理端规则来源与计算规则分离：后台只维护规则，生效判定一律走 rule.go 的纯函数。
package service

import (
	"context"
	"encoding/json"
	"strings"

	"eqray-flavor/app/modules/base/errs"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/promotion/enums"
	promotionErrs "eqray-flavor/app/modules/promotion/errs"
	promotionModel "eqray-flavor/app/modules/promotion/model"
	promotionRepository "eqray-flavor/app/modules/promotion/repository"
)

// CouponTemplateSaveInput 是券模板的保存入参。
type CouponTemplateSaveInput struct {
	// Name 券名称
	Name string
	// Type 券类型
	Type enums.CouponType
	// Description 券说明
	Description string
	// ThresholdAmount 使用门槛（分）
	ThresholdAmount int64
	// DiscountAmount 抵扣金额（分）
	DiscountAmount int64
	// DiscountRate 折扣率（万分比）
	DiscountRate int32
	// MaxDiscountAmount 折扣封顶（分）
	MaxDiscountAmount int64
	// Scope 适用范围
	Scope enums.CouponScope
	// ScopeIDs 适用对象 ID 列表
	ScopeIDs []types.ID
	// ChannelLimit 渠道限制
	ChannelLimit enums.CouponChannelLimit
	// IssueType 发放方式
	IssueType enums.CouponIssueType
	// TotalCount 发行总量，0 表示不限
	TotalCount int
	// PerMemberLimit 每人限领张数
	PerMemberLimit int
	// StartTime 生效开始时间
	StartTime types.DateTime
	// EndTime 生效结束时间
	EndTime types.DateTime
	// ValidDays 领取后有效天数
	ValidDays int
	// Status 启用状态
	Status baseEnums.CommonStatusEnum
}

// CouponTemplateRow 是券模板的输出行。
type CouponTemplateRow struct {
	ID               types.ID                  `json:"id"`
	Name             string                    `json:"name"`
	Type             int8                      `json:"type"`
	TypeText         string                    `json:"typeText"`
	Description      string                    `json:"description"`
	ThresholdAmount  int64                     `json:"thresholdAmount"`
	DiscountAmount   int64                     `json:"discountAmount"`
	DiscountRate     int32                     `json:"discountRate"`
	MaxDiscountAmount int64                    `json:"maxDiscountAmount"`
	Scope            int8                      `json:"scope"`
	ScopeText        string                    `json:"scopeText"`
	ScopeIDs         string                    `json:"scopeIds"`
	ChannelLimit     int8                      `json:"channelLimit"`
	ChannelLimitText string                    `json:"channelLimitText"`
	IssueType        int8                      `json:"issueType"`
	IssueTypeText    string                    `json:"issueTypeText"`
	TotalCount       int                       `json:"totalCount"`
	IssuedCount      int                       `json:"issuedCount"`
	PerMemberLimit   int                       `json:"perMemberLimit"`
	StartTime        types.DateTime            `json:"startTime"`
	EndTime          types.DateTime            `json:"endTime"`
	ValidDays        int                       `json:"validDays"`
	Status           int8                      `json:"status"`
	CreatedAt        types.DateTime            `json:"createdAt"`
}

// CouponTemplateListQuery 是券模板列表查询条件。
type CouponTemplateListQuery struct {
	// Keyword 券名称模糊匹配
	Keyword string
	// Type 券类型，0 表示不限
	Type enums.CouponType
	// Status 启用状态，nil 表示不限
	Status *int8
	// PageNo 页码
	PageNo int
	// PageSize 页大小
	PageSize int
}

// PageCouponTemplates 分页查询券模板。
// 入参：查询条件与分页参数；返回券模板列表与总数。
func (s *PromotionService) PageCouponTemplates(ctx context.Context, q CouponTemplateListQuery) ([]CouponTemplateRow, int64, error) {
	list, total, err := s.templates.Page(ctx, couponTemplateFilter(q), q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	rows := make([]CouponTemplateRow, 0, len(list))
	for i := range list {
		rows = append(rows, buildCouponTemplateRow(&list[i]))
	}

	return rows, total, nil
}

// DetailCouponTemplate 查询券模板详情，不存在返回业务错误码。
func (s *PromotionService) DetailCouponTemplate(ctx context.Context, id types.ID) (*CouponTemplateRow, error) {
	m, err := s.templates.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errs.New(promotionErrs.CouponTemplateNotExists)
	}

	row := buildCouponTemplateRow(m)

	return &row, nil
}

// SaveCouponTemplate 新增或更新券模板。
// 入参：模板 ID（0 表示新增）与保存内容；返回模板 ID。
// 说明：校验券类型、适用范围与金额口径；折扣券必须给折扣率且小于一万。
func (s *PromotionService) SaveCouponTemplate(ctx context.Context, id types.ID, in CouponTemplateSaveInput) (types.ID, error) {
	if err := validateCouponTemplate(in); err != nil {
		return 0, err
	}

	model := &promotionModel.CouponTemplateModel{
		Name:              strings.TrimSpace(in.Name),
		Type:              in.Type,
		Description:       strings.TrimSpace(in.Description),
		ThresholdAmount:   in.ThresholdAmount,
		DiscountAmount:    in.DiscountAmount,
		DiscountRate:      in.DiscountRate,
		MaxDiscountAmount: in.MaxDiscountAmount,
		Scope:             in.Scope,
		ScopeIDs:          joinIDs(in.ScopeIDs),
		ChannelLimit:      in.ChannelLimit,
		IssueType:         in.IssueType,
		TotalCount:        in.TotalCount,
		PerMemberLimit:    in.PerMemberLimit,
		StartTime:         in.StartTime,
		EndTime:           in.EndTime,
		ValidDays:         in.ValidDays,
		Status:            in.Status,
	}

	if id > 0 {
		model.ID = id
		if err := s.templates.Update(ctx, model); err != nil {
			return 0, err
		}

		return id, nil
	}

	if err := s.templates.Create(ctx, model); err != nil {
		return 0, err
	}

	return model.ID, nil
}

// DeleteCouponTemplate 删除券模板。
// 入参：模板 ID；返回错误表示模板不存在。
func (s *PromotionService) DeleteCouponTemplate(ctx context.Context, id types.ID) error {
	m, err := s.templates.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return errs.New(promotionErrs.CouponTemplateNotExists)
	}

	return s.templates.Delete(ctx, id)
}

// CouponListQuery 是会员券列表查询条件。
type CouponListQuery struct {
	// TemplateID 券模板，0 表示不限
	TemplateID types.ID
	// MemberID 会员，0 表示不限
	MemberID types.ID
	// Status 券状态，nil 表示不限
	Status *enums.CouponStatus
	// PageNo 页码
	PageNo int
	// PageSize 页大小
	PageSize int
}

// CouponRow 是会员券的输出行。
type CouponRow struct {
	ID          types.ID                `json:"id"`
	TemplateID  types.ID                `json:"templateId"`
	TemplateName string                 `json:"templateName"`
	MemberID    types.ID                `json:"memberId"`
	Status      int8                    `json:"status"`
	StatusText  string                  `json:"statusText"`
	OrderID     types.ID                `json:"orderId"`
	UseTime     *types.DateTime         `json:"useTime"`
	ExpireTime  types.DateTime          `json:"expireTime"`
	GetTime     types.DateTime          `json:"getTime"`
	CreatedAt   types.DateTime          `json:"createdAt"`
}

// PageCoupons 分页查询会员券，返回列表与总数。
func (s *PromotionService) PageCoupons(ctx context.Context, q CouponListQuery) ([]CouponRow, int64, error) {
	list, total, err := s.coupons.Page(ctx, couponFilter(q), q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	// 批量补模板名称，避免逐行查库
	templateIDs := make([]types.ID, 0, len(list))
	for _, item := range list {
		templateIDs = append(templateIDs, item.TemplateID)
	}
	templates, err := s.templates.FindByIDs(ctx, templateIDs)
	if err != nil {
		return nil, 0, err
	}
	templateName := make(map[types.ID]string, len(templates))
	for _, item := range templates {
		templateName[item.ID] = item.Name
	}

	rows := make([]CouponRow, 0, len(list))
	for i := range list {
		item := &list[i]
		rows = append(rows, CouponRow{
			ID: item.ID, TemplateID: item.TemplateID,
			TemplateName: templateName[item.TemplateID],
			MemberID:     item.MemberID,
			Status:       int8(item.Status), StatusText: item.Status.String(),
			OrderID:  item.OrderID, UseTime: item.UseTime,
			ExpireTime: item.ExpireTime, GetTime: item.GetTime, CreatedAt: item.CreatedAt,
		})
	}

	return rows, total, nil
}

// RevokeCoupon 作废会员券：未使用 → 已过期。
// 入参：券 ID；返回错误表示券不存在或状态不允许作废。
func (s *PromotionService) RevokeCoupon(ctx context.Context, id types.ID) error {
	coupon, err := s.coupons.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if coupon == nil {
		return errs.New(promotionErrs.CouponNotExists)
	}
	if coupon.Status != enums.CouponStatusUnused {
		return errs.New(promotionErrs.CouponStatusInvalid, coupon.Status.String())
	}

	moved, err := s.coupons.TransitOutside(ctx, id, enums.CouponStatusUnused, enums.CouponStatusExpired, map[string]any{
		"updated_at": types.Now(),
	})
	if err != nil {
		return err
	}
	if !moved {
		return errs.New(promotionErrs.CouponStatusInvalid, coupon.Status.String())
	}

	return nil
}

// ActivitySaveInput 是营销活动的保存入参。
type ActivitySaveInput struct {
	// Name 活动名称
	Name string
	// Type 活动玩法
	Type enums.ActivityType
	// Level 作用层级
	Level enums.ActivityLevel
	// Description 活动说明
	Description string
	// Rule 玩法规则
	Rule ActivityRule
	// Scope 适用门店范围
	Scope enums.ActivityScope
	// ScopeIDs 适用门店 ID 列表
	ScopeIDs []types.ID
	// Priority 优先级
	Priority int
	// MutexGroup 互斥组
	MutexGroup string
	// Exclusive 是否独占
	Exclusive bool
	// Stackable 是否可叠加
	Stackable bool
	// QuotaLimit 名额上限
	QuotaLimit int
	// QuotaCycle 名额周期
	QuotaCycle enums.ActivityQuotaCycle
	// StartTime 活动开始时间
	StartTime types.DateTime
	// EndTime 活动结束时间
	EndTime types.DateTime
	// Status 启用状态
	Status enums.ActivityStatus
	// ProductSPUIDs 参与商品的 SPU ID 列表
	ProductSPUIDs []types.ID
}

// ActivityRow 是营销活动的输出行。
type ActivityRow struct {
	ID          types.ID             `json:"id"`
	Name        string               `json:"name"`
	Type        int8                 `json:"type"`
	TypeText    string               `json:"typeText"`
	Level       int8                 `json:"level"`
	LevelText   string               `json:"levelText"`
	Description string               `json:"description"`
	Rule        string               `json:"rule"`
	Scope       int8                 `json:"scope"`
	ScopeText   string               `json:"scopeText"`
	ScopeIDs    string               `json:"scopeIds"`
	Priority    int                  `json:"priority"`
	MutexGroup  string               `json:"mutexGroup"`
	Exclusive   int8                 `json:"exclusive"`
	Stackable   int8                 `json:"stackable"`
	QuotaLimit  int                  `json:"quotaLimit"`
	QuotaCycle  int8                 `json:"quotaCycle"`
	StartTime   types.DateTime       `json:"startTime"`
	EndTime     types.DateTime       `json:"endTime"`
	Status      int8                 `json:"status"`
	StatusText  string               `json:"statusText"`
	ProductSPUIDs []types.ID         `json:"productSpuIds"`
	CreatedAt   types.DateTime       `json:"createdAt"`
}

// ActivityListQuery 是活动列表查询条件。
type ActivityListQuery struct {
	// Keyword 活动名称模糊匹配
	Keyword string
	// Type 活动玩法，0 表示不限
	Type enums.ActivityType
	// Status 启用状态，nil 表示不限
	Status *enums.ActivityStatus
	// PageNo 页码
	PageNo int
	// PageSize 页大小
	PageSize int
}

// PageActivities 分页查询营销活动，返回列表与总数。
func (s *PromotionService) PageActivities(ctx context.Context, q ActivityListQuery) ([]ActivityRow, int64, error) {
	list, total, err := s.activities.Page(ctx, activityFilter(q), q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	ids := make([]types.ID, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.ID)
	}
	productMap, err := s.activities.ListProductSPUIDsByActivityIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	rows := make([]ActivityRow, 0, len(list))
	for i := range list {
		row := buildActivityRow(&list[i])
		row.ProductSPUIDs = productMap[list[i].ID]
		rows = append(rows, row)
	}

	return rows, total, nil
}

// DetailActivity 查询营销活动详情，含参与商品。
func (s *PromotionService) DetailActivity(ctx context.Context, id types.ID) (*ActivityRow, error) {
	m, err := s.activities.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errs.New(promotionErrs.ActivityNotExists)
	}

	row := buildActivityRow(m)
	spuIDs, err := s.activities.ListProductSPUIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	row.ProductSPUIDs = spuIDs

	return &row, nil
}

// SaveActivity 新增或更新营销活动。
// 入参：活动 ID（0 表示新增）与保存内容；返回活动 ID。
// 说明：商品级活动必须配置参与商品；规则按玩法校验必填字段。
func (s *PromotionService) SaveActivity(ctx context.Context, id types.ID, in ActivitySaveInput) (types.ID, error) {
	if err := validateActivity(in); err != nil {
		return 0, err
	}

	ruleJSON, err := json.Marshal(in.Rule)
	if err != nil {
		return 0, err
	}

	model := &promotionModel.ActivityModel{
		Name:        strings.TrimSpace(in.Name),
		Type:        in.Type,
		Level:       in.Level,
		Description: strings.TrimSpace(in.Description),
		Rule:        string(ruleJSON),
		Scope:       in.Scope,
		ScopeIDs:    joinIDs(in.ScopeIDs),
		Priority:    in.Priority,
		MutexGroup:  strings.TrimSpace(in.MutexGroup),
		Exclusive:   boolToInt8(in.Exclusive),
		Stackable:   boolToInt8(in.Stackable),
		QuotaLimit:  in.QuotaLimit,
		QuotaCycle:  in.QuotaCycle,
		StartTime:   in.StartTime,
		EndTime:      in.EndTime,
		Status:      in.Status,
	}

	if id > 0 {
		model.ID = id
		if err := s.activities.Update(ctx, model); err != nil {
			return 0, err
		}
		if err := s.activities.ReplaceProducts(ctx, id, in.ProductSPUIDs); err != nil {
			return 0, err
		}

		return id, nil
	}

	if err := s.activities.Create(ctx, model); err != nil {
		return 0, err
	}
	if err := s.activities.ReplaceProducts(ctx, model.ID, in.ProductSPUIDs); err != nil {
		return 0, err
	}

	return model.ID, nil
}

// DeleteActivity 删除营销活动及其参与商品配置。
func (s *PromotionService) DeleteActivity(ctx context.Context, id types.ID) error {
	m, err := s.activities.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return errs.New(promotionErrs.ActivityNotExists)
	}

	return s.activities.Delete(ctx, id)
}

// validateCouponTemplate 校验券模板的取值口径。
func validateCouponTemplate(in CouponTemplateSaveInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errs.New(promotionErrs.CouponTemplateAmountInvalid)
	}
	if in.Type < enums.CouponTypeFullReduction || in.Type > enums.CouponTypeCash {
		return errs.New(promotionErrs.CouponTemplateTypeInvalid)
	}
	if in.Scope < enums.CouponScopeAll || in.Scope > enums.CouponScopeProduct {
		return errs.New(promotionErrs.CouponTemplateScopeInvalid)
	}
	// 指定门店 / 指定商品必须选对象，否则券永远不可用
	if in.Scope != enums.CouponScopeAll && len(in.ScopeIDs) == 0 {
		return errs.New(promotionErrs.CouponTemplateScopeInvalid)
	}
	if in.IssueType < enums.CouponIssueActive || in.IssueType > enums.CouponIssueDirected {
		return errs.New(promotionErrs.CouponTemplateIssueTypeInvalid)
	}
	if in.ThresholdAmount < 0 || in.DiscountAmount < 0 || in.MaxDiscountAmount < 0 {
		return errs.New(promotionErrs.CouponTemplateAmountInvalid)
	}
	if in.TotalCount < 0 || in.PerMemberLimit < 0 || in.ValidDays < 0 {
		return errs.New(promotionErrs.CouponTemplateQuotaInvalid)
	}
	// 折扣券必须给折扣率且小于一万；其他券必须给抵扣金额
	if in.Type == enums.CouponTypeDiscount {
		if in.DiscountRate <= 0 || in.DiscountRate >= 10000 {
			return errs.New(promotionErrs.CouponTemplateAmountInvalid)
		}
	} else if in.DiscountAmount <= 0 {
		return errs.New(promotionErrs.CouponTemplateAmountInvalid)
	}

	return nil
}

// validateActivity 校验活动的取值口径。
func validateActivity(in ActivitySaveInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errs.New(promotionErrs.ActivityRuleInvalid)
	}
	if in.Type < enums.ActivityTypeFullReduction || in.Type > enums.ActivityTypeAddonPurchase {
		return errs.New(promotionErrs.ActivityTypeInvalid)
	}
	if in.Level != enums.ActivityLevelProduct && in.Level != enums.ActivityLevelOrder {
		return errs.New(promotionErrs.ActivityLevelInvalid)
	}
	if in.Scope < enums.ActivityScopeAll || in.Scope > enums.ActivityScopeStore {
		return errs.New(promotionErrs.ActivityScopeInvalid)
	}
	if in.Scope == enums.ActivityScopeStore && len(in.ScopeIDs) == 0 {
		return errs.New(promotionErrs.ActivityScopeInvalid)
	}
	// 商品级活动必须明确参与商品，否则活动范围不明确
	if in.Level == enums.ActivityLevelProduct && len(in.ProductSPUIDs) == 0 {
		return errs.New(promotionErrs.ActivityProductRequired)
	}
	if in.QuotaLimit < 0 {
		return errs.New(promotionErrs.ActivityQuotaInvalid)
	}

	// 按玩法校验规则的必填字段
	switch in.Type {
	case enums.ActivityTypeFullReduction:
		if len(in.Rule.Tiers) == 0 && in.Rule.DiscountAmount <= 0 {
			return errs.New(promotionErrs.ActivityRuleInvalid)
		}
	case enums.ActivityTypeTimeDiscount:
		if in.Rule.DiscountRate <= 0 || in.Rule.DiscountRate >= 10000 {
			return errs.New(promotionErrs.ActivityRuleInvalid)
		}
	case enums.ActivityTypeCombo:
		if len(in.Rule.ComboItems) == 0 || in.Rule.ComboPrice <= 0 {
			return errs.New(promotionErrs.ActivityRuleInvalid)
		}
	case enums.ActivityTypeAddonPurchase:
		if in.Rule.AddonSpuID == 0 || in.Rule.AddonPrice <= 0 {
			return errs.New(promotionErrs.ActivityRuleInvalid)
		}
	case enums.ActivityTypeSecondHalfPrice:
		// 第二份半价的折扣率缺省按半价处理，不强制必填
	}

	return nil
}

// buildCouponTemplateRow 把券模板模型转成输出行。
func buildCouponTemplateRow(m *promotionModel.CouponTemplateModel) CouponTemplateRow {
	return CouponTemplateRow{
		ID: m.ID, Name: m.Name,
		Type: int8(m.Type), TypeText: m.Type.String(),
		Description:   m.Description,
		ThresholdAmount: m.ThresholdAmount, DiscountAmount: m.DiscountAmount,
		DiscountRate: m.DiscountRate, MaxDiscountAmount: m.MaxDiscountAmount,
		Scope: int8(m.Scope), ScopeText: m.Scope.String(), ScopeIDs: m.ScopeIDs,
		ChannelLimit: int8(m.ChannelLimit), ChannelLimitText: m.ChannelLimit.String(),
		IssueType: int8(m.IssueType), IssueTypeText: m.IssueType.String(),
		TotalCount: m.TotalCount, IssuedCount: m.IssuedCount, PerMemberLimit: m.PerMemberLimit,
		StartTime: m.StartTime, EndTime: m.EndTime, ValidDays: m.ValidDays,
		Status: int8(m.Status), CreatedAt: m.CreatedAt,
	}
}

// buildActivityRow 把活动模型转成输出行。
func buildActivityRow(m *promotionModel.ActivityModel) ActivityRow {
	return ActivityRow{
		ID: m.ID, Name: m.Name,
		Type: int8(m.Type), TypeText: m.Type.String(),
		Level: int8(m.Level), LevelText: m.Level.String(),
		Description: m.Description, Rule: m.Rule,
		Scope: int8(m.Scope), ScopeText: m.Scope.String(), ScopeIDs: m.ScopeIDs,
		Priority: m.Priority, MutexGroup: m.MutexGroup,
		Exclusive: m.Exclusive, Stackable: m.Stackable,
		QuotaLimit: m.QuotaLimit, QuotaCycle: int8(m.QuotaCycle),
		StartTime: m.StartTime, EndTime: m.EndTime,
		Status: int8(m.Status), StatusText: m.Status.String(),
		CreatedAt: m.CreatedAt,
	}
}

// couponTemplateFilter 把列表查询条件转成仓储筛选条件。
func couponTemplateFilter(q CouponTemplateListQuery) promotionRepository.TemplateListFilter {
	return promotionRepository.TemplateListFilter{
		Name:   strings.TrimSpace(q.Keyword),
		Type:   q.Type,
		Status: q.Status,
	}
}

// couponFilter 把列表查询条件转成仓储筛选条件。
func couponFilter(q CouponListQuery) promotionRepository.CouponListFilter {
	return promotionRepository.CouponListFilter{
		TemplateID: q.TemplateID,
		MemberID:   q.MemberID,
		Status:     q.Status,
	}
}

// activityFilter 把列表查询条件转成仓储筛选条件。
func activityFilter(q ActivityListQuery) promotionRepository.ActivityListFilter {
	return promotionRepository.ActivityListFilter{
		Name:   strings.TrimSpace(q.Keyword),
		Type:   q.Type,
		Status: q.Status,
	}
}

// joinIDs 把 ID 列表拼成逗号分隔的字符串。
func joinIDs(ids []types.ID) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			parts = append(parts, id.String())
		}
	}

	return strings.Join(parts, ",")
}

// boolToInt8 把布尔值转成 tinyint。
func boolToInt8(value bool) int8 {
	if value {
		return 1
	}

	return 0
}
