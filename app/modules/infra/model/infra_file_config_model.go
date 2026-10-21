package model

import (
	"eqray-flavor/app/modules/base/model"
)

// InfraFileConfigTable 是表名常量，Model 与 Repository 共用同一来源。
const InfraFileConfigTable = "infra_file_config"

// InfraFileConfigModel 对应 infra_file_config 表（docs 6.4）。
type InfraFileConfigModel struct {
	model.BaseModel
	Name string `db:"name" json:"name"`
	// Storage 存储类型，当前仅支持 local
	Storage string `db:"storage" json:"storage"`
	// Master 是否主配置：1 是 / 0 否
	Master int8 `db:"master" json:"master"`
	// Config 存储配置 JSON，如 {"dir":"uploads"}
	Config string `db:"config" json:"config"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *InfraFileConfigModel) TableName() string {
	return InfraFileConfigTable
}
