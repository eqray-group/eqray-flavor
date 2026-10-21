package service

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

const (
	// memberSignRecentLimit 计算连续签到天数时最多回看的记录条数。
	memberSignRecentLimit = 90
	// memberSignDateLayout 与 member_sign_record.sign_date 的存储格式一致。
	memberSignDateLayout = "2006-01-02"
)

// MemberSignResult 是签到成功的返回，前端据此提示获得积分与经验。
type MemberSignResult struct {
	// Day 本次是连续第几天
	Day             int      `json:"day"`
	Point           int      `json:"point"`
	Experience      int      `json:"experience"`
	TotalPoint      int      `json:"totalPoint"`
	TotalExperience int      `json:"totalExperience"`
	ContinuousDays  int      `json:"continuousDays"`
	LevelUp         bool     `json:"levelUp"`
	LevelID         types.ID `json:"levelId"`
	LevelName       string   `json:"levelName"`
}

// MemberSignStatus 是签到状态，供 C 端渲染签到日历。
type MemberSignStatus struct {
	SignedToday bool `json:"signedToday"`
	// ContinuousDays 截至今天的连续天数：今天已签算今天，未签算到昨天
	ContinuousDays int   `json:"continuousDays"`
	TotalDays      int64 `json:"totalDays"`
	// TodayDay / TodayPoint / TodayExperience 是今天这一档的档位与奖励
	TodayDay        int      `json:"todayDay"`
	TodayPoint      int      `json:"todayPoint"`
	TodayExperience int      `json:"todayExperience"`
	MonthDays       []string `json:"monthDays"`
	// NextDayPoint / NextDayExperience 明天的档位预览，便于前端展示"连续签到奖励"
	NextDayPoint      int `json:"nextDayPoint"`
	NextDayExperience int `json:"nextDayExperience"`
}

// MemberSignService 负责 C 端会员的每日签到与签到奖励档位。
// 奖励档位来自 member_sign_in_config，连续天数超过最大档位时封顶取最高档（对齐 yudao）。
type MemberSignService struct {
	members *repository.MemberRepository
	signs   *repository.MemberSignRepository
	points  *repository.MemberPointRecordRepository
	configs *repository.MemberSignConfigRepository
	levels  *repository.MemberLevelRepository
}

// NewMemberSignService 构造会员签到服务。
func NewMemberSignService() *MemberSignService {
	return &MemberSignService{
		members: repository.NewMemberRepository(),
		signs:   repository.NewMemberSignRepository(),
		points:  repository.NewMemberPointRecordRepository(),
		configs: repository.NewMemberSignConfigRepository(),
		levels:  repository.NewMemberLevelRepository(),
	}
}

// Sign 完成当日签到：写签到记录、加积分、加经验、写积分流水，等级有变化时补一条等级记录。
func (s *MemberSignService) Sign(ctx context.Context, memberID types.ID) (*MemberSignResult, error) {
	m, err := s.loadMember(ctx, memberID)
	if err != nil {
		return nil, err
	}

	today := truncateToDay(time.Now())
	signed, err := s.signs.ExistsByDate(ctx, m.ID, today)
	if err != nil {
		return nil, err
	}
	if signed {
		return nil, baseErrs.New(memberErrs.MemberSignDuplicated)
	}

	recent, err := s.signs.RecentDates(ctx, m.ID, memberSignRecentLimit)
	if err != nil {
		return nil, err
	}
	// 截至昨天的连续天数 +1，就是本次签到是第几天
	day := continuousDays(recent, today.AddDate(0, 0, -1)) + 1

	tier, err := s.tierForDay(ctx, day)
	if err != nil {
		return nil, err
	}

	oldLevelID := m.LevelID
	m.Point += tier.Point
	m.Experience += tier.Experience

	// 经验达到更高等级的门槛时自动升级，等级只升不降
	target, err := s.levelForExperience(ctx, m.Experience)
	if err != nil {
		return nil, err
	}
	levelUp := target != nil && target.ID != oldLevelID && (oldLevelID == 0 || target.Level > s.currentLevelValue(ctx, oldLevelID))
	if levelUp {
		m.LevelID = target.ID
	}

	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.signs.Create(ctx, tx, &model.MemberSignRecordModel{
			UserID:     m.ID,
			SignDate:   types.DateTime(today),
			Day:        day,
			Point:      tier.Point,
			Experience: tier.Experience,
		}); err != nil {
			if support.IsDuplicate(err) {
				return baseErrs.New(memberErrs.MemberSignDuplicated)
			}
			return err
		}

		if err := s.members.UpdatePoint(ctx, tx, m); err != nil {
			return err
		}
		if err := s.members.UpdateLevel(ctx, tx, m); err != nil {
			return err
		}

		if err := s.points.Create(ctx, tx, &model.MemberPointRecordModel{
			UserID:      m.ID,
			BizType:     MemberPointBizSign,
			Title:       "每日签到",
			Description: "签到奖励",
			Point:       tier.Point,
			TotalPoint:  m.Point,
		}); err != nil {
			return err
		}

		if !levelUp {
			return nil
		}

		return s.members.CreateLevelRecord(ctx, tx, &model.MemberLevelRecordModel{
			UserID:      m.ID,
			LevelID:     m.LevelID,
			Experience:  m.Experience,
			Title:       "签到升级",
			Description: "签到获得经验，等级提升",
		})
	})
	if err != nil {
		return nil, err
	}

	res := &MemberSignResult{
		Day:             day,
		Point:           tier.Point,
		Experience:      tier.Experience,
		TotalPoint:      m.Point,
		TotalExperience: m.Experience,
		ContinuousDays:  day,
		LevelUp:         levelUp,
		LevelID:         m.LevelID,
	}
	if levelUp && target != nil {
		res.LevelName = target.Name
	}

	return res, nil
}

