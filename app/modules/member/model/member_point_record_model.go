package model

import (
	baseModel "eqray-flavor/app/modules/base/model"
	"eqray-flavor/app/modules/base/types"
)

// MemberPointRecordTable 是表名常量，Model 与 Repository 共用同一来源。
const MemberPointRecordTable = "member_point_record"

// MemberPointRecordModel 对应 member_point_record 表，结构对齐 yudao 的积分记录。
// point 用正负表示增减方向，total_point 保存变动后的余额，流水中不做更新与删除。
type MemberPointRecordModel struct {
	baseModel.BaseModel
	UserID types.ID `db:"user_id" json:"userId"`
	// BizType 业务类型：0 其它 / 10 后台调整 / 20 签到 / 30 消费
	BizType int8     `db:"biz_type" json:"bizType"`
	BizID   types.ID `db:"biz_id" json:"bizId"`
	Title   string   `db:"title" json:"title"`
	// Description 流水描述
	Description string `db:"description" json:"description"`
	// Point 积分，正数增加 / 负数扣减
	Point int `db:"point" json:"point"`
	// TotalPoint 变动后的积分
	TotalPoint int `db:"total_point" json:"totalPoint"`
}

// TableName 返回表名，供仓储层构造查询时引用。
func (m *MemberPointRecordModel) TableName() string {
	return MemberPointRecordTable
}
