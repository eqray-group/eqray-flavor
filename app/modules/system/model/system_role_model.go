// Package model 存放 system 模块的表映射。
package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/enums"
)

// SystemRoleTable 是表名常量，Model 与 Repository 共用同一来源。
const SystemRoleTable = "system_role"

// SystemRoleModel 对应 system_role 表（docs 6.3）。
type SystemRoleModel struct {
	model.BaseModel
	Name   string                     `db:"name" json:"name"`
	Code   string                     `db:"code" json:"code"`
	Sort   int                        `db:"sort" json:"sort"`
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	Type   enums.RoleTypeEnum         `db:"type" json:"type"`
	Remark string                     `db:"remark" json:"remark"`
}

// TableName 返回表名，供仓储层构造查询时引用（docs 6.7）。
func (m *SystemRoleModel) TableName() string {
	return SystemRoleTable
}

// SystemUserRoleTable 是用户角色关联表名。
const SystemUserRoleTable = "system_user_role"

// SystemUserRoleModel 对应 system_user_role 关联表。
type SystemUserRoleModel struct {
	model.BaseModel
	UserID types.ID `db:"user_id" json:"userId"`
	RoleID types.ID `db:"role_id" json:"roleId"`
}

// TableName 返回表名。
func (m *SystemUserRoleModel) TableName() string {
	return SystemUserRoleTable
}

// SystemRoleMenuTable 是角色菜单关联表名。
const SystemRoleMenuTable = "system_role_menu"

// SystemRoleMenuModel 对应 system_role_menu 关联表。
type SystemRoleMenuModel struct {
	model.BaseModel
	RoleID types.ID `db:"role_id" json:"roleId"`
	MenuID types.ID `db:"menu_id" json:"menuId"`
}

// TableName 返回表名。
func (m *SystemRoleMenuModel) TableName() string {
	return SystemRoleMenuTable
}
