package migrations

import (
	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	memberModel "eqray-flavor/app/modules/member/model"
)

// M20261020000036SeedMemberSignInConfig 写入默认的 1~7 天签到奖励档位。
// 第 7 天奖励最高；连续天数超过 7 天时封顶取第 7 天档位（见 MemberSignService）。
type M20261020000036SeedMemberSignInConfig struct{}

// Signature The unique signature for the migration.
func (r *M20261020000036SeedMemberSignInConfig) Signature() string {
	return "20261020000036_seed_member_sign_in_config"
}

// signInSeed 是默认档位：day / point / experience。
var signInSeed = []struct {
	Day        int
	Point      int
	Experience int
}{
	{1, 1, 5},
	{2, 1, 5},
	{3, 2, 10},
	{4, 2, 10},
	{5, 3, 15},
	{6, 3, 15},
	{7, 5, 20},
}

// Up 幂等写入签到档位，已存在的 day 直接跳过。
func (r *M20261020000036SeedMemberSignInConfig) Up() error {
	if !facades.Schema().HasTable("member_sign_in_config") {
		return nil
	}

	now := types.Now()

	for _, item := range signInSeed {
		var exist memberModel.MemberSignInConfigModel
		if err := facades.DB().Table(memberModel.MemberSignInConfigTable).
			Where("day", item.Day).
			WhereNull("deleted_at").
			First(&exist); err != nil {
			return err
		}
		if exist.ID != 0 {
			continue
		}

		if _, err := facades.DB().Table(memberModel.MemberSignInConfigTable).Insert(map[string]any{
			"id":         support.NewID(),
			"day":        item.Day,
			"point":      item.Point,
			"experience": item.Experience,
			"status":     1,
			"creator":    "system",
			"created_at": now,
			"updater":    "system",
			"updated_at": now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Down 软删默认档位。
func (r *M20261020000036SeedMemberSignInConfig) Down() error {
	if !facades.Schema().HasTable("member_sign_in_config") {
		return nil
	}

	_, err := facades.DB().Table(memberModel.MemberSignInConfigTable).
		Where("creator", "system").
		Update(map[string]any{"deleted_at": types.Now()})

	return err
}
