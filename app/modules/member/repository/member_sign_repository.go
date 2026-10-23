package repository

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/member/model"
)

const signRecordTable = model.MemberSignRecordTable

// dateLayout 是 DATE 列的比较格式，sign_date 在建表时就是 date 类型。
const dateLayout = "2006-01-02"

// MemberSignRepository 是 member_sign_record 表的仓储。
// 同一会员同一天只能有一条记录，由 uk_user_date 唯一键兜底。
type MemberSignRepository struct{}

func NewMemberSignRepository() *MemberSignRepository {
	return &MemberSignRepository{}
}

// Create 在事务内写入签到记录；并发重复签到由唯一键冲突暴露，由 Service 转换错误码。
func (r *MemberSignRepository) Create(ctx context.Context, tx db.Tx, m *model.MemberSignRecordModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(signRecordTable).Insert(map[string]any{
		"id":         m.ID,
		"user_id":    m.UserID,
		"sign_date":  m.SignDate.Time().Format(dateLayout),
		"day":        m.Day,
		"point":      m.Point,
		"experience": m.Experience,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// ExistsByDate 判断某个自然日是否已签到。
func (r *MemberSignRepository) ExistsByDate(ctx context.Context, userID types.ID, date time.Time) (bool, error) {
	var m model.MemberSignRecordModel

	if err := facades.DB().WithContext(ctx).Table(signRecordTable).
		Where("user_id", userID).
		Where("sign_date", date.Format(dateLayout)).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return false, err
	}

	return m.ID != 0, nil
}

// DatesBetween 查询 [from, to) 区间内的签到日期，用于展示当月签到日历。
func (r *MemberSignRepository) DatesBetween(ctx context.Context, userID types.ID, from, to time.Time) ([]types.DateTime, error) {
	list := make([]model.MemberSignRecordModel, 0)

	if err := facades.DB().WithContext(ctx).Table(signRecordTable).
		Where("user_id", userID).
		Where("sign_date >= ?", from.Format(dateLayout)).
		Where("sign_date < ?", to.Format(dateLayout)).
		WhereNull("deleted_at").
		OrderBy("sign_date").
		Get(&list); err != nil {
		return nil, err
	}

	dates := make([]types.DateTime, 0, len(list))
	for i := range list {
		dates = append(dates, list[i].SignDate)
	}

	return dates, nil
}

// RecentDates 取最近的若干条签到日期（倒序），用于计算连续签到天数。
func (r *MemberSignRepository) RecentDates(ctx context.Context, userID types.ID, limit int) ([]types.DateTime, error) {
	list := make([]model.MemberSignRecordModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(signRecordTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		OrderByDesc("sign_date").
		Paginate(1, limit, &list, &total); err != nil {
		return nil, err
	}

	dates := make([]types.DateTime, 0, len(list))
	for i := range list {
		dates = append(dates, list[i].SignDate)
	}

	return dates, nil
}

// Total 返回会员累计签到天数。
func (r *MemberSignRepository) Total(ctx context.Context, userID types.ID) (int64, error) {
	list := make([]model.MemberSignRecordModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(signRecordTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		Paginate(1, 1, &list, &total); err != nil {
		return 0, err
	}

	return total, nil
}
