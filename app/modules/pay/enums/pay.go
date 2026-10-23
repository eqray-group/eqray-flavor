// Package enums 定义支付模块的枚举值，取值与 yudao（ruoyi-vue-pro）pay 模块保持一致。
// 注意：支付模块的「开启状态」沿用 yudao 口径（0 开启 / 1 关闭），
// 与全站 baseEnums.CommonStatusEnum（1 启用 / 0 禁用）相反，模块内统一使用本文件的枚举。
package enums

// PayEnableStatusEnum 是应用 / 渠道 / 充值套餐的开启状态，对齐 yudao CommonStatusEnum。
type PayEnableStatusEnum int8

const (
	// PayEnableStatusEnable 开启
	PayEnableStatusEnable PayEnableStatusEnum = 0
	// PayEnableStatusDisable 关闭
	PayEnableStatusDisable PayEnableStatusEnum = 1
)

// String 返回状态中文描述，便于列表展示与日志定位。
func (s PayEnableStatusEnum) String() string {
	if s == PayEnableStatusEnable {
		return "开启"
	}
	if s == PayEnableStatusDisable {
		return "关闭"
	}

	return "未知"
}

// IsValid 判断开启状态取值是否合法。
func (s PayEnableStatusEnum) IsValid() bool {
	return s == PayEnableStatusEnable || s == PayEnableStatusDisable
}

// 支付渠道编码，对齐 yudao PayChannelEnum。
const (
	// ChannelWxPub 微信公众号支付
	ChannelWxPub = "wx_pub"
	// ChannelWxLite 微信小程序支付
	ChannelWxLite = "wx_lite"
	// ChannelWxApp 微信 App 支付
	ChannelWxApp = "wx_app"
	// ChannelWxNative 微信 Native 扫码支付
	ChannelWxNative = "wx_native"
	// ChannelWxWap 微信 H5 支付
	ChannelWxWap = "wx_wap"
	// ChannelAlipayPc 支付宝电脑网站支付
	ChannelAlipayPc = "alipay_pc"
	// ChannelAlipayWap 支付宝手机网站支付
	ChannelAlipayWap = "alipay_wap"
	// ChannelAlipayApp 支付宝 App 支付
	ChannelAlipayApp = "alipay_app"
	// ChannelAlipayQr 支付宝当面付（扫码）
	ChannelAlipayQr = "alipay_qr"
	// ChannelAlipayBar 支付宝当面付（条码）
	ChannelAlipayBar = "alipay_bar"
	// ChannelMock 模拟支付，仅本地联调使用
	ChannelMock = "mock"
	// ChannelWallet 钱包余额支付：由本模块直接扣减余额完成，不走渠道网关。
	// 它是系统内置渠道，不需要在 pay_channel 里配置，出现在渠道编码清单里供筛选与下单使用。
	ChannelWallet = "wallet"
)

// PayChannelItem 是渠道编码的可选项描述，供下拉与校验共用。
type PayChannelItem struct {
	// Code 渠道编码
	Code string
	// Name 渠道中文名
	Name string
}

// PayChannelItems 返回全部渠道编码，顺序与 yudao 一致，末尾追加内置的钱包余额渠道。
func PayChannelItems() []PayChannelItem {
	return []PayChannelItem{
		{ChannelWxPub, "微信公从号"},
		{ChannelWxLite, "微信小程序"},
		{ChannelWxApp, "微信 App"},
		{ChannelWxNative, "微信 Native"},
		{ChannelWxWap, "微信 H5"},
		{ChannelAlipayPc, "支付宝 PC"},
		{ChannelAlipayWap, "支付宝 WAP"},
		{ChannelAlipayApp, "支付宝 APP"},
		{ChannelAlipayQr, "支付宝二维码"},
		{ChannelAlipayBar, "支付宝条码"},
		{ChannelMock, "模拟支付"},
		{ChannelWallet, "钱包余额"},
	}
}

// IsValidChannel 判断渠道编码是否在支持范围内（含内置的钱包余额渠道）。
func IsValidChannel(code string) bool {
	for _, item := range PayChannelItems() {
		if item.Code == code {
			return true
		}
	}

	return false
}

// ChannelLabel 返回渠道中文名，未知编码原样返回。
func ChannelLabel(code string) string {
	for _, item := range PayChannelItems() {
		if item.Code == code {
			return item.Name
		}
	}

	return code
}

