// rule.go 实现营销规则解析：把活动、优惠券、积分三类规则算成一次可复现的优惠结果。
// 口径唯一来源：docs/design/07-会员与营销设计.md §4 优惠计算顺序、§5 活动引擎判定顺序。
//
// 本文件是纯计算部分，不依赖数据库：入参是装配好的活动与券快照，
// 出参是 RuleResult，由 service 层负责从库里装配入参。
package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/promotion/enums"
)

// ActivityRule 是活动的玩法规则 JSON 结构。
// 不同玩法用不同字段，未使用的字段留空；解析时按 type 取对应字段。
type ActivityRule struct {
	// ThresholdAmount 满减门槛（分），满减玩法使用
	ThresholdAmount int64 `json:"thresholdAmount"`
	// DiscountAmount 减免金额（分），满减与加价购使用
	DiscountAmount int64 `json:"discountAmount"`
	// Tiers 阶梯满减档位，按 thresholdAmount 升序；取满足门槛的最高一档
	Tiers []ActivityTier `json:"tiers"`
	// DiscountRate 折扣率（万分比），限时折扣使用：7500 表示 75 折
	DiscountRate int32 `json:"discountRate"`
	// MaxDiscountAmount 折扣封顶（分），0 表示不封顶
	MaxDiscountAmount int64 `json:"maxDiscountAmount"`
	// SecondHalfRate 第二份折扣率（万分比），第二份半价使用：5000 表示半价
	SecondHalfRate int32 `json:"secondHalfRate"`
	// ComboItems 套餐商品组合，套餐特惠使用
	ComboItems []ComboItem `json:"comboItems"`
	// ComboPrice 套餐价（分），套餐特惠使用
	ComboPrice int64 `json:"comboPrice"`
	// AddonSpuID 加价购商品 SPU ID，加价购使用
	AddonSpuID types.ID `json:"addonSpuId"`
	// AddonPrice 加价购价格（分），加价购使用
	AddonPrice int64 `json:"addonPrice"`
}

// ActivityTier 是阶梯满减的一档。
type ActivityTier struct {
	// ThresholdAmount 该档门槛（分）
	ThresholdAmount int64 `json:"thresholdAmount"`
	// DiscountAmount 该档减免金额（分）
	DiscountAmount int64 `json:"discountAmount"`
}

// ComboItem 是套餐中的一个商品。
type ComboItem struct {
	// SpuID 商品 SPU ID
	SpuID types.ID `json:"spuId"`
	// Quantity 数量
	Quantity int `json:"quantity"`
}

// RuleItem 是参与优惠计算的一行商品。
type RuleItem struct {
	// SpuID 菜品 SPU ID
	SpuID types.ID
	// SkuID 菜品规格 SKU ID
	SkuID types.ID
	// SpuName 菜品名称
	SpuName string
	// UnitPrice 成交单价（分）
	UnitPrice int64
	// Quantity 数量
	Quantity int
	// AddonAmount 加料费合计（分）
	AddonAmount int64
}

// Subtotal 返回本行小计（分）。
func (i RuleItem) Subtotal() int64 {
	return i.UnitPrice*int64(i.Quantity) + i.AddonAmount
}

// RuleActivity 是参与判定的活动快照。
type RuleActivity struct {
	// ID 活动 ID
	ID types.ID
	// Name 活动名称
	Name string
	// Type 玩法
	Type enums.ActivityType
	// Level 作用层级
	Level enums.ActivityLevel
	// Rule 玩法规则
	Rule ActivityRule
	// Priority 优先级
	Priority int
	// MutexGroup 互斥组
	MutexGroup string
	// Exclusive 是否独占
	Exclusive bool
	// Stackable 是否可叠加
	Stackable bool
	// QuotaLimit 名额上限，0 表示不限
	QuotaLimit int
	// QuotaUsed 已占用名额
	QuotaUsed int
	// ProductSPUIDs 参与商品 SPU ID 列表，商品级活动必填
	ProductSPUIDs []types.ID
}

