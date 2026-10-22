// Package model 定义门店域的数据模型。
// 表结构以 database/migrations 为准，字段注释只为说明业务语义。
package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
)

// StoreTable 是门店表名常量，Model 与 Repository 共用同一来源。
const StoreTable = "store"

// StoreModel 对应 store 表，承载总部 / 大区 / 门店三级组织树。
// parent_id 为 0 表示根节点；同表内靠 type 区分层级，不另建层级表。
type StoreModel struct {
	baseModel.BaseModel
	// ParentID 上级节点 ID，0 表示无上级（总部根节点）
	ParentID types.ID `db:"parent_id" json:"parentId"`
	// Type 节点类型：1 总部 / 2 大区 / 3 门店
	Type enums.StoreTypeEnum `db:"type" json:"type"`
	// Code 门店编码，全局唯一，创建后不可修改
	Code string `db:"code" json:"code"`
	// Name 门店名称
	Name string `db:"name" json:"name"`
	// ShortName 简称，小票与看板展示用
	ShortName string `db:"short_name" json:"shortName"`
	// Status 经营状态：1 营业 / 2 休息 / 3 装修 / 4 停业
	Status enums.StoreStatusEnum `db:"status" json:"status"`
	// ContactName 门店联系人
	ContactName string `db:"contact_name" json:"contactName"`
	// ContactMobile 联系电话
	ContactMobile string `db:"contact_mobile" json:"contactMobile"`
	// DistrictID 区县 ID，省市区通过该 ID 反查；0 表示未选择
	DistrictID types.ID `db:"district_id" json:"districtId"`
	// Address 详细地址
	Address string `db:"address" json:"address"`
	// Longitude 经度，配送范围校验用，未配置为 0
	Longitude float64 `db:"longitude" json:"longitude"`
	// Latitude 纬度，配送范围校验用，未配置为 0
	Latitude float64 `db:"latitude" json:"latitude"`
	// OpenTime 营业开始时间，格式 HH:mm（24 小时制），如 09:00
	OpenTime string `db:"open_time" json:"openTime"`
	// CloseTime 营业结束时间，格式 HH:mm；
	// 小于 OpenTime 表示跨夜营业（如 18:00 ~ 02:00），等于 OpenTime 表示全天营业
	CloseTime string `db:"close_time" json:"closeTime"`
	// DeliveryRule 配送规则 JSON，结构见 service.DeliveryRule：
	// {"radius":3000,"minAmount":2000,"deliveryFee":500,"packingFee":100,"etaMinutes":40}
	// 半径单位米（0 表示不限范围），金额单位分，时长单位分钟
	DeliveryRule string `db:"delivery_rule" json:"deliveryRule"`
	// AutoAccept 是否自动接单：1 自动 / 0 人工。
	// 开启后外卖与自提单支付成功即由系统接单，订单从「待接单」直接转「制作中」并推 KDS，
	// KDS 不再显示接单栏（design/19 §6）；关闭则需门店人工接单，超时会被拒单。
	// 堂食下单即进制作中，没有接单环节，不受该开关影响。
	AutoAccept int8 `db:"auto_accept" json:"autoAccept"`
	// OnlineEnabled 是否开放线上：0 否 / 1 是，控制顾客端是否可下单
	OnlineEnabled int8 `db:"online_enabled" json:"onlineEnabled"`
	// DeliveryEnabled 是否开启外卖：0 否 / 1 是；默认关闭，关闭时顾客端不展示外卖入口
	DeliveryEnabled int8 `db:"delivery_enabled" json:"deliveryEnabled"`
	// Sort 同级排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *StoreModel) TableName() string {
	return StoreTable
}
