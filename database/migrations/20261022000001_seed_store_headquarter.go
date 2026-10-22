package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261022000001SeedStoreHeadquarter 初始化品牌总部根节点。
// 设计约定品牌总部全库唯一（design/10 §2），已有总部时直接跳过，保证迁移可重复执行。
type M20261022000001SeedStoreHeadquarter struct{}

// Signature The unique signature for the migration.
func (r *M20261022000001SeedStoreHeadquarter) Signature() string {
	return "20261022000001_seed_store_headquarter"
}

// Up 写入品牌总部根节点：不接单、休息状态、关闭线上与外卖，只作为组织树根。
func (r *M20261022000001SeedStoreHeadquarter) Up() error {
	if !facades.Schema().HasTable("store") {
		return nil
	}

	count, err := facades.DB().Table("store").
		Where("type", 1).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// 编码与已有数据冲突时不再插入，避免种子破坏唯一键
	codeCount, err := facades.DB().Table("store").
		Where("code", "BRAND_HQ").
		Count()
	if err != nil {
		return err
	}
	if codeCount > 0 {
		return nil
	}

	now := time.Now()
	_, err = facades.DB().Table("store").Insert(map[string]any{
		"id":             support.NewID(),
		"parent_id":      0,
		"type":           1,
		"code":           "BRAND_HQ",
		"name":           "品牌总部",
		"short_name":     "总部",
		"status":         2,
		"contact_name":   "",
		"contact_mobile": "",
		"district_id":    0,
		"address":        "",
		"longitude":      0,
		"latitude":       0,
		"open_time":      "00:00",
		"close_time":     "00:00",
		"delivery_rule":  "",
		"auto_accept":    0,
		"online_enabled": 0,
		"sort":           0,
		"remark":         "系统初始化：品牌总部根节点，全库唯一",
		"creator":        "system",
		"created_at":     now,
		"updater":        "system",
		"updated_at":     now,
	})

	return err
}

// Down 删除品牌总部根节点；已有下级节点时保留，避免回滚把业务数据挂空。
func (r *M20261022000001SeedStoreHeadquarter) Down() error {
	if !facades.Schema().HasTable("store") {
		return nil
	}

	row := struct {
		ID int64 `db:"id"`
	}{}
	if err := facades.DB().Table("store").
		Select("id").
		Where("code", "BRAND_HQ").
		Where("type", 1).
		First(&row); err != nil {
		return err
	}
	if row.ID == 0 {
		return nil
	}

	count, err := facades.DB().Table("store").
		Where("parent_id", row.ID).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err = facades.DB().Table("store").Where("id", row.ID).Delete()

	return err
}
