package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	dishErrs "eqray-flavor/app/modules/dish/errs"
	"eqray-flavor/app/modules/dish/model"
	"eqray-flavor/app/modules/dish/repository"
	infraService "eqray-flavor/app/modules/infra/service"
	storeEnums "eqray-flavor/app/modules/store/enums"
	storeService "eqray-flavor/app/modules/store/service"
)

// storePriceFloatConfigKey 是门店价相对品牌价的允许浮动百分比参数键，未配置时不限制。
const storePriceFloatConfigKey = "dish_store_price_float_rate"

// StoreSkuListQuery 是门店菜品配置列表的查询条件。
type StoreSkuListQuery struct {
	// Keyword 菜品名称或编码关键字
	Keyword string
	// CategoryID 分类 ID（含下级），0 表示全部分类
	CategoryID types.ID
	PageNo     int
	PageSize   int
}

// StoreSkuRow 是门店维度的规格配置：品牌侧菜品信息 + 门店可售与价格。
type StoreSkuRow struct {
	SkuID         types.ID                   `json:"skuId"`
	SkuName       string                     `json:"skuName"`
	SkuCode       string                     `json:"skuCode"`
	SpuID         types.ID                   `json:"spuId"`
	SpuName       string                     `json:"spuName"`
	CategoryID    types.ID                   `json:"categoryId"`
	CategoryName  string                     `json:"categoryName"`
	StandardPrice int64                      `json:"standardPrice"`
	MemberPrice   int64                      `json:"memberPrice"`
	SkuStatus     baseEnums.CommonStatusEnum `json:"skuStatus"`
	SpuStatus     baseEnums.CommonStatusEnum `json:"spuStatus"`
	// Configured 该门店是否已配置过此规格，未配置时按不可售处理
	Configured       bool            `json:"configured"`
	Available        int8            `json:"available"`
	SoldOut          int8            `json:"soldOut"`
	OnlineOnly       int8            `json:"onlineOnly"`
	Sort             int             `json:"sort"`
	Remark           string          `json:"remark"`
	StorePrice       int64           `json:"storePrice"`
	StoreMemberPrice int64           `json:"storeMemberPrice"`
	StartAt          *types.DateTime `json:"startAt"`
	EndAt            *types.DateTime `json:"endAt"`
}

// StoreSkuSaveInput 是保存单规格门店配置的入参。
type StoreSkuSaveInput struct {
	Available        int8
	SoldOut          int8
	OnlineOnly       int8
	Sort             int
	Remark           string
	StorePrice       int64
	StoreMemberPrice int64
	StartAt          *types.DateTime
	EndAt            *types.DateTime
}

// StoreSkuBatchInput 是批量调整门店可售状态的入参，只提交需要变更的开关。
type StoreSkuBatchInput struct {
	SkuIDs     []types.ID
	Available  *int8
	SoldOut    *int8
	OnlineOnly *int8
}

// StoreSkuService 负责门店可售与门店价覆盖的维护。
type StoreSkuService struct {
	storeRepo  *repository.DishStoreRepository
	skus       *repository.DishSkuRepository
	spus       *repository.DishSpuRepository
	categories *repository.DishCategoryRepository
	category   *CategoryService
	stores     *storeService.StoreService
	configs    *infraService.ConfigService
}

// NewStoreSkuService 构造门店菜品服务。
func NewStoreSkuService() *StoreSkuService {
	return &StoreSkuService{
		storeRepo:  repository.NewDishStoreRepository(),
		skus:       repository.NewDishSkuRepository(),
		spus:       repository.NewDishSpuRepository(),
		categories: repository.NewDishCategoryRepository(),
		category:   NewCategoryService(),
		stores:     storeService.NewStoreService(),
		configs:    infraService.NewConfigService(),
	}
}

