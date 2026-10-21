// Package model 存放 system 模块的表映射。
package model

import (
	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// SystemUserTable 是表名常量，Model 与 Repository 共用同一来源。
const SystemUserTable = "system_users"

// SystemUserModel 对应 system_users 表（docs 6.3）。
type SystemUserModel struct {
	model.BaseModel
	Username    string                     `db:"username" json:"username"`
	Password    string                     `db:"password" json:"-"`
	Nickname    string                     `db:"nickname" json:"nickname"`
	Remark      string                     `db:"remark" json:"remark"`
	Email       *string                    `db:"email" json:"email"`
	Mobile      *string                    `db:"mobile" json:"mobile"`
	Sex         int8                       `db:"sex" json:"sex"`
	Avatar      string                     `db:"avatar" json:"avatar"`
	Status      baseEnums.CommonStatusEnum `db:"status" json:"status"`
	LoginIP     string                     `db:"login_ip" json:"loginIp"`
	LastLoginAt *types.DateTime            `db:"last_login_at" json:"lastLoginAt"`
}

// TableName 返回表名，供仓储层构造查询时引用（docs 6.7）。
func (m *SystemUserModel) TableName() string {
	return SystemUserTable
}
