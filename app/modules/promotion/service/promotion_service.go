// promotion_service.go 实现 promotion 模块的业务层：规则解析入口、券生命周期与积分抵扣。
// 对外只暴露「算一次优惠」与「券状态流转」两组能力，订单模块通过它们接入营销。
package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/promotion/enums"
	promotionErrs "eqray-flavor/app/modules/promotion/errs"
	promotionModel "eqray-flavor/app/modules/promotion/model"
	promotionRepository "eqray-flavor/app/modules/promotion/repository"
)

// PromotionService 是营销模块的业务入口，聚合券模板、会员券、活动与积分规则仓储。
type PromotionService struct {
	templates      *promotionRepository.CouponTemplateRepository
	coupons        *promotionRepository.CouponRepository
	activities     *promotionRepository.ActivityRepository
	quotas         *promotionRepository.ActivityQuotaRepository
	orderActivities *promotionRepository.OrderActivityRepository
	pointRule      *promotionRepository.PointRuleRepository
}

// NewPromotionService 构造营销服务。
func NewPromotionService() *PromotionService {
	return &PromotionService{
		templates:       promotionRepository.NewCouponTemplateRepository(),
		coupons:         promotionRepository.NewCouponRepository(),
		activities:      promotionRepository.NewActivityRepository(),
		quotas:          promotionRepository.NewActivityQuotaRepository(),
		orderActivities: promotionRepository.NewOrderActivityRepository(),
		pointRule:       promotionRepository.NewPointRuleRepository(),
	}
}

// PricingInput 是营销规则解析的入参，由订单模块装配。
type PricingInput struct {
	// StoreID 门店 ID
	StoreID types.ID
	// MemberID 会员 ID，0 表示散客
	MemberID types.ID
	// Channel 下单渠道：1 小程序 / 2 POS / 3 自助机 / 4 外卖平台
	Channel int8
	// Items 明细行
	Items []RuleItem
	// CouponID 使用的会员券 ID，0 表示不用券
	CouponID types.ID
	// UsePoints 是否使用积分抵扣
	UsePoints bool
}

// PricingOutput 是营销规则解析的结果。
type PricingOutput struct {
	// LineDiscounts 行级单品优惠，与输入明细同序
	LineDiscounts []LineDiscount
	// ActivityAmount 订单级活动优惠合计（分）
	ActivityAmount int64
	// ActivityNames 生效的订单级活动名称
	ActivityNames []string
	// CouponAmount 券抵扣金额（分）
	CouponAmount int64
	// CouponName 生效券名称
	CouponName string
	// CouponThreshold 券使用门槛（分）
	CouponThreshold int64
	// PointsAmount 积分抵扣金额（分）
	PointsAmount int64
	// PointsUsed 实际使用的积分数
	PointsUsed int
	// ExchangeRate 本次计价使用的积分兑换比例：多少积分抵 1 元（分）。
	// 下单时按这个比例把抵扣金额折算成积分数并落单，
	// 支付与退款都用订单上的积分数，避免比例调整后多扣少退。
	ExchangeRate int
	// Notes 解析留档
	Notes []string
	// ActivityIDs 实际产生优惠的活动 ID（商品级 + 订单级），用于下单时占用名额
	ActivityIDs []types.ID
}

