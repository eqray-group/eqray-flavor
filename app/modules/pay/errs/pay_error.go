// Package errs 定义 pay 模块的业务错误码，划分与 yudao pay 模块保持一致。
// 编码规则：1_{模块段}_{子模块}_{序号}，pay 模块段为 008。
package errs

import baseErrs "eqray-flavor/app/modules/base/errs"

// pay 模块业务错误码：1-008-{子模块}-{序号}。
// 子模块占用：app 应用 001、channel 渠道 002、order 支付单 003、refund 退款单 004、
// notify 通知 005、transfer 转账 006、wallet 钱包 007、recharge 充值 008。
var (
	// PayAppNotExists 支付应用不存在
	PayAppNotExists = baseErrs.ErrorCode{Code: 1_008_001_001, Msg: "支付应用不存在"}
	// PayAppIsDisable 支付应用已关闭
	PayAppIsDisable = baseErrs.ErrorCode{Code: 1_008_001_002, Msg: "支付应用【%v】已关闭"}
	// PayAppKeyDuplicate 应用标识 appKey 已被占用
	PayAppKeyDuplicate = baseErrs.ErrorCode{Code: 1_008_001_003, Msg: "应用标识【%v】已存在"}
	// PayAppNotifyURLRequired 应用的回调地址不能为空
	PayAppNotifyURLRequired = baseErrs.ErrorCode{Code: 1_008_001_004, Msg: "支付应用【%v】的回调地址不能为空"}

	// PayChannelNotExists 支付渠道不存在
	PayChannelNotExists = baseErrs.ErrorCode{Code: 1_008_002_001, Msg: "支付渠道不存在"}
	// PayChannelIsDisable 支付渠道已关闭
	PayChannelIsDisable = baseErrs.ErrorCode{Code: 1_008_002_002, Msg: "支付渠道【%v】已关闭"}
	// PayChannelCodeDuplicate 同一应用下渠道编码重复
	PayChannelCodeDuplicate = baseErrs.ErrorCode{Code: 1_008_002_003, Msg: "应用下已存在渠道【%v】"}
	// PayChannelCodeInvalid 渠道编码不在 yudao 支持范围内
	PayChannelCodeInvalid = baseErrs.ErrorCode{Code: 1_008_002_004, Msg: "渠道编码【%v】不支持"}
	// PayChannelConfigMissing 渠道配置缺失或解析失败
	PayChannelConfigMissing = baseErrs.ErrorCode{Code: 1_008_002_005, Msg: "支付渠道【%v】配置不完整：%v"}
	// PayChannelConfigInvalid 渠道配置 JSON 不合法
	PayChannelConfigInvalid = baseErrs.ErrorCode{Code: 1_008_002_006, Msg: "支付渠道配置必须是合法的 JSON"}
	// PayChannelAppImmutable 渠道创建后不能跨应用移动
	PayChannelAppImmutable = baseErrs.ErrorCode{Code: 1_008_002_007, Msg: "支付渠道不能跨应用移动"}

	// PayOrderNotExists 支付单不存在
	PayOrderNotExists = baseErrs.ErrorCode{Code: 1_008_003_001, Msg: "支付单不存在"}
	// PayOrderStatusNotWaiting 支付单不是待支付状态，不能发起支付
	PayOrderStatusNotWaiting = baseErrs.ErrorCode{Code: 1_008_003_002, Msg: "支付单【%v】不是待支付状态"}
	// PayOrderStatusNotSuccess 支付单未支付成功，不能退款或关闭
	PayOrderStatusNotSuccess = baseErrs.ErrorCode{Code: 1_008_003_003, Msg: "支付单【%v】不是已支付状态"}
	// PayOrderPriceInvalid 支付金额不合法
	PayOrderPriceInvalid = baseErrs.ErrorCode{Code: 1_008_003_004, Msg: "支付金额必须是大于 0 的整数（单位分）"}
	// PayOrderExtensionNotExists 支付订单拓展不存在
	PayOrderExtensionNotExists = baseErrs.ErrorCode{Code: 1_008_003_005, Msg: "支付订单拓展不存在"}
	// PayOrderSubmitFail 调用渠道下单失败
	PayOrderSubmitFail = baseErrs.ErrorCode{Code: 1_008_003_006, Msg: "调用支付渠道失败：%v"}
	// PayOrderNotifySignInvalid 渠道回调验签失败
	PayOrderNotifySignInvalid = baseErrs.ErrorCode{Code: 1_008_003_007, Msg: "支付回调验签失败"}
	// PayOrderNotifyPriceMismatch 回调金额与支付单金额不一致
	PayOrderNotifyPriceMismatch = baseErrs.ErrorCode{Code: 1_008_003_008, Msg: "回调金额【%v】与支付单金额【%v】不一致"}
	// PayOrderMerchantOrderIDDuplicate 商户订单号已存在有效支付单
	PayOrderMerchantOrderIDDuplicate = baseErrs.ErrorCode{Code: 1_008_003_009, Msg: "商户订单号【%v】已存在未关闭的支付单"}
	// PayOrderExpired 支付单已过期
	PayOrderExpired = baseErrs.ErrorCode{Code: 1_008_003_010, Msg: "支付单【%v】已过期，请重新下单"}
	// PayOrderQueryFail 查询渠道订单失败
	PayOrderQueryFail = baseErrs.ErrorCode{Code: 1_008_003_011, Msg: "查询渠道订单失败：%v"}
	// PayOrderCloseFail 关闭渠道订单失败
	PayOrderCloseFail = baseErrs.ErrorCode{Code: 1_008_003_012, Msg: "关闭渠道订单失败：%v"}

	// PayRefundNotExists 退款单不存在
	PayRefundNotExists = baseErrs.ErrorCode{Code: 1_008_004_001, Msg: "退款单不存在"}
	// PayRefundPriceInvalid 退款金额超出可退范围
	PayRefundPriceInvalid = baseErrs.ErrorCode{Code: 1_008_004_002, Msg: "退款金额必须是大于 0 的整数，且不超过可退金额【%v】"}
	// PayRefundStatusInvalid 退款单当前状态不允许该操作
	PayRefundStatusInvalid = baseErrs.ErrorCode{Code: 1_008_004_003, Msg: "退款单状态【%v】不允许该操作"}
	// PayRefundOrderStatusInvalid 支付单不是已支付状态，不能退款
	PayRefundOrderStatusInvalid = baseErrs.ErrorCode{Code: 1_008_004_004, Msg: "支付单【%v】不是已支付状态，不能退款"}
	// PayRefundMerchantRefundIDDuplicate 商户退款单号已存在
	PayRefundMerchantRefundIDDuplicate = baseErrs.ErrorCode{Code: 1_008_004_005, Msg: "商户退款单号【%v】已存在"}
	// PayRefundSubmitFail 调用渠道退款失败
	PayRefundSubmitFail = baseErrs.ErrorCode{Code: 1_008_004_006, Msg: "调用渠道退款失败：%v"}
	// PayRefundExceedOrderPrice 累计退款金额将超过支付金额；并发退款时由条件更新兜底
	PayRefundExceedOrderPrice = baseErrs.ErrorCode{Code: 1_008_004_007, Msg: "累计退款金额将超过支付金额，本次退款未生效"}

	// PayNotifyTaskNotExists 回调通知任务不存在
	PayNotifyTaskNotExists = baseErrs.ErrorCode{Code: 1_008_005_001, Msg: "回调通知任务不存在"}
	// PayNotifySendFail 回调通知发送失败
	PayNotifySendFail = baseErrs.ErrorCode{Code: 1_008_005_002, Msg: "回调通知发送失败：%v"}

	// PayTransferNotExists 转账单不存在
	PayTransferNotExists = baseErrs.ErrorCode{Code: 1_008_006_001, Msg: "转账单不存在"}
	// PayTransferPriceInvalid 转账金额不合法
	PayTransferPriceInvalid = baseErrs.ErrorCode{Code: 1_008_006_002, Msg: "转账金额必须是大于 0 的整数（单位分）"}
	// PayTransferStatusInvalid 转账单当前状态不允许该操作
	PayTransferStatusInvalid = baseErrs.ErrorCode{Code: 1_008_006_003, Msg: "转账单状态【%v】不允许该操作"}
	// PayTransferMerchantTransferIDDuplicate 商户转账单号已存在
	PayTransferMerchantTransferIDDuplicate = baseErrs.ErrorCode{Code: 1_008_006_004, Msg: "商户转账单号【%v】已存在"}

	// PayWalletNotExists 会员钱包不存在
	PayWalletNotExists = baseErrs.ErrorCode{Code: 1_008_007_001, Msg: "会员钱包不存在"}
	// PayWalletBalanceNotEnough 钱包余额不足
	PayWalletBalanceNotEnough = baseErrs.ErrorCode{Code: 1_008_007_002, Msg: "钱包余额不足，当前可用余额【%v】"}
	// PayWalletFreezePriceNotEnough 钱包冻结金额不足
	PayWalletFreezePriceNotEnough = baseErrs.ErrorCode{Code: 1_008_007_003, Msg: "钱包冻结金额不足，当前冻结金额【%v】"}
	// PayWalletTransactionNotExists 钱包流水不存在
	PayWalletTransactionNotExists = baseErrs.ErrorCode{Code: 1_008_007_004, Msg: "钱包流水不存在"}
	// PayWalletPriceInvalid 钱包变动金额不合法
	PayWalletPriceInvalid = baseErrs.ErrorCode{Code: 1_008_007_005, Msg: "钱包变动金额必须是不为 0 的整数（单位分）"}
	// PayWalletTransactionTypeInvalid 钱包流水收支类型不合法
	PayWalletTransactionTypeInvalid = baseErrs.ErrorCode{Code: 1_008_007_006, Msg: "钱包流水收支类型只能是 1（收入）或 2（支出）"}

	// PayWalletRechargePackageNotExists 充值套餐不存在
	PayWalletRechargePackageNotExists = baseErrs.ErrorCode{Code: 1_008_008_001, Msg: "充值套餐不存在"}
	// PayWalletRechargeNotExists 充值记录不存在
	PayWalletRechargeNotExists = baseErrs.ErrorCode{Code: 1_008_008_002, Msg: "充值记录不存在"}
	// PayWalletRechargeStatusInvalid 充值记录当前状态不允许该操作
	PayWalletRechargeStatusInvalid = baseErrs.ErrorCode{Code: 1_008_008_003, Msg: "充值记录状态【%v】不允许该操作"}
	// PayWalletRechargePriceInvalid 充值金额不合法
	PayWalletRechargePriceInvalid = baseErrs.ErrorCode{Code: 1_008_008_004, Msg: "充值金额必须是大于 0 的整数（单位分）"}
)