// IsWechatChannel 判断是否微信系渠道，决定调用微信适配器。
func IsWechatChannel(code string) bool {
	switch code {
	case ChannelWxPub, ChannelWxLite, ChannelWxApp, ChannelWxNative, ChannelWxWap:
		return true
	default:
		return false
	}
}

// IsAlipayChannel 判断是否支付宝系渠道。
func IsAlipayChannel(code string) bool {
	switch code {
	case ChannelAlipayPc, ChannelAlipayWap, ChannelAlipayApp, ChannelAlipayQr, ChannelAlipayBar:
		return true
	default:
		return false
	}
}

// 微信交易状态取值，用于查单结果判断。
const (
	// WechatTradeStateSuccess 支付成功
	WechatTradeStateSuccess = "SUCCESS"
)

// 微信退款单状态取值，用于判断退款结果是已到账还是仍在处理。
const (
	// WechatRefundStatusSuccess 退款成功（资金已退到用户账户）
	WechatRefundStatusSuccess = "SUCCESS"
	// WechatRefundStatusProcessing 退款处理中，需等退款回调或查询确认
	WechatRefundStatusProcessing = "PROCESSING"
)

// PayOrderStatusEnum 是支付单状态，取值与 yudao PayOrderStatusEnum 完全一致：
// 待支付(0) → 支付成功(10)；待支付 → 支付关闭(20)。
// 注意：yudao 没有「已退款」状态。全额退款不改变支付单状态，只把退款金额累加到
// pay_order.refund_price，是否退完由 refund_price >= price 判定；支付关闭只表示
// 「未付款超时关闭 / 主动关单」（yudao 注释里同时把全额退款归入关闭语义，但代码未改状态）。
type PayOrderStatusEnum int8

const (
	// PayOrderStatusWaiting 待支付
	PayOrderStatusWaiting PayOrderStatusEnum = 0
	// PayOrderStatusSuccess 支付成功
	PayOrderStatusSuccess PayOrderStatusEnum = 10
	// PayOrderStatusClosed 支付关闭
	PayOrderStatusClosed PayOrderStatusEnum = 20
)

// String 返回支付单状态中文描述。
func (s PayOrderStatusEnum) String() string {
	switch s {
	case PayOrderStatusWaiting:
		return "待支付"
	case PayOrderStatusSuccess:
		return "支付成功"
	case PayOrderStatusClosed:
		return "支付关闭"
	default:
		return "未知状态"
	}
}

// IsValid 判断支付单状态取值是否合法。
func (s PayOrderStatusEnum) IsValid() bool {
	switch s {
	case PayOrderStatusWaiting, PayOrderStatusSuccess, PayOrderStatusClosed:
		return true
	default:
		return false
	}
}

// CanRefund 判断支付单是否允许发起退款：只有已支付且未退完的单可以退。
func (s PayOrderStatusEnum) CanRefund() bool {
	return s == PayOrderStatusSuccess
}

// PayRefundStatusEnum 是退款单状态，取值与 yudao PayRefundStatusEnum 完全一致：
// 退款订单生成(0) → 退款成功(1) / 退款失败(2)；退款关闭(99)。
// 说明：退款是「先建单再调渠道」，0 表示已受理但结果未定，等同旧实现的「等待退款」。
type PayRefundStatusEnum int8

const (
	// PayRefundStatusCreate 退款订单生成
	PayRefundStatusCreate PayRefundStatusEnum = 0
	// PayRefundStatusSuccess 退款成功
	PayRefundStatusSuccess PayRefundStatusEnum = 1
	// PayRefundStatusFail 退款失败
	PayRefundStatusFail PayRefundStatusEnum = 2
	// PayRefundStatusClose 退款关闭
	PayRefundStatusClose PayRefundStatusEnum = 99
)

// String 返回退款单状态中文描述。
func (s PayRefundStatusEnum) String() string {
	switch s {
	case PayRefundStatusCreate:
		return "退款订单生成"
	case PayRefundStatusSuccess:
		return "退款成功"
	case PayRefundStatusFail:
		return "退款失败"
	case PayRefundStatusClose:
		return "退款关闭"
	default:
		return "未知状态"
	}
}

// PayRefundAuditStatusEnum 是退款审核状态。
// 与退款状态解耦：yudao 的退款状态机保持原样，审核进度只体现在这一列。
// 无需审核(0) 表示建单即提交渠道，与改造前的行为一致，历史数据全部按此处理。
type PayRefundAuditStatusEnum int8