// RuleCoupon 是参与计算的券快照。
type RuleCoupon struct {
	// ID 会员券 ID
	ID types.ID
	// Name 券名称
	Name string
	// Type 券类型
	Type enums.CouponType
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
}

// RuleInput 是规则解析的入参。
type RuleInput struct {
	// StoreID 门店 ID
	StoreID types.ID
	// Channel 下单渠道，1 小程序 / 2 POS / 3 自助机 / 4 外卖平台
	Channel int8
	// Items 明细行
	Items []RuleItem
	// Activities 候选活动
	Activities []RuleActivity
	// Coupon 使用的券，未使用传零值
	Coupon *RuleCoupon
	// PointsAmount 会员拟用积分抵扣的金额（分），0 表示不用积分
	PointsAmount int64
	// PointMaxDeduct 单笔积分抵扣上限（分），0 表示不限制
	PointMaxDeduct int64
	// PointMinOrder 使用积分抵扣的最低订单金额（分），0 表示无门槛
	PointMinOrder int64
}

// LineDiscount 是一行商品上的单品级优惠。
type LineDiscount struct {
	// SpuID 菜品 SPU ID
	SpuID types.ID
	// SkuID 菜品规格 SKU ID
	SkuID types.ID
	// Amount 优惠金额（分）
	Amount int64
	// ActivityName 来源活动名称，用于留档
	ActivityName string
}

// RuleResult 是规则解析的结果。
type RuleResult struct {
	// LineDiscounts 行级单品优惠，按明细行顺序
	LineDiscounts []LineDiscount
	// ActivityAmount 订单级活动优惠合计（分）
	ActivityAmount int64
	// ActivityNames 生效的订单级活动名称
	ActivityNames []string
	// ActivityIDs 实际产生优惠的活动 ID（商品级 + 订单级），供下单占用名额、取消时释放
	ActivityIDs []types.ID
	// CouponAmount 券抵扣金额（分）
	CouponAmount int64
	// CouponName 生效券名称
	CouponName string
	// CouponThreshold 券使用门槛（分），用于计价引擎判断门槛
	CouponThreshold int64
	// PointsAmount 积分抵扣金额（分）
	PointsAmount int64
	// Notes 解析留档，逐条说明判定过程
	Notes []string
}

// Resolve 执行一次完整的营销规则解析。
// 判定顺序严格按 07 §5：同类型互斥取优先级最高 → 排除组互斥 → 独占活动生效即终止
// → 可叠加活动累加 → 名额用完跳过 → 累加不超过商品小计。
// 券与积分在活动之后计算，券门槛按活动后金额比较，与计价引擎口径一致。
func Resolve(in RuleInput) *RuleResult {
	result := &RuleResult{
		LineDiscounts: make([]LineDiscount, len(in.Items)),
		ActivityNames: make([]string, 0),
		ActivityIDs:   make([]types.ID, 0),
		Notes:         make([]string, 0),
	}

	goodsAmount := int64(0)
	for _, item := range in.Items {
		goodsAmount += item.Subtotal()
	}

	// ---------- 活动判定 ----------
	selected, notes := selectActivities(in, goodsAmount)
	result.Notes = append(result.Notes, notes...)

	// 逐个应用选中的活动
	for _, activity := range selected {
		if activity.Level == enums.ActivityLevelProduct {
			// 商品级活动落到行级优惠
			discounts, discount, note := applyProductActivity(activity, in.Items, result.LineDiscounts)
			for i := range discounts {
				result.LineDiscounts[i] = discounts[i]
			}
			if discount > 0 {
				result.ActivityIDs = append(result.ActivityIDs, activity.ID)
				result.Notes = append(result.Notes, note)
			}
			continue
		}

		// 订单级活动累加到 ActivityAmount
		discount, note := applyOrderActivity(activity, in.Items, goodsAmount)
		if discount > 0 {
			result.ActivityAmount += discount
			result.ActivityNames = append(result.ActivityNames, activity.Name)
			result.ActivityIDs = append(result.ActivityIDs, activity.ID)
			result.Notes = append(result.Notes, note)
		}
	}

	// 活动优惠总额不得超过商品小计
	if result.ActivityAmount > goodsAmount {
		result.ActivityAmount = goodsAmount
	}

	// ---------- 优惠券 ----------
	if in.Coupon != nil {
		amount, threshold, note := resolveCoupon(in.Coupon, in, goodsAmount, result.ActivityAmount)
		result.CouponAmount = amount
		result.CouponThreshold = threshold
		if amount > 0 {
			result.CouponName = in.Coupon.Name
		}
		result.Notes = append(result.Notes, note)
	}

	// ---------- 积分抵扣 ----------
	points, note := resolvePoints(in, goodsAmount, result.ActivityAmount, result.CouponAmount)
	result.PointsAmount = points
	result.Notes = append(result.Notes, note)

	return result
}

