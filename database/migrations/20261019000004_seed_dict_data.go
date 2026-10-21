package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

type M20261019000004SeedDictData struct{}

// Signature The unique signature for the migration.
func (r *M20261019000004SeedDictData) Signature() string {
	return "20261019000004_seed_dict_data"
}

// Up 初始化常用字典：通用状态、用户性别。
func (r *M20261019000004SeedDictData) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()

	types := []struct {
		name   string
		code   string
		remark string
	}{
		{"通用状态", "common_status", "启用 / 禁用"},
		{"用户性别", "user_gender", "0 未知 / 1 男 / 2 女"},
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
		{"common_status", "禁用", "0", 1, "danger"},
		{"common_status", "启用", "1", 2, "success"},
		{"user_gender", "未知", "0", 1, "info"},
		{"user_gender", "男", "1", 2, "primary"},
		{"user_gender", "女", "2", 3, "success"},
	}
	for _, item := range items {
		if err := r.ensureData(item.dictType, item.label, item.value, item.sort, item.colorType, now); err != nil {
			return err
		}
	}

	return nil
}

// ensureType 按编码查找字典类型，不存在时插入。
func (r *M20261019000004SeedDictData) ensureType(name, code, remark string, now time.Time) (int64, error) {
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

// ensureData 按 (dict_type, value) 查找字典数据，不存在时插入。
func (r *M20261019000004SeedDictData) ensureData(dictType, label, value string, sort int, colorType string, now time.Time) error {
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

// Down 清空初始化字典。
func (r *M20261019000004SeedDictData) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	if _, err := facades.DB().Table("system_dict_data").
		WhereIn("dict_type", []any{"common_status", "user_gender"}).
		Delete(); err != nil {
		return err
	}

	_, err := facades.DB().Table("system_dict_type").
		WhereIn("type", []any{"common_status", "user_gender"}).
		Delete()

	return err
}
