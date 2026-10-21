package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"eqray-flavor/database/migrations"
)

// Migrations 返回全部迁移，顺序即执行顺序。
// 约定：一个表对应一个 create_*_table 迁移；数据填充（菜单、初始账号、默认配置等）
// 单独放 seed_* 迁移，不混在建表迁移里。
func Migrations() []schema.Migration {
	return []schema.Migration{
		// ---------- system / base ----------
		&migrations.M20261017000001CreateSystemUsersTable{},
		&migrations.M20261017000002CreateSystemMenuTable{},
		&migrations.M20261017000003CreateSystemRoleTable{},
		&migrations.M20261017000004CreateSystemUserRoleTable{},
		&migrations.M20261017000005CreateSystemRoleMenuTable{},
		&migrations.M20261017000006SeedSystemMenu{},
		&migrations.M20261017000007SeedSystemUsers{},
		&migrations.M20261017000008SeedSystemRole{},
		&migrations.M20261017000009SeedMenuIconAndDashboard{},
		&migrations.M20261017000010SeedSystemUserRole{},
		&migrations.M20261019000001CreateAuthLoginLogTable{},
		&migrations.M20261019000002CreateSystemDictTypeTable{},
		&migrations.M20261019000003CreateSystemDictDataTable{},
		&migrations.M20261019000004SeedDictData{},
		&migrations.M20261019000005SeedDictAndLoginLogMenu{},

		// ---------- infra ----------
		&migrations.M20261020000001CreateInfraFileTable{},
		&migrations.M20261020000002CreateSystemOperateLogTable{},
		&migrations.M20261020000003SeedOperateLogMenu{},
		&migrations.M20261020000004CreateInfraConfigTable{},
		&migrations.M20261020000005SeedInfraConfigMenu{},
		&migrations.M20261020000006SeedInfraFileMenu{},
		&migrations.M20261020000007CreateInfraFileConfigTable{},
		&migrations.M20261020000008CreateInfraApiAccessLogTable{},
		&migrations.M20261020000009CreateInfraApiErrorLogTable{},
		&migrations.M20261020000012SeedInfraOpsMenus{},
		&migrations.M20261020000013SeedInfraFileConfig{},
		&migrations.M20261020000030CreateInfraSmsChannelTable{},
		&migrations.M20261020000031CreateInfraSmsTemplateTable{},
		&migrations.M20261020000032SeedInfraSms{},
		&migrations.M20261020000033SeedInfraSmsMenu{},

		// ---------- member ----------
		&migrations.M20261020000014CreateMemberUserTable{},
		&migrations.M20261020000015SeedMemberMenu{},
		&migrations.M20261020000016CreateMemberLevelTable{},
		&migrations.M20261020000017SeedMemberLevels{},
		&migrations.M20261020000018CreateMemberPointRecordTable{},
		&migrations.M20261020000019SeedMemberOpsMenu{},
		&migrations.M20261020000020CreateMemberLevelRecordTable{},
		&migrations.M20261020000021CreateMemberSignRecordTable{},
		&migrations.M20261020000022SeedMemberTagGroupMenu{},
		&migrations.M20261020000023CreateMemberTagTable{},
		&migrations.M20261020000024CreateMemberGroupTable{},
		&migrations.M20261020000025CreateMemberSocialClientTable{},
		&migrations.M20261020000026SeedMemberSocialClientMenu{},
		&migrations.M20261020000027CreateMemberAddressTable{},
		&migrations.M20261020000028CreateMemberSocialUserTable{},
		&migrations.M20261020000029CreateMemberSocialBindingTable{},
		&migrations.M20261020000034CreateMemberSignInConfigTable{},
		&migrations.M20261020000036SeedMemberSignInConfig{},
		&migrations.M20261020000037SeedMemberSignConfigMenu{},
	}
}
