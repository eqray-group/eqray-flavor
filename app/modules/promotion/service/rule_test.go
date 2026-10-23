// rule_test.go 覆盖营销规则解析的判定顺序与金额计算，对应测试设计 PR-01 ~ PR-08。
package service

import (
	"testing"

	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/promotion/enums"
)

// 断言金额相等，失败时输出期望值与实际值。
func assertAmount(t *testing.T, name string, got, want int64) {
	t.Helper()
	if got != want {
		t.Fatalf("%s：期望 %d，实际 %d", name, want, got)
	}
}

// 构造一行普通商品：单价 2000 分（20 元），数量 1，无加料。
func normalItem(spuID types.ID, quantity int) RuleItem {
	return RuleItem{SpuID: spuID, SkuID: spuID * 10, UnitPrice: 2000, Quantity: quantity}
}

// TestFullReductionHit 满减达门槛时生效，优惠金额取减免额。
func TestFullReductionHit(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)}, // 小计 6000 分（60 元），达 50 元门槛
		Activities: []RuleActivity{{
			ID: 1, Name: "满50减10", Type: enums.ActivityTypeFullReduction,
			Rule: ActivityRule{ThresholdAmount: 5000, DiscountAmount: 1000},
		}},
	}

	result := Resolve(in)
	assertAmount(t, "满减优惠", result.ActivityAmount, 1000)
}

// TestFullReductionNotHit 未达门槛时不生效。
func TestFullReductionNotHit(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 1)},
		Activities: []RuleActivity{{
			ID: 1, Name: "满50减10", Type: enums.ActivityTypeFullReduction,
			Rule: ActivityRule{ThresholdAmount: 5000, DiscountAmount: 1000},
		}},
	}

	result := Resolve(in)
	assertAmount(t, "未达门槛优惠", result.ActivityAmount, 0)
}

// TestFullReductionTiers 阶梯满减取满足门槛的最高一档。
func TestFullReductionTiers(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)}, // 小计 6000 分
		Activities: []RuleActivity{{
			ID: 1, Name: "阶梯满减", Type: enums.ActivityTypeFullReduction,
			Rule: ActivityRule{Tiers: []ActivityTier{
				{ThresholdAmount: 3000, DiscountAmount: 500},
				{ThresholdAmount: 5000, DiscountAmount: 800},
				{ThresholdAmount: 10000, DiscountAmount: 1500},
			}},
		}},
	}

	result := Resolve(in)
	assertAmount(t, "阶梯满减优惠", result.ActivityAmount, 800)
}

// TestSameTypeMutex 同类型活动只取优先级最高的一个。
func TestSameTypeMutex(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)}, // 小计 6000
		Activities: []RuleActivity{
			{ID: 1, Name: "低优先级满减", Type: enums.ActivityTypeFullReduction, Priority: 1,
				Rule: ActivityRule{ThresholdAmount: 1000, DiscountAmount: 300}},
			{ID: 2, Name: "高优先级满减", Type: enums.ActivityTypeFullReduction, Priority: 10,
				Rule: ActivityRule{ThresholdAmount: 1000, DiscountAmount: 600}},
		},
	}

	result := Resolve(in)
	assertAmount(t, "同类型互斥后优惠", result.ActivityAmount, 600)
}

// TestMutexGroup 排除组内只生效一个。
func TestMutexGroup(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)},
		Activities: []RuleActivity{
			{ID: 1, Name: "A组活动", Type: enums.ActivityTypeFullReduction, Priority: 5, MutexGroup: "A",
				Rule: ActivityRule{ThresholdAmount: 1000, DiscountAmount: 400}},
			{ID: 2, Name: "B组活动", Type: enums.ActivityTypeTimeDiscount, Priority: 9, MutexGroup: "A",
				Rule: ActivityRule{DiscountRate: 9000}},
		},
	}

	result := Resolve(in)
	// 高优先级的是 9 折：6000 - 6000*9000/10000 = 600
	assertAmount(t, "排除组互斥后优惠", result.ActivityAmount, 600)
}