const (
	// PayRefundAuditNone 无需审核：建单后立即提交渠道
	PayRefundAuditNone PayRefundAuditStatusEnum = 0
	// PayRefundAuditPending 待审核：退款单已生成，尚未提交渠道
	PayRefundAuditPending PayRefundAuditStatusEnum = 1
	// PayRefundAuditApproved 审核通过：已提交渠道，结果以退款状态为准
	PayRefundAuditApproved PayRefundAuditStatusEnum = 2
	// PayRefundAuditRejected 审核驳回：退款单直接关闭，不会动钱
	PayRefundAuditRejected PayRefundAuditStatusEnum = 3
)

// String 返回审核状态中文描述。
func (s PayRefundAuditStatusEnum) String() string {
	switch s {
	case PayRefundAuditNone:
		return "无需审核"
	case PayRefundAuditPending:
		return "待审核"
	case PayRefundAuditApproved:
		return "审核通过"
	case PayRefundAuditRejected:
		return "审核驳回"
	default:
		return "未知状态"
	}
}

// IsValid 判断审核状态取值是否合法。
func (s PayRefundAuditStatusEnum) IsValid() bool {
	switch s {
	case PayRefundAuditNone, PayRefundAuditPending, PayRefundAuditApproved, PayRefundAuditRejected:
		return true
	default:
		return false
	}
}

// IsValid 判断退款单状态取值是否合法。
func (s PayRefundStatusEnum) IsValid() bool {
	switch s {
	case PayRefundStatusCreate, PayRefundStatusSuccess, PayRefundStatusFail, PayRefundStatusClose:
		return true
	default:
		return false
	}
}

// PayTransferStatusEnum 是转账单状态，对齐 yudao PayTransferStatusEnum。
type PayTransferStatusEnum int8

const (
	// PayTransferStatusWaiting 等待转账
	PayTransferStatusWaiting PayTransferStatusEnum = 0
	// PayTransferStatusSuccess 转账成功
	PayTransferStatusSuccess PayTransferStatusEnum = 10
	// PayTransferStatusFail 转账失败
	PayTransferStatusFail PayTransferStatusEnum = 20
)

// String 返回转账单状态中文描述。
func (s PayTransferStatusEnum) String() string {
	switch s {
	case PayTransferStatusWaiting:
		return "等待转账"
	case PayTransferStatusSuccess:
		return "转账成功"
	case PayTransferStatusFail:
		return "转账失败"
	default:
		return "未知状态"
	}
}

// PayNotifyTypeEnum 是回调通知的类型，取值与 yudao PayNotifyTypeEnum 一致。
// yudao 只对「支付单」和「退款单」通知业务方，转账不产生通知任务，因此这里只有两个值。
type PayNotifyTypeEnum int8

const (
	// PayNotifyTypeOrder 支付单通知
	PayNotifyTypeOrder PayNotifyTypeEnum = 1
	// PayNotifyTypeRefund 退款单通知
	PayNotifyTypeRefund PayNotifyTypeEnum = 2
)

// String 返回通知类型中文描述。
func (s PayNotifyTypeEnum) String() string {
	switch s {
	case PayNotifyTypeOrder:
		return "支付单"
	case PayNotifyTypeRefund:
		return "退款单"
	default:
		return "未知类型"
	}
}

// PayNotifyStatusEnum 是回调通知任务状态，取值与 yudao PayNotifyStatusEnum 一致。
// 语义区分「请求是否发出去」与「业务方是否认账」：
// 网络失败 / 非 2xx 记请求失败(5)，2xx 但响应不是 success 记请求成功但结果失败(4)。
// 只有成功(2) 是终态，其余都会按重试频率继续尝试，直到达到最大次数转为失败(3)。
type PayNotifyStatusEnum int8

const (
	// PayNotifyStatusWaiting 等待通知
	PayNotifyStatusWaiting PayNotifyStatusEnum = 1
	// PayNotifyStatusSuccess 通知成功
	PayNotifyStatusSuccess PayNotifyStatusEnum = 2
	// PayNotifyStatusFail 通知失败（多次尝试后的最终失败）
	PayNotifyStatusFail PayNotifyStatusEnum = 3
	// PayNotifyStatusRequestSuccess 请求成功，但业务方返回结果失败
	PayNotifyStatusRequestSuccess PayNotifyStatusEnum = 4
	// PayNotifyStatusRequestFail 请求失败（网络异常或 HTTP 非 2xx）
	PayNotifyStatusRequestFail PayNotifyStatusEnum = 5
)

