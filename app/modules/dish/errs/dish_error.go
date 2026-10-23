// Package errs 定义 dish 模块的业务错误码。
// 编码规则：1_{模块段}_{子模块}_{序号}，dish 模块段为 005。
package errs

import baseErrs "eqray-flavor/app/modules/base/errs"

// dish 模块业务错误码：1-005-{子模块}-{序号}。
// 子模块占用：category 分类 001、spu 菜品 002、sku 规格 003、addon 加料 004、store 门店菜单 005。
var (
	// DishCategoryNotExists 分类不存在或已删除
	DishCategoryNotExists = baseErrs.ErrorCode{Code: 1_005_001_001, Msg: "菜品分类不存在"}
	// DishCategoryExists 同一上级下分类名称重复
	DishCategoryExists = baseErrs.ErrorCode{Code: 1_005_001_002, Msg: "分类名称【%v】已存在"}
	// DishCategoryHasChildren 仍有下级分类时禁止删除
	DishCategoryHasChildren = baseErrs.ErrorCode{Code: 1_005_001_003, Msg: "分类【%v】下仍有子分类，不可删除"}
	// DishCategoryHasDish 分类下仍有菜品时禁止删除
	DishCategoryHasDish = baseErrs.ErrorCode{Code: 1_005_001_004, Msg: "分类【%v】下仍有【%v】个菜品，请先移动或删除"}
	// DishCategoryParentCycle 上级分类指向自身或自身下级，会形成环
	DishCategoryParentCycle = baseErrs.ErrorCode{Code: 1_005_001_005, Msg: "上级分类不能是自身或下级分类"}
	// DishCategoryStatusInvalid 分类启停状态取值不合法
	DishCategoryStatusInvalid = baseErrs.ErrorCode{Code: 1_005_001_006, Msg: "分类状态取值不合法"}

	// DishSpuNotExists 菜品不存在或已删除
	DishSpuNotExists = baseErrs.ErrorCode{Code: 1_005_002_001, Msg: "菜品不存在"}
	// DishSpuNameExists 同一分类下菜品名称重复
	DishSpuNameExists = baseErrs.ErrorCode{Code: 1_005_002_002, Msg: "菜品名称【%v】已存在"}
	// DishSpuCodeExists 菜品编码重复，编码非空时全局唯一
	DishSpuCodeExists = baseErrs.ErrorCode{Code: 1_005_002_003, Msg: "菜品编码【%v】已存在"}
	// DishSpuStatusInvalid 菜品上下架状态取值不合法
	DishSpuStatusInvalid = baseErrs.ErrorCode{Code: 1_005_002_004, Msg: "菜品上下架状态取值不合法"}
	// DishSpuCategoryNotExists 菜品所属分类不存在
	DishSpuCategoryNotExists = baseErrs.ErrorCode{Code: 1_005_002_005, Msg: "所属分类不存在"}
	// DishSpuHasSku 菜品下仍有规格时禁止直接删除
	DishSpuHasSku = baseErrs.ErrorCode{Code: 1_005_002_006, Msg: "菜品【%v】下仍有规格，不可删除"}

	// DishSkuNotExists 规格不存在或已删除
	DishSkuNotExists = baseErrs.ErrorCode{Code: 1_005_003_001, Msg: "规格不存在"}
	// DishSkuNameExists 同一菜品下规格名称重复
	DishSkuNameExists = baseErrs.ErrorCode{Code: 1_005_003_002, Msg: "规格名称【%v】已存在"}
	// DishSkuCodeExists 规格编码重复，编码非空时全局唯一
	DishSkuCodeExists = baseErrs.ErrorCode{Code: 1_005_003_003, Msg: "规格编码【%v】已存在"}
	// DishSkuPriceInvalid 标准价必须大于 0，会员价与成本价不能为负
	DishSkuPriceInvalid = baseErrs.ErrorCode{Code: 1_005_003_004, Msg: "规格价格不合法：标准价必须大于 0，会员价与成本价不能为负"}
	// DishSkuStatusInvalid 规格上下架状态取值不合法
	DishSkuStatusInvalid = baseErrs.ErrorCode{Code: 1_005_003_005, Msg: "规格上下架状态取值不合法"}
	// DishSkuRequired 菜品至少保留一个规格
	DishSkuRequired = baseErrs.ErrorCode{Code: 1_005_003_006, Msg: "菜品至少需要保留一个规格"}

	// DishAddonGroupNotExists 加料组不存在或已删除
	DishAddonGroupNotExists = baseErrs.ErrorCode{Code: 1_005_004_001, Msg: "加料组不存在"}
	// DishAddonGroupExists 加料组名称重复，名称全局唯一
	DishAddonGroupExists = baseErrs.ErrorCode{Code: 1_005_004_002, Msg: "加料组【%v】已存在"}
	// DishAddonSelectRangeInvalid 加料选择数量范围不合法
	DishAddonSelectRangeInvalid = baseErrs.ErrorCode{Code: 1_005_004_003, Msg: "加料选择数量不合法：必选时最少为 1，最多不小于最少"}
	// DishAddonItemNotExists 加料项不存在或已删除
	DishAddonItemNotExists = baseErrs.ErrorCode{Code: 1_005_004_004, Msg: "加料项不存在"}
	// DishAddonItemExists 同一加料组内加料项名称重复
	DishAddonItemExists = baseErrs.ErrorCode{Code: 1_005_004_005, Msg: "加料项【%v】已存在"}
	// DishAddonItemPriceInvalid 加料项加价不能为负
	DishAddonItemPriceInvalid = baseErrs.ErrorCode{Code: 1_005_004_006, Msg: "加料项加价不能为负数"}
	// DishAddonGroupInUse 加料组已挂到菜品时禁止删除
	DishAddonGroupInUse = baseErrs.ErrorCode{Code: 1_005_004_007, Msg: "加料组【%v】已挂到【%v】个菜品，不可删除"}

	// DishStoreSkuExists 该门店已配置过此规格
	DishStoreSkuExists = baseErrs.ErrorCode{Code: 1_005_005_001, Msg: "该门店已存在此规格的可售配置"}
	// DishStoreSkuNotExists 门店可售配置不存在
	DishStoreSkuNotExists = baseErrs.ErrorCode{Code: 1_005_005_002, Msg: "门店可售配置不存在"}
	// DishStorePriceInvalid 门店价必须大于 0
	DishStorePriceInvalid = baseErrs.ErrorCode{Code: 1_005_005_003, Msg: "门店价必须大于 0，不覆盖时请留空"}
	// DishStorePriceRangeInvalid 门店价生效区间不合法
	DishStorePriceRangeInvalid = baseErrs.ErrorCode{Code: 1_005_005_004, Msg: "生效开始时间必须早于结束时间"}
	// DishStorePriceFloatExceeded 门店价超出品牌价的浮动区间
	DishStorePriceFloatExceeded = baseErrs.ErrorCode{Code: 1_005_005_005, Msg: "门店价【%v】超出品牌价【%v】允许的浮动区间"}
	// DishSkuNotSellable 规格未在门店上架，不能加入菜单
	DishSkuNotSellable = baseErrs.ErrorCode{Code: 1_005_005_006, Msg: "该规格未在当前门店上架"}
)
