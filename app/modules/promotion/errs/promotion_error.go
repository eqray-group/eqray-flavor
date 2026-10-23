// Package errs 定义 promotion 模块的业务错误码。
// 编码规则：1_{模块段}_{子模块}_{序号}，promotion 模块段为 007。
package errs

import baseErrs "eqray-flavor/app/modules/base/errs"

// promotion 模块业务错误码：1-007-{子模块}-{序号}。
// 子模块占用：coupon-template 券模板 001、coupon 会员券 002、activity 活动 003、points 积分 004。
var (
	// CouponTemplateNotExists 券模板不存在或已删除
	CouponTemplateNotExists = baseErrs.ErrorCode{Code: 1_007_001_001, Msg: "优惠券模板不存在"}
	// CouponTemplateDisabled 券模板已停用
	CouponTemplateDisabled = baseErrs.ErrorCode{Code: 1_007_001_002, Msg: "优惠券模板已停用"}
	// CouponTemplateTypeInvalid 券类型取值不合法
	CouponTemplateTypeInvalid = baseErrs.ErrorCode{Code: 1_007_001_003, Msg: "优惠券类型取值不合法"}
	// CouponTemplateScopeInvalid 适用范围与 scope_ids 不匹配
	CouponTemplateScopeInvalid = baseErrs.ErrorCode{Code: 1_007_001_004, Msg: "指定门店 / 商品的适用范围必须选择对应的对象"}
	// CouponTemplateAmountInvalid 面额、折扣率或门槛取值不合法
	CouponTemplateAmountInvalid = baseErrs.ErrorCode{Code: 1_007_001_005, Msg: "优惠券面额、折扣率或门槛取值不合法"}
	// CouponTemplateQuotaInvalid 发行量或每人限领取值不合法
	CouponTemplateQuotaInvalid = baseErrs.ErrorCode{Code: 1_007_001_006, Msg: "优惠券发行量或每人限领取值不合法"}
	// CouponTemplateIssueTypeInvalid 发放方式取值不合法
	CouponTemplateIssueTypeInvalid = baseErrs.ErrorCode{Code: 1_007_001_007, Msg: "优惠券发放方式取值不合法"}

	// CouponNotExists 会员券不存在或已删除
	CouponNotExists = baseErrs.ErrorCode{Code: 1_007_002_001, Msg: "优惠券不存在"}
	// CouponStatusInvalid 券状态不允许该操作
	CouponStatusInvalid = baseErrs.ErrorCode{Code: 1_007_002_002, Msg: "优惠券当前状态【%v】不允许该操作"}
	// CouponNotOwned 优惠券不属于当前会员
	CouponNotOwned = baseErrs.ErrorCode{Code: 1_007_002_003, Msg: "优惠券不属于当前账号"}
	// CouponExpired 优惠券已过期
	CouponExpired = baseErrs.ErrorCode{Code: 1_007_002_004, Msg: "优惠券已过期"}
	// CouponThresholdNotReached 未达使用门槛
	CouponThresholdNotReached = baseErrs.ErrorCode{Code: 1_007_002_005, Msg: "未达到优惠券使用门槛"}
	// CouponScopeNotMatch 优惠券不适用于当前门店或商品
	CouponScopeNotMatch = baseErrs.ErrorCode{Code: 1_007_002_006, Msg: "优惠券不适用于当前门店或商品"}
	// CouponChannelNotMatch 优惠券不适用于当前下单渠道
	CouponChannelNotMatch = baseErrs.ErrorCode{Code: 1_007_002_007, Msg: "优惠券不适用于当前下单渠道"}
	// CouponIssueLimitReached 超出每人限领张数
	CouponIssueLimitReached = baseErrs.ErrorCode{Code: 1_007_002_008, Msg: "已达到该优惠券的领取上限"}
	// CouponTemplateSoldOut 优惠券已发完
	CouponTemplateSoldOut = baseErrs.ErrorCode{Code: 1_007_002_009, Msg: "优惠券已领完"}
	// CouponDiscountInvalid 券面额或折扣计算结果不合法
	CouponDiscountInvalid = baseErrs.ErrorCode{Code: 1_007_002_010, Msg: "优惠券优惠金额不合法"}

	// ActivityNotExists 活动不存在或已删除
	ActivityNotExists = baseErrs.ErrorCode{Code: 1_007_003_001, Msg: "营销活动不存在"}
	// ActivityDisabled 活动未启用
	ActivityDisabled = baseErrs.ErrorCode{Code: 1_007_003_002, Msg: "营销活动未启用"}
	// ActivityNotInTime 活动不在有效期内
	ActivityNotInTime = baseErrs.ErrorCode{Code: 1_007_003_003, Msg: "营销活动不在有效期内"}
	// ActivityTypeInvalid 活动玩法取值不合法
	ActivityTypeInvalid = baseErrs.ErrorCode{Code: 1_007_003_004, Msg: "营销活动玩法取值不合法"}
	// ActivityRuleInvalid 活动规则取值不合法
	ActivityRuleInvalid = baseErrs.ErrorCode{Code: 1_007_003_005, Msg: "营销活动规则取值不合法"}
	// ActivityScopeInvalid 活动范围与 scope_ids 不匹配
	ActivityScopeInvalid = baseErrs.ErrorCode{Code: 1_007_003_006, Msg: "指定门店的活动必须选择对应的门店"}
	// ActivityProductRequired 商品级活动必须选择参与商品
	ActivityProductRequired = baseErrs.ErrorCode{Code: 1_007_003_007, Msg: "商品级活动必须选择参与商品"}
	// ActivityQuotaInvalid 活动名额取值不合法
	ActivityQuotaInvalid = baseErrs.ErrorCode{Code: 1_007_003_008, Msg: "营销活动名额取值不合法"}
	// ActivityLevelInvalid 活动层级取值不合法
	ActivityLevelInvalid = baseErrs.ErrorCode{Code: 1_007_003_009, Msg: "营销活动层级取值不合法"}
	// ActivityQuotaNotEnough 活动名额已被抢完，需重新试算
	ActivityQuotaNotEnough = baseErrs.ErrorCode{Code: 1_007_003_010, Msg: "营销活动【%v】名额已用完，请重新下单"}

	// PointsNotEnough 会员积分不足
	PointsNotEnough = baseErrs.ErrorCode{Code: 1_007_004_001, Msg: "积分不足，当前可用积分【%v】"}
	// PointsRuleInvalid 积分抵扣规则取值不合法
	PointsRuleInvalid = baseErrs.ErrorCode{Code: 1_007_004_002, Msg: "积分抵扣规则取值不合法"}
	// PointsMemberRequired 积分抵扣必须绑定会员
	PointsMemberRequired = baseErrs.ErrorCode{Code: 1_007_004_003, Msg: "积分抵扣需要登录会员账号"}
)
