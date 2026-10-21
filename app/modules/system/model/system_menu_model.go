// Package model 存放 system 模块的表映射。
package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/enums"
)

// SystemMenuTable 是表名常量，Model 与 Repository 共用同一来源。
const SystemMenuTable = "system_menu"

// SystemMenuModel 对应 system_menu 表，目录 / 菜单 / 按钮同一张表（docs 6.3）。
type SystemMenuModel struct {
	model.BaseModel
	Name          string                     `db:"name" json:"name"`
	Permission    string                     `db:"permission" json:"permission"`
	Type          enums.MenuTypeEnum         `db:"type" json:"type"`
	Sort          int                        `db:"sort" json:"sort"`
	ParentID      types.ID                   `db:"parent_id" json:"parentId"`
	Path          string                     `db:"path" json:"path"`
	Icon          string                     `db:"icon" json:"icon"`
	Component     string                     `db:"component" json:"component"`
	ComponentName string                     `db:"component_name" json:"componentName"`
	Status        baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Visible       int8                       `db:"visible" json:"visible"`
	KeepAlive     int8                       `db:"keep_alive" json:"keepAlive"`
	AlwaysShow    int8                       `db:"always_show" json:"alwaysShow"`
}

// TableName 返回表名，供仓储层构造查询时引用（docs 6.7）。
func (m *SystemMenuModel) TableName() string {
	return SystemMenuTable
}
