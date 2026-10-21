// Package model 存放 infra 模块的表映射。
package model

import (
	"eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// InfraFileTable 是表名常量，Model 与 Repository 共用同一来源。
const InfraFileTable = "infra_file"

// InfraFileModel 对应 infra_file 表（docs 6.5）。
type InfraFileModel struct {
	model.BaseModel
	ConfigID types.ID `db:"config_id" json:"configId"`
	Name     string   `db:"name" json:"name"`
	Path     string   `db:"path" json:"path"`
	Url      string   `db:"url" json:"url"`
	Type     string   `db:"type" json:"type"`
	Size     int64    `db:"size" json:"size"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *InfraFileModel) TableName() string {
	return InfraFileTable
}
