// pricing.go 实现订单价格试算，是纯函数包，不依赖数据库与会话。
// 口径唯一来源：docs/design/07-会员与营销设计.md §4 优惠计算顺序——
// 1. 菜品原价（门店价 / 会员价 / 标准价，调用方按门店上下文把「原价」解析成门店价后传入）
// 2. 单品级优惠（单品券、加价购）→ 行级 LineDiscount，调用方预算
// 3+4. 商品级 / 订单级活动 → 合并为 ActivityAmount，由调用方（营销模块）按互斥规则选优后传入
// 5. 优惠券（校验门槛）
// 6. 积分抵扣（受上限约束）
// 7. 储值卡抵扣 → 不在本函数内，属于支付环节的组合支付
//
// 全程整型分（AM-06）；整单优惠按明细小计占比分摊，最后一批用「应付 − 已付」消尾差。
// 每一步都产生 PricingStep 留档，可完整复现当时结果。
package service

import (
	"eqray-flavor/app/modules/base/types"
)

// PricingAddonInput 是试算输入里的一行加料。
type PricingAddonInput struct {
	// AddonItemID 加料项 ID
	AddonItemID types.ID
	// Name 加料名称快照
	Name string
	// Price 单份加价（分）
	Price int64
	// Quantity 份数
	Quantity int
}

// PricingLineInput 是试算输入里的一行明细。
type PricingLineInput struct {
	// SpuID / SkuID 菜品与规格 ID
	SpuID types.ID
	SkuID types.ID
	// SpuName / SkuName 名称快照
	SpuName string
	SkuName string
	// OriginalPrice 单价（分）：调用方已按门店上下文解析出的门店价或标准价
	OriginalPrice int64
	// MemberPrice 会员价（分），0 表示该规格未配置会员价
	MemberPrice int64
	// Quantity 数量，必须大于 0
	Quantity int
	// Addons 已选加料
	Addons []PricingAddonInput
	// LineDiscount 单品级优惠总额（分），来自单品券 / 加价购，由营销模块预算后传入
	LineDiscount int64
}

// PricingRuleInput 是营销规则解析结果，营销模块建成前由调用方直接填金额。
type PricingRuleInput struct {
	// IsMember 是否会员，决定是否启用会员价
	IsMember bool
	// ActivityAmount 商品级 + 订单级活动优惠合计（分），已按互斥规则选优
	ActivityAmount int64
	// ActivityName 生效活动名称，用于留档
	ActivityName string
	// CouponAmount 优惠券面额（分）
	CouponAmount int64
	// CouponName 券名称，用于留档
	CouponName string
	// CouponThreshold 券使用门槛（分），0 表示无门槛；门槛与活动优惠后的金额比较
	CouponThreshold int64
	// PointsAmount 拟用积分折算的金额（分），引擎会按剩余应付截断
	PointsAmount int64
}

// PricingStep 是计价留档的一步：步骤序号、名称、金额与说明。
type PricingStep struct {
	// Step 步骤序号，与 07 §4 的顺序一致
	Step int
	// Name 步骤名称，如「商品级/订单级活动」「优惠券」
	Name string
	// Amount 本步优惠金额（分），0 表示本步未生效
	Amount int64
	// Detail 说明，含未生效原因
	Detail string
}

// PricingLineResult 是一行明细的计价结果。
type PricingLineResult struct {
	SpuID   types.ID
	SkuID   types.ID
	SpuName string
	SkuName string
	Addons  []PricingAddonInput
	// UnitPrice 成交单价（分）：会员价或原价
	UnitPrice int64
	// Quantity 数量
	Quantity int
	// AddonAmount 本行加料费合计（分）
	AddonAmount int64
	// LineDiscount 单品级优惠（分）
	LineDiscount int64
	// Subtotal 本行小计（分）= unit_price * quantity + addon_amount - line_discount
	Subtotal int64
	// OrderDiscount 本行分摊到的整单优惠（分）
	OrderDiscount int64
	// Payable 本行应付（分）= subtotal - order_discount
	Payable int64
}

