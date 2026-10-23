// Package errs 定义 order 模块的业务错误码。
// 编码规则：1_{模块段}_{子模块}_{序号}，order 模块段为 006。
package errs

import baseErrs "eqray-flavor/app/modules/base/errs"

// order 模块业务错误码：1-006-{子模块}-{序号}。
// 子模块占用：order 订单 001、cart 购物车 002。
var (
	// OrderNotExists 订单不存在或已删除
	OrderNotExists = baseErrs.ErrorCode{Code: 1_006_001_001, Msg: "订单不存在"}
	// OrderStoreMismatch 订单不属于当前门店
	OrderStoreMismatch = baseErrs.ErrorCode{Code: 1_006_001_002, Msg: "订单不属于当前门店"}
	// OrderStatusInvalidTransition 非法状态流转
	OrderStatusInvalidTransition = baseErrs.ErrorCode{Code: 1_006_001_003, Msg: "订单当前状态【%v】不允许流转到【%v】"}
	// OrderItemsEmpty 下单至少需要一个菜品明细
	OrderItemsEmpty = baseErrs.ErrorCode{Code: 1_006_001_004, Msg: "下单至少需要选择一个菜品"}
	// OrderSkuNotSellable 规格未在门店上架或已沽清
	OrderSkuNotSellable = baseErrs.ErrorCode{Code: 1_006_001_005, Msg: "菜品【%v】未在当前门店上架或已沽清"}
	// OrderSkuNotFound 规格不存在
	OrderSkuNotFound = baseErrs.ErrorCode{Code: 1_006_001_006, Msg: "菜品规格不存在"}
	// OrderAddonNotInGroup 加料项不属于该规格允许的加料组
	OrderAddonNotInGroup = baseErrs.ErrorCode{Code: 1_006_001_007, Msg: "加料【%v】不属于菜品【%v】的可选加料"}
	// OrderAddonRangeExceeded 加料数量超出加料组限制
	OrderAddonRangeExceeded = baseErrs.ErrorCode{Code: 1_006_001_008, Msg: "菜品【%v】的加料【%v】数量超出限制"}
	// OrderQuantityInvalid 明细数量不合法
	OrderQuantityInvalid = baseErrs.ErrorCode{Code: 1_006_001_009, Msg: "菜品【%v】的数量必须大于 0"}
	// OrderTypeInvalid 订单类型取值不合法
	OrderTypeInvalid = baseErrs.ErrorCode{Code: 1_006_001_010, Msg: "订单类型取值不合法"}
	// OrderDeliveryInfoRequired 外卖订单必须填写收货信息
	OrderDeliveryInfoRequired = baseErrs.ErrorCode{Code: 1_006_001_011, Msg: "外卖订单必须填写收货人、电话与地址"}
	// OrderStoreNotOpen 门店未营业，不能下单
	OrderStoreNotOpen = baseErrs.ErrorCode{Code: 1_006_001_012, Msg: "门店未营业，暂时不能下单"}
	// OrderAlreadyExists 同一请求号已生成订单，幂等命中
	OrderAlreadyExists = baseErrs.ErrorCode{Code: 1_006_001_013, Msg: "订单已提交，请勿重复下单"}
	// OrderCancelNotAllowed 当前状态不允许取消
	OrderCancelNotAllowed = baseErrs.ErrorCode{Code: 1_006_001_014, Msg: "订单当前状态【%v】不允许取消"}
	// OrderCancelByStaffOnly 待接单后才可由店员取消
	OrderCancelNotAllowedForStaff = baseErrs.ErrorCode{Code: 1_006_001_015, Msg: "订单当前状态【%v】不允许店员取消"}
	// OrderTableRequired 堂食订单必须选择桌台
	OrderTableRequired = baseErrs.ErrorCode{Code: 1_006_001_016, Msg: "堂食订单必须选择桌台"}
	// OrderPayAmountMissing 订单应付金额为 0，不能发起支付
	OrderPayAmountMissing = baseErrs.ErrorCode{Code: 1_006_001_017, Msg: "订单应付金额为 0，无需支付"}
	// OrderMemberRequired C 端下单必须携带会员身份
	OrderMemberRequired = baseErrs.ErrorCode{Code: 1_006_001_018, Msg: "请先登录后再下单"}

	// CartNotExists 购物车不存在
	CartNotExists = baseErrs.ErrorCode{Code: 1_006_002_001, Msg: "购物车不存在"}
	// CartItemNotFound 购物车明细不存在
	CartItemNotFound = baseErrs.ErrorCode{Code: 1_006_002_002, Msg: "购物车明细不存在"}
	// CartItemQuantityInvalid 购物车明细数量不合法
	CartItemQuantityInvalid = baseErrs.ErrorCode{Code: 1_006_002_003, Msg: "购物车数量必须大于 0"}
)