// selectActivities 按 07 §5 的判定顺序筛选生效活动，返回选中的活动与判定留档。
func selectActivities(in RuleInput, goodsAmount int64) ([]RuleActivity, []string) {
	notes := make([]string, 0)

	// 排序：优先级高的先判定，同优先级按 ID 升序保证结果稳定
	candidates := make([]RuleActivity, 0, len(in.Activities))
	for _, activity := range in.Activities {
		if activity.QuotaLimit > 0 && activity.QuotaUsed >= activity.QuotaLimit {
			notes = append(notes, "活动["+activity.Name+"]名额已用完，跳过")
			continue
		}
		candidates = append(candidates, activity)
	}
	sortActivities(candidates)

	selected := make([]RuleActivity, 0, len(candidates))
	usedTypes := make(map[enums.ActivityType]bool)
	usedMutexGroups := make(map[string]bool)

	for _, activity := range candidates {
		// 规则 1：同类型互斥，只取优先级最高（排序后第一个即最高优先级）
		if usedTypes[activity.Type] {
			notes = append(notes, "活动["+activity.Name+"]与同类型高优先级活动互斥，跳过")
			continue
		}

		// 规则 2：排除组互斥，同组只生效一个
		if activity.MutexGroup != "" && usedMutexGroups[activity.MutexGroup] {
			notes = append(notes, "活动["+activity.Name+"]所在排除组["+activity.MutexGroup+"]已有活动生效，跳过")
			continue
		}

		// 判定该活动在本单是否真的产生优惠
		discount, _ := activityDiscount(activity, in.Items, goodsAmount)
		if discount <= 0 {
			notes = append(notes, "活动["+activity.Name+"]本单未达条件，不生效")
			continue
		}

		selected = append(selected, activity)
		usedTypes[activity.Type] = true
		if activity.MutexGroup != "" {
			usedMutexGroups[activity.MutexGroup] = true
		}

		// 规则 3：不可叠加的活动独占，生效即终止后续判定
		if activity.Exclusive {
			notes = append(notes, "活动["+activity.Name+"]为独占活动，其余活动终止判定")
			break
		}
	}

	return selected, notes
}

// applyOrderActivity 计算订单级活动的优惠金额。
func applyOrderActivity(activity RuleActivity, items []RuleItem, goodsAmount int64) (int64, string) {
	discount, _ := activityDiscount(activity, items, goodsAmount)

	return discount, "活动[" + activity.Name + "]优惠 " + formatAmount(discount) + " 元"
}

// applyProductActivity 计算商品级活动的优惠，按行落优惠。
// 返回：逐行优惠、优惠合计、留档说明。
func applyProductActivity(activity RuleActivity, items []RuleItem, current []LineDiscount) ([]LineDiscount, int64, string) {
	discounts := make([]LineDiscount, len(items))
	copy(discounts, current)

	total := int64(0)
	for index, item := range items {
		if !activityCoversSpu(activity, item.SpuID) {
			continue
		}
		amount := activityLineDiscount(activity, item)
		if amount <= 0 {
			continue
		}
		// 同一行已有其他行级优惠时取较大者，避免叠加超出行小计
		if amount > discounts[index].Amount {
			discounts[index] = LineDiscount{
				SpuID:        item.SpuID,
				SkuID:        item.SkuID,
				Amount:       amount,
				ActivityName: activity.Name,
			}
			total += amount
		}
	}

	return discounts, total, "商品级活动[" + activity.Name + "]优惠 " + formatAmount(total) + " 元"
}

