package service

import (
	"context"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

// MemberSignConfigMaxDay 是签到档位允许的最大天数，避免管理员填出极端值。
const MemberSignConfigMaxDay = 31

// MemberSignConfigInput 是后台新增 / 编辑签到档位的入参。
type MemberSignConfigInput struct {
	Day        int
	Point      int
	Experience int
	Status     int8
}

// MemberSignConfigListQuery 是签到档位列表的查询条件。
type MemberSignConfigListQuery struct {
	Day      *int
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// MemberSignConfigService 负责签到奖励档位的后台维护。
type MemberSignConfigService struct {
	configs *repository.MemberSignConfigRepository
}

// NewMemberSignConfigService 构造签到奖励档位服务。
func NewMemberSignConfigService() *MemberSignConfigService {
	return &MemberSignConfigService{configs: repository.NewMemberSignConfigRepository()}
}

// Page 分页查询档位，按第几天升序，支持按天数、状态筛选。
func (s *MemberSignConfigService) Page(ctx context.Context, q MemberSignConfigListQuery) ([]model.MemberSignInConfigModel, int64, error) {
	return s.configs.Paginate(ctx, repository.MemberSignConfigFilter{
		Day:    q.Day,
		Status: q.Status,
	}, q.PageNo, q.PageSize)
}

// Create 新增档位：同一天只能有一条。
func (s *MemberSignConfigService) Create(ctx context.Context, in MemberSignConfigInput) error {
	if err := normalizeSignConfigInput(&in); err != nil {
		return err
	}

	exists, err := s.configs.ExistsByDay(ctx, in.Day, 0)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(memberErrs.MemberSignConfigDayExists, in.Day)
	}

	return s.configs.Create(ctx, &model.MemberSignInConfigModel{
		Day:        in.Day,
		Point:      in.Point,
		Experience: in.Experience,
		Status:     baseEnums.CommonStatusEnum(in.Status),
	})
}

// Update 编辑档位，同样保证第几天不重复。
func (s *MemberSignConfigService) Update(ctx context.Context, id types.ID, in MemberSignConfigInput) error {
	if err := normalizeSignConfigInput(&in); err != nil {
		return err
	}

	m, err := s.configs.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(memberErrs.MemberSignConfigNotExists, id)
	}

	exists, err := s.configs.ExistsByDay(ctx, in.Day, id)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(memberErrs.MemberSignConfigDayExists, in.Day)
	}

	m.Day = in.Day
	m.Point = in.Point
	m.Experience = in.Experience
	m.Status = baseEnums.CommonStatusEnum(in.Status)

	return s.configs.Update(ctx, m)
}

// Delete 删除档位；至少保留一条启用档位，否则签到时没有奖励可发。
func (s *MemberSignConfigService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.configs.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(memberErrs.MemberSignConfigNotExists, id)
	}

	enabled, err := s.configs.Enabled(ctx)
	if err != nil {
		return err
	}
	if m.Status == baseEnums.CommonStatusEnabled && len(enabled) <= 1 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "至少保留一条启用中的签到档位")
	}

	return s.configs.Delete(ctx, m)
}

// normalizeSignConfigInput 校验档位字段。
func normalizeSignConfigInput(in *MemberSignConfigInput) error {
	if in.Day < 1 || in.Day > MemberSignConfigMaxDay {
		return baseErrs.WithMsg(baseErrs.BadRequest, "签到天数取值范围 1 ~ 31")
	}
	if in.Point < 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "奖励积分不能为负数")
	}
	if in.Experience < 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "奖励经验不能为负数")
	}
	if !baseEnums.CommonStatusEnum(in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "状态取值不合法")
	}

	return nil
}
