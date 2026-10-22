package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261022000002SeedStoreDicts 初始化 store 模板块需要的业务字典。
// 取值与 app/modules/store/enums/store.go 保持一致，重复执行不会重复插入。
type M20261022000002SeedStoreDicts struct{}

// Signature The unique signature for the migration.
func (r *M20261022000002SeedStoreDicts) Signature() string {
	return "20261022000002_seed_store_dicts"
}

// Up 写入 store_type / store_status 字典类型与字典项。
func (r *M20261022000002SeedStoreDicts) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()

	types := []struct {
		name   string
		code   string
		remark string
	}{
		{"门店类型", "store_type", "总部 / 大区 / 门店"},
		{"门店经营状态", "store_status", "营业 / 休息 / 装修 / 停业"},
	}
	for _, item := range types {
		if _, err := r.ensureType(item.name, item.code, item.remark, now); err != nil {
			return err
		}
	}

	items := []struct {
		dictType  string
		label     string
		value     string
		sort      int
		colorType string
	}{
		{"store_type", "总部", "1", 1, "primary"},
		{"store_type", "大区", "2", 2, "warning"},
		{"store_type", "门店", "3", 3, "success"},
		{"store_status", "营业", "1", 1, "success"},
		{"store_status", "休息", "2", 2, "info"},
		{"store_status", "装修", "3", 3, "warning"},
		{"store_status", "停业", "4", 4, "danger"},
	}
	for _, item := range items {
		if err := r.ensureData(item.dictType, item.label, item.value, item.sort, item.colorType, now); err != nil {
			return err
		}
	}

	return nil
}

// ensureType 按编码查找字典类型，不存在时插入。
// 入参：name 名称、code 编码、remark 备注、now 时间；返回字典类型 ID。
func (r *M20261022000002SeedStoreDicts) ensureType(name, code, remark string, now time.Time) (int64, error) {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_type").
		Select("id").
		Where("type", code).
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
		"name":       name,
		"type":       code,
		"status":     1,
		"remark":     remark,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return id, err
}

// ensureData 按（字典类型, 值）查找字典项，不存在时插入。
// 入参：dictType 字典编码、label/value/sort/colorType 字典项字段、now 时间；返回错误。
func (r *M20261022000002SeedStoreDicts) ensureData(dictType, label, value string, sort int, colorType string, now time.Time) error {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_data").
		Select("id").
		Where("dict_type", dictType).
		Where("value", value).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return err
	}
	if row.ID > 0 {
		return nil
	}

	_, err := facades.DB().Table("system_dict_data").Insert(map[string]any{
		"id":         support.NewID(),
		"sort":       sort,
		"label":      label,
		"value":      value,
		"dict_type":  dictType,
		"status":     1,
		"color_type": colorType,
		"css_class":  "",
		"remark":     "",
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除 store 模块字典；只删除本迁移写入的编码。
func (r *M20261022000002SeedStoreDicts) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	if _, err := facades.DB().Table("system_dict_data").
		WhereIn("dict_type", []any{"store_type", "store_status"}).
		Delete(); err != nil {
		return err
	}

	_, err := facades.DB().Table("system_dict_type").
		WhereIn("type", []any{"store_type", "store_status"}).
		Delete()

	return err
}
