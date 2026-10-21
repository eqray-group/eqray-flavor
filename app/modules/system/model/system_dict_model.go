package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/model"
)

// SystemDictTypeTable 是字典类型表名。
const SystemDictTypeTable = "system_dict_type"

// SystemDictTypeModel 对应 system_dict_type 表（docs 6.3）。
type SystemDictTypeModel struct {
	model.BaseModel
	Name   string                     `db:"name" json:"name"`
	Type   string                     `db:"type" json:"type"`
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Remark string                     `db:"remark" json:"remark"`
}

// TableName 返回表名。
func (m *SystemDictTypeModel) TableName() string {
	return SystemDictTypeTable
}

// SystemDictDataTable 是字典数据表名。
const SystemDictDataTable = "system_dict_data"

// SystemDictDataModel 对应 system_dict_data 表（docs 6.3）。
type SystemDictDataModel struct {
	model.BaseModel
	Sort      int                        `db:"sort" json:"sort"`
	Label     string                     `db:"label" json:"label"`
	Value     string                     `db:"value" json:"value"`
	DictType  string                     `db:"dict_type" json:"dictType"`
	Status    baseEnums.CommonStatusEnum `db:"status" json:"status"`
	ColorType string                     `db:"color_type" json:"colorType"`
	CSSClass  string                     `db:"css_class" json:"cssClass"`
	Remark    string                     `db:"remark" json:"remark"`
}

// TableName 返回表名。
func (m *SystemDictDataModel) TableName() string {
	return SystemDictDataTable
}