// activityDiscount 计算活动在本单能产生的优惠总额。
func activityDiscount(activity RuleActivity, items []RuleItem, goodsAmount int64) (int64, string) {
	switch activity.Type {
	case enums.ActivityTypeFullReduction:
		return fullReductionDiscount(activity, goodsAmount), ""
	case enums.ActivityTypeTimeDiscount:
		return timeDiscountAmount(activity, goodsAmount), ""
	case enums.ActivityTypeSecondHalfPrice:
		return secondHalfPriceDiscount(activity, items), ""
	case enums.ActivityTypeCombo:
		return comboDiscount(activity, items), ""
	case enums.ActivityTypeAddonPurchase:
		return addonPurchaseDiscount(activity, items, goodsAmount), ""
	default:
		return 0, ""
	}
}

// fullReductionDiscount 计算满减 / 阶梯满减优惠。
// 有阶梯档位时取满足门槛的最高一档，没有档位时用单门槛单减免。
func fullReductionDiscount(activity RuleActivity, goodsAmount int64) int64 {
	rule := activity.Rule
	if len(rule.Tiers) > 0 {
		discount := int64(0)
		for _, tier := range rule.Tiers {
			if goodsAmount >= tier.ThresholdAmount && tier.DiscountAmount > discount {
				discount = tier.DiscountAmount
			}
		}

		return discount
	}

	if goodsAmount < rule.ThresholdAmount {
		return 0
	}

	return rule.DiscountAmount
}

// timeDiscountAmount 计算限时折扣优惠：按折扣率打折，可封顶。
func timeDiscountAmount(activity RuleActivity, goodsAmount int64) int64 {
	rule := activity.Rule
	if rule.DiscountRate <= 0 || rule.DiscountRate >= 10000 || goodsAmount <= 0 {
		return 0
	}

	// 折后应付 = 原价 × 折扣率 / 10000，优惠 = 原价 − 折后应付
	discount := goodsAmount - goodsAmount*int64(rule.DiscountRate)/10000
	if rule.MaxDiscountAmount > 0 && discount > rule.MaxDiscountAmount {
		discount = rule.MaxDiscountAmount
	}
	if discount < 0 {
		discount = 0
	}

	return discount
}

// secondHalfPriceDiscount 计算第二份半价优惠：同款第 2 份按折扣率计价。
func secondHalfPriceDiscount(activity RuleActivity, items []RuleItem) int64 {
	rate := activity.Rule.SecondHalfRate
	if rate <= 0 {
		rate = 5000
	}

	// 按 SPU 聚合数量：同一款商品的数量决定能省几份
	quantityBySpu := make(map[types.ID]int, len(items))
	unitPriceBySpu := make(map[types.ID]int64, len(items))
	for _, item := range items {
		quantityBySpu[item.SpuID] += item.Quantity
		if _, ok := unitPriceBySpu[item.SpuID]; !ok {
			unitPriceBySpu[item.SpuID] = item.UnitPrice
		}
	}

	total := int64(0)
	for spuID, quantity := range quantityBySpu {
		if quantity < 2 {
			continue
		}
		// 每 2 份可省 1 份的差价
		groups := quantity / 2
		unitPrice := unitPriceBySpu[spuID]
		savePerGroup := unitPrice - unitPrice*int64(rate)/10000
		if savePerGroup < 0 {
			savePerGroup = 0
		}
		total += int64(groups) * savePerGroup
	}

	return total
}