// TestExclusiveActivity 独占活动生效后其他活动终止。
func TestExclusiveActivity(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)},
		Activities: []RuleActivity{
			{ID: 1, Name: "独占活动", Type: enums.ActivityTypeFullReduction, Priority: 100, Exclusive: true,
				Rule: ActivityRule{ThresholdAmount: 1000, DiscountAmount: 500}},
			{ID: 2, Name: "另一活动", Type: enums.ActivityTypeTimeDiscount, Priority: 1,
				Rule: ActivityRule{DiscountRate: 8000}},
		},
	}

	result := Resolve(in)
	assertAmount(t, "独占后优惠", result.ActivityAmount, 500)
}

// TestStackableActivities 不同类型的可叠加活动优惠累加。
func TestStackableActivities(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)}, // 小计 6000
		Activities: []RuleActivity{
			{ID: 1, Name: "满减", Type: enums.ActivityTypeFullReduction, Priority: 10, Stackable: true,
				Rule: ActivityRule{ThresholdAmount: 1000, DiscountAmount: 500}},
			{ID: 2, Name: "限时折扣", Type: enums.ActivityTypeTimeDiscount, Priority: 5, Stackable: true,
				Rule: ActivityRule{DiscountRate: 9000}},
		},
	}

	result := Resolve(in)
	// 500 + (6000 - 5400) = 1100
	assertAmount(t, "叠加优惠", result.ActivityAmount, 1100)
}

// TestActivityQuotaExhausted 名额用完的活动被跳过。
func TestActivityQuotaExhausted(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)},
		Activities: []RuleActivity{
			{ID: 1, Name: "已满名额", Type: enums.ActivityTypeFullReduction, Priority: 10,
				QuotaLimit: 10, QuotaUsed: 10,
				Rule: ActivityRule{ThresholdAmount: 1000, DiscountAmount: 500}},
			{ID: 2, Name: "有名额", Type: enums.ActivityTypeTimeDiscount, Priority: 1,
				QuotaLimit: 10, QuotaUsed: 1,
				Rule: ActivityRule{DiscountRate: 9000}},
		},
	}

	result := Resolve(in)
	assertAmount(t, "名额用完后的优惠", result.ActivityAmount, 600)
}

// TestActivityNotExceedSubtotal 累加优惠不超过商品小计。
func TestActivityNotExceedSubtotal(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 1)}, // 小计 2000
		Activities: []RuleActivity{
			{ID: 1, Name: "满减一", Type: enums.ActivityTypeFullReduction, Priority: 10, Stackable: true,
				Rule: ActivityRule{ThresholdAmount: 100, DiscountAmount: 1500}},
			{ID: 2, Name: "满减二", Type: enums.ActivityTypeFullReduction, Priority: 5, Stackable: true,
				Rule: ActivityRule{ThresholdAmount: 100, DiscountAmount: 1500}},
		},
	}

	result := Resolve(in)
	// 同类型互斥，只生效一个 1500，但不超过小计 2000
	assertAmount(t, "优惠封顶", result.ActivityAmount, 1500)
}

// TestCouponThresholdUsesAmountAfterActivity 券门槛按活动后金额比较。
func TestCouponThresholdUsesAmountAfterActivity(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)}, // 小计 6000
		Activities: []RuleActivity{
			{ID: 1, Name: "满减", Type: enums.ActivityTypeFullReduction, Priority: 10,
				Rule: ActivityRule{ThresholdAmount: 1000, DiscountAmount: 2000}},
		},
		Coupon: &RuleCoupon{
			ID: 1, Name: "满50减10", Type: enums.CouponTypeFullReduction,
			ThresholdAmount: 5000, DiscountAmount: 1000,
		},
	}

	result := Resolve(in)
	// 活动后金额 4000 < 门槛 5000，券不生效
	assertAmount(t, "券抵扣", result.CouponAmount, 0)
	assertAmount(t, "活动优惠", result.ActivityAmount, 2000)
}

