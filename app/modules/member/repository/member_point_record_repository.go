package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/member/model"
)

const pointRecordTable = model.MemberPointRecordTable

// MemberPointRecordRepository 是 member_point_record 表的仓储。
// 流水只增不改，因此不提供 Update / Delete。
type MemberPointRecordRepository struct{}

func NewMemberPointRecordRepository() *MemberPointRecordRepository {
	return &MemberPointRecordRepository{}
}

// Create 在事务内写入一条积分记录。
func (r *MemberPointRecordRepository) Create(ctx context.Context, tx db.Tx, m *model.MemberPointRecordModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(pointRecordTable).Insert(map[string]any{
		"id":          m.ID,
		"user_id":     m.UserID,
		"biz_type":    m.BizType,
		"biz_id":      m.BizID,
		"title":       m.Title,
		"description": m.Description,
		"point":       m.Point,
		"total_point": m.TotalPoint,
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})

	return err
}

// Paginate 按会员分页查询积分记录，最新的在前。
func (r *MemberPointRecordRepository) Paginate(ctx context.Context, userID types.ID, pageNo, pageSize int) ([]model.MemberPointRecordModel, int64, error) {
	list := make([]model.MemberPointRecordModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(pointRecordTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		OrderByDesc("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
