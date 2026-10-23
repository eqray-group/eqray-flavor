// Package request 定义 promotion 管理端接口的请求结构。
package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// CouponTemplatePageRequest 是券模板列表查询入参。
type CouponTemplatePageRequest struct {
	// Keyword 券名称模糊匹配
	Keyword string `form:"keyword" json:"keyword"`
	// Type 券类型，0 表示不限
	Type int8 `form:"type" json:"type"`
	// Status 启用状态，缺省不限
	Status int8 `form:"status" json:"status"`
	// PageNo 页码
	PageNo int `form:"pageNo" json:"pageNo"`
	// PageSize 页大小
	PageSize int `form:"pageSize" json:"pageSize"`
}

// Rules 返回校验规则。
func (r *CouponTemplatePageRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"keyword": "max:64",
		"type":    "integer|min:0|max:5",
	}
}

// Messages 返回中文提示。
func (r *CouponTemplatePageRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

// CouponTemplateSaveRequest 是券模板新增 / 编辑入参。
type CouponTemplateSaveRequest struct {
	// ID 模板 ID，0 表示新增；编辑时由前端回填
	ID types.ID `form:"id" json:"id"`
	// Name 券名称
	Name string `form:"name" json:"name"`
	// Type 券类型：1 满减 / 2 折扣 / 3 单品 / 4 兑换 / 5 代金
	Type int8 `form:"type" json:"type"`
	// Description 券说明
	Description string `form:"description" json:"description"`
	// ThresholdAmount 使用门槛（分）
	ThresholdAmount int64 `form:"thresholdAmount" json:"thresholdAmount"`
	// DiscountAmount 抵扣金额（分）
	DiscountAmount int64 `form:"discountAmount" json:"discountAmount"`
	// DiscountRate 折扣率（万分比），7500 表示 75 折
	DiscountRate int32 `form:"discountRate" json:"discountRate"`
	// MaxDiscountAmount 折扣封顶（分）
	MaxDiscountAmount int64 `form:"maxDiscountAmount" json:"maxDiscountAmount"`
	// Scope 适用范围：1 全品牌 / 2 指定门店 / 3 指定商品
	Scope int8 `form:"scope" json:"scope"`
	// ScopeIDs 适用对象 ID 列表
	ScopeIDs []types.ID `form:"scopeIds" json:"scopeIds"`
	// ChannelLimit 渠道限制：0 不限 / 1 仅线上 / 2 仅线下
	ChannelLimit int8 `form:"channelLimit" json:"channelLimit"`
	// IssueType 发放方式：1 主动领取 / 2 消费返券 / 3 新人礼 / 4 定向发券
	IssueType int8 `form:"issueType" json:"issueType"`
	// TotalCount 发行总量，0 表示不限
	TotalCount int `form:"totalCount" json:"totalCount"`
	// PerMemberLimit 每人限领张数，0 表示不限
	PerMemberLimit int `form:"perMemberLimit" json:"perMemberLimit"`
	// StartTime 生效开始时间
	StartTime string `form:"startTime" json:"startTime"`
	// EndTime 生效结束时间
	EndTime string `form:"endTime" json:"endTime"`
	// ValidDays 领取后有效天数
	ValidDays int `form:"validDays" json:"validDays"`
	// Status 启用状态：0 关闭 / 1 开启
	Status int8 `form:"status" json:"status"`
}

// Rules 返回校验规则。
func (r *CouponTemplateSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":              "required|max:64",
		"type":              "required|integer|min:1|max:5",
		"description":       "max:255",
		"thresholdAmount":   "integer|min:0",
		"discountAmount":    "integer|min:0",
		"discountRate":      "integer|min:0|max:9999",
		"maxDiscountAmount": "integer|min:0",
		"scope":             "required|integer|min:1|max:3",
		"channelLimit":      "integer|min:0|max:2",
		"issueType":         "required|integer|min:1|max:4",
		"totalCount":        "integer|min:0",
		"perMemberLimit":    "integer|min:0",
		"startTime":         "required",
		"endTime":           "required",
		"validDays":         "integer|min:0|max:3650",
		"status":            "integer|min:0|max:1",
	}
}

// Messages 返回中文提示。
func (r *CouponTemplateSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":      "券名称不能为空",
		"name.max":           "券名称不能超过 64 个字符",
		"type.required":      "请选择券类型",
		"type.integer":       "券类型必须是整数",
		"scope.required":     "请选择适用范围",
		"issueType.required": "请选择发放方式",
		"startTime.required": "生效开始时间不能为空",
		"endTime.required":   "生效结束时间不能为空",
	}
}

