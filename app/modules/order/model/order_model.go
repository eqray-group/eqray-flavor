// Package model 定义 order 模块的数据模型，对应交易域的订单主表与从表。
package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/order/enums"
)

// CaterOrderTable 是订单主表表名常量。
const CaterOrderTable = "cater_order"

// CaterOrderModel 对应 cater_order 表，表示一次消费凭证。
// 状态与支付状态分离：status 走订单状态机，pay_status 与 pay 模块对齐。
// 金额一律整型分；优惠分项与合计分开存，discount_snapshot 保留完整计算过程。
type CaterOrderModel struct {
	baseModel.BaseModel
	// No 订单号，对外展示，全局唯一
	No string `db:"no" json:"no"`
	// StoreID 门店 ID
	StoreID types.ID `db:"store_id" json:"storeId"`
	// MemberID 会员 ID，0 表示散客
	MemberID types.ID `db:"member_id" json:"memberId"`
	// Type 订单类型：1 堂食 / 2 外卖 / 3 自提
	Type enums.OrderType `db:"type" json:"type"`
	// Channel 下单渠道：1 小程序 / 2 POS / 3 自助机 / 4 外卖平台
	Channel enums.OrderChannel `db:"channel" json:"channel"`
	// TableID 堂食桌台 ID，非堂食为 0
	TableID types.ID `db:"table_id" json:"tableId"`
	// TableTitle 桌台名称快照（Go 字段避开 TableName() 方法，列名仍是 table_name）
	TableTitle string `db:"table_name" json:"tableName"`
	// RequestNo 幂等请求号，(store_id, request_no) 唯一
	RequestNo string `db:"request_no" json:"-"`
	// Status 订单状态：100 待支付 … 800 已退款
	Status enums.OrderStatus `db:"status" json:"status"`
	// PayStatus 支付状态：0 未支付 / 10 已支付 / 20 已退款
	PayStatus enums.PayStatus `db:"pay_status" json:"payStatus"`
	// ItemAmount 商品小计（分），不含加料
	ItemAmount int64 `db:"item_amount" json:"itemAmount"`
	// AddonAmount 加料费合计（分）
	AddonAmount int64 `db:"addon_amount" json:"addonAmount"`
	// LineDiscountAmount 单品级优惠合计（分），来自单品券 / 加价购
	LineDiscountAmount int64 `db:"line_discount_amount" json:"lineDiscountAmount"`
	// ActivityAmount 活动优惠合计（分）
	ActivityAmount int64 `db:"activity_amount" json:"activityAmount"`
	// CouponAmount 优惠券抵扣（分）
	CouponAmount int64 `db:"coupon_amount" json:"couponAmount"`
	// PointsAmount 积分抵扣（分）
	PointsAmount int64 `db:"points_amount" json:"pointsAmount"`
	// PointsUsed 实际使用的积分数。
	// 下单时按当时生效的兑换比例把 PointsAmount 折算成积分数并落库，
	// 支付时按这个数扣减、退款时按这个数返还，
	// 避免兑换比例调整后出现「扣的积分与退的积分对不上」。
	PointsUsed int `db:"points_used" json:"pointsUsed"`
	// PayAmount 应付金额（分）= item_amount + addon_amount - activity_amount - coupon_amount - points_amount
	PayAmount int64 `db:"pay_amount" json:"payAmount"`
	// RefundAmount 已退款金额（分）
	RefundAmount int64 `db:"refund_amount" json:"refundAmount"`
	// DiscountSnapshot 优惠计算过程快照 JSON，可完整复现计价结果
	DiscountSnapshot string `db:"discount_snapshot" json:"discountSnapshot"`
	// PickupCode 取餐码
	PickupCode string `db:"pickup_code" json:"pickupCode"`
	// ReceiverName 收货人姓名，外卖订单使用
	ReceiverName string `db:"receiver_name" json:"receiverName"`
	// ReceiverPhone 收货人电话
	ReceiverPhone string `db:"receiver_phone" json:"receiverPhone"`
	// ReceiverAddress 收货地址
	ReceiverAddress string `db:"receiver_address" json:"receiverAddress"`
	// Remark 订单备注
	Remark string `db:"remark" json:"remark"`
	// MerchantRemark 商家备注，与顾客备注分列存放，店员改动不影响顾客原始诉求
	MerchantRemark string `db:"merchant_remark" json:"merchantRemark"`
	// AdjustAmount 改价差额（分）：正数加价、负数减免，pay_amount 已包含该差额
	AdjustAmount int64 `db:"adjust_amount" json:"adjustAmount"`
	// RiderID 配送骑手 ID，0 表示未派单
	RiderID types.ID `db:"rider_id" json:"riderId"`
	// DispatchTime 派单时间
	DispatchTime *types.DateTime `db:"dispatch_time" json:"dispatchTime"`
	// PayOrderID 关联支付单 ID
	PayOrderID types.ID `db:"pay_order_id" json:"payOrderId"`
	// PayTime 支付成功时间
	PayTime *types.DateTime `db:"pay_time" json:"payTime"`
	// AcceptTime 接单时间
	AcceptTime *types.DateTime `db:"accept_time" json:"acceptTime"`
	// CompleteTime 完成时间
	CompleteTime *types.DateTime `db:"complete_time" json:"completeTime"`
	// CancelTime 取消时间
	CancelTime *types.DateTime `db:"cancel_time" json:"cancelTime"`
	// CancelReason 取消 / 拒单原因
	CancelReason string `db:"cancel_reason" json:"cancelReason"`
	// ExpireTime 待支付超时时间，超时由定时任务关单
	ExpireTime types.DateTime `db:"expire_time" json:"expireTime"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *CaterOrderModel) TableName() string {
	return CaterOrderTable
}

// CaterOrderItemTable 是订单明细表名常量。
const CaterOrderItemTable = "cater_order_item"

// CaterOrderItemModel 对应 cater_order_item 表，是下单时的菜品快照。
// 名称与价格全部落库，菜品后续改价不影响历史订单；item_discount_amount 是整单优惠按小计占比分摊的结果。
type CaterOrderItemModel struct {
	baseModel.BaseModel
	// OrderID 订单 ID
	OrderID types.ID `db:"order_id" json:"orderId"`
	// SpuID 菜品 SPU ID
	SpuID types.ID `db:"spu_id" json:"spuId"`
	// SkuID 菜品规格 SKU ID
	SkuID types.ID `db:"sku_id" json:"skuId"`
	// SpuName 菜品名称快照
	SpuName string `db:"spu_name" json:"spuName"`
	// SkuName 规格名称快照
	SkuName string `db:"sku_name" json:"skuName"`
	// OriginalPrice 下单时标准价 / 门店价（分），用于对账
	OriginalPrice int64 `db:"original_price" json:"originalPrice"`
	// UnitPrice 成交单价（分），含会员价与单品级优惠后的价格
	UnitPrice int64 `db:"unit_price" json:"unitPrice"`
	// Quantity 数量
	Quantity int `db:"quantity" json:"quantity"`
	// AddonAmount 加料费合计（分）
	AddonAmount int64 `db:"addon_amount" json:"addonAmount"`
	// ItemDiscountAmount 分摊到本明细的整单优惠（分）
	ItemDiscountAmount int64 `db:"item_discount_amount" json:"itemDiscountAmount"`
	// Subtotal 小计（分）= unit_price * quantity + addon_amount - item_discount_amount
	Subtotal int64 `db:"subtotal" json:"subtotal"`
	// Status 明细状态：1 正常 / 2 已退
	Status int8 `db:"status" json:"status"`
	// Orderer 点菜人，堂食多人点菜时使用
	Orderer string `db:"orderer" json:"orderer"`
	// Remark 明细备注，如少辣
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *CaterOrderItemModel) TableName() string {
	return CaterOrderItemTable
}

// CaterOrderItemAddonTable 是明细加料快照表名常量。
const CaterOrderItemAddonTable = "cater_order_item_addon"

// CaterOrderItemAddonModel 对应 cater_order_item_addon 表，记录明细上的加料快照。
type CaterOrderItemAddonModel struct {
	baseModel.BaseModel
	// OrderID 订单 ID
	OrderID types.ID `db:"order_id" json:"orderId"`
	// OrderItemID 订单明细 ID
	OrderItemID types.ID `db:"order_item_id" json:"orderItemId"`
	// AddonItemID 加料项 ID
	AddonItemID types.ID `db:"addon_item_id" json:"addonItemId"`
	// AddonName 加料名称快照
	AddonName string `db:"addon_name" json:"addonName"`
	// Price 单份加价（分）
	Price int64 `db:"price" json:"price"`
	// Quantity 份数
	Quantity int `db:"quantity" json:"quantity"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *CaterOrderItemAddonModel) TableName() string {
	return CaterOrderItemAddonTable
}

