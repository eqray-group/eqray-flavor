package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261022000016SeedStorePrintDicts 写入打印相关字典：
// cater_printer_type、cater_printer_access、cater_print_type、cater_print_status。
// 取值与 app/modules/store/enums 保持一致，重复执行不会重复插入。
type M20261022000016SeedStorePrintDicts struct{}

// Signature The unique signature for the migration.
func (r *M20261022000016SeedStorePrintDicts) Signature() string {
	return "20261022000016_seed_store_print_dicts"
}

// printDictItem 是待写入的单个字典项。
type printDictItem struct {
	Label     string
	Value     string
	Sort      int
	ColorType string
}

// printDictDefinition 是一个字典类型及其字典项。
type printDictDefinition struct {
	Name   string
	Code   string
	Remark string
	Items  []printDictItem
}

// Up 写入打印设备、接入方式、打印类型与打印任务状态字典。
func (r *M20261022000016SeedStorePrintDicts) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()
	definitions := []printDictDefinition{
		{
			Name:   "打印机类型",
			Code:   "cater_printer_type",
			Remark: "前台小票 / 后厨 / 标签 / 叫号",
			Items: []printDictItem{
				{Label: "前台小票", Value: "1", Sort: 1, ColorType: "primary"},
				{Label: "后厨", Value: "2", Sort: 2, ColorType: "warning"},
				{Label: "标签", Value: "3", Sort: 3, ColorType: "success"},
				{Label: "叫号", Value: "4", Sort: 4, ColorType: "info"},
			},
		},
		{
			Name:   "打印机接入方式",
			Code:   "cater_printer_access",
			Remark: "云打印机 / 局域网 / 本地客户端",
			Items: []printDictItem{
				{Label: "云打印机", Value: "1", Sort: 1, ColorType: "primary"},
				{Label: "局域网", Value: "2", Sort: 2, ColorType: "warning"},
				{Label: "本地客户端", Value: "3", Sort: 3, ColorType: "info"},
			},
		},
		{
			Name:   "打印类型",
			Code:   "cater_print_type",
			Remark: "结账单 / 预结单 / 厨打 / 退菜 / 标签 / 退款单 / 交班 / 日结",
			Items: []printDictItem{
				{Label: "结账单", Value: "1", Sort: 1, ColorType: "primary"},
				{Label: "预结单", Value: "2", Sort: 2, ColorType: "primary"},
				{Label: "厨打", Value: "3", Sort: 3, ColorType: "warning"},
				{Label: "退菜", Value: "4", Sort: 4, ColorType: "danger"},
				{Label: "标签", Value: "5", Sort: 5, ColorType: "success"},
				{Label: "退款单", Value: "6", Sort: 6, ColorType: "danger"},
				{Label: "交班", Value: "7", Sort: 7, ColorType: "info"},
				{Label: "日结", Value: "8", Sort: 8, ColorType: "info"},
			},
		},
		{
			Name:   "打印任务状态",
			Code:   "cater_print_status",
			Remark: "待打印 / 成功 / 失败",
			Items: []printDictItem{
				{Label: "待打印", Value: "1", Sort: 1, ColorType: "info"},
				{Label: "成功", Value: "2", Sort: 2, ColorType: "success"},
				{Label: "失败", Value: "3", Sort: 3, ColorType: "danger"},
			},
		},
	}

	for _, definition := range definitions {
		if _, err := r.ensureType(definition, now); err != nil {
			return err
		}
		for _, item := range definition.Items {
			if err := r.ensureData(definition.Code, item, now); err != nil {
				return err
			}
		}
	}

	return nil
}

// ensureType 按编码查找字典类型，不存在时插入；返回字典类型 ID。
func (r *M20261022000016SeedStorePrintDicts) ensureType(definition printDictDefinition, now time.Time) (int64, error) {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_type").
		Select("id").
		Where("type", definition.Code).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, err
	}
	if row.ID > 0 {
		return row.ID, nil
	}

	id := int64(support.NewID())
	_, err := facades.DB().Table("system_dict_type").Insert(map[string]any{
		"id":         id,
		"name":       definition.Name,
		"type":       definition.Code,
		"status":     1,
		"remark":     definition.Remark,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return id, err
}

// ensureData 按（字典类型, 值）查找字典项，不存在时插入。
func (r *M20261022000016SeedStorePrintDicts) ensureData(code string, item printDictItem, now time.Time) error {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_data").
		Select("id").
		Where("dict_type", code).
		Where("value", item.Value).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return err
	}
	if row.ID > 0 {
		return nil
	}

	_, err := facades.DB().Table("system_dict_data").Insert(map[string]any{
		"id":         support.NewID(),
		"sort":       item.Sort,
		"label":      item.Label,
		"value":      item.Value,
		"dict_type":  code,
		"status":     1,
		"color_type": item.ColorType,
		"css_class":  "",
		"remark":     "",
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除本次写入的打印相关字典类型与字典项。
func (r *M20261022000016SeedStorePrintDicts) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	codes := []any{
		"cater_printer_type",
		"cater_printer_access",
		"cater_print_type",
		"cater_print_status",
	}
	if _, err := facades.DB().Table("system_dict_data").WhereIn("dict_type", codes).Delete(); err != nil {
		return err
	}
	_, err := facades.DB().Table("system_dict_type").WhereIn("type", codes).Delete()

	return err
}
