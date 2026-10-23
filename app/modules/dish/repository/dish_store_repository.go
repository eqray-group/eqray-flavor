package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	dishErrs "eqray-flavor/app/modules/dish/errs"
	"eqray-flavor/app/modules/dish/model"
)

// DishStoreRepository 是门店可售与门店价格两张门店级表的仓储。
// 两张表都按 (store_id, sku_id) 唯一，写入一律走「查存在 → 更新或插入」，避免依赖 MySQL upsert 方言。
type DishStoreRepository struct{}

// NewDishStoreRepository 构造门店菜品仓储。
func NewDishStoreRepository() *DishStoreRepository {
	return &DishStoreRepository{}
}

// FindStoreSkus 批量查询门店对指定规格的可售配置。
func (r *DishStoreRepository) FindStoreSkus(ctx context.Context, storeID types.ID, skuIDs []types.ID) ([]model.DishStoreSkuModel, error) {
	list := make([]model.DishStoreSkuModel, 0)
	values := idValues(uniqueIDs(skuIDs))
	if storeID <= 0 || len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishStoreSkuTable).
		Where("store_id", storeID).
		WhereIn("sku_id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// FindStorePrices 批量查询门店对指定规格的价格覆盖配置。
func (r *DishStoreRepository) FindStorePrices(ctx context.Context, storeID types.ID, skuIDs []types.ID) ([]model.DishStorePriceModel, error) {
	list := make([]model.DishStorePriceModel, 0)
	values := idValues(uniqueIDs(skuIDs))
	if storeID <= 0 || len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishStorePriceTable).
		Where("store_id", storeID).
		WhereIn("sku_id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// SaveStoreSku 在事务内写入门店可售配置：已存在则更新三个开关与排序，否则插入。
func (r *DishStoreRepository) SaveStoreSku(ctx context.Context, tx db.Tx, m *model.DishStoreSkuModel) error {
	var existing model.DishStoreSkuModel
	if err := tx.Table(dishStoreSkuTable).
		Where("store_id", m.StoreID).
		Where("sku_id", m.SkuID).
		WhereNull("deleted_at").
		First(&existing); err != nil {
		return err
	}

	if existing.ID > 0 {
		m.FillUpdate(ctx)
		_, err := tx.Table(dishStoreSkuTable).
			Where("id", existing.ID).
			Update(map[string]any{
				"available":   m.Available,
				"sold_out":    m.SoldOut,
				"online_only": m.OnlineOnly,
				"sort":        m.Sort,
				"remark":      m.Remark,
				"updater":     m.Updater,
				"updated_at":  m.UpdatedAt,
			})
		if err != nil {
			return err
		}
		m.ID = existing.ID

		return nil
	}

	m.FillCreate(ctx)
	_, err := tx.Table(dishStoreSkuTable).Insert(map[string]any{
		"id":          m.ID,
		"store_id":    m.StoreID,
		"sku_id":      m.SkuID,
		"available":   m.Available,
		"sold_out":    m.SoldOut,
		"online_only": m.OnlineOnly,
		"sort":        m.Sort,
		"remark":      m.Remark,
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishStoreSkuExists)
	}

	return err
}

// SaveStorePrice 在事务内写入门店价格：已存在则更新价格与区间，否则插入。
func (r *DishStoreRepository) SaveStorePrice(ctx context.Context, tx db.Tx, m *model.DishStorePriceModel) error {
	var existing model.DishStorePriceModel
	if err := tx.Table(dishStorePriceTable).
		Where("store_id", m.StoreID).
		Where("sku_id", m.SkuID).
		WhereNull("deleted_at").
		First(&existing); err != nil {
		return err
	}

	if existing.ID > 0 {
		m.FillUpdate(ctx)
		_, err := tx.Table(dishStorePriceTable).
			Where("id", existing.ID).
			Update(map[string]any{
				"store_price":        m.StorePrice,
				"store_member_price": m.StoreMemberPrice,
				"start_at":           m.StartAt,
				"end_at":             m.EndAt,
				"remark":             m.Remark,
				"updater":            m.Updater,
				"updated_at":         m.UpdatedAt,
			})
		if err != nil {
			return err
		}
		m.ID = existing.ID

		return nil
	}

	m.FillCreate(ctx)
	_, err := tx.Table(dishStorePriceTable).Insert(map[string]any{
		"id":                 m.ID,
		"store_id":           m.StoreID,
		"sku_id":             m.SkuID,
		"store_price":        m.StorePrice,
		"store_member_price": m.StoreMemberPrice,
		"start_at":           m.StartAt,
		"end_at":             m.EndAt,
		"remark":             m.Remark,
		"creator":            m.Creator,
		"created_at":         m.CreatedAt,
		"updater":            m.Updater,
		"updated_at":         m.UpdatedAt,
	})

	return err
}

// DeleteStorePrice 在事务内删除门店价格覆盖，恢复按品牌价取值。
func (r *DishStoreRepository) DeleteStorePrice(ctx context.Context, tx db.Tx, storeID, skuID types.ID) error {
	_, err := tx.Table(dishStorePriceTable).
		Where("store_id", storeID).
		Where("sku_id", skuID).
		Delete()

	return err
}

// DistinctStoreIDs 返回配置过门店菜品数据的门店 ID，用于品牌级变更后批量失效菜单缓存。
func (r *DishStoreRepository) DistinctStoreIDs(ctx context.Context) ([]types.ID, error) {
	rows := make([]model.DishStoreSkuModel, 0)
	if err := facades.DB().WithContext(ctx).Table(dishStoreSkuTable).
		Select("store_id").
		WhereNull("deleted_at").
		GroupBy("store_id").
		Get(&rows); err != nil {
		return nil, err
	}

	ids := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.StoreID)
	}

	return ids, nil
}

// CountStoreConfigs 统计某规格被多少门店配置过，菜品删除前的完整性校验用。
func (r *DishStoreRepository) CountStoreConfigs(ctx context.Context, skuID types.ID) (int64, error) {
	return facades.DB().WithContext(ctx).Table(dishStoreSkuTable).
		Where("sku_id", skuID).
		Count()
}
