package service

import (
	"context"
	"encoding/json"
	"time"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	dishErrs "eqray-flavor/app/modules/dish/errs"
	"eqray-flavor/app/modules/dish/model"
	"eqray-flavor/app/modules/dish/repository"
	storeEnums "eqray-flavor/app/modules/store/enums"
	storeService "eqray-flavor/app/modules/store/service"
)

// MenuSnapshot 是门店菜单快照：保存品牌价与门店价原值，取价时按会员身份计算。
// 缓存按门店维度存一份快照，避免会员价与非会员价各缓存一份（docs 15 §2）。
type MenuSnapshot struct {
	StoreID    types.ID        `json:"storeId"`
	BuiltAt    types.DateTime  `json:"builtAt"`
	Categories []*MenuCategory `json:"categories"`
}

// MenuCategory 是菜单分类节点，子节点用指针保证深层分类在组树后仍保留自己的下级。
type MenuCategory struct {
	ID       types.ID        `json:"id"`
	Name     string          `json:"name"`
	Image    string          `json:"image"`
	Children []*MenuCategory `json:"children"`
	Items    []MenuSpu       `json:"items"`
}

// MenuSpu 是菜单中的菜品，含可售规格与加料组。
type MenuSpu struct {
	ID          types.ID         `json:"id"`
	Name        string           `json:"name"`
	Unit        string           `json:"unit"`
	Description string           `json:"description"`
	Image       string           `json:"image"`
	Tags        string           `json:"tags"`
	Allergens   string           `json:"allergens"`
	Skus        []MenuSku        `json:"skus"`
	AddonGroups []MenuAddonGroup `json:"addonGroups"`
}

// MenuSku 是菜单中的规格：保留三档原价，输出时填充最终价。
type MenuSku struct {
	ID               types.ID        `json:"id"`
	Name             string          `json:"name"`
	Status           int8            `json:"status"`
	SoldOut          int8            `json:"soldOut"`
	OnlineOnly       int8            `json:"onlineOnly"`
	StandardPrice    int64           `json:"standardPrice"`
	BrandMemberPrice int64           `json:"brandMemberPrice"`
	StorePrice       int64           `json:"storePrice"`
	StoreMemberPrice int64           `json:"storeMemberPrice"`
	StartAt          *types.DateTime `json:"startAt"`
	EndAt            *types.DateTime `json:"endAt"`
	// Price 是最终展示价（分），按会员身份与生效区间计算
	Price int64 `json:"price"`
}

// MenuAddonGroup 是菜单菜品上的加料组。
type MenuAddonGroup struct {
	ID        types.ID        `json:"id"`
	Name      string          `json:"name"`
	Required  int8            `json:"required"`
	MinSelect int             `json:"minSelect"`
	MaxSelect int             `json:"maxSelect"`
	Items     []MenuAddonItem `json:"items"`
}

// MenuAddonItem 是加料项，价格单位为分。
type MenuAddonItem struct {
	ID    types.ID `json:"id"`
	Name  string   `json:"name"`
	Price int64    `json:"price"`
}

// MenuService 负责门店菜单聚合（T-DISH-04）。
type MenuService struct {
	storeRepo  *repository.DishStoreRepository
	skus       *repository.DishSkuRepository
	spus       *repository.DishSpuRepository
	categories *repository.DishCategoryRepository
	addons     *repository.DishAddonRepository
	stores     *storeService.StoreService
}

// NewMenuService 构造门店菜单服务。
func NewMenuService() *MenuService {
	return &MenuService{
		storeRepo:  repository.NewDishStoreRepository(),
		skus:       repository.NewDishSkuRepository(),
		spus:       repository.NewDishSpuRepository(),
		categories: repository.NewDishCategoryRepository(),
		addons:     repository.NewDishAddonRepository(),
		stores:     storeService.NewStoreService(),
	}
}

// Menu 返回指定门店的菜单，价格按 isMember 计算。
// 链路：菜品上架 → 规格上架 → 门店可售（未配置视为不可售）→ 价格取值链 → 沽清标记。
func (s *MenuService) Menu(ctx context.Context, storeID types.ID, isMember bool) (*MenuSnapshot, error) {
	if storeID <= 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "请指定门店")
	}

	store, err := s.stores.FindByID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store == nil || store.Type != storeEnums.StoreTypeShop {
		return nil, baseErrs.New(dishErrs.DishSkuNotSellable)
	}

	snapshot, err := s.loadSnapshot(ctx, storeID)
	if err != nil {
		return nil, err
	}

	applyMenuItemPrices(snapshot, isMember, time.Now())

	return snapshot, nil
}

// loadSnapshot 优先读门店菜单缓存，未命中时构建并回写。
func (s *MenuService) loadSnapshot(ctx context.Context, storeID types.ID) (*MenuSnapshot, error) {
	key := storeMenuCacheKey(storeID)
	if cached := facades.Cache().GetString(key); cached != "" {
		var snapshot MenuSnapshot
		if err := json.Unmarshal([]byte(cached), &snapshot); err == nil {
			return &snapshot, nil
		}
	}

	snapshot, err := s.buildSnapshot(ctx, storeID)
	if err != nil {
		return nil, err
	}

	if payload, err := json.Marshal(snapshot); err == nil {
		_ = facades.Cache().Put(key, string(payload), menuCacheTTL)
	}

	return snapshot, nil
}

