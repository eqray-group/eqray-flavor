package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261023000014SeedPayDicts 写入支付模块的五个字典，取值与 app/modules/pay/enums 一致。
// 支付字典此前只写在设计文档里，没有落库，导致前端只能靠后端返回的 statusName 展示；
// 这里补齐后，筛选下拉、标签颜色都可以走字典。种子幂等，重复执行不会重复插入。
type M20261023000014SeedPayDicts struct{}

// Signature The unique signature for the migration.
func (r *M20261023000014SeedPayDicts) Signature() string {
	return "20261023000014_seed_pay_dicts"
}

// payDictItem 是一个字典项：标签、取值、排序与标签颜色。
type payDictItem struct {
	label     string
	value     string
	sort      int
	colorType string
}

// payDictGroup 是一个字典类型及其全部字典项。
type payDictGroup struct {
	name   string
	code   string
	remark string
	items  []payDictItem
}

// Up 写入支付相关字典类型与字典项。
func (r *M20261023000014SeedPayDicts) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()
	groups := []payDictGroup{
		{
			name: "支付单状态", code: "pay_order_status", remark: "待支付 / 支付成功 / 支付关闭，对齐 yudao",
			items: []payDictItem{
				{"待支付", "0", 1, "info"},
				{"支付成功", "10", 2, "success"},
				{"支付关闭", "20", 3, "info"},
			},
		},
		{
			name: "退款单状态", code: "pay_refund_status", remark: "退款订单生成 / 退款成功 / 退款失败 / 退款关闭，对齐 yudao",
			items: []payDictItem{
				{"退款订单生成", "0", 1, "info"},
				{"退款成功", "1", 2, "success"},
				{"退款失败", "2", 3, "danger"},
				{"退款关闭", "99", 4, "info"},
			},
		},
		{
			name: "支付通知状态", code: "pay_notify_status", remark: "等待 / 成功 / 失败 / 请求成功但结果失败 / 请求失败，对齐 yudao",
			items: []payDictItem{
				{"等待通知", "1", 1, "info"},
				{"通知成功", "2", 2, "success"},
				{"通知失败", "3", 3, "danger"},
				{"请求成功但结果失败", "4", 4, "warning"},
				{"请求失败", "5", 5, "danger"},
			},
		},
		{
			name: "支付通知类型", code: "pay_notify_type", remark: "支付单 / 退款单，对齐 yudao",
			items: []payDictItem{
				{"支付单", "1", 1, "primary"},
				{"退款单", "2", 2, "warning"},
			},
		},
		{
			name: "钱包流水业务类型", code: "pay_wallet_biz_type", remark: "充值 / 充值退款 / 支付 / 支付退款 / 手工调整",
			items: []payDictItem{
				{"充值", "1", 1, "success"},
				{"充值退款", "2", 2, "warning"},
				{"支付", "3", 3, "primary"},
				{"支付退款", "4", 4, "warning"},
				{"手工调整", "5", 5, "info"},
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
func (r *M20261023000014SeedPayDicts) ensureType(group payDictGroup, now time.Time) error {
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
func (r *M20261023000014SeedPayDicts) ensureData(code string, item payDictItem, now time.Time) error {
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

// Down 删除本次写入的五个字典。
func (r *M20261023000014SeedPayDicts) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	codes := []string{"pay_order_status", "pay_refund_status", "pay_notify_status", "pay_notify_type", "pay_wallet_biz_type"}
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