// Pricing 解析一次营销规则：装配候选活动、校验券、读积分规则后交给纯函数 Resolve。
// 入参：门店、会员、渠道、明细与用券 / 用积分意向；返回优惠拆分与留档。
// 说明：本方法只读不写，名额占用与券锁定发生在下单时。
func (s *PromotionService) Pricing(ctx context.Context, in PricingInput) (*PricingOutput, error) {
	now := types.Now()

	// ---------- 装配候选活动 ----------
	activities, err := s.activities.ListEnabled(ctx, in.StoreID, now)
	if err != nil {
		return nil, err
	}

	activityIDs := make([]types.ID, 0, len(activities))
	for _, activity := range activities {
		activityIDs = append(activityIDs, activity.ID)
	}
	productMap, err := s.activities.ListProductSPUIDsByActivityIDs(ctx, activityIDs)
	if err != nil {
		return nil, err
	}

	ruleActivities := make([]RuleActivity, 0, len(activities))
	for i := range activities {
		activity := &activities[i]
		quotaUsed := 0
		periodKey := quotaPeriodKey(activity, now)
		if activity.QuotaLimit > 0 {
			quotaUsed, err = s.quotas.GetUsedCount(ctx, activity.ID, periodKey)
			if err != nil {
				return nil, err
			}
		}
		ruleActivities = append(ruleActivities, RuleActivity{
			ID:            activity.ID,
			Name:          activity.Name,
			Type:          activity.Type,
			Level:         activity.Level,
			Rule:          ParseActivityRule(activity.Rule),
			Priority:      activity.Priority,
			MutexGroup:    activity.MutexGroup,
			Exclusive:     activity.Exclusive == 1,
			Stackable:     activity.Stackable == 1,
			QuotaLimit:    activity.QuotaLimit,
			QuotaUsed:     quotaUsed,
			ProductSPUIDs: productMap[activity.ID],
		})
	}

	// ---------- 装配券 ----------
	var coupon *RuleCoupon
	if in.CouponID > 0 {
		coupon, err = s.buildCoupon(ctx, in.CouponID, in.MemberID, in.StoreID, in.Channel)
		if err != nil {
			return nil, err
		}
	}

	// ---------- 装配积分 ----------
	pointsAmount := int64(0)
	exchangeRate := 0
	maxDeduct := int64(0)
	minOrder := int64(0)
	if in.MemberID > 0 {
		rule, ruleErr := s.pointRule.FindEnabled(ctx)
		if ruleErr != nil {
			return nil, ruleErr
		}
		if rule != nil {
			exchangeRate = rule.ExchangeRate
			maxDeduct = rule.MaxDeductAmount
			minOrder = rule.MinOrderAmount
			// 会员可用积分决定拟抵扣金额
			if in.UsePoints {
				available, pointErr := s.memberAvailablePoints(ctx, in.MemberID)
				if pointErr != nil {
					return nil, pointErr
				}
				pointsAmount = pointsToAmount(available, rule.ExchangeRate)
			}
		}
	}

	result := Resolve(RuleInput{
		StoreID:        in.StoreID,
		Channel:        in.Channel,
		Items:          in.Items,
		Activities:     ruleActivities,
		Coupon:         coupon,
		PointsAmount:   pointsAmount,
		PointMaxDeduct: maxDeduct,
		PointMinOrder:  minOrder,
	})

	out := &PricingOutput{
		LineDiscounts:   result.LineDiscounts,
		ActivityAmount:  result.ActivityAmount,
		ActivityNames:   result.ActivityNames,
		ActivityIDs:     result.ActivityIDs,
		CouponAmount:    result.CouponAmount,
		CouponName:      result.CouponName,
		CouponThreshold: result.CouponThreshold,
		PointsAmount:    result.PointsAmount,
		Notes:           result.Notes,
	}

	// 积分抵扣金额按兑换比例回算实际使用积分数
	if out.PointsAmount > 0 && exchangeRate > 0 {
		out.PointsUsed = int(out.PointsAmount * int64(exchangeRate) / 100)
	}
	out.ExchangeRate = exchangeRate

	return out, nil
}