// comboDiscount 计算套餐特惠优惠：套餐组合的实际售价低于单品之和的差额。
func comboDiscount(activity RuleActivity, items []RuleItem) int64 {
	rule := activity.Rule
	if len(rule.ComboItems) == 0 || rule.ComboPrice <= 0 {
		return 0
	}

	// 套餐要求的商品必须全部在购物篮里，且数量足够
	required := make(map[types.ID]int, len(rule.ComboItems))
	for _, combo := range rule.ComboItems {
		required[combo.SpuID] += combo.Quantity
	}

	available := make(map[types.ID]int, len(items))
	for _, item := range items {
		available[item.SpuID] += item.Quantity
	}
	for spuID, need := range required {
		if available[spuID] < need {
			return 0
		}
	}

	// 单品之和按套餐内商品的实际成交单价算
	unitPriceBySpu := make(map[types.ID]int64, len(items))
	for _, item := range items {
		if _, ok := unitPriceBySpu[item.SpuID]; !ok {
			unitPriceBySpu[item.SpuID] = item.UnitPrice
		}
	}

	singleSum := int64(0)
	for spuID, need := range required {
		singleSum += unitPriceBySpu[spuID] * int64(need)
	}

	discount := singleSum - rule.ComboPrice
	if discount <= 0 {
		return 0
	}

	return discount
}

// addonPurchaseDiscount 计算加价购优惠：满额后指定商品按加价购价抵扣差额。
func addonPurchaseDiscount(activity RuleActivity, items []RuleItem, goodsAmount int64) int64 {
	rule := activity.Rule
	if rule.AddonSpuID == 0 || rule.AddonPrice <= 0 {
		return 0
	}
	if goodsAmount < rule.ThresholdAmount {
		return 0
	}

	// 购物篮里必须有加价购商品，且加价购价低于原价才产生优惠
	for _, item := range items {
		if item.SpuID != rule.AddonSpuID {
			continue
		}
		discount := (item.UnitPrice - rule.AddonPrice) * int64(item.Quantity)
		if discount > 0 {
			return discount
		}
	}

	return 0
}

// activityLineDiscount 计算商品级活动在单行上的优惠金额。
func activityLineDiscount(activity RuleActivity, item RuleItem) int64 {
	switch activity.Type {
	case enums.ActivityTypeTimeDiscount:
		rate := activity.Rule.DiscountRate
		if rate <= 0 || rate >= 10000 {
			return 0
		}
		discount := item.Subtotal() - item.Subtotal()*int64(rate)/10000
		if activity.Rule.MaxDiscountAmount > 0 && discount > activity.Rule.MaxDiscountAmount {
			discount = activity.Rule.MaxDiscountAmount
		}
		if discount < 0 {
			discount = 0
		}

		return discount
	case enums.ActivityTypeSecondHalfPrice:
		rate := activity.Rule.SecondHalfRate
		if rate <= 0 {
			rate = 5000
		}
		// 单行内每 2 份省 1 份差价
		groups := item.Quantity / 2
		if groups <= 0 {
			return 0
		}
		savePerGroup := item.UnitPrice - item.UnitPrice*int64(rate)/10000
		if savePerGroup < 0 {
			savePerGroup = 0
		}

		return int64(groups) * savePerGroup
	case enums.ActivityTypeFullReduction:
		// 满减按行小计判断门槛
		if item.Subtotal() < activity.Rule.ThresholdAmount {
			return 0
		}

		return activity.Rule.DiscountAmount
	default:
		return 0
	}
}

// activityCoversSpu 判断活动是否覆盖该商品：未配置参与商品时视为覆盖全部。
func activityCoversSpu(activity RuleActivity, spuID types.ID) bool {
	if len(activity.ProductSPUIDs) == 0 {
		return true
	}
	for _, id := range activity.ProductSPUIDs {
		if id == spuID {
			return true
		}
	}

	return false
}

