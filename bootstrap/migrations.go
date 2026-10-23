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
		&migrations.M20261022000018CreateInfraDistrictTable{},
		&migrations.M20261022000019SeedInfraDistrictMenu{},
		&migrations.M20261022000020SeedInfraAmapConfig{},

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
		// 补写会员等级：早期种子误用旧表名 member_levels，从未落库
		&migrations.M20261023000038SeedMemberLevels{},

		// ---------- store ----------
		&migrations.M20261021000001CreateStoreTable{},
		&migrations.M20261021000002SeedStoreMenu{},
		&migrations.M20261022000001SeedStoreHeadquarter{},
		&migrations.M20261022000002SeedStoreDicts{},
		&migrations.M20261022000003AlterStoreAddDeliveryEnabled{},
		&migrations.M20261022000004CreateStoreEmployeeTable{},
		&migrations.M20261022000005SeedStoreRoleDict{},
		&migrations.M20261022000006SeedStoreEmployeeMenu{},
		&migrations.M20261022000007CreateStoreTableAreaTable{},
		&migrations.M20261022000008CreateStoreTableTable{},
		&migrations.M20261022000009SeedStoreTableStatusDict{},
		&migrations.M20261022000010SeedStoreTableMenu{},
		&migrations.M20261022000011CreateStoreRiderTable{},
		&migrations.M20261022000012SeedStoreRiderStatusDict{},
		&migrations.M20261022000013SeedStoreRiderMenu{},
		&migrations.M20261022000014CreateStorePrinterTable{},
		&migrations.M20261022000015CreateStorePrintTemplateTable{},
		&migrations.M20261022000016SeedStorePrintDicts{},
		&migrations.M20261022000017SeedStorePrintMenu{},
		// 打印任务：设备与模板之前只配不用，这里补上任务表与后台重打入口
		&migrations.M20261023000049CreateStorePrintTaskTable{},
		&migrations.M20261023000050SeedStorePrintTaskMenu{},

		// ---------- dish ----------
		&migrations.M20261022000021CreateDishCategoryTable{},
		&migrations.M20261022000022CreateDishSpuTable{},
		&migrations.M20261022000023CreateDishSkuTable{},
		&migrations.M20261022000024CreateDishAddonGroupTable{},
		&migrations.M20261022000025CreateDishAddonItemTable{},
		&migrations.M20261022000026CreateDishSpuAddonTable{},
		&migrations.M20261022000027CreateDishStoreSkuTable{},
		&migrations.M20261022000028CreateDishStorePriceTable{},
		&migrations.M20261022000029SeedDishMenu{},
		&migrations.M20261022000030SeedDishPriceFloatConfig{},
		&migrations.M20261022000031SeedDishStoreSkuMenu{},
		&migrations.M20261022000032EnableDishMenu{},
		// ---------- pay（yudao 结构） ----------
		&migrations.M20261022000060DropPayTablesBeforeYudao{},
		&migrations.M20261022000061CreatePayAppTable{},
		&migrations.M20261022000062CreatePayChannelTable{},
		&migrations.M20261022000063CreatePayOrderTable{},
		&migrations.M20261022000064CreatePayOrderExtensionTable{},
		&migrations.M20261022000065CreatePayRefundTable{},
		&migrations.M20261022000066CreatePayNotifyTaskTable{},
		&migrations.M20261022000067CreatePayNotifyLogTable{},
		&migrations.M20261022000068CreatePayTransferTable{},
		&migrations.M20261022000069CreatePayWalletTable{},
		&migrations.M20261022000070CreatePayWalletTransactionTable{},
		&migrations.M20261022000071CreatePayWalletRechargePackageTable{},
		&migrations.M20261022000072CreatePayWalletRechargeTable{},
		&migrations.M20261022000073RemovePayMenusBeforeYudao{},
		&migrations.M20261023000001SeedPayMenu{},
		&migrations.M20261023000002SeedPayRefundMenu{},
		&migrations.M20261023000003SeedPayNotifyMenu{},
		&migrations.M20261023000004SeedPayOrderSyncMenu{},
		&migrations.M20261023000005SeedPayWalletMenu{},
		// 对齐 yudao 口径：补字段、改枚举类型、修正历史状态值、补字典
		&migrations.M20261023000011AlterPayWalletTransactionBizType{},
		&migrations.M20261023000012AlignPayStatusValues{},
		&migrations.M20261023000013AddPayCompositeIndexes{},
		&migrations.M20261023000014SeedPayDicts{},
		&migrations.M20261023000015WidenPayOrderTextColumns{},

		// 订单模块：交易域核心表、字典与菜单
		&migrations.M20261023000020CreateCaterOrderTable{},
		&migrations.M20261023000021CreateCaterOrderItemTable{},
		&migrations.M20261023000022CreateCaterOrderItemAddonTable{},
		&migrations.M20261023000023CreateCaterOrderLogTable{},
		&migrations.M20261023000024CreateCaterCartTable{},
		&migrations.M20261023000025SeedOrderDicts{},
		&migrations.M20261023000026SeedOrderMenu{},
		// 放宽订单状态列：百位段位状态超出 tinyint 范围
		&migrations.M20261023000039WidenOrderStatusColumns{},
		// 补记订单实际使用的积分数：结算与退款都按此数出入，避免兑换比例调整后错账
		&migrations.M20261023000041AddOrderPointsUsed{},
		// 补放订单日志原状态列：上一个迁移漏改 from_status，写日志会溢出
		&migrations.M20261023000042WidenOrderLogFromStatus{},
		// 订单后台经营字段：商家备注、改价差额、骑手与派单时间
		&migrations.M20261023000044AddOrderAdminFields{},
		// 备注 / 改价 / 派单三个按钮权限点
		&migrations.M20261023000045SeedOrderOpsMenu{},
		// 会员收货地址后台维护按钮
		&migrations.M20261023000046SeedMemberAddressMenu{},
		// 退款审核字段与审核按钮：退款改为可先申请、再审核放款
		&migrations.M20261023000047AddPayRefundAuditFields{},
		&migrations.M20261023000048SeedPayRefundAuditMenu{},

		// 营销模块：优惠券模板、会员券、活动与名额、积分规则，附字典与菜单
		&migrations.M20261023000030CreatePromotionCouponTemplateTable{},
		&migrations.M20261023000031CreatePromotionCouponTable{},
		&migrations.M20261023000032CreatePromotionActivityTable{},
		&migrations.M20261023000033CreatePromotionActivityProductTable{},
		&migrations.M20261023000034CreatePromotionActivityQuotaTable{},
		&migrations.M20261023000035CreatePromotionPointRuleTable{},
		&migrations.M20261023000036SeedPromotionDicts{},
		&migrations.M20261023000037SeedPromotionMenu{},
		// 订单占用活动名额的关联表：名额计数行无法反查订单，跨天释放会退错池子
		&migrations.M20261023000040CreatePromotionOrderActivityTable{},
		// 积分抵扣规则页面：接口早已就绪，但一直没有菜单入口
		&migrations.M20261023000043SeedPromotionPointRuleMenu{},
	}
}
