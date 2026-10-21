package service

import (
	"context"
	"strconv"
	"strings"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

// MemberProfile 是 C 端「我的资料」的返回结构。
// 只暴露会员自己能看懂的字段，后台用的 mark / 分组 ID / 登录 IP 等不回传。
type MemberProfile struct {
	ID        types.ID        `json:"id"`
	Mobile    string          `json:"mobile"`
	Nickname  string          `json:"nickname"`
	Avatar    string          `json:"avatar"`
	Name      string          `json:"name"`
	Sex       int8            `json:"sex"`
	Birthday  *types.DateTime `json:"birthday"`
	AreaID    int             `json:"areaId"`
	LevelID   types.ID        `json:"levelId"`
	Level     int             `json:"level"`
	LevelName string          `json:"levelName"`
	Point     int             `json:"point"`
	// Experience 当前经验值，NextLevelExperience 升级所需经验（满级时与 Experience 相同）
	Experience          int      `json:"experience"`
	NextLevelExperience int      `json:"nextLevelExperience"`
	NextLevelName       string   `json:"nextLevelName"`
	GroupName           string   `json:"groupName"`
	TagIDs              string   `json:"tagIds"`
	TagNames            []string `json:"tagNames"`
	CreatedAt           string   `json:"createdAt"`
}

// MemberProfileUpdateInput 是 C 端可自助修改的资料字段。
// 手机号需要走换绑验证码流程，积分 / 等级 / 状态由后台维护，都不在这里。
type MemberProfileUpdateInput struct {
	Nickname string
	Avatar   string
	Name     string
	Sex      int8
	Birthday *types.DateTime
	AreaID   int
}

// MemberProfileService 负责 C 端会员资料的自助查询与修改。
type MemberProfileService struct {
	members *repository.MemberRepository
	levels  *repository.MemberLevelRepository
	tags    *repository.MemberTagRepository
	groups  *repository.MemberGroupRepository
}

// NewMemberProfileService 构造会员资料服务，装配会员、等级、标签与分组仓储。
func NewMemberProfileService() *MemberProfileService {
	return &MemberProfileService{
		members: repository.NewMemberRepository(),
		levels:  repository.NewMemberLevelRepository(),
		tags:    repository.NewMemberTagRepository(),
		groups:  repository.NewMemberGroupRepository(),
	}
}

// Profile 组装当前会员的资料，一次性带上等级名、分组名与标签名。
func (s *MemberProfileService) Profile(ctx context.Context, memberID types.ID) (*MemberProfile, error) {
	m, err := s.mustGet(ctx, memberID)
	if err != nil {
		return nil, err
	}

	profile := &MemberProfile{
		ID:         m.ID,
		Mobile:     m.Mobile,
		Nickname:   m.Nickname,
		Avatar:     m.Avatar,
		Name:       m.Name,
		Sex:        m.Sex,
		Birthday:   m.Birthday,
		AreaID:     m.AreaID,
		LevelID:    m.LevelID,
		Point:      m.Point,
		Experience: m.Experience,
		TagIDs:     m.TagIDs,
		TagNames:   []string{},
		CreatedAt:  m.CreatedAt.String(),
	}

	enabled, err := s.levels.Enabled(ctx)
	if err != nil {
		return nil, err
	}
	for _, level := range enabled {
		if level.ID == m.LevelID {
			profile.Level = level.Level
			profile.LevelName = level.Name
		}
	}
	// 下一等级 = 序号大于当前等级里最小的那个；等级表按 sort 排序，不能假设 level 递增
	var next *model.MemberLevelModel
	for i := range enabled {
		if enabled[i].Level <= profile.Level {
			continue
		}
		if next == nil || enabled[i].Level < next.Level {
			next = &enabled[i]
		}
	}
	if next != nil {
		profile.NextLevelExperience = next.Experience
		profile.NextLevelName = next.Name
	} else {
		profile.NextLevelExperience = m.Experience
	}

	profile.TagNames = s.tagNames(ctx, m.TagIDs)

	if m.GroupID > 0 {
		group, err := s.groups.FindByID(ctx, m.GroupID)
		if err != nil {
			return nil, err
		}
		if group != nil {
			profile.GroupName = group.Name
		}
	}

	return profile, nil
}

// Update 修改会员自助可改的字段，只写 nickname / avatar / name / sex / birthday / area_id。
func (s *MemberProfileService) Update(ctx context.Context, memberID types.ID, in MemberProfileUpdateInput) error {
	m, err := s.mustGet(ctx, memberID)
	if err != nil {
		return err
	}

	in.Nickname = strings.TrimSpace(in.Nickname)
	in.Avatar = strings.TrimSpace(in.Avatar)
	in.Name = strings.TrimSpace(in.Name)

	if in.Nickname == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "昵称不能为空")
	}
	if len([]rune(in.Nickname)) > 32 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "昵称不能超过 32 个字符")
	}
	if len(in.Avatar) > 255 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "头像地址过长")
	}
	if in.Sex < 0 || in.Sex > 2 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "性别取值不合法")
	}
	if in.AreaID < 0 {
		in.AreaID = 0
	}
	if in.Birthday != nil && !in.Birthday.IsZero() && in.Birthday.Time().After(types.Now().Time()) {
		return baseErrs.WithMsg(baseErrs.BadRequest, "生日不能晚于今天")
	}

	m.Nickname = in.Nickname
	m.Avatar = in.Avatar
	m.Name = in.Name
	m.Sex = in.Sex
	m.Birthday = in.Birthday
	m.AreaID = in.AreaID

	return s.members.UpdateProfile(ctx, m)
}

// mustGet 取当前会员并校验状态，禁用会员不允许访问任何自助接口。
func (s *MemberProfileService) mustGet(ctx context.Context, memberID types.ID) (*model.MemberUserModel, error) {
	if memberID == 0 {
		return nil, baseErrs.New(memberErrs.MemberNotExists, memberID)
	}

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

// tagNames 把逗号分隔的标签 ID 转成标签名，忽略已删除或不存在的标签。
func (s *MemberProfileService) tagNames(ctx context.Context, tagIDs string) []string {
	names := make([]string, 0)

	ids := make([]types.ID, 0)
	for _, part := range strings.Split(tagIDs, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		ids = append(ids, types.ID(id))
	}
	if len(ids) == 0 {
		return names
	}

	list, err := s.tags.FindByIDs(ctx, ids)
	if err != nil {
		return names
	}
	for i := range list {
		names = append(names, list[i].Name)
	}

	return names
}