// CouponTemplateIssueRequest 是发券入参。
type CouponTemplateIssueRequest struct {
	// MemberID 目标会员 ID
	MemberID types.ID `form:"memberId" json:"memberId"`
	// Count 发放张数，缺省 1
	Count int `form:"count" json:"count"`
}

// Rules 返回校验规则。
func (r *CouponTemplateIssueRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"memberId": "required",
		"count":    "integer|min:1|max:100",
	}
}

// Messages 返回中文提示。
func (r *CouponTemplateIssueRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"memberId.required": "请选择会员",
	}
}

// CouponPageRequest 是会员券列表查询入参。
type CouponPageRequest struct {
	// TemplateID 券模板，0 表示不限
	TemplateID types.ID `form:"templateId" json:"templateId"`
	// MemberID 会员，0 表示不限
	MemberID types.ID `form:"memberId" json:"memberId"`
	// Status 券状态，缺省不限
	Status int8 `form:"status" json:"status"`
	// PageNo 页码
	PageNo int `form:"pageNo" json:"pageNo"`
	// PageSize 页大小
	PageSize int `form:"pageSize" json:"pageSize"`
}

// Rules 返回校验规则。
func (r *CouponPageRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"status": "integer|min:1|max:4",
	}
}

// Messages 返回中文提示。
func (r *CouponPageRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

// ActivityPageRequest 是营销活动列表查询入参。
type ActivityPageRequest struct {
	// Keyword 活动名称模糊匹配
	Keyword string `form:"keyword" json:"keyword"`
	// Type 活动玩法，0 表示不限
	Type int8 `form:"type" json:"type"`
	// Status 启用状态，缺省不限
	Status int8 `form:"status" json:"status"`
	// PageNo 页码
	PageNo int `form:"pageNo" json:"pageNo"`
	// PageSize 页大小
	PageSize int `form:"pageSize" json:"pageSize"`
}

// Rules 返回校验规则。
func (r *ActivityPageRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"keyword": "max:64",
		"type":    "integer|min:0|max:5",
	}
}