// String 返回通知状态中文描述。
func (s PayNotifyStatusEnum) String() string {
	switch s {
	case PayNotifyStatusWaiting:
		return "等待通知"
	case PayNotifyStatusSuccess:
		return "通知成功"
	case PayNotifyStatusFail:
		return "通知失败"
	case PayNotifyStatusRequestSuccess:
		return "请求成功但结果失败"
	case PayNotifyStatusRequestFail:
		return "请求失败"
	default:
		return "未知状态"
	}
}

// IsRetryable 判断任务是否还需要继续重试：等待中、请求失败、结果失败都要重试。
func (s PayNotifyStatusEnum) IsRetryable() bool {
	switch s {
	case PayNotifyStatusWaiting, PayNotifyStatusRequestSuccess, PayNotifyStatusRequestFail:
		return true
	default:
		return false
	}
}

// PayWalletTransactionTypeEnum 是钱包流水的收支方向，对齐 yudao PayWalletTransactionTypeEnum。
type PayWalletTransactionTypeEnum int8

const (
	// PayWalletTransactionTypeIncome 收入
	PayWalletTransactionTypeIncome PayWalletTransactionTypeEnum = 1
	// PayWalletTransactionTypeExpense 支出
	PayWalletTransactionTypeExpense PayWalletTransactionTypeEnum = 2
)

// String 返回收支方向中文描述。
func (s PayWalletTransactionTypeEnum) String() string {
	switch s {
	case PayWalletTransactionTypeIncome:
		return "收入"
	case PayWalletTransactionTypeExpense:
		return "支出"
	default:
		return "未知类型"
	}
}

// PayWalletBizTypeEnum 是钱包流水的业务类型，取值与 yudao PayWalletBizTypeEnum 一致：
// 充值(1) / 充值退款(2) / 支付(3) / 支付退款(4)。
// 说明：yudao 用整型存 biz_type，且把「充值退款」与「支付退款」分开记，便于对账时区分
// 资金是回到余额还是退出余额；本模块沿用数值口径，另加 5 表示后台手工调整（项目扩展）。
type PayWalletBizTypeEnum int8

const (
	// WalletBizRecharge 充值
	WalletBizRecharge PayWalletBizTypeEnum = 1
	// WalletBizRechargeRefund 充值退款
	WalletBizRechargeRefund PayWalletBizTypeEnum = 2
	// WalletBizPay 支付
	WalletBizPay PayWalletBizTypeEnum = 3
	// WalletBizPayRefund 支付退款
	WalletBizPayRefund PayWalletBizTypeEnum = 4
	// WalletBizAdjust 后台手工调整余额（yudao 无此类型，本项目用于对账留痕）
	WalletBizAdjust PayWalletBizTypeEnum = 5
)

// String 返回钱包流水业务类型中文描述。
func (s PayWalletBizTypeEnum) String() string {
	switch s {
	case WalletBizRecharge:
		return "充值"
	case WalletBizRechargeRefund:
		return "充值退款"
	case WalletBizPay:
		return "支付"
	case WalletBizPayRefund:
		return "支付退款"
	case WalletBizAdjust:
		return "手工调整"
	default:
		return "未知类型"
	}
}

// PayWalletRechargeStatusEnum 是钱包充值记录状态，对齐 yudao PayWalletRechargeStatusEnum。
type PayWalletRechargeStatusEnum int8

const (
	// PayWalletRechargeStatusUnpaid 未支付
	PayWalletRechargeStatusUnpaid PayWalletRechargeStatusEnum = 0
	// PayWalletRechargeStatusPaid 已支付
	PayWalletRechargeStatusPaid PayWalletRechargeStatusEnum = 10
	// PayWalletRechargeStatusRefund 已退款
	PayWalletRechargeStatusRefund PayWalletRechargeStatusEnum = 20
)

// String 返回充值记录状态中文描述。
func (s PayWalletRechargeStatusEnum) String() string {
	switch s {
	case PayWalletRechargeStatusUnpaid:
		return "未支付"
	case PayWalletRechargeStatusPaid:
		return "已支付"
	case PayWalletRechargeStatusRefund:
		return "已退款"
	default:
		return "未知状态"
	}
}
