package model

import (
	"eqray-flavor/app/modules/base/model"
)

// InfraConfigTable 是表名常量，Model 与 Repository 共用同一来源。
const InfraConfigTable = "infra_config"

// InfraConfigModel 对应 infra_config 表（docs 6.4）。
type InfraConfigModel struct {
	model.BaseModel
	Category  string `db:"category" json:"category"`
	Name      string `db:"name" json:"name"`
	ConfigKey string `db:"config_key" json:"configKey"`
	Value     string `db:"value" json:"value"`
	// Type 1 系统内置 / 2 自定义
	Type int8 `db:"type" json:"type"`
	// Visible 0 否 / 1 是，决定是否允许通过公开接口读取
	Visible int8   `db:"visible" json:"visible"`
	Remark  string `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *InfraConfigModel) TableName() string {
	return InfraConfigTable
}
