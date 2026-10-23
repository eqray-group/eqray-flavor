// Package model 定义 promotion 模块的数据模型：优惠券模板、会员券、活动与活动关联对象。
// 表名统一用 promotion_ 前缀；金额一律整型分（docs/design/07 §4 口径 AM-06）。
package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/promotion/enums"
)

// CouponTemplateTable 是券模板表名常量。
const CouponTemplateTable = "promotion_coupon_template"

// CouponTemplateModel 对应 promotion_coupon_template 表，描述一种可发放的优惠券。
// 券模板是「规则」的载体：会员券只是模板的一次发放实例，规则变更不影响已发出的券。
type CouponTemplateModel struct {
	baseModel.BaseModel
	// Name 券名称
	Name string `db:"name" json:"name"`
	// Type 券类型：1 满减 / 2 折扣 / 3 单品 / 4 兑换 / 5 代金
	Type enums.CouponType `db:"type" json:"type"`
	// Description 券说明
	Description string `db:"description" json:"description"`
	// ThresholdAmount 使用门槛（分），0 表示无门槛
	ThresholdAmount int64 `db:"threshold_amount" json:"thresholdAmount"`
	// DiscountAmount 抵扣金额（分）：满减、代金、兑换、单品券使用
	DiscountAmount int64 `db:"discount_amount" json:"discountAmount"`
	// DiscountRate 折扣率（万分比）：7500 表示 75 折，只有折扣券使用
	DiscountRate int32 `db:"discount_rate" json:"discountRate"`
	// MaxDiscountAmount 折扣封顶（分），0 表示不封顶
	MaxDiscountAmount int64 `db:"max_discount_amount" json:"maxDiscountAmount"`
	// Scope 适用范围：1 全品牌 / 2 指定门店 / 3 指定商品
	Scope enums.CouponScope `db:"scope" json:"scope"`
	// ScopeIDs 适用对象 ID 列表，逗号分隔：门店 ID 或商品 SPU ID
	ScopeIDs string `db:"scope_ids" json:"scopeIds"`
	// ChannelLimit 渠道限制：0 不限 / 1 仅线上 / 2 仅线下
	ChannelLimit enums.CouponChannelLimit `db:"channel_limit" json:"channelLimit"`
	// IssueType 发放方式：1 主动领取 / 2 消费返券 / 3 新人礼 / 4 定向发券
	IssueType enums.CouponIssueType `db:"issue_type" json:"issueType"`
	// TotalCount 发行总量，0 表示不限量
	TotalCount int `db:"total_count" json:"totalCount"`
	// IssuedCount 已发放数量
	IssuedCount int `db:"issued_count" json:"issuedCount"`
	// PerMemberLimit 每人限领张数，0 表示不限
	PerMemberLimit int `db:"per_member_limit" json:"perMemberLimit"`
	// StartTime 生效开始时间
	StartTime types.DateTime `db:"start_time" json:"startTime"`
	// EndTime 生效结束时间
	EndTime types.DateTime `db:"end_time" json:"endTime"`
	// ValidDays 领取后有效天数，0 表示长期有效直到 end_time
	ValidDays int `db:"valid_days" json:"validDays"`
	// Status 启用状态
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *CouponTemplateModel) TableName() string {
	return CouponTemplateTable
}

// CouponTable 是会员券表名常量。
const CouponTable = "promotion_coupon"

// CouponModel 对应 promotion_coupon 表，表示一张已发放给会员的券。
// 状态流转必须条件更新，防并发重复使用（docs/design/07 §3）。
type CouponModel struct {
	baseModel.BaseModel
	// TemplateID 券模板 ID
	TemplateID types.ID `db:"template_id" json:"templateId"`
	// MemberID 持有会员 ID
	MemberID types.ID `db:"member_id" json:"memberId"`
	// Status 券状态：1 未使用 / 2 锁定 / 3 已使用 / 4 已过期
	Status enums.CouponStatus `db:"status" json:"status"`
	// OrderID 占用该券的订单 ID，锁定后写入
	OrderID types.ID `db:"order_id" json:"orderId"`
	// UseTime 核销时间
	UseTime *types.DateTime `db:"use_time" json:"useTime"`
	// ExpireTime 券过期时间
	ExpireTime types.DateTime `db:"expire_time" json:"expireTime"`
	// GetTime 领取时间
	GetTime types.DateTime `db:"get_time" json:"getTime"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *CouponModel) TableName() string {
	return CouponTable
}

// ActivityTable 是活动表名常量。
const ActivityTable = "promotion_activity"

// ActivityModel 对应 promotion_activity 表，描述一条营销活动规则。
// 规则细节（阶梯档位、套餐组合、加价购商品等）存在 rule JSON 里，
// 由 service 解析成强类型结构后参与计算，字段口径见 docs/design/07 §5。
type ActivityModel struct {
	baseModel.BaseModel
	// Name 活动名称
	Name string `db:"name" json:"name"`
	// Type 活动玩法：1 满减 / 2 第二份半价 / 3 限时折扣 / 4 套餐特惠 / 5 加价购
	Type enums.ActivityType `db:"type" json:"type"`
	// Level 作用层级：1 商品级 / 2 订单级
	Level enums.ActivityLevel `db:"level" json:"level"`
	// Description 活动说明
	Description string `db:"description" json:"description"`
	// Rule 玩法规则 JSON，结构见 service 层的规则解析类型
	Rule string `db:"rule" json:"rule"`
	// Scope 适用门店范围：1 全品牌 / 2 指定门店
	Scope enums.ActivityScope `db:"scope" json:"scope"`
	// ScopeIDs 适用门店 ID 列表，逗号分隔
	ScopeIDs string `db:"scope_ids" json:"scopeIds"`
	// Priority 优先级，数值大的先判定；同类型互斥时取优先级最高
	Priority int `db:"priority" json:"priority"`
	// MutexGroup 互斥组，组内活动只生效一个
	MutexGroup string `db:"mutex_group" json:"mutexGroup"`
	// Exclusive 是否独占：独占活动生效后其他活动全部终止（07 §5 判定顺序 3）
	Exclusive int8 `db:"exclusive" json:"exclusive"`
	// Stackable 是否可叠加：可叠加活动优惠金额累加
	Stackable int8 `db:"stackable" json:"stackable"`
	// QuotaLimit 名额上限，0 表示不限
	QuotaLimit int `db:"quota_limit" json:"quotaLimit"`
	// QuotaCycle 名额计数周期：1 整个活动期 / 2 每日重置
	QuotaCycle enums.ActivityQuotaCycle `db:"quota_cycle" json:"quotaCycle"`
	// StartTime 活动开始时间
	StartTime types.DateTime `db:"start_time" json:"startTime"`
	// EndTime 活动结束时间
	EndTime types.DateTime `db:"end_time" json:"endTime"`
	// Status 启用状态
	Status enums.ActivityStatus `db:"status" json:"status"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *ActivityModel) TableName() string {
	return ActivityTable
}