// Page 分页查询当前门店的规格配置，按菜品关键字与分类筛选，未配置的规格也可列出。
func (s *StoreSkuService) Page(ctx context.Context, q StoreSkuListQuery) ([]StoreSkuRow, int64, error) {
	storeID, err := s.requireStore(ctx)
	if err != nil {
		return nil, 0, err
	}

	skuFilter := repository.DishSkuFilter{}
	keyword := strings.TrimSpace(q.Keyword)
	if keyword != "" || q.CategoryID > 0 {
		spuFilter := repository.DishSpuFilter{Keyword: keyword}
		if q.CategoryID > 0 {
			subtree, err := s.category.DescendantIDs(ctx, q.CategoryID)
			if err != nil {
				return nil, 0, err
			}
			ids := make([]types.ID, 0, len(subtree))
			for id := range subtree {
				ids = append(ids, id)
			}
			spuFilter.CategoryIDs = ids
		}

		spuIDs, err := s.spus.FindIDsByFilter(ctx, spuFilter)
		if err != nil {
			return nil, 0, err
		}
		skuFilter.SpuIDs = spuIDs
		skuFilter.RestrictSpu = true
	}

	skus, total, err := s.skus.Paginate(ctx, skuFilter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list, err := s.decorate(ctx, storeID, skus)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Save 保存单规格的门店配置与门店价覆盖。
// 约束：门店价必须在品牌价的允许浮动区间内；会员价不高于门店价；生效区间开始早于结束。
func (s *StoreSkuService) Save(ctx context.Context, skuID types.ID, in StoreSkuSaveInput) error {
	storeID, err := s.requireStore(ctx)
	if err != nil {
		return err
	}

	sku, err := s.skus.FindByID(ctx, skuID)
	if err != nil {
		return err
	}
	if sku == nil {
		return baseErrs.New(dishErrs.DishSkuNotExists)
	}

	if err := s.validatePrice(ctx, sku, in); err != nil {
		return err
	}

	storeSku := &model.DishStoreSkuModel{
		StoreID:    storeID,
		SkuID:      sku.ID,
		Available:  in.Available,
		SoldOut:    in.SoldOut,
		OnlineOnly: in.OnlineOnly,
		Sort:       in.Sort,
		Remark:     in.Remark,
	}

	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.storeRepo.SaveStoreSku(ctx, tx, storeSku); err != nil {
			return err
		}

		// 门店价与会员价都为 0 且没有区间时，视为不覆盖品牌价，直接清理历史覆盖
		if in.StorePrice <= 0 && in.StoreMemberPrice <= 0 && in.StartAt == nil && in.EndAt == nil {
			return s.storeRepo.DeleteStorePrice(ctx, tx, storeID, sku.ID)
		}

		return s.storeRepo.SaveStorePrice(ctx, tx, &model.DishStorePriceModel{
			StoreID:          storeID,
			SkuID:            sku.ID,
			StorePrice:       in.StorePrice,
			StoreMemberPrice: in.StoreMemberPrice,
			StartAt:          in.StartAt,
			EndAt:            in.EndAt,
			Remark:           in.Remark,
		})
	})
	if err != nil {
		return err
	}

	InvalidateStoreMenuCache(storeID)

	return nil
}

// BatchUpdate 批量调整门店可售开关（上架 / 下架 / 沽清），只改提交的字段。
// 入参：skuIds + available / soldOut / onlineOnly 中任意组合；返回 null。
func (s *StoreSkuService) BatchUpdate(ctx context.Context, in StoreSkuBatchInput) error {
	storeID, err := s.requireStore(ctx)
	if err != nil {
		return err
	}

	skuIDs := uniqueIDs(in.SkuIDs)
	if len(skuIDs) == 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "请至少选择一个规格")
	}
	if in.Available == nil && in.SoldOut == nil && in.OnlineOnly == nil {
		return baseErrs.WithMsg(baseErrs.BadRequest, "请至少提交一个要变更的开关")
	}

	current, err := s.storeRepo.FindStoreSkus(ctx, storeID, skuIDs)
	if err != nil {
		return err
	}
	existing := make(map[types.ID]*model.DishStoreSkuModel, len(current))
	for index := range current {
		existing[current[index].SkuID] = &current[index]
	}

	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		for _, skuID := range skuIDs {
			m, ok := existing[skuID]
			if !ok {
				m = &model.DishStoreSkuModel{StoreID: storeID, SkuID: skuID, Available: 1}
			}
			if in.Available != nil {
				m.Available = *in.Available
			}
			if in.SoldOut != nil {
				m.SoldOut = *in.SoldOut
			}
			if in.OnlineOnly != nil {
				m.OnlineOnly = *in.OnlineOnly
			}
			if err := s.storeRepo.SaveStoreSku(ctx, tx, m); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	InvalidateStoreMenuCache(storeID)

	return nil
}

// requireStore 取当前门店上下文并校验门店存在且为门店节点。
func (s *StoreSkuService) requireStore(ctx context.Context) (types.ID, error) {
	storeID := support.StoreID(ctx)
	if storeID <= 0 {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "请先在页面上选择门店")
	}

	store, err := s.stores.FindByID(ctx, storeID)
	if err != nil {
		return 0, err
	}
	if store == nil || store.Type != storeEnums.StoreTypeShop {
		return 0, baseErrs.New(dishErrs.DishSkuNotSellable)
	}

	return storeID, nil
}

