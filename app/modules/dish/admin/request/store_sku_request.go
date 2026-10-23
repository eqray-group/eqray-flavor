package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// StoreSkuSaveRequest 是保存单规格门店配置的入参。
// 门店由请求头 X-Store-Id 决定，入参中不接受 storeId，避免越权改到其它门店。
type StoreSkuSaveRequest struct {
	// Available 门店是否可售：0 下架 / 1 可售
	Available int8 `form:"available" json:"available"`
	// SoldOut 沽清标记：0 正常 / 1 已沽清
	SoldOut int8 `form:"soldOut" json:"soldOut"`
	// OnlineOnly 是否仅线上可售：0 不限 / 1 仅线上
	OnlineOnly int8 `form:"onlineOnly" json:"onlineOnly"`
	// Sort 门店内排序
	Sort int `form:"sort" json:"sort"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
	// StorePrice 门店价（分），0 表示不覆盖品牌标准价
	StorePrice int64 `form:"storePrice" json:"storePrice"`
	// StoreMemberPrice 门店会员价（分），0 表示不覆盖品牌会员价
	StoreMemberPrice int64 `form:"storeMemberPrice" json:"storeMemberPrice"`
	// StartAt 生效开始时间（yyyy-MM-dd HH:mm:ss），空表示立即生效
	StartAt string `form:"startAt" json:"startAt"`
	// EndAt 生效结束时间（yyyy-MM-dd HH:mm:ss），空表示长期有效
	EndAt string `form:"endAt" json:"endAt"`
}

// Rules 返回门店配置入参的校验规则；价格区间与浮动由服务层校验。
func (r *StoreSkuSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"available":        "required|integer|in:0,1",
		"soldOut":          "integer|in:0,1",
		"onlineOnly":       "integer|in:0,1",
		"sort":             "integer|min:0",
		"remark":           "max:255",
		"storePrice":       "integer|min:0",
		"storeMemberPrice": "integer|min:0",
		"startAt":          "max:32",
		"endAt":            "max:32",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *StoreSkuSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"available.required":       "是否可售不能为空",
		"available.integer":        "是否可售必须是整数",
		"available.in":             "是否可售只能是 0 或 1",
		"soldOut.integer":          "沽清标记必须是整数",
		"soldOut.in":               "沽清标记只能是 0 或 1",
		"onlineOnly.integer":       "仅线上标记必须是整数",
		"onlineOnly.in":            "仅线上标记只能是 0 或 1",
		"sort.integer":             "排序必须是整数",
		"sort.min":                 "排序不能小于 0",
		"remark.max":               "备注不能超过 255 个字符",
		"storePrice.integer":       "门店价必须是整数（单位分）",
		"storePrice.min":           "门店价不能为负数",
		"storeMemberPrice.integer": "门店会员价必须是整数（单位分）",
		"storeMemberPrice.min":     "门店会员价不能为负数",
		"startAt.max":              "生效开始时间格式不正确",
		"endAt.max":                "生效结束时间格式不正确",
	}
}

// StoreSkuBatchRequest 是批量调整门店可售开关的入参。
type StoreSkuBatchRequest struct {
	// SkuIDs 要调整的规格 ID 列表
	SkuIDs []types.ID `form:"skuIds" json:"skuIds"`
	// Available 可售开关，nil 表示不改
	Available *int8 `form:"available" json:"available"`
	// SoldOut 沽清开关，nil 表示不改
	SoldOut *int8 `form:"soldOut" json:"soldOut"`
	// OnlineOnly 仅线上开关，nil 表示不改
	OnlineOnly *int8 `form:"onlineOnly" json:"onlineOnly"`
}

// Rules 返回批量调整入参的校验规则。
func (r *StoreSkuBatchRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"skuIds":     "required",
		"available":  "integer|in:0,1",
		"soldOut":    "integer|in:0,1",
		"onlineOnly": "integer|in:0,1",
	}
}

// Messages 覆盖框架默认英文提示。
func (r *StoreSkuBatchRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"skuIds.required": "请至少选择一个规格",
		"available.in":    "是否可售只能是 0 或 1",
		"soldOut.in":      "沽清标记只能是 0 或 1",
		"onlineOnly.in":   "仅线上标记只能是 0 或 1",
	}
}