// PricingResult 是一次试算的完整结果。
type PricingResult struct {
	// Lines 逐行计价结果
	Lines []PricingLineResult
	// ItemAmount 商品小计（分），不含加料，按成交单价计
	ItemAmount int64
	// AddonAmount 加料费合计（分）
	AddonAmount int64
	// LineDiscountAmount 单品级优惠合计（分），已在行小计中扣减
	LineDiscountAmount int64
	// ActivityAmount 实际生效的活动优惠（分）
	ActivityAmount int64
	// CouponAmount 实际生效的券抵扣（分）
	CouponAmount int64
	// PointsAmount 实际生效的积分抵扣（分）
	PointsAmount int64
	// PayAmount 应付金额（分）= item_amount + addon_amount - activity - coupon - points
	PayAmount int64
	// Steps 计价过程留档
	Steps []PricingStep
}

// CalculatePricing 按唯一口径计算一笔订单的价格。
// 入参：明细行与营销规则；返回计价结果。金额全部为整型分。
// 约定：各输入金额非负；本函数不校验库存与营业状态，只管钱。
func CalculatePricing(lines []PricingLineInput, rule PricingRuleInput) *PricingResult {
	result := &PricingResult{Lines: make([]PricingLineResult, 0, len(lines)), Steps: make([]PricingStep, 0, 6)}

	// 步骤 1：菜品原价选择——会员且配置了会员价时用会员价，否则用调用方解析的原价
	for _, line := range lines {
		unitPrice := line.OriginalPrice
		if rule.IsMember && line.MemberPrice > 0 && line.MemberPrice < unitPrice {
			unitPrice = line.MemberPrice
		}

		addonAmount := int64(0)
		for _, addon := range line.Addons {
			if addon.Price > 0 && addon.Quantity > 0 {
				addonAmount += addon.Price * int64(addon.Quantity)
			}
		}

		lineDiscount := line.LineDiscount
		if lineDiscount < 0 {
			lineDiscount = 0
		}
		subtotal := unitPrice*int64(line.Quantity) + addonAmount - lineDiscount
		if subtotal < 0 {
			// 单品级优惠超过行金额时只抵到 0，避免负数小计污染分摊
			lineDiscount += subtotal
			subtotal = 0
		}

		result.Lines = append(result.Lines, PricingLineResult{
			SpuID: line.SpuID, SkuID: line.SkuID,
			SpuName: line.SpuName, SkuName: line.SkuName,
			Addons:    line.Addons,
			UnitPrice: unitPrice, Quantity: line.Quantity,
			AddonAmount: addonAmount, LineDiscount: lineDiscount,
			Subtotal: subtotal,
		})

		result.ItemAmount += unitPrice * int64(line.Quantity)
		result.AddonAmount += addonAmount
		result.LineDiscountAmount += lineDiscount
	}
	result.Steps = append(result.Steps, PricingStep{Step: 1, Name: "菜品原价", Amount: 0, Detail: detailOfPrice(rule, result.ItemAmount)})

	goodsAmount := result.ItemAmount + result.AddonAmount

	// 步骤 3+4：商品级 / 订单级活动（互斥选优由营销模块完成，这里只记账）
	activity := clampNonNegative(rule.ActivityAmount)
	if activity > goodsAmount {
		activity = goodsAmount
	}
	result.ActivityAmount = activity
	result.Steps = append(result.Steps, PricingStep{Step: 3, Name: "商品级/订单级活动", Amount: activity, Detail: activityDetail(rule, activity)})

	// 步骤 5：优惠券——门槛与活动优惠后的金额比较，面额最多抵到剩余应付
	afterActivity := goodsAmount - activity
	coupon := clampNonNegative(rule.CouponAmount)
	couponDetail := ""
	switch {
	case coupon == 0:
		couponDetail = "未使用优惠券"
	case afterActivity < rule.CouponThreshold:
		// AM-03 边界：门槛 100 元对 99.99 元不可用
		coupon = 0
		couponDetail = "未达使用门槛，券不生效"
	default:
		if coupon > afterActivity {
			coupon = afterActivity
			couponDetail = "面额超过应付，只抵到剩余应付"
		}
	}
	result.CouponAmount = coupon
	result.Steps = append(result.Steps, PricingStep{Step: 5, Name: "优惠券", Amount: coupon, Detail: couponDetail})

	// 步骤 6：积分抵扣——只抵到剩余应付
	afterCoupon := afterActivity - coupon
	points := clampNonNegative(rule.PointsAmount)
	pointsDetail := "未使用积分"
	if points > 0 {
		if points > afterCoupon {
			points = afterCoupon
			pointsDetail = "超过上限，只抵到剩余应付"
		} else {
			pointsDetail = "按拟抵扣金额生效"
		}
	}
	result.PointsAmount = points
	result.Steps = append(result.Steps, PricingStep{Step: 6, Name: "积分抵扣", Amount: points, Detail: pointsDetail})

	// 应付金额 = 商品 + 加料 − 单品级优惠 − 活动 − 券 − 积分
	// （行小计已扣单品级优惠，这里显式列出各项，保证与优惠快照逐项可对）
	result.PayAmount = result.ItemAmount + result.AddonAmount - result.LineDiscountAmount - result.ActivityAmount - result.CouponAmount - result.PointsAmount
	if result.PayAmount < 0 {
		result.PayAmount = 0
	}

	allocateOrderDiscount(result)

	return result
}

