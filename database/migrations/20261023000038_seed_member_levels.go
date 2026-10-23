package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/member/model"
)

// M20261023000038SeedMemberLevels 补写默认会员等级。
// 背景：20261020000017 误用旧表名 member_levels 与旧列名 growth / discount，
// 建表迁移改为 member_level 后该种子静默跳过，等级表一直是空的，
// 导致新增会员因「等级必须存在且启用」而无法保存。本迁移按最终表结构补齐三档等级。
type M20261023000038SeedMemberLevels struct{}

// Signature The unique signature for the migration.
func (r *M20261023000038SeedMemberLevels) Signature() string {
	return "20261023000038_seed_member_levels"
}

// Up 写入默认的三档会员等级：按 level 唯一，存在则跳过，保证可重复执行。
func (r *M20261023000038SeedMemberLevels) Up() error {
	if !facades.Schema().HasTable(model.MemberLevelTable) {
		return nil
	}

	levels := []struct {
		level       int
		name        string
		experience  int
		discount    int
		description string
		sort        int
	}{
		{1, "普通会员", 0, 100, "注册即可获得的默认等级", 1},
		{2, "银卡会员", 1000, 98, "成长值满 1000 自动升级，享 98 折", 2},
		{3, "金卡会员", 5000, 95, "成长值满 5000 自动升级，享 95 折", 3},
	}

	now := time.Now()

	for _, item := range levels {
		var existing model.MemberLevelModel
		if err := facades.DB().Table(model.MemberLevelTable).
			Select("id").
			Where("level", item.level).
			First(&existing); err != nil {
			return err
		}
		if existing.ID > 0 {
			continue
		}

		if _, err := facades.DB().Table(model.MemberLevelTable).Insert(map[string]any{
			"id":               support.NewID(),
			"level":            item.level,
			"name":             item.name,
			"icon":             "",
			"experience":       item.experience,
			"discount_percent": item.discount,
			"description":      item.description,
			"status":           1,
			"sort":             item.sort,
			"creator":          "system",
			"created_at":       now,
			"updater":          "system",
			"updated_at":       now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Down 保留种子等级：已产生会员引用，回滚时不做删除。
func (r *M20261023000038SeedMemberLevels) Down() error {
	return nil
}
