package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261022000005SeedStoreRoleDict 写入门店员工角色字典 store_role_type。
// 取值与 app/modules/store/enums/employee.go 保持一致，重复执行不会重复插入。
type M20261022000005SeedStoreRoleDict struct{}

// Signature The unique signature for the migration.
func (r *M20261022000005SeedStoreRoleDict) Signature() string {
	return "20261022000005_seed_store_role_dict"
}

// Up 写入 store_role_type 字典类型与五个角色字典项。
func (r *M20261022000005SeedStoreRoleDict) Up() error {
	if !facades.Schema().HasTable("system_dict_type") || !facades.Schema().HasTable("system_dict_data") {
		return nil
	}

	now := time.Now()
	if _, err := r.ensureType("门店员工角色", "store_role_type", "店长 / 收银 / 后厨 / 服务员 / 骑手", now); err != nil {
		return err
	}

	items := []struct {
		label     string
		value     string
		sort      int
		colorType string
	}{
		{"店长", "1", 1, "primary"},
		{"收银", "2", 2, "success"},
		{"后厨", "3", 3, "warning"},
		{"服务员", "4", 4, "info"},
		{"骑手", "5", 5, "danger"},
	}
	for _, item := range items {
		if err := r.ensureData(item.label, item.value, item.sort, item.colorType, now); err != nil {
			return err
		}
	}

	return nil
}

// ensureType 按编码查找字典类型，不存在时插入；返回字典类型 ID。
func (r *M20261022000005SeedStoreRoleDict) ensureType(name, code, remark string, now time.Time) (int64, error) {
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
func (r *M20261022000005SeedStoreRoleDict) ensureData(label, value string, sort int, colorType string, now time.Time) error {
	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("system_dict_data").
		Select("id").
		Where("dict_type", "store_role_type").
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
		"dict_type":  "store_role_type",
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

// Down 删除 store_role_type 字典类型与字典项。
func (r *M20261022000005SeedStoreRoleDict) Down() error {
	if !facades.Schema().HasTable("system_dict_data") || !facades.Schema().HasTable("system_dict_type") {
		return nil
	}

	if _, err := facades.DB().Table("system_dict_data").
		Where("dict_type", "store_role_type").
		Delete(); err != nil {
		return err
	}

	_, err := facades.DB().Table("system_dict_type").
		Where("type", "store_role_type").
		Delete()

	return err
}
