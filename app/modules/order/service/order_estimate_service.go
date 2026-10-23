// order_estimate_service.go 实现订单价格试算：把门店、菜品、加料的持久化数据
// 装配成纯函数 CalculatePricing 的入参。装配口径与 C 端菜单完全一致——
// 价格取值链复用 dish service 的 ResolveSkuPrice，避免菜单与订单两处漂移。
package service

import (
	"context"
	"time"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	dishModel "eqray-flavor/app/modules/dish/model"
	dishService "eqray-flavor/app/modules/dish/service"
	orderErrs "eqray-flavor/app/modules/order/errs"
)

// EstimateAddonInput 是试算输入里的一行加料选择。
type EstimateAddonInput struct {
	// AddonItemID 加料项 ID
	AddonItemID types.ID
	// Quantity 份数，缺省按 1
	Quantity int
}

// EstimateItemInput 是试算输入里的一行明细。
type EstimateItemInput struct {
	// SkuID 菜品规格 ID
	SkuID types.ID
	// Quantity 数量，必须大于 0
	Quantity int
	// Addons 已选加料
	Addons []EstimateAddonInput
}

// EstimateInput 是一次试算的输入。
type EstimateInput struct {
	// StoreID 门店 ID
	StoreID types.ID
	// IsMember 是否会员
	IsMember bool
	// Items 明细行
	Items []EstimateItemInput
}

// EstimateAddonResult 是试算结果里的一行加料。
type EstimateAddonResult struct {
	AddonItemID types.ID `json:"addonItemId"`
	Name        string   `json:"name"`
	Price       int64    `json:"price"`
	Quantity    int      `json:"quantity"`
}

// EstimateLineResult 是试算结果里的一行明细。
type EstimateLineResult struct {
	SkuID         types.ID              `json:"skuId"`
	SpuID         types.ID              `json:"spuId"`
	SpuName       string                `json:"spuName"`
	SkuName       string                `json:"skuName"`
	Addons        []EstimateAddonResult `json:"addons"`
	UnitPrice     int64                 `json:"unitPrice"`
	Quantity      int                   `json:"quantity"`
	AddonAmount   int64                 `json:"addonAmount"`
	LineDiscount  int64                 `json:"lineDiscount"`
	Subtotal      int64                 `json:"subtotal"`
	OrderDiscount int64                 `json:"orderDiscount"`
	Payable       int64                 `json:"payable"`
}

// EstimateResult 是试算的完整结果，C 端直接展示。
type EstimateResult struct {
	Lines              []EstimateLineResult `json:"lines"`
	ItemAmount         int64                `json:"itemAmount"`
	AddonAmount        int64                `json:"addonAmount"`
	LineDiscountAmount int64                `json:"lineDiscountAmount"`
	ActivityAmount     int64                `json:"activityAmount"`
	CouponAmount       int64                `json:"couponAmount"`
	PointsAmount       int64                `json:"pointsAmount"`
	PayAmount          int64                `json:"payAmount"`
	// Steps 计价过程留档，前端可展开给顾客看
	Steps []PricingStep `json:"steps"`
}

// Estimate 按门店上下文试算一笔购物篮的价格。
// 入参：门店、会员身份与明细行；返回逐行计价与合计。只读操作，不落库。
func (s *OrderService) Estimate(ctx context.Context, in EstimateInput) (*EstimateResult, error) {
	lines, err := s.resolveLines(ctx, in.StoreID, in.IsMember, in.Items)
	if err != nil {
		return nil, err
	}

	// 营销模块未建成：活动 / 券 / 积分暂无来源，规则入参只带会员身份
	result := CalculatePricing(lines, PricingRuleInput{IsMember: in.IsMember})

	out := &EstimateResult{
		Lines:              make([]EstimateLineResult, 0, len(result.Lines)),
		ItemAmount:         result.ItemAmount,
		AddonAmount:        result.AddonAmount,
		LineDiscountAmount: result.LineDiscountAmount,
		ActivityAmount:     result.ActivityAmount,
		CouponAmount:       result.CouponAmount,
		PointsAmount:       result.PointsAmount,
		PayAmount:          result.PayAmount,
		Steps:              result.Steps,
	}
	for _, line := range result.Lines {
		addons := make([]EstimateAddonResult, 0, len(line.Addons))
		for _, addon := range line.Addons {
			addons = append(addons, EstimateAddonResult{
				AddonItemID: addon.AddonItemID,
				Name:        addon.Name,
				Price:       addon.Price,
				Quantity:    addon.Quantity,
			})
		}
		out.Lines = append(out.Lines, EstimateLineResult{
			SkuID: line.SkuID, SpuID: line.SpuID,
			SpuName: line.SpuName, SkuName: line.SkuName,
			Addons:    addons,
			UnitPrice: line.UnitPrice, Quantity: line.Quantity,
			AddonAmount: line.AddonAmount, LineDiscount: line.LineDiscount,
			Subtotal: line.Subtotal, OrderDiscount: line.OrderDiscount, Payable: line.Payable,
		})
	}

	return out, nil
}

