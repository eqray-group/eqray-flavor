package service

import (
	"testing"
	"time"

	"eqray-flavor/app/modules/base/types"
)

// dateTime 把标准库时间转成项目的 DateTime，便于构造生效区间用例。
func dateTime(t time.Time) *types.DateTime {
	value := types.DateTime(t)
	return &value
}

// TestResolveSkuPrice 覆盖价格取值链：时段价 > 门店会员价 > 门店价 > 品牌会员价 > 品牌标准价。
func TestResolveSkuPrice(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.Local)
	past := dateTime(now.Add(-2 * time.Hour))
	future := dateTime(now.Add(2 * time.Hour))
	expired := dateTime(now.Add(-1 * time.Hour))

	cases := []struct {
		name     string
		sku      MenuSku
		isMember bool
		want     int64
	}{
		{
			name:     "无门店价时非会员取品牌标准价",
			sku:      MenuSku{StandardPrice: 2000, BrandMemberPrice: 1800},
			isMember: false,
			want:     2000,
		},
		{
			name:     "无门店价时会员取品牌会员价",
			sku:      MenuSku{StandardPrice: 2000, BrandMemberPrice: 1800},
			isMember: true,
			want:     1800,
		},
		{
			name:     "长期门店价非会员取门店价",
			sku:      MenuSku{StandardPrice: 2000, StorePrice: 2600},
			isMember: false,
			want:     2600,
		},
		{
			name:     "长期门店价会员优先门店会员价",
			sku:      MenuSku{StandardPrice: 2000, StorePrice: 2600, StoreMemberPrice: 2300},
			isMember: true,
			want:     2300,
		},
		{
			name:     "区间内时段价最高优先",
			sku:      MenuSku{StandardPrice: 2000, StorePrice: 2600, StoreMemberPrice: 2300, StartAt: past, EndAt: future},
			isMember: true,
			want:     2600,
		},
		{
			name:     "区间内时段价对非会员同样生效",
			sku:      MenuSku{StandardPrice: 2000, StorePrice: 2600, StartAt: past, EndAt: future},
			isMember: false,
			want:     2600,
		},
		{
			name:     "区间结束后回落品牌价",
			sku:      MenuSku{StandardPrice: 2000, StorePrice: 2600, StartAt: past, EndAt: expired},
			isMember: false,
			want:     2000,
		},
		{
			name:     "区间结束后会员取品牌会员价",
			sku:      MenuSku{StandardPrice: 2000, BrandMemberPrice: 1800, StorePrice: 2600, EndAt: expired},
			isMember: true,
			want:     1800,
		},
		{
			name:     "只配门店会员价时非会员仍取品牌标准价",
			sku:      MenuSku{StandardPrice: 2000, StoreMemberPrice: 1500},
			isMember: false,
			want:     2000,
		},
		{
			name:     "只配门店会员价时会员取门店会员价",
			sku:      MenuSku{StandardPrice: 2000, StoreMemberPrice: 1500},
			isMember: true,
			want:     1500,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveSkuPrice(&tc.sku, tc.isMember, now); got != tc.want {
				t.Fatalf("resolveSkuPrice() = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestApplyMenuItemPrices 验证分类树与子分类下的规格都会被重新出价。
func TestApplyMenuItemPrices(t *testing.T) {
	now := time.Now()
	snapshot := &MenuSnapshot{
		Categories: []*MenuCategory{
			{
				ID: 1,
				Items: []MenuSpu{
					{Skus: []MenuSku{{StandardPrice: 1000, BrandMemberPrice: 900}}},
				},
				Children: []*MenuCategory{
					{
						ID: 2,
						Items: []MenuSpu{
							{Skus: []MenuSku{{StandardPrice: 500, StorePrice: 600}}},
						},
					},
				},
			},
		},
	}

	applyMenuItemPrices(snapshot, true, now)

	if got := snapshot.Categories[0].Items[0].Skus[0].Price; got != 900 {
		t.Fatalf("根分类菜品价格 = %d, want 900", got)
	}
	if got := snapshot.Categories[0].Children[0].Items[0].Skus[0].Price; got != 600 {
		t.Fatalf("子分类菜品价格 = %d, want 600", got)
	}
}
