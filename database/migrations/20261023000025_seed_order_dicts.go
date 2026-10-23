package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261023000025SeedOrderDicts 写入订单模块的五个字典，取值与 app/modules/order/enums 一致。
// 订单状态用百位段位：100 起步，便于后续在状态之间插入子状态而不改历史数据。
type M20261023000025SeedOrderDicts struct{}

// Signature The unique signature for the migration.
func (r *M20261023000025SeedOrderDicts) Signature() string {
	return "20261023000025_seed_order_dicts"
}

// orderDictItem 是一个字典项：标签、取值、排序与标签颜色。
type orderDictItem struct {
	label     string
	value     string
	sort      int
	colorType string
}

// orderDictGroup 是一个字典类型及其全部字典项。
type orderDictGroup struct {
	name   string
	code   string
	remark string
	items  []orderDictItem
}

// Up 写入订单相关字典类型与字典项，幂等可重跑。
func (r *M20261023000025SeedOrderDicts) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()
	groups := []orderDictGroup{
		{
			name: "订单类型", code: "cater_order_type", remark: "堂食 / 外卖 / 自提",
			items: []orderDictItem{
				{"堂食", "1", 1, "primary"},
				{"外卖", "2", 2, "warning"},
				{"自提", "3", 3, "success"},
			},
		},
		{
			name: "下单渠道", code: "cater_order_channel", remark: "小程序 / POS / 自助机 / 外卖平台",
			items: []orderDictItem{
				{"小程序", "1", 1, "success"},
				{"POS", "2", 2, "primary"},
				{"自助机", "3", 3, "info"},
				{"外卖平台", "4", 4, "warning"},
			},
		},
		{
			name: "订单状态", code: "cater_order_status", remark: "百位段位：待支付 100 起，便于插入子状态",
			items: []orderDictItem{
				{"待支付", "100", 1, "warning"},
				{"待接单", "200", 2, "info"},
				{"制作中", "300", 3, "primary"},
				{"待取餐", "400", 4, "primary"},
				{"已完成", "500", 5, "success"},
				{"已取消", "600", 6, "info"},
				{"退款中", "700", 7, "danger"},
				{"已退款", "800", 8, "danger"},
			},
		},
		{
			name: "订单支付状态", code: "cater_pay_status", remark: "与 pay 模块对齐：0 未支付 / 10 已支付 / 20 已退款",
			items: []orderDictItem{
				{"未支付", "0", 1, "info"},
				{"已支付", "10", 2, "success"},
				{"已退款", "20", 3, "danger"},
			},
		},
		{
			name: "订单明细状态", code: "cater_item_status", remark: "正常 / 已退",
			items: []orderDictItem{
				{"正常", "1", 1, "success"},
				{"已退", "2", 2, "danger"},
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
func (r *M20261023000025SeedOrderDicts) ensureType(group orderDictGroup, now time.Time) error {
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
func (r *M20261023000025SeedOrderDicts) ensureData(code string, item orderDictItem, now time.Time) error {
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

// Down 删除本次写入的订单字典。
func (r *M20261023000025SeedOrderDicts) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	codes := []string{"cater_order_type", "cater_order_channel", "cater_order_status", "cater_pay_status", "cater_item_status"}
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
