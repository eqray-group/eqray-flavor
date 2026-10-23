// pricing_test.go 用测试设计的 AM 红线用例锁定计价口径。
// 对应 docs/design/13-测试设计.md §5 优惠与金额。
package service

import (
	"testing"

	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/order/enums"
)

// lineOf 构造一行测试明细。
func lineOf(skuID types.ID, name string, price int64, qty int, addons ...PricingAddonInput) PricingLineInput {
	return PricingLineInput{
		SpuID: skuID, SkuID: skuID,
		SpuName: name, SkuName: "默认规格",
		OriginalPrice: price, Quantity: qty, Addons: addons,
	}
}

// TestCalculatePricingAM01 复现红线用例：原价 108 + 满 100 减 15 + 券满 50 减 10 = 应付 83，顺序正确。
func TestCalculatePricingAM01(t *testing.T) {
	result := CalculatePricing(
		[]PricingLineInput{lineOf(1, "招牌菜", 10_800, 1)},
		PricingRuleInput{ActivityAmount: 1_500, ActivityName: "满100减15", CouponAmount: 1_000, CouponName: "满50减10", CouponThreshold: 5_000},
	)

	if result.PayAmount != 8_300 {
		t.Fatalf("应付 %d 分，期望 8300 分", result.PayAmount)
	}
	if result.ActivityAmount != 1_500 || result.CouponAmount != 1_000 {
		t.Fatalf("活动 %d / 券 %d，期望 1500 / 1000", result.ActivityAmount, result.CouponAmount)
	}
	// 顺序：先活动后券；券门槛按活动后金额 9300 比较，9300 ≥ 5000 通过
	if result.Steps[1].Step != 3 || result.Steps[2].Step != 5 {
		t.Fatalf("计价步骤顺序错误：%v", result.Steps)
	}
	line := result.Lines[0]
	if line.Payable != 8_300 {
		t.Fatalf("行应付 %d 分，期望 8300 分", line.Payable)
	}
}

// TestCalculatePricingAM03 复现红线用例：券门槛边界 99.99 对满 100 不可用。
func TestCalculatePricingAM03(t *testing.T) {
	result := CalculatePricing(
		[]PricingLineInput{lineOf(1, "小食", 9_999, 1)},
		PricingRuleInput{CouponAmount: 1_000, CouponThreshold: 10_000},
	)

	if result.CouponAmount != 0 {
		t.Fatalf("未达门槛时券不应生效，实际抵扣 %d 分", result.CouponAmount)
	}
	if result.PayAmount != 9_999 {
		t.Fatalf("应付 %d 分，期望 9999 分", result.PayAmount)
	}
}

// TestCalculatePricingAM04 复现红线用例：积分抵扣超过上限时只抵到剩余应付。
func TestCalculatePricingAM04(t *testing.T) {
	result := CalculatePricing(
		[]PricingLineInput{lineOf(1, "饮品", 2_000, 1)},
		PricingRuleInput{PointsAmount: 99_999},
	)

	if result.PointsAmount != 2_000 {
		t.Fatalf("积分抵扣 %d 分，期望只抵到应付 2000 分", result.PointsAmount)
	}
	if result.PayAmount != 0 {
		t.Fatalf("应付 %d 分，期望 0 分", result.PayAmount)
	}
}

// TestCalculatePricingAM06 验证全程整型分且明细合计等于订单合计：多行分摊后行应付之和 = 订单应付。
func TestCalculatePricingAM06(t *testing.T) {
	result := CalculatePricing(
		[]PricingLineInput{
			lineOf(1, "菜品A", 3_333, 3), // 9999
			lineOf(2, "菜品B", 2_222, 2), // 4444
			lineOf(3, "菜品C", 1_111, 1), // 1111
		},
		PricingRuleInput{ActivityAmount: 1_234, CouponAmount: 567, PointsAmount: 89},
	)

	paySum := int64(0)
	discountSum := int64(0)
	subtotalSum := int64(0)
	for _, line := range result.Lines {
		paySum += line.Payable
		discountSum += line.OrderDiscount
		subtotalSum += line.Subtotal
	}

	goods := int64(9_999 + 4_444 + 1_111)
	if result.ItemAmount != goods {
		t.Fatalf("商品小计 %d，期望 %d", result.ItemAmount, goods)
	}
	if result.PayAmount != goods-1_234-567-89 {
		t.Fatalf("应付 %d，期望 %d", result.PayAmount, goods-1_234-567-89)
	}
	if paySum != result.PayAmount {
		t.Fatalf("行应付合计 %d ≠ 订单应付 %d，分摊未消尾差", paySum, result.PayAmount)
	}
	if discountSum != 1_234+567+89 {
		t.Fatalf("行分摊优惠合计 %d ≠ 整单优惠 1890", discountSum)
	}
	if subtotalSum != goods {
		t.Fatalf("行小计合计 %d ≠ 商品金额 %d", subtotalSum, goods)
	}
}