// allocateOrderDiscount 把整单优惠（活动 + 券 + 积分）按明细小计占比分摊到行。
// 前面的行按比例分，最后一行用「应付合计 − 已分摊」消尾差，保证行应付之和等于订单应付（AM-06）。
// 行级（单品级）优惠已在行小计中扣减，不参与这里的分摊。
func allocateOrderDiscount(result *PricingResult) {
	totalDiscount := result.ActivityAmount + result.CouponAmount + result.PointsAmount
	if totalDiscount == 0 || len(result.Lines) == 0 {
		for i := range result.Lines {
			result.Lines[i].OrderDiscount = 0
			result.Lines[i].Payable = result.Lines[i].Subtotal
		}

		return
	}

	goodsAmount := int64(0)
	for i := range result.Lines {
		goodsAmount += result.Lines[i].Subtotal
	}
	if goodsAmount <= 0 {
		return
	}

	allocated := int64(0)
	lastIndex := len(result.Lines) - 1
	for i := range result.Lines {
		line := &result.Lines[i]
		if i == lastIndex {
			// 消尾差：最后一行承担剩余全部优惠
			line.OrderDiscount = totalDiscount - allocated
		} else {
			line.OrderDiscount = line.Subtotal * totalDiscount / goodsAmount
			allocated += line.OrderDiscount
		}

		line.Payable = line.Subtotal - line.OrderDiscount
		if line.Payable < 0 {
			// 比例分摊理论上不会为负，防御性兜底
			line.Payable = 0
		}
	}
}

// clampNonNegative 把负数金额收敛为 0。
func clampNonNegative(amount int64) int64 {
	if amount < 0 {
		return 0
	}

	return amount
}

// detailOfPrice 生成步骤 1 的留档说明。
func detailOfPrice(rule PricingRuleInput, itemAmount int64) string {
	if rule.IsMember {
		return "会员身份已生效，逐行按会员价与原价取低"
	}

	return "非会员，按原价计"
}

// activityDetail 生成活动步骤的留档说明。
func activityDetail(rule PricingRuleInput, applied int64) string {
	if rule.ActivityName != "" {
		return "生效活动：" + rule.ActivityName
	}
	if applied > 0 {
		return "活动优惠已生效"
	}

	return "无活动优惠"
}