// buildCoupon 装配券快照并校验归属、有效期、适用范围与渠道限制。
// 入参：券 ID、会员 ID、门店 ID、渠道；返回券快照，校验不通过时返回业务错误。
func (s *PromotionService) buildCoupon(ctx context.Context, couponID, memberID, storeID types.ID, channel int8) (*RuleCoupon, error) {
	coupon, err := s.coupons.FindByID(ctx, couponID)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, errs.New(promotionErrs.CouponNotExists)
	}
	if memberID > 0 && coupon.MemberID != memberID {
		return nil, errs.New(promotionErrs.CouponNotOwned)
	}
	if coupon.Status != enums.CouponStatusUnused {
		return nil, errs.New(promotionErrs.CouponStatusInvalid, coupon.Status.String())
	}
	if coupon.ExpireTime.Time().Before(types.Now().Time()) {
		return nil, errs.New(promotionErrs.CouponExpired)
	}

	template, err := s.templates.FindByID(ctx, coupon.TemplateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errs.New(promotionErrs.CouponTemplateNotExists)
	}
	if template.Status != baseEnums.CommonStatusEnabled {
		return nil, errs.New(promotionErrs.CouponTemplateDisabled)
	}

	// 渠道限制：线上渠道（小程序 / 外卖平台）与线下渠道（POS / 自助机）
	if !channelAllowed(template.ChannelLimit, channel) {
		return nil, errs.New(promotionErrs.CouponChannelNotMatch)
	}

	// 适用范围：指定门店时校验当前门店
	if template.Scope == enums.CouponScopeStore && !scopeContains(template.ScopeIDs, storeID) {
		return nil, errs.New(promotionErrs.CouponScopeNotMatch)
	}

	return &RuleCoupon{
		ID:                coupon.ID,
		Name:              template.Name,
		Type:              template.Type,
		ThresholdAmount:   template.ThresholdAmount,
		DiscountAmount:    template.DiscountAmount,
		DiscountRate:      template.DiscountRate,
		MaxDiscountAmount: template.MaxDiscountAmount,
		Scope:             template.Scope,
		ScopeIDs:          parseScopeIDs(template.ScopeIDs),
	}, nil
}

// IssueInput 是发券入参。
type IssueInput struct {
	// TemplateID 券模板 ID
	TemplateID types.ID
	// MemberID 目标会员 ID
	MemberID types.ID
	// Count 发放张数，缺省 1
	Count int
}

// Issue 向会员发放优惠券。
// 入参：模板、会员与张数；返回实际发放张数。
// 并发保护：发行量按「已发放 + 本次 <= 发行总量」条件更新，每人限领在同事务内校验。
func (s *PromotionService) Issue(ctx context.Context, in IssueInput) (int, error) {
	template, err := s.templates.FindByID(ctx, in.TemplateID)
	if err != nil {
		return 0, err
	}
	if template == nil {
		return 0, errs.New(promotionErrs.CouponTemplateNotExists)
	}
	if template.Status != baseEnums.CommonStatusEnabled {
		return 0, errs.New(promotionErrs.CouponTemplateDisabled)
	}

	now := types.Now()
	if now.Time().Before(template.StartTime.Time()) || now.Time().After(template.EndTime.Time()) {
		return 0, errs.New(promotionErrs.CouponTemplateDisabled)
	}

	count := in.Count
	if count <= 0 {
		count = 1
	}

	issued := 0
	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		// 每人限领校验：已持有数量 + 本次不能超过限领
		if template.PerMemberLimit > 0 {
			owned, countErr := s.coupons.CountByTemplateAndMember(ctx, template.ID, in.MemberID)
			if countErr != nil {
				return countErr
			}
			if owned+count > template.PerMemberLimit {
				return errs.New(promotionErrs.CouponIssueLimitReached)
			}
		}

		// 发行量原子扣减，限量模板发完即拒绝
		ok, incErr := s.templates.IncrementIssuedCount(ctx, tx, template.ID, count)
		if incErr != nil {
			return incErr
		}
		if !ok {
			return errs.New(promotionErrs.CouponTemplateSoldOut)
		}

		// 有效期：优先用 valid_days，否则到模板结束时间
		expire := template.EndTime
		if template.ValidDays > 0 {
			expire = types.DateTime(now.Time().AddDate(0, 0, template.ValidDays))
		}

		for i := 0; i < count; i++ {
			coupon := &promotionModel.CouponModel{
				TemplateID: template.ID,
				MemberID:   in.MemberID,
				Status:     enums.CouponStatusUnused,
				ExpireTime: expire,
				GetTime:    now,
			}
			if createErr := s.coupons.Create(ctx, tx, coupon); createErr != nil {
				return createErr
			}
		}
		issued = count

		return nil
	})
	if err != nil {
		return 0, err
	}

	return issued, nil
}

