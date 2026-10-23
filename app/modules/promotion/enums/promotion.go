// Package enums 定义 promotion 模块的枚举，取值与字典 promotion_* 一致。
// 口径来源：docs/design/07-会员与营销设计.md §3 优惠券、§5 活动引擎。
package enums

// CouponType 是优惠券类型，对应 07 §3「满减、折扣（可封顶）、单品、兑换、代金」。
type CouponType int8

const (
	// CouponTypeFullReduction 满减券：满门槛直接减固定金额
	CouponTypeFullReduction CouponType = 1
	// CouponTypeDiscount 折扣券：按折扣率打折，可封顶
	CouponTypeDiscount CouponType = 2
	// CouponTypeProduct 单品券：只对指定商品生效
	CouponTypeProduct CouponType = 3
	// CouponTypeExchange 兑换券：兑换指定商品，抵扣固定金额
	CouponTypeExchange CouponType = 4
	// CouponTypeCash 代金券：满门槛减固定金额，可跨门店
	CouponTypeCash CouponType = 5
)

// String 返回优惠券类型中文描述。
func (t CouponType) String() string {
	switch t {
	case CouponTypeFullReduction:
		return "满减券"
	case CouponTypeDiscount:
		return "折扣券"
	case CouponTypeProduct:
		return "单品券"
	case CouponTypeExchange:
		return "兑换券"
	case CouponTypeCash:
		return "代金券"
	default:
		return "未知券类型"
	}
}

// CouponStatus 是会员券状态，对应 07 §3 的生命周期图。
// 流转必须条件更新，防并发重复使用。
type CouponStatus int8

const (
	// CouponStatusUnused 未使用：已领取待使用
	CouponStatusUnused CouponStatus = 1
	// CouponStatusLocked 锁定：下单占用，等待收款核销
	CouponStatusLocked CouponStatus = 2
	// CouponStatusUsed 已使用：收款后核销
	CouponStatusUsed CouponStatus = 3
	// CouponStatusExpired 已过期：超过有效期
	CouponStatusExpired CouponStatus = 4
)

// String 返回券状态中文描述。
func (s CouponStatus) String() string {
	switch s {
	case CouponStatusUnused:
		return "未使用"
	case CouponStatusLocked:
		return "锁定"
	case CouponStatusUsed:
		return "已使用"
	case CouponStatusExpired:
		return "已过期"
	default:
		return "未知状态"
	}
}

// CouponScope 是优惠券适用范围，对应 07 §3「全品牌 / 指定门店 / 指定商品」。
type CouponScope int8

const (
	// CouponScopeAll 全品牌：所有门店可用
	CouponScopeAll CouponScope = 1
	// CouponScopeStore 指定门店：仅 scope_ids 中的门店可用
	CouponScopeStore CouponScope = 2
	// CouponScopeProduct 指定商品：仅 scope_ids 中的商品可用
	CouponScopeProduct CouponScope = 3
)

// String 返回适用范围中文描述。
func (s CouponScope) String() string {
	switch s {
	case CouponScopeAll:
		return "全品牌"
	case CouponScopeStore:
		return "指定门店"
	case CouponScopeProduct:
		return "指定商品"
	default:
		return "未知范围"
	}
}

// CouponChannelLimit 是优惠券的渠道限制，对应 07 §3「不限 / 仅线上 / 仅线下」。
type CouponChannelLimit int8

const (
	// CouponChannelUnlimited 不限渠道
	CouponChannelUnlimited CouponChannelLimit = 0
	// CouponChannelOnlineOnly 仅线上渠道
	CouponChannelOnlineOnly CouponChannelLimit = 1
	// CouponChannelOfflineOnly 仅线下渠道
	CouponChannelOfflineOnly CouponChannelLimit = 2
)

// String 返回渠道限制中文描述。
func (c CouponChannelLimit) String() string {
	switch c {
	case CouponChannelUnlimited:
		return "不限渠道"
	case CouponChannelOnlineOnly:
		return "仅线上"
	case CouponChannelOfflineOnly:
		return "仅线下"
	default:
		return "未知限制"
	}
}

// CouponIssueType 是优惠券发放方式，对应 07 §3「主动领取 / 消费返券 / 新人礼 / 定向发券」。
type CouponIssueType int8

const (
	// CouponIssueActive 主动领取：会员自己领，受每人限领约束
	CouponIssueActive CouponIssueType = 1
	// CouponIssueConsume 消费返券：消费达标后系统发放
	CouponIssueConsume CouponIssueType = 2
	// CouponIssueNewcomer 新人礼：新会员注册发放
	CouponIssueNewcomer CouponIssueType = 3
	// CouponIssueDirected 定向发券：后台指定会员发放
	CouponIssueDirected CouponIssueType = 4
)

// String 返回发放方式中文描述。
func (i CouponIssueType) String() string {
	switch i {
	case CouponIssueActive:
		return "主动领取"
	case CouponIssueConsume:
		return "消费返券"
	case CouponIssueNewcomer:
		return "新人礼"
	case CouponIssueDirected:
		return "定向发券"
	default:
		return "未知方式"
	}
}

// ActivityType 是活动玩法，对应 07 §5 的五种玩法。
type ActivityType int8