// ActivityProductTable 是活动参与商品表名常量。
const ActivityProductTable = "promotion_activity_product"

// ActivityProductModel 对应 promotion_activity_product 表，记录活动参与的商品范围。
// 商品级活动必须配置参与商品，否则活动范围不明确。
type ActivityProductModel struct {
	baseModel.BaseModel
	// ActivityID 活动 ID
	ActivityID types.ID `db:"activity_id" json:"activityId"`
	// SpuID 参与活动的菜品 SPU ID
	SpuID types.ID `db:"spu_id" json:"spuId"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *ActivityProductModel) TableName() string {
	return ActivityProductTable
}

// ActivityQuotaTable 是活动名额计数表名常量。
const ActivityQuotaTable = "promotion_activity_quota"

// ActivityQuotaModel 对应 promotion_activity_quota 表，记录「活动 + 周期」的名额占用。
// 周期键 daily 用 YYYY-MM-DD，整期用 fixed，跨天自动重置且无需定时清零（07 §5）。
type ActivityQuotaModel struct {
	baseModel.BaseModel
	// ActivityID 活动 ID
	ActivityID types.ID `db:"activity_id" json:"activityId"`
	// PeriodKey 周期键：整期为 fixed，每日为 YYYY-MM-DD
	PeriodKey string `db:"period_key" json:"periodKey"`
	// UsedCount 已占用名额
	UsedCount int `db:"used_count" json:"usedCount"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *ActivityQuotaModel) TableName() string {
	return ActivityQuotaTable
}

// OrderActivityTable 是订单占用活动名额的关联表名常量。
const OrderActivityTable = "promotion_order_activity"

// OrderActivityModel 对应 promotion_order_activity 表，记录某订单占用了哪些活动、哪个周期的名额。
// 名额计数行只有 (activity_id, period_key)，无法反查订单；每日周期的活动跨天后
// 用当前日期算周期键会退错池子，因此下单时把「订单 + 活动 + 周期键」三者落库，
// 取消 / 超时关单 / 整单退款时据此精确回退 used_count。
type OrderActivityModel struct {
	baseModel.BaseModel
	// OrderID 订单 ID
	OrderID types.ID `db:"order_id" json:"orderId"`
	// ActivityID 活动 ID
	ActivityID types.ID `db:"activity_id" json:"activityId"`
	// PeriodKey 占用时的周期键：整期为 fixed，每日为 YYYY-MM-DD
	PeriodKey string `db:"period_key" json:"periodKey"`
	// Status 占用状态：1 占用中 / 2 已释放
	Status int8 `db:"status" json:"status"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *OrderActivityModel) TableName() string {
	return OrderActivityTable
}

// PointRuleTable 是积分抵扣规则表名常量。
const PointRuleTable = "promotion_point_rule"

// PointRuleModel 对应 promotion_point_rule 表，定义积分抵扣的兑换比例与单笔上限。
// 规则全品牌统一，只有一条生效记录，改动即时生效（07 §2）。
type PointRuleModel struct {
	baseModel.BaseModel
	// ExchangeRate 兑换比例：多少积分抵 1 元（分），如 100 表示 100 积分抵 1 元
	ExchangeRate int `db:"exchange_rate" json:"exchangeRate"`
	// MaxDeductAmount 单笔抵扣上限（分），0 表示不限制
	MaxDeductAmount int64 `db:"max_deduct_amount" json:"maxDeductAmount"`
	// MinOrderAmount 使用积分抵扣的最低订单金额（分），0 表示无门槛
	MinOrderAmount int64 `db:"min_order_amount" json:"minOrderAmount"`
	// Status 启用状态
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *PointRuleModel) TableName() string {
	return PointRuleTable
}
