package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
)

// StorePrinterTable 是打印设备表名常量。
const StorePrinterTable = "store_printer"

// StorePrinterModel 对应 store_printer 表，表示门店绑定的打印设备。
// 同一门店内设备号唯一；密钥列只保存 AES 密文，接口只返回脱敏结果。
type StorePrinterModel struct {
	baseModel.BaseModel
	// StoreID 门店 ID，只能指向 type=3 的门店节点
	StoreID types.ID `db:"store_id" json:"storeId"`
	// Name 设备名称，便于门店识别，如「前台小票机」
	Name string `db:"name" json:"name"`
	// DeviceNo 设备号；同一门店内唯一，用于云打印机绑定与局域网识别
	DeviceNo string `db:"device_no" json:"deviceNo"`
	// Type 设备类型：1 前台小票 / 2 后厨 / 3 标签 / 4 叫号
	Type enums.PrinterTypeEnum `db:"type" json:"type"`
	// AccessMode 接入方式：1 云打印机 / 2 局域网 / 3 本地客户端
	AccessMode enums.PrinterAccessModeEnum `db:"access_mode" json:"accessMode"`
	// Config 非敏感配置 JSON，如厂商、地址、端口、客户端标识；禁止写入密钥
	Config string `db:"config" json:"config"`
	// Secret 设备密钥的 AES 密文，禁止通过接口返回
	Secret string `db:"secret" json:"-"`
	// SecretMask 密钥脱敏展示值，如 ****abcd；只保留末四位
	SecretMask string `db:"secret_mask" json:"secretMask"`
	// Copies 默认打印份数，1 ~ 9
	Copies int `db:"copies" json:"copies"`
	// Status 启停状态：0 禁用 / 1 启用
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *StorePrinterModel) TableName() string {
	return StorePrinterTable
}