// LockCoupons 锁定订单占用的券：未使用 → 锁定。
// 入参：事务、订单 ID、券 ID 列表；返回成功锁定的券 ID 列表。
// 并发保护：状态流转用带原状态的条件更新，失败说明已被其他订单占用。
func (s *PromotionService) LockCoupons(ctx context.Context, tx db.Tx, orderID types.ID, couponIDs []types.ID) ([]types.ID, error) {
	locked := make([]types.ID, 0, len(couponIDs))
	for _, couponID := range couponIDs {
		ok, err := s.coupons.Transit(ctx, tx, couponID, enums.CouponStatusUnused, enums.CouponStatusLocked, map[string]any{
			"order_id":   orderID,
			"updated_at": types.Now(),
		})
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errs.New(promotionErrs.CouponStatusInvalid, enums.CouponStatusUnused.String())
		}
		locked = append(locked, couponID)
	}

	return locked, nil
}

// UseCoupons 核销订单的券：锁定 → 已使用。
// 入参：事务、订单 ID；返回核销的券 ID 列表。
func (s *PromotionService) UseCoupons(ctx context.Context, tx db.Tx, orderID types.ID) ([]types.ID, error) {
	coupons, err := s.coupons.ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	used := make([]types.ID, 0, len(coupons))
	for i := range coupons {
		coupon := &coupons[i]
		if coupon.Status != enums.CouponStatusLocked {
			continue
		}
		ok, transitErr := s.coupons.Transit(ctx, tx, coupon.ID, enums.CouponStatusLocked, enums.CouponStatusUsed, map[string]any{
			"use_time":   types.Now(),
			"updated_at": types.Now(),
		})
		if transitErr != nil {
			return nil, transitErr
		}
		if ok {
			used = append(used, coupon.ID)
		}
	}

	return used, nil
}

// ReleaseCoupons 释放订单占用的券：锁定 → 未使用（取消 / 超时关单时调用）。
// 入参：事务、订单 ID；返回释放的券 ID 列表。
func (s *PromotionService) ReleaseCoupons(ctx context.Context, tx db.Tx, orderID types.ID) ([]types.ID, error) {
	coupons, err := s.coupons.ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	released := make([]types.ID, 0, len(coupons))
	for i := range coupons {
		coupon := &coupons[i]
		if coupon.Status != enums.CouponStatusLocked {
			continue
		}
		ok, transitErr := s.coupons.Transit(ctx, tx, coupon.ID, enums.CouponStatusLocked, enums.CouponStatusUnused, map[string]any{
			"order_id":   0,
			"updated_at": types.Now(),
		})
		if transitErr != nil {
			return nil, transitErr
		}
		if ok {
			released = append(released, coupon.ID)
		}
	}

	return released, nil
}

// ReturnCoupons 整单退款返还券：已使用 → 未使用（部分退款不返还）。
// 入参：事务、订单 ID；返回返还的券 ID 列表。
func (s *PromotionService) ReturnCoupons(ctx context.Context, tx db.Tx, orderID types.ID) ([]types.ID, error) {
	coupons, err := s.coupons.ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	returned := make([]types.ID, 0, len(coupons))
	for i := range coupons {
		coupon := &coupons[i]
		if coupon.Status != enums.CouponStatusUsed {
			continue
		}
		ok, transitErr := s.coupons.Transit(ctx, tx, coupon.ID, enums.CouponStatusUsed, enums.CouponStatusUnused, map[string]any{
			"order_id":   0,
			"use_time":   nil,
			"updated_at": types.Now(),
		})
		if transitErr != nil {
			return nil, transitErr
		}
		if ok {
			returned = append(returned, coupon.ID)
		}
	}

	return returned, nil
}