// buildSnapshot 按上架、门店可售、沽清与加料组聚合门店菜单。
func (s *MenuService) buildSnapshot(ctx context.Context, storeID types.ID) (*MenuSnapshot, error) {
	on := int8(baseEnums.CommonStatusEnabled)
	spuIDs, err := s.spus.FindIDsByFilter(ctx, repository.DishSpuFilter{Status: &on})
	if err != nil {
		return nil, err
	}
	if len(spuIDs) == 0 {
		return &MenuSnapshot{StoreID: storeID, BuiltAt: types.Now(), Categories: []*MenuCategory{}}, nil
	}

	spus, err := s.spus.FindByIDs(ctx, spuIDs)
	if err != nil {
		return nil, err
	}

	skus, err := s.skus.FindBySpuIDs(ctx, spuIDs)
	if err != nil {
		return nil, err
	}
	skuIDs := make([]types.ID, 0, len(skus))
	for _, sku := range skus {
		skuIDs = append(skuIDs, sku.ID)
	}

	configs, err := s.storeRepo.FindStoreSkus(ctx, storeID, skuIDs)
	if err != nil {
		return nil, err
	}
	available := make(map[types.ID]*model.DishStoreSkuModel, len(configs))
	for index := range configs {
		available[configs[index].SkuID] = &configs[index]
	}

	prices, err := s.storeRepo.FindStorePrices(ctx, storeID, skuIDs)
	if err != nil {
		return nil, err
	}
	priceMap := make(map[types.ID]*model.DishStorePriceModel, len(prices))
	for index := range prices {
		priceMap[prices[index].SkuID] = &prices[index]
	}

	groupsBySpu, err := s.addons.FindSpuAddonMap(ctx, spuIDs)
	if err != nil {
		return nil, err
	}
	menuGroups, err := s.buildAddonGroups(ctx, groupsBySpu)
	if err != nil {
		return nil, err
	}

	// 按菜品聚合可售规格
	skusBySpu := make(map[types.ID][]MenuSku, len(spus))
	for _, sku := range skus {
		if sku.Status != baseEnums.CommonStatusEnabled {
			continue
		}
		config, ok := available[sku.ID]
		if !ok || config.Available != 1 {
			continue
		}

		item := MenuSku{
			ID:               sku.ID,
			Name:             sku.Name,
			Status:           int8(sku.Status),
			SoldOut:          config.SoldOut,
			OnlineOnly:       config.OnlineOnly,
			StandardPrice:    sku.StandardPrice,
			BrandMemberPrice: sku.MemberPrice,
		}
		if price, ok := priceMap[sku.ID]; ok {
			item.StorePrice = price.StorePrice
			item.StoreMemberPrice = price.StoreMemberPrice
			item.StartAt = price.StartAt
			item.EndAt = price.EndAt
		}
		skusBySpu[sku.SpuID] = append(skusBySpu[sku.SpuID], item)
	}

	spuMap := make(map[types.ID]MenuSpu, len(spus))
	for _, spu := range spus {
		items, ok := skusBySpu[spu.ID]
		if !ok || len(items) == 0 {
			continue
		}
		spuMap[spu.ID] = MenuSpu{
			ID:          spu.ID,
			Name:        spu.Name,
			Unit:        spu.Unit,
			Description: spu.Description,
			Image:       spu.Image,
			Tags:        spu.Tags,
			Allergens:   spu.Allergens,
			Skus:        items,
			AddonGroups: menuGroups[spu.ID],
		}
	}

	categories, err := s.categories.All(ctx)
	if err != nil {
		return nil, err
	}

	// 组树并挂菜品；空分类直接丢弃，避免前端出现空目录
	nodes := make(map[types.ID]*MenuCategory, len(categories))
	for _, category := range categories {
		if category.Status != baseEnums.CommonStatusEnabled {
			continue
		}
		nodes[category.ID] = &MenuCategory{
			ID:       category.ID,
			Name:     category.Name,
			Image:    category.Image,
			Children: make([]*MenuCategory, 0),
			Items:    make([]MenuSpu, 0),
		}
	}

	for _, spu := range spus {
		item, ok := spuMap[spu.ID]
		if !ok {
			continue
		}
		if category, ok := nodes[spu.CategoryID]; ok {
			category.Items = append(category.Items, item)
		}
	}

	// 分类 All 已按 sort、id 排序，这里按同样顺序组树，保持后台配置的展示次序
	roots := make([]*MenuCategory, 0)
	for _, category := range categories {
		node, ok := nodes[category.ID]
		if !ok {
			continue
		}
		if parent, ok := nodes[category.ParentID]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}

	result := make([]*MenuCategory, 0, len(roots))
	for _, root := range roots {
		if len(root.Items) == 0 && len(root.Children) == 0 {
			continue
		}
		result = append(result, root)
	}

	return &MenuSnapshot{StoreID: storeID, BuiltAt: types.Now(), Categories: result}, nil
}