// resolveCoupon 校验并计算券优惠。
// 门槛按「活动后金额」比较，与计价引擎的券步骤口径一致。
func resolveCoupon(coupon *RuleCoupon, in RuleInput, goodsAmount, activityAmount int64) (int64, int64, string) {
	afterActivity := goodsAmount - activityAmount

	// 适用范围：指定商品时购物篮必须命中该商品
	if coupon.Scope == enums.CouponScopeProduct {
		if !couponHitsProducts(coupon, in.Items) {
			return 0, coupon.ThresholdAmount, "优惠券[" + coupon.Name + "]不适用于当前商品，未生效"
		}
	}

	// 门槛校验：门槛按活动后金额比较
	if coupon.ThresholdAmount > 0 && afterActivity < coupon.ThresholdAmount {
		return 0, coupon.ThresholdAmount, "优惠券[" + coupon.Name + "]未达使用门槛，未生效"
	}

	var amount int64
	switch coupon.Type {
	case enums.CouponTypeDiscount:
		rate := coupon.DiscountRate
		if rate <= 0 || rate >= 10000 {
			return 0, coupon.ThresholdAmount, "优惠券[" + coupon.Name + "]折扣率不合法，未生效"
		}
		amount = afterActivity - afterActivity*int64(rate)/10000
		if coupon.MaxDiscountAmount > 0 && amount > coupon.MaxDiscountAmount {
			amount = coupon.MaxDiscountAmount
		}
	default:
		amount = coupon.DiscountAmount
	}

	if amount <= 0 {
		return 0, coupon.ThresholdAmount, "优惠券[" + coupon.Name + "]抵扣金额为 0，未生效"
	}

	// 面额最多抵到剩余应付
	if amount > afterActivity {
		amount = afterActivity
	}

	return amount, coupon.ThresholdAmount, "优惠券[" + coupon.Name + "]抵扣 " + formatAmount(amount) + " 元"
}

// couponHitsProducts 判断购物篮是否命中券指定的商品。
func couponHitsProducts(coupon *RuleCoupon, items []RuleItem) bool {
	for _, item := range items {
		for _, spuID := range coupon.ScopeIDs {
			if spuID == item.SpuID {
				return true
			}
		}
	}

	return false
}

// resolvePoints 计算积分抵扣金额：受单笔上限与剩余应付双重约束。
func resolvePoints(in RuleInput, goodsAmount, activityAmount, couponAmount int64) (int64, string) {
	if in.PointsAmount <= 0 {
		return 0, "未使用积分"
	}
	if in.PointMinOrder > 0 && goodsAmount < in.PointMinOrder {
		return 0, "未达积分抵扣门槛，未生效"
	}

	points := in.PointsAmount
	// 单笔上限约束
	if in.PointMaxDeduct > 0 && points > in.PointMaxDeduct {
		points = in.PointMaxDeduct
	}
	// 剩余应付约束
	remaining := goodsAmount - activityAmount - couponAmount
	if points > remaining {
		points = remaining
	}
	if points < 0 {
		points = 0
	}

	return points, "积分抵扣 " + formatAmount(points) + " 元"
}

// sortActivities 按优先级降序、ID 升序排序，保证相同输入产出相同结果。
func sortActivities(activities []RuleActivity) {
	for i := 1; i < len(activities); i++ {
		for j := i; j > 0; j-- {
			less := activities[j].Priority > activities[j-1].Priority ||
				(activities[j].Priority == activities[j-1].Priority && activities[j].ID < activities[j-1].ID)
			if !less {
				break
			}
			activities[j], activities[j-1] = activities[j-1], activities[j]
		}
	}
}

// ParseActivityRule 解析活动的规则 JSON；解析失败或为空时返回零值规则。
func ParseActivityRule(raw string) ActivityRule {
	rule := ActivityRule{}
	if strings.TrimSpace(raw) == "" {
		return rule
	}
	if err := json.Unmarshal([]byte(raw), &rule); err != nil {
		return ActivityRule{}
	}

	return rule
}

// formatAmount 把分转成「x.xx 元」用于留档说明。
func formatAmount(amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	return sign + strconv.FormatInt(amount/100, 10) + "." + fmt.Sprintf("%02d", amount%100)
}
