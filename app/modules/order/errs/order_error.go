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
	// OrderPaidCannotAdjust 已支付订单不允许改价，避免与支付单金额不一致
	OrderPaidCannotAdjust = baseErrs.ErrorCode{Code: 1_006_001_019, Msg: "订单已支付，不能改价"}
	// OrderAdjustAmountInvalid 改价金额不合法
	OrderAdjustAmountInvalid = baseErrs.ErrorCode{Code: 1_006_001_020, Msg: "改价金额不能为 0，且改价后应付金额不能小于 0"}
	// OrderStatusNotAdjustable 当前状态不允许改价
	OrderStatusNotAdjustable = baseErrs.ErrorCode{Code: 1_006_001_021, Msg: "订单当前状态【%v】不允许改价"}
	// OrderNotDeliverable 仅外卖订单支持派单
	OrderNotDeliverable = baseErrs.ErrorCode{Code: 1_006_001_022, Msg: "仅外卖订单支持派单"}
	// OrderRiderNotExists 骑手不存在或不属于该门店
	OrderRiderNotExists = baseErrs.ErrorCode{Code: 1_006_001_023, Msg: "骑手不存在或不属于该门店"}
	// OrderRiderUnavailable 骑手当前不可接单
	OrderRiderUnavailable = baseErrs.ErrorCode{Code: 1_006_001_024, Msg: "骑手当前不可接单或已达接单上限"}
	// OrderAlreadyDispatched 订单已派单
	OrderAlreadyDispatched = baseErrs.ErrorCode{Code: 1_006_001_025, Msg: "该订单已派单，请先取消原派单"}
	// OrderStatusNotDispatchable 当前状态不允许派单
	OrderStatusNotDispatchable = baseErrs.ErrorCode{Code: 1_006_001_026, Msg: "订单当前状态【%v】不允许派单"}

	// CartNotExists 购物车不存在
	CartNotExists = baseErrs.ErrorCode{Code: 1_006_002_001, Msg: "购物车不存在"}
	// CartItemNotFound 购物车明细不存在
	CartItemNotFound = baseErrs.ErrorCode{Code: 1_006_002_002, Msg: "购物车明细不存在"}
	// CartItemQuantityInvalid 购物车明细数量不合法
	CartItemQuantityInvalid = baseErrs.ErrorCode{Code: 1_006_002_003, Msg: "购物车数量必须大于 0"}

	// OrderPosStoreRequired POS 与门店绑定，未带门店上下文时不允许下单与收款
	OrderPosStoreRequired = baseErrs.ErrorCode{Code: 1_006_003_001, Msg: "请先选择门店"}
	// OrderCashSettleStatusInvalid 当前状态不能现金收款
	OrderCashSettleStatusInvalid = baseErrs.ErrorCode{Code: 1_006_003_002, Msg: "订单当前状态【%v】不能收款"}
	// OrderAlreadyPaid 订单已收款，禁止重复结账
	OrderAlreadyPaid = baseErrs.ErrorCode{Code: 1_006_003_003, Msg: "订单已收款，请勿重复结账"}
	// OrderPosPayAppNotConfigured 系统内没有启用的支付应用
	OrderPosPayAppNotConfigured = baseErrs.ErrorCode{Code: 1_006_003_004, Msg: "尚未配置支付应用，请先在支付管理中新增"}
	// OrderPosPayAppNotUnique 系统内启用了多个支付应用
	OrderPosPayAppNotUnique = baseErrs.ErrorCode{Code: 1_006_003_005, Msg: "检测到【%v】个启用的支付应用，本系统按单应用运行，请只保留一个"}
	// OrderPosPayChannelMissing 支付应用下缺少对应渠道
	OrderPosPayChannelMissing = baseErrs.ErrorCode{Code: 1_006_003_006, Msg: "未配置【%v】支付渠道，请先在支付渠道中添加"}
	// OrderPosPayAuthCodeRequired 支付宝条码支付必须扫码
	OrderPosPayAuthCodeRequired = baseErrs.ErrorCode{Code: 1_006_003_007, Msg: "请扫描顾客的支付宝付款码"}
	// OrderPosPayRejected 顾客未完成付款
	OrderPosPayRejected = baseErrs.ErrorCode{Code: 1_006_003_008, Msg: "顾客尚未完成付款，请重试或改用其它方式"}
	// OrderPayAmountMismatch 支付单金额与订单应付不一致，拒绝入账
	OrderPayAmountMismatch = baseErrs.ErrorCode{Code: 1_006_003_009, Msg: "支付金额【%v】与订单应付【%v】不一致，已拒绝入账"}
	// OrderAppendNotAllowed 当前订单状态不支持加菜
	OrderAppendNotAllowed = baseErrs.ErrorCode{Code: 1_006_003_010, Msg: "订单当前状态【%v】不能加菜"}
	// OrderMergeNotAllowed 并台条件不满足
	OrderMergeNotAllowed = baseErrs.ErrorCode{Code: 1_006_003_011, Msg: "只有同一门店、同一时间段的未支付堂食单才能并台"}
	// OrderAppendCouponLocked 已用券的订单不允许加菜
	OrderAppendCouponLocked = baseErrs.ErrorCode{Code: 1_006_003_012, Msg: "该单已使用优惠券，加菜会破坏优惠分摊，请取消后重新下单"}
)
