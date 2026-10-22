package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261022000009SeedStoreTableStatusDict 写入桌台状态字典 table_status。
// 取值与 app/modules/store/enums/table.go 保持一致，重复执行不会重复插入。
type M20261022000009SeedStoreTableStatusDict struct{}

// Signature The unique signature for the migration.
func (r *M20261022000009SeedStoreTableStatusDict) Signature() string {
	return "20261022000009_seed_store_table_status_dict"
}

// Up 写入 table_status 字典类型与空闲 / 使用中 / 待清台三个字典项。
func (r *M20261022000009SeedStoreTableStatusDict) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()
	if _, err := r.ensureType("桌台状态", "table_status", "空闲 / 使用中 / 待清台", now); err != nil {
		return err
	}

	items := []struct {
		label     string
		value     string
		sort      int
		colorType string
	}{
		{"空闲", "1", 1, "info"},
		{"使用中", "2", 2, "success"},
		{"待清台", "3", 3, "warning"},
	}
	for _, item := range items {
		if err := r.ensureData(item.label, item.value, item.sort, item.colorType, now); err != nil {
			return err
		}
	}

	return nil
}

// ensureType 按编码查找字典类型，不存在时插入；返回字典类型 ID。
func (r *M20261022000009SeedStoreTableStatusDict) ensureType(name, code, remark string, now time.Time) (int64, error) {
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
func (r *M20261022000009SeedStoreTableStatusDict) ensureData(label, value string, sort int, colorType string, now time.Time) error {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_data").
		Select("id").
		Where("dict_type", "table_status").
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
		"dict_type":  "table_status",
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

// Down 删除 table_status 字典类型与字典项。
func (r *M20261022000009SeedStoreTableStatusDict) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	if _, err := facades.DB().Table("system_dict_data").
		Where("dict_type", "table_status").
		Delete(); err != nil {
		return err
	}

	_, err := facades.DB().Table("system_dict_type").
		Where("type", "table_status").
		Delete()

	return err
}
