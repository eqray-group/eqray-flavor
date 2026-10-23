// Package request 定义订单 C 端接口的请求结构。
package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// ClientAddonItem 是 C 端明细上的一条加料选择。
type ClientAddonItem struct {
	// AddonItemID 加料项 ID，雪花 ID 按字符串提交
	AddonItemID types.ID `form:"addonItemId" json:"addonItemId"`
	// Quantity 份数，缺省 1
	Quantity int `form:"quantity" json:"quantity"`
}

// ClientLine 是 C 端的一行明细。
type ClientLine struct {
	// SkuID 菜品规格 ID，雪花 ID 按字符串提交
	SkuID types.ID `form:"skuId" json:"skuId"`
	// Quantity 数量
	Quantity int `form:"quantity" json:"quantity"`
	// Addons 已选加料
	Addons []ClientAddonItem `form:"addons" json:"addons"`
	// Orderer 点菜人，堂食多人点菜时使用
	Orderer string `form:"orderer" json:"orderer"`
	// Remark 明细备注
	Remark string `form:"remark" json:"remark"`
}

// EstimateRequest 是价格试算入参。
type EstimateRequest struct {
	// StoreID 门店 ID
	StoreID types.ID `form:"storeId" json:"storeId"`
	// Items 明细行
	Items []ClientLine `form:"items" json:"items"`
}

// Rules 返回校验规则。
func (r *EstimateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId": "required",
	}
}

// Messages 返回中文提示。
func (r *EstimateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.required": "门店不能为空",
	}
}

// CreateRequest 是 C 端下单入参。
type CreateRequest struct {
	// StoreID 门店 ID
	StoreID types.ID `form:"storeId" json:"storeId"`
	// Type 订单类型：1 堂食 / 2 外卖 / 3 自提
	Type int8 `form:"type" json:"type"`
	// TableID 堂食桌台 ID
	TableID types.ID `form:"tableId" json:"tableId"`
	// RequestNo 幂等请求号，客户端生成一次后重试复用
	RequestNo string `form:"requestNo" json:"requestNo"`
	// Remark 订单备注
	Remark string `form:"remark" json:"remark"`
	// ReceiverName / ReceiverPhone / ReceiverAddress 外卖收货信息
	ReceiverName    string `form:"receiverName" json:"receiverName"`
	ReceiverPhone   string `form:"receiverPhone" json:"receiverPhone"`
	ReceiverAddress string `form:"receiverAddress" json:"receiverAddress"`
	// Items 明细行
	Items []ClientLine `form:"items" json:"items"`
}

// Rules 返回校验规则。
func (r *CreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"storeId":         "required",
		"type":            "required|integer|min:1|max:3",
		"requestNo":       "max:64",
		"remark":          "max:255",
		"receiverName":    "max:64",
		"receiverPhone":   "max:20",
		"receiverAddress": "max:255",
	}
}

// Messages 返回中文提示。
func (r *CreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"storeId.required": "门店不能为空",
		"type.required":    "订单类型不能为空",
	}
}

// CancelRequest 是取消订单入参。
type CancelRequest struct {
	// Reason 取消原因
	Reason string `form:"reason" json:"reason"`
}

// Rules 返回校验规则。
func (r *CancelRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"reason": "max:255",
	}
}

// Messages 返回中文提示。
func (r *CancelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

// PayRequest 是为订单发起支付的入参。
type PayRequest struct {
	// AppID 支付应用 ID，按品牌配置
	AppID types.ID `form:"appId" json:"appId"`
	// ChannelCode 支付渠道编码
	ChannelCode string `form:"channelCode" json:"channelCode"`
	// OpenID 微信 JSAPI / 小程序支付必填
	OpenID string `form:"openId" json:"openId"`
}

// Rules 返回校验规则。
func (r *PayRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"appId":       "required",
		"channelCode": "required|max:32",
		"openId":      "max:128",
	}
}

// Messages 返回中文提示。
func (r *PayRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"appId.required":       "支付应用不能为空",
		"channelCode.required": "支付渠道不能为空",
	}
}