// Messages 返回中文提示。
func (r *ActivityPageRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

// ActivityTierRequest 是阶梯满减的一档。
type ActivityTierRequest struct {
	// ThresholdAmount 该档门槛（分）
	ThresholdAmount int64 `form:"thresholdAmount" json:"thresholdAmount"`
	// DiscountAmount 该档减免金额（分）
	DiscountAmount int64 `form:"discountAmount" json:"discountAmount"`
}

// ActivityComboItemRequest 是套餐中的一个商品。
type ActivityComboItemRequest struct {
	// SpuID 商品 SPU ID
	SpuID types.ID `form:"spuId" json:"spuId"`
	// Quantity 数量
	Quantity int `form:"quantity" json:"quantity"`
}

// ActivitySaveRequest 是营销活动新增 / 编辑入参。
type ActivitySaveRequest struct {
	// ID 活动 ID，0 表示新增；编辑时由前端回填
	ID types.ID `form:"id" json:"id"`
	// Name 活动名称
	Name string `form:"name" json:"name"`
	// Type 活动玩法：1 满减 / 2 第二份半价 / 3 限时折扣 / 4 套餐特惠 / 5 加价购
	Type int8 `form:"type" json:"type"`
	// Level 作用层级：1 商品级 / 2 订单级
	Level int8 `form:"level" json:"level"`
	// Description 活动说明
	Description string `form:"description" json:"description"`
	// ThresholdAmount 满减门槛（分）
	ThresholdAmount int64 `form:"thresholdAmount" json:"thresholdAmount"`
	// DiscountAmount 减免金额（分）
	DiscountAmount int64 `form:"discountAmount" json:"discountAmount"`
	// Tiers 阶梯满减档位
	Tiers []ActivityTierRequest `form:"tiers" json:"tiers"`
	// DiscountRate 折扣率（万分比）
	DiscountRate int32 `form:"discountRate" json:"discountRate"`
	// MaxDiscountAmount 折扣封顶（分）
	MaxDiscountAmount int64 `form:"maxDiscountAmount" json:"maxDiscountAmount"`
	// SecondHalfRate 第二份折扣率（万分比）
	SecondHalfRate int32 `form:"secondHalfRate" json:"secondHalfRate"`
	// ComboItems 套餐商品组合
	ComboItems []ActivityComboItemRequest `form:"comboItems" json:"comboItems"`
	// ComboPrice 套餐价（分）
	ComboPrice int64 `form:"comboPrice" json:"comboPrice"`
	// AddonSpuID 加价购商品 SPU ID
	AddonSpuID types.ID `form:"addonSpuId" json:"addonSpuId"`
	// AddonPrice 加价购价格（分）
	AddonPrice int64 `form:"addonPrice" json:"addonPrice"`
	// Scope 适用门店范围：1 全品牌 / 2 指定门店
	Scope int8 `form:"scope" json:"scope"`
	// ScopeIDs 适用门店 ID 列表
	ScopeIDs []types.ID `form:"scopeIds" json:"scopeIds"`
	// Priority 优先级
	Priority int `form:"priority" json:"priority"`
	// MutexGroup 互斥组
	MutexGroup string `form:"mutexGroup" json:"mutexGroup"`
	// Exclusive 是否独占
	Exclusive bool `form:"exclusive" json:"exclusive"`
	// Stackable 是否可叠加
	Stackable bool `form:"stackable" json:"stackable"`
	// QuotaLimit 名额上限，0 表示不限
	QuotaLimit int `form:"quotaLimit" json:"quotaLimit"`
	// QuotaCycle 名额周期：1 整个活动期 / 2 每日重置
	QuotaCycle int8 `form:"quotaCycle" json:"quotaCycle"`
	// StartTime 活动开始时间
	StartTime string `form:"startTime" json:"startTime"`
	// EndTime 活动结束时间
	EndTime string `form:"endTime" json:"endTime"`
	// Status 启用状态：0 未启用 / 1 已启用
	Status int8 `form:"status" json:"status"`
	// ProductSPUIDs 参与商品的 SPU ID 列表
	ProductSPUIDs []types.ID `form:"productSpuIds" json:"productSpuIds"`
}

// Rules 返回校验规则。
func (r *ActivitySaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":              "required|max:64",
		"type":              "required|integer|min:1|max:5",
		"level":             "required|integer|min:1|max:2",
		"description":       "max:255",
		"thresholdAmount":   "integer|min:0",
		"discountAmount":    "integer|min:0",
		"discountRate":      "integer|min:0|max:9999",
		"maxDiscountAmount": "integer|min:0",
		"secondHalfRate":    "integer|min:0|max:9999",
		"comboPrice":        "integer|min:0",
		"addonPrice":        "integer|min:0",
		"scope":             "required|integer|min:1|max:2",
		"priority":          "integer|min:0|max:9999",
		"mutexGroup":        "max:64",
		"quotaLimit":        "integer|min:0",
		"quotaCycle":        "integer|min:1|max:2",
		"startTime":         "required",
		"endTime":           "required",
		"status":            "integer|min:0|max:1",
	}
}

// Messages 返回中文提示。
func (r *ActivitySaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":      "活动名称不能为空",
		"name.max":           "活动名称不能超过 64 个字符",
		"type.required":      "请选择活动玩法",
		"level.required":     "请选择作用层级",
		"scope.required":     "请选择适用门店范围",
		"startTime.required": "活动开始时间不能为空",
		"endTime.required":   "活动结束时间不能为空",
	}
}

// PointRuleSaveRequest 是积分抵扣规则保存入参。
type PointRuleSaveRequest struct {
	// ExchangeRate 兑换比例：多少积分抵 1 元（分）
	ExchangeRate int `form:"exchangeRate" json:"exchangeRate"`
	// MaxDeductAmount 单笔抵扣上限（分），0 表示不限制
	MaxDeductAmount int64 `form:"maxDeductAmount" json:"maxDeductAmount"`
	// MinOrderAmount 使用积分抵扣的最低订单金额（分）
	MinOrderAmount int64 `form:"minOrderAmount" json:"minOrderAmount"`
	// Status 启用状态：0 关闭 / 1 开启
	Status int8 `form:"status" json:"status"`
}

// Rules 返回校验规则。
func (r *PointRuleSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"exchangeRate":     "required|integer|min:1",
		"maxDeductAmount":  "integer|min:0",
		"minOrderAmount":   "integer|min:0",
		"status":           "integer|min:0|max:1",
	}
}

// Messages 返回中文提示。
func (r *PointRuleSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"exchangeRate.required": "兑换比例不能为空",
		"exchangeRate.min":      "兑换比例必须大于 0",
	}
}