// buildAddonGroups 批量构建菜品上的加料组，停用的组与项不进入菜单。
func (s *MenuService) buildAddonGroups(ctx context.Context, groupsBySpu map[types.ID][]types.ID) (map[types.ID][]MenuAddonGroup, error) {
	groupIDSet := make([]types.ID, 0)
	seen := make(map[types.ID]bool, len(groupsBySpu))
	for _, ids := range groupsBySpu {
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			groupIDSet = append(groupIDSet, id)
		}
	}

	groups, err := s.addons.FindGroupsByIDs(ctx, groupIDSet)
	if err != nil {
		return nil, err
	}
	items, err := s.addons.FindItemsByGroupIDs(ctx, groupIDSet)
	if err != nil {
		return nil, err
	}

	itemsByGroup := make(map[types.ID][]MenuAddonItem, len(groups))
	for _, item := range items {
		if item.Status != baseEnums.CommonStatusEnabled {
			continue
		}
		itemsByGroup[item.GroupID] = append(itemsByGroup[item.GroupID], MenuAddonItem{
			ID:    item.ID,
			Name:  item.Name,
			Price: item.Price,
		})
	}

	groupMap := make(map[types.ID]MenuAddonGroup, len(groups))
	for _, group := range groups {
		if group.Status != baseEnums.CommonStatusEnabled {
			continue
		}
		groupMap[group.ID] = MenuAddonGroup{
			ID:        group.ID,
			Name:      group.Name,
			Required:  group.Required,
			MinSelect: group.MinSelect,
			MaxSelect: group.MaxSelect,
			Items:     itemsByGroup[group.ID],
		}
	}

	result := make(map[types.ID][]MenuAddonGroup, len(groupsBySpu))
	for spuID, ids := range groupsBySpu {
		groups := make([]MenuAddonGroup, 0, len(ids))
		for _, id := range ids {
			if group, ok := groupMap[id]; ok {
				groups = append(groups, group)
			}
		}
		result[spuID] = groups
	}

	return result, nil
}

// applyMenuItemPrices 按会员身份与生效区间为快照中的每个规格计算最终价。
// 取值链（高到低）：门店时段价 > 门店会员价（会员）> 门店价 > 品牌会员价（会员）> 品牌标准价。
func applyMenuItemPrices(snapshot *MenuSnapshot, isMember bool, now time.Time) {
	for categoryIndex := range snapshot.Categories {
		applyCategoryPrices(snapshot.Categories[categoryIndex], isMember, now)
	}
}

// applyCategoryPrices 递归处理分类树，为分类与子分类下的菜品出价。
func applyCategoryPrices(category *MenuCategory, isMember bool, now time.Time) {
	for itemIndex := range category.Items {
		item := &category.Items[itemIndex]
		for skuIndex := range item.Skus {
			sku := &item.Skus[skuIndex]
			sku.Price = resolveSkuPrice(sku, isMember, now)
		}
	}
	for childIndex := range category.Children {
		applyCategoryPrices(category.Children[childIndex], isMember, now)
	}
}

// resolveSkuPrice 按价格取值链算出规格价格（分）。
// 门店价记录带区间时视为时段价，仅区间内生效且优先级最高；区间外整体回落品牌价。
func resolveSkuPrice(sku *MenuSku, isMember bool, now time.Time) int64 {
	return ResolveSkuPrice(sku.StandardPrice, sku.BrandMemberPrice, sku.StorePrice, sku.StoreMemberPrice, sku.StartAt, sku.EndAt, isMember, now)
}

// ResolveSkuPrice 是价格取值链的导出口径，菜单展示与订单试算共用，避免两处漂移。
// 优先级：时段门店价 > 门店会员价 > 门店价 > 品牌会员价 > 品牌标准价；
// 门店价记录带区间时视为时段价，仅区间内生效且优先级最高，区间外整体回落品牌价。
func ResolveSkuPrice(standardPrice, brandMemberPrice, storePrice, storeMemberPrice int64, startAt, endAt *types.DateTime, isMember bool, now time.Time) int64 {
	if inPriceRange(storePrice, storeMemberPrice, startAt, endAt, now) {
		hasRange := startAt != nil || endAt != nil
		if hasRange && storePrice > 0 {
			return storePrice
		}
		if isMember && storeMemberPrice > 0 {
			return storeMemberPrice
		}
		if storePrice > 0 {
			return storePrice
		}
	}

	if isMember && brandMemberPrice > 0 {
		return brandMemberPrice
	}

	return standardPrice
}

// inPriceRange 判断门店价记录当前是否生效；未配置区间即长期有效。
func inPriceRange(storePrice, storeMemberPrice int64, startAt, endAt *types.DateTime, now time.Time) bool {
	if storePrice <= 0 && storeMemberPrice <= 0 {
		return false
	}
	if startAt != nil && now.Before(startAt.Time()) {
		return false
	}
	if endAt != nil && now.After(endAt.Time()) {
		return false
	}

	return true
}