// TestCalculatePricingMemberPrice 验证步骤 1：会员且配置会员价时用会员价，且只对配置了的行生效。
func TestCalculatePricingMemberPrice(t *testing.T) {
	normal := lineOf(1, "无会员价", 2_000, 1)
	member := lineOf(2, "有会员价", 3_000, 1)
	member.MemberPrice = 2_500

	result := CalculatePricing([]PricingLineInput{normal, member}, PricingRuleInput{IsMember: true})

	if result.Lines[0].UnitPrice != 2_000 {
		t.Fatalf("未配置会员价的行应使用原价，实际 %d", result.Lines[0].UnitPrice)
	}
	if result.Lines[1].UnitPrice != 2_500 {
		t.Fatalf("配置会员价的行应使用会员价，实际 %d", result.Lines[1].UnitPrice)
	}
	if result.PayAmount != 4_500 {
		t.Fatalf("应付 %d，期望 4500", result.PayAmount)
	}
}

// TestCalculatePricingAddon 验证加料费计入小计并参与整单优惠分摊。
func TestCalculatePricingAddon(t *testing.T) {
	result := CalculatePricing(
		[]PricingLineInput{lineOf(1, "面", 1_000, 1, PricingAddonInput{AddonItemID: 11, Name: "加蛋", Price: 200, Quantity: 2})},
		PricingRuleInput{CouponAmount: 300},
	)

	if result.AddonAmount != 400 {
		t.Fatalf("加料费 %d，期望 400", result.AddonAmount)
	}
	if result.PayAmount != 1_100 {
		t.Fatalf("应付 %d，期望 1100", result.PayAmount)
	}
	if result.Lines[0].Payable != 1_100 {
		t.Fatalf("行应付 %d，期望 1100", result.Lines[0].Payable)
	}
}

// TestCalculatePricingSingleLineDiscount 验证单品级优惠在整单优惠之前生效且不会把小计减成负数。
func TestCalculatePricingSingleLineDiscount(t *testing.T) {
	line := lineOf(1, "小菜", 500, 1)
	line.LineDiscount = 800

	result := CalculatePricing([]PricingLineInput{line}, PricingRuleInput{})

	if result.Lines[0].Subtotal != 0 {
		t.Fatalf("单品级优惠超过行金额时应只抵到 0，实际小计 %d", result.Lines[0].Subtotal)
	}
	if result.PayAmount != 0 {
		t.Fatalf("应付 %d，期望 0", result.PayAmount)
	}
}

// TestCalculatePricingCouponCapped 验证券面额超过应付时只抵到剩余应付。
func TestCalculatePricingCouponCapped(t *testing.T) {
	result := CalculatePricing(
		[]PricingLineInput{lineOf(1, "茶", 800, 1)},
		PricingRuleInput{CouponAmount: 5_000, CouponThreshold: 500},
	)

	if result.CouponAmount != 800 {
		t.Fatalf("券抵扣 %d，期望 800", result.CouponAmount)
	}
	if result.PayAmount != 0 {
		t.Fatalf("应付 %d，期望 0", result.PayAmount)
	}
}

// TestOrderStatusTransitions 锁定状态机合法流转，与测试设计 SM-00 ~ SM-07 对应。
func TestOrderStatusTransitions(t *testing.T) {
	cases := []struct {
		from, to int16
		ok       bool
	}{
		{100, 200, true},  // 外卖/自提支付成功
		{100, 300, true},  // 堂食支付即制作
		{100, 600, true},  // 超时/用户取消
		{200, 300, true},  // 接单
		{200, 700, true},  // 拒单全额退
		{300, 400, true},  // 出餐待取
		{300, 500, true},  // 堂食结清
		{300, 700, true},  // 未出餐取消
		{400, 500, true},  // 取餐完成
		{500, 700, true},  // 售后退款
		{700, 800, true},  // 退款成功
		{700, 500, true},  // 退款驳回
		{100, 500, false}, // 未支付直接完成
		{600, 100, false}, // 取消后重新支付
		{800, 300, false}, // 已退款又制作
		{600, 700, false}, // 已取消不能退款
		{800, 700, false}, // 已退款是终态
	}
	for _, c := range cases {
		if got := enums.OrderStatus(c.from).CanTransit(enums.OrderStatus(c.to)); got != c.ok {
			t.Fatalf("流转 %d → %d 合法性应为 %v", c.from, c.to, c.ok)
		}
	}
}
