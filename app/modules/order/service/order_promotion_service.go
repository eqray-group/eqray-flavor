// order_promotion_service.go 是订单与营销模块之间的桥接层。
// 职责：把已装配好的明细行交给营销模块解析活动 / 券 / 积分，再把结果写回计价引擎的入参。
// 试算与下单共用本文件，保证「试算价 = 下单价」——计价口径只有一处。
package service

import (
	"context"

	"eqray-flavor/app/modules/base/types"
	promotionService "eqray-flavor/app/modules/promotion/service"
)

// marketingContext 是一次营销解析的结果：计价引擎入参 + 营销侧明细。
type marketingContext struct {
	// Rule 计价引擎的营销规则入参
	Rule PricingRuleInput
	// Output 营销模块的原始结果，用于下单时锁券与占用名额
	Output *promotionService.PricingOutput
}

// buildPricingRule 解析营销规则并生成计价引擎的入参。
// 入参：门店、会员、下单渠道、使用的券、是否用积分、已装配的明细行；返回计价规则与营销明细。
// 说明：单品级优惠（商品级活动）回写到明细行的 LineDiscount，其余优惠进整单规则。
func (s *OrderService) buildPricingRule(
	ctx context.Context,
	storeID, memberID types.ID,
	channel int8,
	couponID types.ID,
	usePoints bool,
	lines []PricingLineInput,
) (*marketingContext, error) {
	// 明细行转营销模块的输入结构：成交单价与加料费与计价引擎保持同一口径
	items := make([]promotionService.RuleItem, 0, len(lines))
	for _, line := range lines {
		addonAmount := int64(0)
		for _, addon := range line.Addons {
			if addon.Price > 0 && addon.Quantity > 0 {
				addonAmount += addon.Price * int64(addon.Quantity)
			}
		}
		items = append(items, promotionService.RuleItem{
			SpuID:       line.SpuID,
			SkuID:       line.SkuID,
			SpuName:     line.SpuName,
			UnitPrice:   line.OriginalPrice,
			Quantity:    line.Quantity,
			AddonAmount: addonAmount,
		})
	}

	output, err := s.promotions.Pricing(ctx, promotionService.PricingInput{
		StoreID:   storeID,
		MemberID:  memberID,
		Channel:   channel,
		Items:     items,
		CouponID:  couponID,
		UsePoints: usePoints,
	})
	if err != nil {
		return nil, err
	}

	// 单品级优惠回写到明细行：计价引擎的行小计会扣掉这部分
	for index, discount := range output.LineDiscounts {
		if index >= len(lines) {
			break
		}
		lines[index].LineDiscount = discount.Amount
	}

	rule := PricingRuleInput{
		IsMember:      memberID > 0,
		ActivityAmount: output.ActivityAmount,
		ActivityName:   joinActivityNames(output.ActivityNames),
		CouponAmount:   output.CouponAmount,
		CouponName:     output.CouponName,
		CouponThreshold: output.CouponThreshold,
		PointsAmount:   output.PointsAmount,
	}

	return &marketingContext{Rule: rule, Output: output}, nil
}

// joinActivityNames 把生效的活动名称拼成一句，用于优惠留档。
func joinActivityNames(names []string) string {
	if len(names) == 0 {
		return ""
	}

	result := names[0]
	for _, name := range names[1:] {
		result += "、" + name
	}

	return result
}
