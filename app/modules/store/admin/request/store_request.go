// Package request 定义门店后台接口的入参与校验规则。
package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// StoreSaveRequest 是新增 / 编辑门店的入参。
// code 仅新增时生效，编辑时后端会忽略并沿用库中编码。
// parentId 是雪花 ID，前端必须按字符串提交，避免 JS Number 丢精度。
type StoreSaveRequest struct {
	ParentID      types.ID `form:"parentId" json:"parentId"`
	Type          int8     `form:"type" json:"type"`
	Code          string   `form:"code" json:"code"`
	Name          string   `form:"name" json:"name"`
	ShortName     string   `form:"shortName" json:"shortName"`
	Status        int8     `form:"status" json:"status"`
	ContactName   string   `form:"contactName" json:"contactName"`
	ContactMobile string   `form:"contactMobile" json:"contactMobile"`
	// DistrictID 区县 ID，同样是雪花 ID，必须按字符串提交
	DistrictID   types.ID `form:"districtId" json:"districtId"`
	Address      string   `form:"address" json:"address"`
	Longitude    float64  `form:"longitude" json:"longitude"`
	Latitude     float64  `form:"latitude" json:"latitude"`
	OpenTime     string   `form:"openTime" json:"openTime"`
	CloseTime    string   `form:"closeTime" json:"closeTime"`
	DeliveryRule string   `form:"deliveryRule" json:"deliveryRule"`
	// AutoAccept 是否自动接单：1 外卖/自提单支付成功即自动接单，0 需门店人工接单
	AutoAccept int8 `form:"autoAccept" json:"autoAccept"`
	// OnlineEnabled 是否开放线上：1 顾客端可下单，0 只保留线下渠道
	OnlineEnabled int8 `form:"onlineEnabled" json:"onlineEnabled"`
	// DeliveryEnabled 是否开启外卖：1 开启，0 关闭；默认关闭，关闭时顾客端不展示外卖入口
	DeliveryEnabled int8   `form:"deliveryEnabled" json:"deliveryEnabled"`
	Sort            int    `form:"sort" json:"sort"`
	Remark          string `form:"remark" json:"remark"`
}

// Rules 返回校验规则，由 Controller 交给 param.Bind 执行。
func (r *StoreSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		// parentId / districtId 是雪花 ID，按字符串提交，合法性由 types.ID 反序列化兜底，
		// 不能加 integer 规则，否则 19 位字符串会被判为非法
		"type":          "required|integer|in:1,2,3",
		"code":          "required|alpha_dash|max:32",
		"name":          "required|max:64",
		"shortName":     "max:32",
		"status":        "required|integer|in:1,2,3,4",
		"contactName":   "max:32",
		"contactMobile": "max:20",
		"address":       "max:255",
		"longitude":     "numeric|between:-180,180",
		"latitude":      "numeric|between:-90,90",
		// 具体 HH:mm 合法性由 Service 用 time.Parse 校验，避免正则里的竖线与规则分隔符冲突
		"openTime":        "required|max:5",
		"closeTime":       "required|max:5",
		"deliveryRule":    "max:1000",
		"autoAccept":      "integer|in:0,1",
		"onlineEnabled":   "integer|in:0,1",
		"deliveryEnabled": "integer|in:0,1",
		"sort":            "integer|min:0",
		"remark":          "max:255",
	}
}

// Messages 覆盖框架默认的英文提示，保证接口文案统一为中文。
func (r *StoreSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.required":           "门店类型不能为空",
		"type.integer":            "门店类型必须是整数",
		"type.in":                 "门店类型取值不合法",
		"code.required":           "门店编码不能为空",
		"code.alpha_dash":         "门店编码只能包含字母、数字、下划线与短横线",
		"code.max":                "门店编码不能超过 32 个字符",
		"name.required":           "门店名称不能为空",
		"name.max":                "门店名称不能超过 64 个字符",
		"shortName.max":           "门店简称不能超过 32 个字符",
		"status.required":         "门店状态不能为空",
		"status.integer":          "门店状态必须是整数",
		"status.in":               "门店状态取值不合法",
		"contactName.max":         "联系人不能超过 32 个字符",
		"contactMobile.max":       "联系电话不能超过 20 个字符",
		"districtId.integer":      "区县 ID 必须是整数",
		"districtId.min":          "区县 ID 不能为负数",
		"address.max":             "详细地址不能超过 255 个字符",
		"longitude.numeric":       "经度必须是数字",
		"longitude.between":       "经度需在 -180 ~ 180 之间",
		"latitude.numeric":        "纬度必须是数字",
		"latitude.between":        "纬度需在 -90 ~ 90 之间",
		"openTime.required":       "营业开始时间不能为空",
		"openTime.max":            "营业开始时间格式应为 HH:mm",
		"closeTime.required":      "营业结束时间不能为空",
		"closeTime.max":           "营业结束时间格式应为 HH:mm",
		"deliveryRule.max":        "配送规则不能超过 1000 个字符",
		"autoAccept.integer":      "是否自动接单必须是整数",
		"autoAccept.in":           "是否自动接单取值不合法",
		"onlineEnabled.integer":   "是否开放线上必须是整数",
		"onlineEnabled.in":        "是否开放线上取值不合法",
		"deliveryEnabled.integer": "是否开启外卖必须是整数",
		"deliveryEnabled.in":      "是否开启外卖取值不合法",
		"sort.integer":            "排序必须是整数",
		"sort.min":                "排序不能小于 0",
		"remark.max":              "备注不能超过 255 个字符",
	}
}
