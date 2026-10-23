// pos_request.go 定义 POS 收银端接口的请求结构。
// 门店不在这里传：一律取请求头 X-Store-Id，避免越权操作其它门店的订单。
package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-flavor/app/modules/base/types"
)

// PosAddonItem 是 POS 明细上的一条加料。
type PosAddonItem struct {
	// AddonItemID 加料项 ID，雪花 ID 按字符串提交
	AddonItemID types.ID `form:"addonItemId" json:"addonItemId"`
	// Quantity 份数，缺省 1
	Quantity int `form:"quantity" json:"quantity"`
}

// PosLine 是 POS 的一行明细，与 C 端同构。
type PosLine struct {
	// SkuID 菜品规格 ID
	SkuID types.ID `form:"skuId" json:"skuId"`
	// Quantity 数量
	Quantity int `form:"quantity" json:"quantity"`
	// Addons 已选加料
	Addons []PosAddonItem `form:"addons" json:"addons"`
	// Orderer 点菜人，堂食多人点菜时使用
	Orderer string `form:"orderer" json:"orderer"`
	// Remark 明细备注
	Remark string `form:"remark" json:"remark"`
}

// PosEstimateRequest 是收银端代客试算的入参。
type PosEstimateRequest struct {
	// MemberID 会员 ID，0 表示散客
	MemberID types.ID `form:"memberId" json:"memberId"`
	// Items 明细行
	Items []PosLine `form:"items" json:"items"`
	// CouponID 使用的优惠券 ID，0 表示不用券
	CouponID types.ID `form:"couponId" json:"couponId"`
	// UsePoints 是否使用积分抵扣
	UsePoints bool `form:"usePoints" json:"usePoints"`
}

// Rules 返回校验规则。
func (r *PosEstimateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"items": "required",
	}
}

// Messages 返回中文提示。
func (r *PosEstimateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"items.required": "请先选择菜品",
	}
}

// PosCreateRequest 是收银端代客下单的入参。
type PosCreateRequest struct {
	// MemberID 会员 ID，0 表示散客
	MemberID types.ID `form:"memberId" json:"memberId"`
	// Type 订单类型：1 堂食 / 2 外卖 / 3 自提
	Type int8 `form:"type" json:"type"`
	// TableID 堂食桌台 ID
	TableID types.ID `form:"tableId" json:"tableId"`
	// RequestNo 幂等请求号，收银端生成一次后重试复用
	RequestNo string `form:"requestNo" json:"requestNo"`
	// Remark 订单备注
	Remark string `form:"remark" json:"remark"`
	// ReceiverName / ReceiverPhone / ReceiverAddress 外卖收货信息
	ReceiverName    string `form:"receiverName" json:"receiverName"`
	ReceiverPhone   string `form:"receiverPhone" json:"receiverPhone"`
	ReceiverAddress string `form:"receiverAddress" json:"receiverAddress"`
	// Items 明细行
	Items []PosLine `form:"items" json:"items"`
	// CouponID 使用的优惠券 ID，0 表示不用券
	CouponID types.ID `form:"couponId" json:"couponId"`
	// UsePoints 是否使用积分抵扣
	UsePoints bool `form:"usePoints" json:"usePoints"`
}

// Rules 返回校验规则。
func (r *PosCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"type":            "required|integer|min:1|max:3",
		"items":           "required",
		"requestNo":       "max:64",
		"remark":          "max:255",
		"receiverName":    "max:64",
		"receiverPhone":   "max:20",
		"receiverAddress": "max:255",
	}
}

// Messages 返回中文提示。
func (r *PosCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.required":  "请选择订单类型",
		"items.required": "请先选择菜品",
	}
}

// PosPayRequest 是 POS 扫码收款的入参。
type PosPayRequest struct {
	// PayType 收款方式：1 微信扫码 / 2 支付宝条码
	PayType int8 `form:"payType" json:"payType"`
	// AuthCode 顾客付款码，支付宝条码必填
	AuthCode string `form:"authCode" json:"authCode"`
}

// Rules 返回校验规则。
func (r *PosPayRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"payType":  "required|integer|min:1|max:2",
		"authCode": "max:64",
	}
}

// Messages 返回中文提示。
func (r *PosPayRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"payType.required": "请选择收款方式",
	}
}