// TestCouponApplied 券生效时抵扣面额。
func TestCouponApplied(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 3)}, // 小计 6000
		Coupon: &RuleCoupon{
			ID: 1, Name: "减10元", Type: enums.CouponTypeFullReduction,
			ThresholdAmount: 5000, DiscountAmount: 1000,
		},
	}

	result := Resolve(in)
	assertAmount(t, "券抵扣", result.CouponAmount, 1000)
}

// TestCouponCappedByRemain 面额超过剩余应付时只抵到剩余应付。
func TestCouponCappedByRemain(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 1)}, // 小计 2000
		Coupon: &RuleCoupon{
			ID: 1, Name: "减50元", Type: enums.CouponTypeCash, DiscountAmount: 5000,
		},
	}

	result := Resolve(in)
	assertAmount(t, "券抵扣封顶", result.CouponAmount, 2000)
}

// TestCouponDiscountRate 折扣券按折扣率计算并受封顶约束。
func TestCouponDiscountRate(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 1)}, // 小计 2000
		Coupon: &RuleCoupon{
			ID: 1, Name: "八折券", Type: enums.CouponTypeDiscount,
			DiscountRate: 8000, MaxDiscountAmount: 300,
		},
	}

	result := Resolve(in)
	// 2000 - 1600 = 400，受封顶 300 约束
	assertAmount(t, "折扣券抵扣", result.CouponAmount, 300)
}

// TestCouponScopeProductNotMatched 指定商品的券不命中购物篮时不生效。
func TestCouponScopeProductNotMatched(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 1)},
		Coupon: &RuleCoupon{
			ID: 1, Name: "单品券", Type: enums.CouponTypeProduct,
			DiscountAmount: 500, Scope: enums.CouponScopeProduct, ScopeIDs: []types.ID{999},
		},
	}

	result := Resolve(in)
	assertAmount(t, "范围不匹配券抵扣", result.CouponAmount, 0)
}

// TestPointsCappedByMaxDeduct 积分抵扣受单笔上限约束。
func TestPointsCappedByMaxDeduct(t *testing.T) {
	in := RuleInput{
		Items:           []RuleItem{normalItem(1, 10)}, // 小计 20000
		PointsAmount:    10000,
		PointMaxDeduct:  3000,
	}

	result := Resolve(in)
	assertAmount(t, "积分抵扣", result.PointsAmount, 3000)
}

// TestPointsCappedByRemain 积分抵扣不超过剩余应付。
func TestPointsCappedByRemain(t *testing.T) {
	in := RuleInput{
		Items:        []RuleItem{normalItem(1, 1)}, // 小计 2000
		PointsAmount: 10000,
	}

	result := Resolve(in)
	assertAmount(t, "积分抵扣封顶", result.PointsAmount, 2000)
}

// TestSecondHalfPrice 同款第二份半价按每两份省一份差价。
func TestSecondHalfPrice(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 2)}, // 两份，单价 2000
		Activities: []RuleActivity{
			{ID: 1, Name: "第二份半价", Type: enums.ActivityTypeSecondHalfPrice, Priority: 10,
				Rule: ActivityRule{SecondHalfRate: 5000}},
		},
	}

	result := Resolve(in)
	// 省一份差价：2000 - 1000 = 1000
	assertAmount(t, "第二份半价优惠", result.ActivityAmount, 1000)
}

// TestProductLevelActivityLandedOnLine 商品级活动优惠落到行级。
func TestProductLevelActivityLandedOnLine(t *testing.T) {
	in := RuleInput{
		Items: []RuleItem{normalItem(1, 1), normalItem(2, 1)},
		Activities: []RuleActivity{
			{ID: 1, Name: "单品限时折扣", Type: enums.ActivityTypeTimeDiscount, Level: enums.ActivityLevelProduct, Priority: 10,
				Rule:        ActivityRule{DiscountRate: 5000},
				ProductSPUIDs: []types.ID{1}},
		},
	}

	result := Resolve(in)
	assertAmount(t, "行级优惠", result.LineDiscounts[0].Amount, 1000)
	assertAmount(t, "未参与行优惠", result.LineDiscounts[1].Amount, 0)
	assertAmount(t, "订单级活动额", result.ActivityAmount, 0)
}
