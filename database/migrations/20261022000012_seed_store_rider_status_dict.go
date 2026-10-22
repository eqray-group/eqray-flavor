package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261022000012SeedStoreRiderStatusDict 写入骑手状态字典 store_rider_status。
// 取值与 app/modules/store/enums/rider.go 保持一致，重复执行不会重复插入。
type M20261022000012SeedStoreRiderStatusDict struct{}

// Signature The unique signature for the migration.
func (r *M20261022000012SeedStoreRiderStatusDict) Signature() string {
	return "20261022000012_seed_store_rider_status_dict"
}

// Up 写入 store_rider_status 字典类型与五个字典项。
func (r *M20261022000012SeedStoreRiderStatusDict) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()
	if _, err := r.ensureType("骑手状态", "store_rider_status", "空闲 / 接单中 / 配送中 / 休息 / 停用", now); err != nil {
		return err
	}

	items := []struct {
		label     string
		value     string
		sort      int
		colorType string
	}{
		{"空闲", "1", 1, "info"},
		{"接单中", "2", 2, "success"},
		{"配送中", "3", 3, "primary"},
		{"休息", "4", 4, "warning"},
		{"停用", "5", 5, "danger"},
	}
	for _, item := range items {
		if err := r.ensureData(item.label, item.value, item.sort, item.colorType, now); err != nil {
			return err
		}
	}

	return nil
}

// ensureType 按编码查找字典类型，不存在时插入；返回字典类型 ID。
func (r *M20261022000012SeedStoreRiderStatusDict) ensureType(name, code, remark string, now time.Time) (int64, error) {
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
func (r *M20261022000012SeedStoreRiderStatusDict) ensureData(label, value string, sort int, colorType string, now time.Time) error {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_data").
		Select("id").
		Where("dict_type", "store_rider_status").
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
		"dict_type":  "store_rider_status",
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

// Down 删除 store_rider_status 字典类型与字典项。
func (r *M20261022000012SeedStoreRiderStatusDict) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	if _, err := facades.DB().Table("system_dict_data").
		Where("dict_type", "store_rider_status").
		Delete(); err != nil {
		return err
	}

	_, err := facades.DB().Table("system_dict_type").
		Where("type", "store_rider_status").
		Delete()

	return err
}