// validatePrice 校验门店价：非负、不高于/不低于浮动区间、会员价不高于门店价、区间合法。
func (s *StoreSkuService) validatePrice(ctx context.Context, sku *model.DishSkuModel, in StoreSkuSaveInput) error {
	if in.StorePrice < 0 || in.StoreMemberPrice < 0 {
		return baseErrs.New(dishErrs.DishStorePriceInvalid)
	}
	if in.StoreMemberPrice > 0 && in.StorePrice > 0 && in.StoreMemberPrice > in.StorePrice {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店会员价不能高于门店价")
	}
	if in.StartAt != nil && in.EndAt != nil && !in.StartAt.Time().Before(in.EndAt.Time()) {
		return baseErrs.New(dishErrs.DishStorePriceRangeInvalid)
	}

	if in.StorePrice <= 0 || sku.StandardPrice <= 0 {
		return nil
	}

	raw := strings.TrimSpace(s.configs.GetString(ctx, storePriceFloatConfigKey, ""))
	if raw == "" {
		return nil
	}

	rate, err := strconv.ParseFloat(raw, 64)
	if err != nil || rate <= 0 {
		return nil
	}

	delta := math.Abs(float64(in.StorePrice-sku.StandardPrice)) / float64(sku.StandardPrice) * 100
	if delta > rate {
		return baseErrs.New(
			dishErrs.DishStorePriceFloatExceeded,
			formatCent(in.StorePrice),
			formatCent(sku.StandardPrice),
		)
	}

	return nil
}

// decorate 为规格列表补齐菜品、分类与门店配置信息。
func (s *StoreSkuService) decorate(ctx context.Context, storeID types.ID, skus []model.DishSkuModel) ([]StoreSkuRow, error) {
	list := make([]StoreSkuRow, 0, len(skus))
	if len(skus) == 0 {
		return list, nil
	}

	skuIDs := make([]types.ID, 0, len(skus))
	spuIDs := make([]types.ID, 0, len(skus))
	for _, sku := range skus {
		skuIDs = append(skuIDs, sku.ID)
		spuIDs = append(spuIDs, sku.SpuID)
	}

	spus, err := s.spus.FindByIDs(ctx, spuIDs)
	if err != nil {
		return nil, err
	}
	spuMap := make(map[types.ID]*model.DishSpuModel, len(spus))
	categoryIDs := make([]types.ID, 0, len(spus))
	for index := range spus {
		spuMap[spus[index].ID] = &spus[index]
		categoryIDs = append(categoryIDs, spus[index].CategoryID)
	}

	categories, err := s.categories.FindByIDs(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}
	categoryMap := make(map[types.ID]string, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category.Name
	}

	configs, err := s.storeRepo.FindStoreSkus(ctx, storeID, skuIDs)
	if err != nil {
		return nil, err
	}
	configMap := make(map[types.ID]*model.DishStoreSkuModel, len(configs))
	for index := range configs {
		configMap[configs[index].SkuID] = &configs[index]
	}

	prices, err := s.storeRepo.FindStorePrices(ctx, storeID, skuIDs)
	if err != nil {
		return nil, err
	}
	priceMap := make(map[types.ID]*model.DishStorePriceModel, len(prices))
	for index := range prices {
		priceMap[prices[index].SkuID] = &prices[index]
	}

	for _, sku := range skus {
		row := StoreSkuRow{
			SkuID:         sku.ID,
			SkuName:       sku.Name,
			SkuCode:       sku.Code,
			StandardPrice: sku.StandardPrice,
			MemberPrice:   sku.MemberPrice,
			SkuStatus:     sku.Status,
			SpuID:         sku.SpuID,
			Configured:    false,
			Available:     0,
			SoldOut:       0,
		}
		if spu, ok := spuMap[sku.SpuID]; ok {
			row.SpuName = spu.Name
			row.CategoryID = spu.CategoryID
			row.CategoryName = categoryMap[spu.CategoryID]
			row.SpuStatus = spu.Status
		}
		if config, ok := configMap[sku.ID]; ok {
			row.Configured = true
			row.Available = config.Available
			row.SoldOut = config.SoldOut
			row.OnlineOnly = config.OnlineOnly
			row.Sort = config.Sort
			row.Remark = config.Remark
		}
		if price, ok := priceMap[sku.ID]; ok {
			row.StorePrice = price.StorePrice
			row.StoreMemberPrice = price.StoreMemberPrice
			row.StartAt = price.StartAt
			row.EndAt = price.EndAt
		}
		list = append(list, row)
	}

	return list, nil
}

// formatCent 把金额（分）格式化成两位小数的元，用于错误提示。
func formatCent(cent int64) string {
	return fmt.Sprintf("%.2f", float64(cent)/100)
}