// OccupyActivityQuotas 占用活动名额：按「活动 + 周期键」计数。
// 入参：事务、活动 ID 列表、当前时间；返回成功占用的活动 ID 列表。
// 占用失败（名额已被并发抢完）直接报错：活动优惠已计入订单金额，
// 此时若跳过占用会让活动超发，因此整单回滚、让用户重新试算。
// 说明：不限名额的活动没有计数行可占，也不写关联记录，释放时自然跳过。
func (s *PromotionService) OccupyActivityQuotas(ctx context.Context, tx db.Tx, activityIDs []types.ID, now types.DateTime) ([]types.ID, error) {
	occupied := make([]types.ID, 0, len(activityIDs))
	// 同一活动可能既产生单品级优惠又产生订单级优惠，占用前先去重：
	// 重复占用会多扣名额，也会撞上 (order_id, activity_id, period_key) 唯一键。
	seen := make(map[types.ID]bool, len(activityIDs))
	for _, activityID := range activityIDs {
		if activityID == 0 || seen[activityID] {
			continue
		}
		seen[activityID] = true

		activity, err := s.activities.FindByID(ctx, activityID)
		if err != nil {
			return nil, err
		}
		if activity == nil {
			continue
		}
		if activity.QuotaLimit <= 0 {
			// 不限名额：没有计数行可占，也就没有名额可释放
			continue
		}

		periodKey := quotaPeriodKey(activity, now)
		ok, occupyErr := s.quotas.Occupy(ctx, tx, activityID, periodKey, activity.QuotaLimit)
		if occupyErr != nil {
			return nil, occupyErr
		}
		if !ok {
			// 名额已被并发抢完：优惠已计入订单金额，不能让这单成立
			return nil, errs.New(promotionErrs.ActivityQuotaNotEnough, activity.Name)
		}
		occupied = append(occupied, activityID)
	}

	return occupied, nil
}

// OccupyActivityQuotasForOrder 占用活动名额并记录订单 ↔ 活动 ↔ 周期键的关联。
// 入参：事务、订单 ID、活动 ID 列表、当前时间；返回成功占用的活动 ID 列表。
// 说明：关联记录是释放名额的依据——每日周期的活动跨天后，
// 用当前日期算周期键会退错池子，必须按下单时记下的周期键回退。
func (s *PromotionService) OccupyActivityQuotasForOrder(
	ctx context.Context,
	tx db.Tx,
	orderID types.ID,
	activityIDs []types.ID,
	now types.DateTime,
) ([]types.ID, error) {
	occupied, err := s.OccupyActivityQuotas(ctx, tx, activityIDs, now)
	if err != nil {
		return nil, err
	}
	if len(occupied) == 0 {
		return occupied, nil
	}

	for _, activityID := range occupied {
		activity, findErr := s.activities.FindByID(ctx, activityID)
		if findErr != nil {
			return nil, findErr
		}
		if activity == nil || activity.QuotaLimit <= 0 {
			continue
		}

		record := &promotionModel.OrderActivityModel{
			OrderID:    orderID,
			ActivityID: activityID,
			PeriodKey:  quotaPeriodKey(activity, now),
			Status:     int8(enums.OrderActivityStatusOccupied),
		}
		if createErr := s.orderActivities.Create(ctx, tx, record); createErr != nil {
			return nil, createErr
		}
	}

	return occupied, nil
}

// ReleaseActivityQuotas 释放订单占用的活动名额：取消 / 超时关单 / 整单退款时调用。
// 入参：事务、订单 ID；返回释放的活动 ID 列表。
// 说明：按关联记录里记下的周期键逐条回退 used_count，
// 关联记录用带原状态的条件更新流转，保证同一名额不会被重复释放。
func (s *PromotionService) ReleaseActivityQuotas(ctx context.Context, tx db.Tx, orderID types.ID) ([]types.ID, error) {
	records, err := s.orderActivities.ListByOrderID(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}

	released := make([]types.ID, 0, len(records))
	for i := range records {
		record := &records[i]
		if record.Status != int8(enums.OrderActivityStatusOccupied) {
			continue
		}

		moved, releaseErr := s.orderActivities.Release(ctx, tx, record.ID)
		if releaseErr != nil {
			return nil, releaseErr
		}
		if !moved {
			// 并发下已被释放过，跳过，避免重复回退计数
			continue
		}

		if releaseQuotaErr := s.quotas.Release(ctx, tx, record.ActivityID, record.PeriodKey); releaseQuotaErr != nil {
			return nil, releaseQuotaErr
		}
		released = append(released, record.ActivityID)
	}

	return released, nil
}

// PointRuleOutput 是积分抵扣规则的对外视图。
type PointRuleOutput struct {
	// ExchangeRate 兑换比例：多少积分抵 1 元（分）
	ExchangeRate int `json:"exchangeRate"`
	// MaxDeductAmount 单笔抵扣上限（分）
	MaxDeductAmount int64 `json:"maxDeductAmount"`
	// MinOrderAmount 使用积分抵扣的最低订单金额（分）
	MinOrderAmount int64 `json:"minOrderAmount"`
	// Status 启用状态
	Status int8 `json:"status"`
}

