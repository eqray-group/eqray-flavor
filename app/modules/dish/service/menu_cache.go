package service

import (
	"context"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/dish/repository"
)

const (
	// menuCachePrefix 是门店菜单缓存键前缀，完整键为 dish:menu:{storeId}（docs 15 §2）。
	menuCachePrefix = "dish:menu:"
	// menuCacheTTL 是门店菜单缓存过期时间，取设计区间 10~30 分钟的中间值。
	menuCacheTTL = 20 * time.Minute
)

// storeMenuCacheKey 拼出门店菜单缓存键。
func storeMenuCacheKey(storeID types.ID) string {
	return menuCachePrefix + storeID.String()
}

// InvalidateStoreMenuCache 清除指定门店的菜单缓存；改价、上下架、沽清后必须调用。
func InvalidateStoreMenuCache(storeIDs ...types.ID) {
	for _, storeID := range storeIDs {
		if storeID <= 0 {
			continue
		}
		_ = facades.Cache().Forget(storeMenuCacheKey(storeID))
	}
}

// InvalidateAllMenus 失效所有已配置门店的菜单缓存。
// 品牌级数据（菜品、分类、加料组）变更会影响多个门店，无法定位到单个门店时使用；
// 只清理配置过门店菜品的门店，避免全表扫描缓存键。
func InvalidateAllMenus(ctx context.Context) {
	storeIDs, err := repository.NewDishStoreRepository().DistinctStoreIDs(ctx)
	if err != nil || len(storeIDs) == 0 {
		return
	}

	InvalidateStoreMenuCache(storeIDs...)
}
