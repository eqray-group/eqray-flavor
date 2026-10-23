package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261023000036SeedPromotionDicts 写入 promotion 模块的字典，取值与 app/modules/promotion/enums 一致。
type M20261023000036SeedPromotionDicts struct{}

// Signature The unique signature for the migration.
func (r *M20261023000036SeedPromotionDicts) Signature() string {
	return "20261023000036_seed_promotion_dicts"
}

// promotionDictItem 是一个字典项：标签、取值、排序与标签颜色。
type promotionDictItem struct {
	label     string
	value     string
	sort      int
	colorType string
}

// promotionDictGroup 是一个字典类型及其全部字典项。
type promotionDictGroup struct {
	name   string
	code   string
	remark string
	items  []promotionDictItem
}

// Up 写入 promotion 相关字典类型与字典项，幂等可重跑。
func (r *M20261023000036SeedPromotionDicts) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()
	groups := []promotionDictGroup{
		{
			name: "优惠券类型", code: "promotion_coupon_type", remark: "满减 / 折扣 / 单品 / 兑换 / 代金",
			items: []promotionDictItem{
				{"满减券", "1", 1, "primary"},
				{"折扣券", "2", 2, "success"},
				{"单品券", "3", 3, "warning"},
				{"兑换券", "4", 4, "info"},
				{"代金券", "5", 5, "danger"},
			},
		},
		{
			name: "优惠券状态", code: "promotion_coupon_status", remark: "未使用 / 锁定 / 已使用 / 已过期",
			items: []promotionDictItem{
				{"未使用", "1", 1, "success"},
				{"锁定", "2", 2, "warning"},
				{"已使用", "3", 3, "info"},
				{"已过期", "4", 4, "info"},
			},
		},
		{
			name: "优惠券适用范围", code: "promotion_coupon_scope", remark: "全品牌 / 指定门店 / 指定商品",
			items: []promotionDictItem{
				{"全品牌", "1", 1, "primary"},
				{"指定门店", "2", 2, "success"},
				{"指定商品", "3", 3, "warning"},
			},
		},
		{
			name: "优惠券渠道限制", code: "promotion_coupon_channel", remark: "不限 / 仅线上 / 仅线下",
			items: []promotionDictItem{
				{"不限渠道", "0", 1, "info"},
				{"仅线上", "1", 2, "primary"},
				{"仅线下", "2", 3, "warning"},
			},
		},
		{
			name: "优惠券发放方式", code: "promotion_coupon_issue_type", remark: "主动领取 / 消费返券 / 新人礼 / 定向发券",
			items: []promotionDictItem{
				{"主动领取", "1", 1, "primary"},
				{"消费返券", "2", 2, "success"},
				{"新人礼", "3", 3, "warning"},
				{"定向发券", "4", 4, "info"},
			},
		},
		{
			name: "营销活动玩法", code: "promotion_activity_type", remark: "满减 / 第二份半价 / 限时折扣 / 套餐特惠 / 加价购",
			items: []promotionDictItem{
				{"满减", "1", 1, "primary"},
				{"第二份半价", "2", 2, "success"},
				{"限时折扣", "3", 3, "warning"},
				{"套餐特惠", "4", 4, "info"},
				{"加价购", "5", 5, "danger"},
			},
		},
		{
			name: "营销活动层级", code: "promotion_activity_level", remark: "商品级 / 订单级",
			items: []promotionDictItem{
				{"商品级", "1", 1, "primary"},
				{"订单级", "2", 2, "success"},
			},
		},
		{
			name: "营销活动范围", code: "promotion_activity_scope", remark: "全品牌 / 指定门店",
			items: []promotionDictItem{
				{"全品牌", "1", 1, "primary"},
				{"指定门店", "2", 2, "success"},
			},
		},
		{
			name: "营销活动名额周期", code: "promotion_activity_quota_cycle", remark: "整个活动期 / 每日重置",
			items: []promotionDictItem{
				{"整个活动期", "1", 1, "primary"},
				{"每日重置", "2", 2, "warning"},
			},
		},
		{
			name: "营销活动状态", code: "promotion_activity_status", remark: "未启用 / 已启用",
			items: []promotionDictItem{
				{"已启用", "1", 1, "success"},
				{"未启用", "0", 2, "info"},
			},
		},
	}

	for _, group := range groups {
		if err := r.ensureType(group, now); err != nil {
			return err
		}
		for _, item := range group.items {
			if err := r.ensureData(group.code, item, now); err != nil {
				return err
			}
		}
	}

	return nil
}

// ensureType 按编码查找字典类型，不存在时插入。
func (r *M20261023000036SeedPromotionDicts) ensureType(group promotionDictGroup, now time.Time) error {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_type").
		Select("id").
		Where("type", group.code).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return err
	}
	if row.ID > 0 {
		return nil
	}

	_, err := facades.DB().Table("system_dict_type").Insert(map[string]any{
		"id":         int64(support.NewID()),
		"name":       group.name,
		"type":       group.code,
		"status":     1,
		"remark":     group.remark,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// ensureData 按（字典类型, 取值）查找字典项，不存在时插入。
func (r *M20261023000036SeedPromotionDicts) ensureData(code string, item promotionDictItem, now time.Time) error {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_data").
		Select("id").
		Where("dict_type", code).
		Where("value", item.value).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return err
	}
	if row.ID > 0 {
		return nil
	}

	_, err := facades.DB().Table("system_dict_data").Insert(map[string]any{
		"id":         support.NewID(),
		"sort":       item.sort,
		"label":      item.label,
		"value":      item.value,
		"dict_type":  code,
		"status":     1,
		"color_type": item.colorType,
		"css_class":  "",
		"remark":     "",
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除本次写入的 promotion 字典。
func (r *M20261023000036SeedPromotionDicts) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	codes := []string{
		"promotion_coupon_type", "promotion_coupon_status", "promotion_coupon_scope",
		"promotion_coupon_channel", "promotion_coupon_issue_type",
		"promotion_activity_type", "promotion_activity_level", "promotion_activity_scope",
		"promotion_activity_quota_cycle", "promotion_activity_status",
	}
	for _, code := range codes {
		if _, err := facades.DB().Table("system_dict_data").Where("dict_type", code).Delete(); err != nil {
			return err
		}
		if _, err := facades.DB().Table("system_dict_type").Where("type", code).Delete(); err != nil {
			return err
		}
	}

	return nil
}