const (
	// ActivityTypeFullReduction 满减：阶梯满减
	ActivityTypeFullReduction ActivityType = 1
	// ActivityTypeSecondHalfPrice 第二份半价：同款第二份折扣
	ActivityTypeSecondHalfPrice ActivityType = 2
	// ActivityTypeTimeDiscount 限时折扣：指定时段折扣
	ActivityTypeTimeDiscount ActivityType = 3
	// ActivityTypeCombo 套餐特惠：组合价低于单品之和
	ActivityTypeCombo ActivityType = 4
	// ActivityTypeAddonPurchase 加价购：满额后加价换购
	ActivityTypeAddonPurchase ActivityType = 5
)

// String 返回活动玩法中文描述。
func (t ActivityType) String() string {
	switch t {
	case ActivityTypeFullReduction:
		return "满减"
	case ActivityTypeSecondHalfPrice:
		return "第二份半价"
	case ActivityTypeTimeDiscount:
		return "限时折扣"
	case ActivityTypeCombo:
		return "套餐特惠"
	case ActivityTypeAddonPurchase:
		return "加价购"
	default:
		return "未知玩法"
	}
}

// ActivityLevel 是活动作用层级，对应 07 §4 优惠顺序里的「商品级 / 订单级」。
type ActivityLevel int8

const (
	// ActivityLevelProduct 商品级：只作用于参与活动的商品行
	ActivityLevelProduct ActivityLevel = 1
	// ActivityLevelOrder 订单级：作用于整单
	ActivityLevelOrder ActivityLevel = 2
)

// String 返回活动层级中文描述。
func (l ActivityLevel) String() string {
	switch l {
	case ActivityLevelProduct:
		return "商品级"
	case ActivityLevelOrder:
		return "订单级"
	default:
		return "未知层级"
	}
}

// ActivityScope 是活动适用门店范围。
type ActivityScope int8

const (
	// ActivityScopeAll 全品牌门店
	ActivityScopeAll ActivityScope = 1
	// ActivityScopeStore 指定门店
	ActivityScopeStore ActivityScope = 2
)

// String 返回活动范围中文描述。
func (s ActivityScope) String() string {
	switch s {
	case ActivityScopeAll:
		return "全品牌"
	case ActivityScopeStore:
		return "指定门店"
	default:
		return "未知范围"
	}
}

// ActivityQuotaCycle 是活动名额的计数周期，对应 07 §5「按活动 + 周期计数」。
type ActivityQuotaCycle int8

const (
	// ActivityQuotaCycleWhole 整个活动期共用一个名额池
	ActivityQuotaCycleWhole ActivityQuotaCycle = 1
	// ActivityQuotaCycleDaily 每日一个名额池，跨天自动重置
	ActivityQuotaCycleDaily ActivityQuotaCycle = 2
)

// String 返回名额周期中文描述。
func (c ActivityQuotaCycle) String() string {
	switch c {
	case ActivityQuotaCycleWhole:
		return "整个活动期"
	case ActivityQuotaCycleDaily:
		return "每日重置"
	default:
		return "未知周期"
	}
}

// ActivityStatus 是活动状态。
type ActivityStatus int8

const (
	// ActivityStatusDisabled 未启用
	ActivityStatusDisabled ActivityStatus = 0
	// ActivityStatusEnabled 已启用
	ActivityStatusEnabled ActivityStatus = 1
)

// String 返回活动状态中文描述。
func (s ActivityStatus) String() string {
	switch s {
	case ActivityStatusEnabled:
		return "已启用"
	case ActivityStatusDisabled:
		return "未启用"
	default:
		return "未知状态"
	}
}

// OrderActivityStatus 是订单占用活动名额的关联状态。
type OrderActivityStatus int8

const (
	// OrderActivityStatusOccupied 占用中：名额已计入活动配额，尚未释放
	OrderActivityStatusOccupied OrderActivityStatus = 1
	// OrderActivityStatusReleased 已释放：名额已退回活动配额
	OrderActivityStatusReleased OrderActivityStatus = 2
)

// String 返回占用状态中文描述。
func (s OrderActivityStatus) String() string {
	switch s {
	case OrderActivityStatusOccupied:
		return "占用中"
	case OrderActivityStatusReleased:
		return "已释放"
	default:
		return "未知状态"
	}
}

// CouponStatus 的合法流转表：key 是原状态，value 是允许的新状态集合。
// 未使用 → 锁定（占用）→ 已使用（核销）；未使用 → 已过期；
// 锁定 → 未使用（取消 / 超时释放）；已使用 → 未使用（整单退款返还）。
var couponTransitions = map[CouponStatus][]CouponStatus{
	CouponStatusUnused:  {CouponStatusLocked, CouponStatusExpired},
	CouponStatusLocked:  {CouponStatusUsed, CouponStatusUnused},
	CouponStatusUsed:    {CouponStatusUnused},
	CouponStatusExpired: {},
}

// CanTransit 判断券状态流转是否合法。
func (s CouponStatus) CanTransit(target CouponStatus) bool {
	for _, next := range couponTransitions[s] {
		if next == target {
			return true
		}
	}

	return false
}