// CaterOrderLogTable 是订单日志表名常量。
const CaterOrderLogTable = "cater_order_log"

// CaterOrderLogModel 对应 cater_order_log 表，每次状态流转写一条，只增不改。
type CaterOrderLogModel struct {
	baseModel.BaseModel
	// OrderID 订单 ID
	OrderID types.ID `db:"order_id" json:"orderId"`
	// FromStatus 原状态，0 表示初始创建
	FromStatus int16 `db:"from_status" json:"fromStatus"`
	// ToStatus 新状态
	ToStatus int16 `db:"to_status" json:"toStatus"`
	// OperatorType 操作方：1 顾客 / 2 店员 / 3 系统
	OperatorType int8 `db:"operator_type" json:"operatorType"`
	// OperatorID 操作人 ID，系统为 0
	OperatorID types.ID `db:"operator_id" json:"operatorId"`
	// OperatorName 操作人名称或系统动作描述
	OperatorName string `db:"operator_name" json:"operatorName"`
	// Reason 变更原因，如拒单理由、超时关单
	Reason string `db:"reason" json:"reason"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *CaterOrderLogModel) TableName() string {
	return CaterOrderLogTable
}

// CaterCartTable 是购物车表名常量。
const CaterCartTable = "cater_cart"

// CaterCartModel 对应 cater_cart 表，一个会员在一个门店只有一张购物车。
type CaterCartModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID
	StoreID types.ID `db:"store_id" json:"storeId"`
	// MemberID 会员 ID
	MemberID types.ID `db:"member_id" json:"memberId"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *CaterCartModel) TableName() string {
	return CaterCartTable
}

// CaterCartItemTable 是购物车明细表名常量。
const CaterCartItemTable = "cater_cart_item"

// CaterCartItemModel 对应 cater_cart_item 表，记录购物车里的一次选购。
type CaterCartItemModel struct {
	baseModel.BaseModel
	// CartID 购物车 ID
	CartID types.ID `db:"cart_id" json:"cartId"`
	// SkuID 菜品规格 SKU ID
	SkuID types.ID `db:"sku_id" json:"skuId"`
	// Quantity 数量
	Quantity int `db:"quantity" json:"quantity"`
	// Addons 已选加料 JSON：[{addonItemId, quantity}]
	Addons string `db:"addons" json:"addons"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *CaterCartItemModel) TableName() string {
	return CaterCartItemTable
}