// PointRule 读取当前生效的积分抵扣规则；未配置时返回默认规则。
func (s *PromotionService) PointRule(ctx context.Context) (*PointRuleOutput, error) {
	rule, err := s.pointRule.FindEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return &PointRuleOutput{ExchangeRate: 100, MaxDeductAmount: 5000}, nil
	}

	return &PointRuleOutput{
		ExchangeRate:     rule.ExchangeRate,
		MaxDeductAmount: rule.MaxDeductAmount,
		MinOrderAmount:  rule.MinOrderAmount,
		Status:           int8(rule.Status),
	}, nil
}

// UpdatePointRule 保存积分抵扣规则。
// 入参：兑换比例、单笔上限、最低订单金额与启用状态；返回错误表示规则不合法。
func (s *PromotionService) UpdatePointRule(ctx context.Context, in PointRuleOutput) error {
	if in.ExchangeRate <= 0 {
		return errs.New(promotionErrs.PointsRuleInvalid)
	}
	if in.MaxDeductAmount < 0 || in.MinOrderAmount < 0 {
		return errs.New(promotionErrs.PointsRuleInvalid)
	}

	current, err := s.pointRule.FindEnabled(ctx)
	if err != nil {
		return err
	}

	rule := &promotionModel.PointRuleModel{
		ExchangeRate:     in.ExchangeRate,
		MaxDeductAmount: in.MaxDeductAmount,
		MinOrderAmount:  in.MinOrderAmount,
		Status:           1,
	}
	if current != nil {
		rule.ID = current.ID
	}
	if in.Status == 0 {
		rule.Status = 0
	}

	return s.pointRule.Save(ctx, rule)
}

// memberAvailablePoints 读取会员当前可用积分。
// 入参：会员 ID；返回可用积分数，读取失败时返回错误。
func (s *PromotionService) memberAvailablePoints(ctx context.Context, memberID types.ID) (int, error) {
	row := struct {
		Point int `db:"point"`
	}{}
	if err := facades.DB().WithContext(ctx).Table("member_user").
		Select("point").
		Where("id", memberID).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, err
	}

	return row.Point, nil
}

// quotaPeriodKey 生成活动名额的周期键：整期用 fixed，每日用 YYYY-MM-DD。
func quotaPeriodKey(activity *promotionModel.ActivityModel, now types.DateTime) string {
	if activity.QuotaCycle == enums.ActivityQuotaCycleDaily {
		return now.Time().Format("2006-01-02")
	}

	return "fixed"
}

// pointsToAmount 把积分数折算成抵扣金额（分）：积分数 / 兑换比例 × 100。
func pointsToAmount(points, exchangeRate int) int64 {
	if points <= 0 || exchangeRate <= 0 {
		return 0
	}

	return int64(points) * 100 / int64(exchangeRate)
}

// channelAllowed 判断券的渠道限制是否允许当前下单渠道。
// 线上渠道为小程序与外卖平台，线下渠道为 POS 与自助机。
func channelAllowed(limit enums.CouponChannelLimit, channel int8) bool {
	switch limit {
	case enums.CouponChannelUnlimited:
		return true
	case enums.CouponChannelOnlineOnly:
		return channel == 1 || channel == 4
	case enums.CouponChannelOfflineOnly:
		return channel == 2 || channel == 3
	default:
		return true
	}
}

// scopeContains 判断逗号分隔的 ID 列表是否含目标 ID。
func scopeContains(scopeIDs string, target types.ID) bool {
	if scopeIDs == "" || target == 0 {
		return false
	}
	for _, part := range strings.Split(scopeIDs, ",") {
		if strings.TrimSpace(part) == target.String() {
			return true
		}
	}

	return false
}

// parseScopeIDs 把逗号分隔的 ID 列表解析成 ID 切片。
func parseScopeIDs(scopeIDs string) []types.ID {
	ids := make([]types.ID, 0)
	for _, part := range strings.Split(scopeIDs, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parsed, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, types.ID(parsed))
	}

	return ids
}