// skuRow 是装配期的规格上下文：主表字段 + 门店可售与门店价配置。
type skuRow struct {
	sku              *dishModel.DishSkuModel
	available        bool
	storePrice       int64
	storeMemberPrice int64
	startAt          *types.DateTime
	endAt            *types.DateTime
}

// resolveLines 把请求里的明细行解析成计价纯函数的入参：
// 校验门店营业、规格上架可售、加料归属与数量限制，并按价格取值链解析原价。
// 试算与下单共用，保证「试算价 = 下单价」。
func (s *OrderService) resolveLines(ctx context.Context, storeID types.ID, isMember bool, items []EstimateItemInput) ([]PricingLineInput, error) {
	if len(items) == 0 {
		return nil, errs.New(orderErrs.OrderItemsEmpty)
	}

	// 门店营业校验：经营状态 1（营业）才允许交易
	store, err := s.stores.FindByID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store == nil || int8(store.Status) != 1 {
		return nil, errs.New(orderErrs.OrderStoreNotOpen)
	}

	// 聚合去重后的 SKU 集合
	skuIDs := make([]types.ID, 0, len(items))
	skuSeen := make(map[types.ID]bool, len(items))
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, errs.New(orderErrs.OrderQuantityInvalid)
		}
		if !skuSeen[item.SkuID] {
			skuSeen[item.SkuID] = true
			skuIDs = append(skuIDs, item.SkuID)
		}
	}

	// 品牌规格与 SPU：必须存在且上架
	skus, err := s.skus.FindByIDs(ctx, skuIDs)
	if err != nil {
		return nil, err
	}
	skuMap := make(map[types.ID]*skuRow, len(skus))
	spuIDs := make([]types.ID, 0, len(skus))
	spuSeen := make(map[types.ID]bool, len(skus))
	for i := range skus {
		sku := &skus[i]
		if sku.Status != baseEnums.CommonStatusEnabled {
			return nil, errs.New(orderErrs.OrderSkuNotFound)
		}
		skuMap[sku.ID] = &skuRow{sku: sku}
		if !spuSeen[sku.SpuID] {
			spuSeen[sku.SpuID] = true
			spuIDs = append(spuIDs, sku.SpuID)
		}
	}
	if len(skuMap) != len(skuIDs) {
		return nil, errs.New(orderErrs.OrderSkuNotFound)
	}

	spus, err := s.spus.FindByIDs(ctx, spuIDs)
	if err != nil {
		return nil, err
	}
	spuName := make(map[types.ID]string, len(spus))
	for _, spu := range spus {
		if spu.Status != baseEnums.CommonStatusEnabled {
			return nil, errs.New(orderErrs.OrderSkuNotFound)
		}
		spuName[spu.ID] = spu.Name
	}

	// 门店可售与沽清校验：未配置可售视为不可售
	storeSkus, err := s.dishStores.FindStoreSkus(ctx, storeID, skuIDs)
	if err != nil {
		return nil, err
	}
	for _, storeSku := range storeSkus {
		if row, ok := skuMap[storeSku.SkuID]; ok {
			row.available = storeSku.Available == 1 && storeSku.SoldOut == 0
		}
	}
	for _, row := range skuMap {
		if !row.available {
			return nil, errs.New(orderErrs.OrderSkuNotSellable, row.sku.Name)
		}
	}

	// 门店价（含时段区间），供价格取值链使用
	storePrices, err := s.dishStores.FindStorePrices(ctx, storeID, skuIDs)
	if err != nil {
		return nil, err
	}
	for _, price := range storePrices {
		if row, ok := skuMap[price.SkuID]; ok {
			row.storePrice = price.StorePrice
			row.storeMemberPrice = price.StoreMemberPrice
			row.startAt = price.StartAt
			row.endAt = price.EndAt
		}
	}

	if err := s.validateAddons(ctx, items, skuMap); err != nil {
		return nil, err
	}

	// 装配计价入参：会员价不单独传，由 ResolveSkuPrice 的取值链统一决定原价
	now := time.Now()
	lines := make([]PricingLineInput, 0, len(items))
	for _, item := range items {
		row := skuMap[item.SkuID]
		originalPrice := dishService.ResolveSkuPrice(
			row.sku.StandardPrice, row.sku.MemberPrice,
			row.storePrice, row.storeMemberPrice,
			row.startAt, row.endAt, isMember, now,
		)

		addons := make([]PricingAddonInput, 0, len(item.Addons))
		for _, chosen := range item.Addons {
			addonItem, err := s.addons.FindItemByID(ctx, chosen.AddonItemID)
			if err != nil {
				return nil, err
			}
			if addonItem == nil {
				return nil, errs.New(orderErrs.OrderAddonNotInGroup, "", row.sku.Name)
			}
			quantity := chosen.Quantity
			if quantity <= 0 {
				quantity = 1
			}
			addons = append(addons, PricingAddonInput{
				AddonItemID: addonItem.ID,
				Name:        addonItem.Name,
				Price:       addonItem.Price,
				Quantity:    quantity,
			})
		}

		lines = append(lines, PricingLineInput{
			SpuID:         row.sku.SpuID,
			SkuID:         row.sku.ID,
			SpuName:       spuName[row.sku.SpuID],
			SkuName:       row.sku.Name,
			OriginalPrice: originalPrice,
			MemberPrice:   0,
			Quantity:      item.Quantity,
			Addons:        addons,
		})
	}

	return lines, nil
}