// Status 返回今日签到状态、连续天数、累计天数、本月已签日期与档位预览。
func (s *MemberSignService) Status(ctx context.Context, memberID types.ID) (*MemberSignStatus, error) {
	if _, err := s.loadMember(ctx, memberID); err != nil {
		return nil, err
	}

	today := truncateToDay(time.Now())
	signed, err := s.signs.ExistsByDate(ctx, memberID, today)
	if err != nil {
		return nil, err
	}

	total, err := s.signs.Total(ctx, memberID)
	if err != nil {
		return nil, err
	}

	recent, err := s.signs.RecentDates(ctx, memberID, memberSignRecentLimit)
	if err != nil {
		return nil, err
	}
	continuous := continuousDays(recent, today)
	todayDay := continuousDays(recent, today.AddDate(0, 0, -1)) + 1

	todayTier, err := s.tierForDay(ctx, todayDay)
	if err != nil {
		return nil, err
	}
	nextTier, err := s.tierForDay(ctx, todayDay+1)
	if err != nil {
		return nil, err
	}

	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	monthDates, err := s.signs.DatesBetween(ctx, memberID, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	monthDays := make([]string, 0, len(monthDates))
	for i := range monthDates {
		monthDays = append(monthDays, monthDates[i].Time().Format(memberSignDateLayout))
	}

	return &MemberSignStatus{
		SignedToday:       signed,
		ContinuousDays:    continuous,
		TotalDays:         total,
		TodayDay:          todayDay,
		TodayPoint:        todayTier.Point,
		TodayExperience:   todayTier.Experience,
		MonthDays:         monthDays,
		NextDayPoint:      nextTier.Point,
		NextDayExperience: nextTier.Experience,
	}, nil
}

// tierForDay 取连续第 day 天对应的档位；超过最大档位时封顶取最高档。
func (s *MemberSignService) tierForDay(ctx context.Context, day int) (*model.MemberSignInConfigModel, error) {
	if day < 1 {
		day = 1
	}

	list, err := s.configs.Enabled(ctx)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, baseErrs.New(memberErrs.MemberSignConfigMissing)
	}

	// Enabled 已按 day 升序，取第一个大于等于 day 的档位；没有则用最后一个
	picked := &list[len(list)-1]
	for i := range list {
		if list[i].Day >= day {
			picked = &list[i]
			break
		}
	}

	return picked, nil
}

// levelForExperience 按累计经验匹配等级：取经验门槛不超过当前经验里最高的那一档。
func (s *MemberSignService) levelForExperience(ctx context.Context, experience int) (*model.MemberLevelModel, error) {
	levels, err := s.levels.Enabled(ctx)
	if err != nil {
		return nil, err
	}

	var target *model.MemberLevelModel
	for i := range levels {
		if levels[i].Experience > experience {
			continue
		}
		if target == nil || levels[i].Experience > target.Experience ||
			(levels[i].Experience == target.Experience && levels[i].Level > target.Level) {
			target = &levels[i]
		}
	}

	return target, nil
}

// currentLevelValue 返回当前等级序号，取不到时返回 0（视为最低等级）。
func (s *MemberSignService) currentLevelValue(ctx context.Context, levelID types.ID) int {
	if levelID == 0 {
		return 0
	}
	level, err := s.levels.FindByID(ctx, levelID)
	if err != nil || level == nil {
		return 0
	}

	return level.Level
}

// loadMember 取当前会员并校验状态，禁用会员不允许签到。
func (s *MemberSignService) loadMember(ctx context.Context, memberID types.ID) (*model.MemberUserModel, error) {
	m, err := s.members.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(memberErrs.MemberNotExists, memberID)
	}
	if m.Status != baseEnums.CommonStatusEnabled {
		return nil, baseErrs.New(memberErrs.MemberDisabled)
	}

	return m, nil
}

// continuousDays 计算连续签到天数：今天签了从今天往前数，没签则从昨天往前数。
func continuousDays(dates []types.DateTime, today time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	signed := make(map[string]struct{}, len(dates))
	for i := range dates {
		signed[dates[i].Time().Format(memberSignDateLayout)] = struct{}{}
	}

	cursor := today
	if _, ok := signed[cursor.Format(memberSignDateLayout)]; !ok {
		cursor = cursor.AddDate(0, 0, -1)
	}

	days := 0
	for {
		if _, ok := signed[cursor.Format(memberSignDateLayout)]; !ok {
			return days
		}
		days++
		cursor = cursor.AddDate(0, 0, -1)
	}
}

// truncateToDay 取当天 00:00:00，与 date 类型的 sign_date 对齐。
func truncateToDay(t time.Time) time.Time {
	t = t.In(types.Location())

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
