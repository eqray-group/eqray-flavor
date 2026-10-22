package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// InfraDistrictTable 是表名常量，Model 与 Repository 共用同一来源。
const InfraDistrictTable = "infra_district"

// InfraDistrictModel 对应 infra_district 表，承载省 / 市 / 区县 / 街道四级行政区划。
// 同一张表用 pid 自引用成树，level 由上级推导，area_code 与第三方数据对齐。
type InfraDistrictModel struct {
	baseModel.BaseModel
	// Pid 上级区划 ID，0 表示顶级（省）
	Pid types.ID `db:"pid" json:"pid"`
	// AreaName 区划名称，如「杭州市」
	AreaName string `db:"area_name" json:"areaName"`
	// ShortName 简称，如「杭州」
	ShortName string `db:"short_name" json:"shortName"`
	// AreaCode 行政编码，如 330100
	AreaCode string `db:"area_code" json:"areaCode"`
	// ParentAreaCode 父级行政编码，顶级为空串
	ParentAreaCode string `db:"parent_area_code" json:"parentAreaCode"`
	// Level 行政级别：1 省 / 2 市 / 3 区县 / 4 街道，由上级自动推导
	Level int `db:"level" json:"level"`
	// Sort 同级排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *InfraDistrictModel) TableName() string {
	return InfraDistrictTable
}