// validateAddons 校验加料选择：加料项必须挂在该 SKU 所属 SPU 的加料组上，且数量满足组限制。
func (s *OrderService) validateAddons(ctx context.Context, items []EstimateItemInput, skuMap map[types.ID]*skuRow) error {
	// 收集所有被选加料项
	allAddonIDs := make([]types.ID, 0)
	addonSeen := make(map[types.ID]bool)
	for _, item := range items {
		for _, chosen := range item.Addons {
			if !addonSeen[chosen.AddonItemID] {
				addonSeen[chosen.AddonItemID] = true
				allAddonIDs = append(allAddonIDs, chosen.AddonItemID)
			}
		}
	}
	if len(allAddonIDs) == 0 {
		return nil
	}

	// 加载加料项与组
	addonItems, err := s.addons.FindItemsByIDs(ctx, allAddonIDs)
	if err != nil {
		return err
	}
	addonItemMap := make(map[types.ID]*dishModel.DishAddonItemModel, len(addonItems))
	for i := range addonItems {
		item := &addonItems[i]
		if item.Status != baseEnums.CommonStatusEnabled {
			return errs.New(orderErrs.OrderAddonNotInGroup, item.Name, "")
		}
		addonItemMap[item.ID] = item
	}

	// SPU → 加料组归属
	spuIDs := make([]types.ID, 0, len(skuMap))
	for _, row := range skuMap {
		spuIDs = append(spuIDs, row.sku.SpuID)
	}
	spuGroupMap, err := s.addons.FindSpuAddonMap(ctx, spuIDs)
	if err != nil {
		return err
	}

	// 加料组全集：光加载「被选中的加料项所属的组」不够，
	// 必选组可能一个都没选，必须按 SPU 的加料组全集加载才能校验下限
	allGroupIDs := make([]types.ID, 0)
	groupSeen := make(map[types.ID]bool)
	for _, groupIDs := range spuGroupMap {
		for _, groupID := range groupIDs {
			if !groupSeen[groupID] {
				groupSeen[groupID] = true
				allGroupIDs = append(allGroupIDs, groupID)
			}
		}
	}
	groups, err := s.addons.FindGroupsByIDs(ctx, allGroupIDs)
	if err != nil {
		return err
	}
	groupMap := make(map[types.ID]*dishModel.DishAddonGroupModel, len(groups))
	for i := range groups {
		groupMap[groups[i].ID] = &groups[i]
	}

	// 逐行校验归属与数量范围
	for _, item := range items {
		row, ok := skuMap[item.SkuID]
		if !ok {
			continue
		}
		allowedGroups := make(map[types.ID]bool)
		for _, groupID := range spuGroupMap[row.sku.SpuID] {
			allowedGroups[groupID] = true
		}

		countByGroup := make(map[types.ID]int)
		for _, chosen := range item.Addons {
			addonItem, ok := addonItemMap[chosen.AddonItemID]
			if !ok {
				return errs.New(orderErrs.OrderAddonNotInGroup, "", row.sku.Name)
			}
			if !allowedGroups[addonItem.GroupID] {
				return errs.New(orderErrs.OrderAddonNotInGroup, addonItem.Name, row.sku.Name)
			}
			quantity := chosen.Quantity
			if quantity <= 0 {
				quantity = 1
			}
			countByGroup[addonItem.GroupID] += quantity
		}

		// 必选组（required=1）与最小可选数：一个都没选也必须拦下来，
		// 否则「必选加料」形同虚设，用户可以不下单就跳过
		for groupID := range allowedGroups {
			group, ok := groupMap[groupID]
			if !ok {
				continue
			}
			count := countByGroup[groupID]
			minimum := group.MinSelect
			if group.Required == 1 && minimum <= 0 {
				minimum = 1
			}
			if count < minimum {
				return errs.New(orderErrs.OrderAddonRangeExceeded, row.sku.Name, group.Name)
			}
			if group.MaxSelect > 0 && count > group.MaxSelect {
				return errs.New(orderErrs.OrderAddonRangeExceeded, row.sku.Name, group.Name)
			}
		}
	}

	return nil
}
