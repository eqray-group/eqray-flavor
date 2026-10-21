package migrations

import (
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/member/model"
)

type M20261020000017SeedMemberLevels struct{}

// Signature The unique signature for the migration.
func (r *M20261020000017SeedMemberLevels) Signature() string {
	return "20261020000017_seed_member_levels"
}

// Up 写入默认的三档会员等级，方便开箱即用地新增会员。
// 已存在同序号等级时跳过，避免重复执行冲突。
func (r *M20261020000017SeedMemberLevels) Up() error {
	if !facades.Schema().HasTable("member_levels") {
		return nil
	}

	levels := []struct {
		level       int
		name        string
		growth      int
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
			"id":          support.NewID(),
			"level":       item.level,
			"name":        item.name,
			"icon":        "",
			"growth":      item.growth,
			"discount":    item.discount,
			"description": item.description,
			"status":      1,
			"sort":        item.sort,
			"creator":     "system",
			"created_at":  now,
			"updater":     "system",
			"updated_at":  now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Down 保留种子等级：已产生会员引用，回滚时不做删除。
func (r *M20261020000017SeedMemberLevels) Down() error {
	return nil
}
