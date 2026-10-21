package service

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
	"eqray-flavor/app/modules/member/repository"
)

// MemberAddressMaxCount 是单个会员可保存的收货地址上限，防止刷接口。
const MemberAddressMaxCount = 20

// MemberAddressInput 是 C 端新增 / 编辑收货地址的入参。
// DefaultStatus 只在新增时使用：1 表示创建后设为默认；没有任何地址时首条自动默认。
type MemberAddressInput struct {
	Name          string
	Mobile        string
	AreaID        int
	DetailAddress string
	DefaultStatus int8
}

// MemberAddressService 负责会员自助维护收货地址。
// 所有方法都必须带 memberID，仓储层查询同样带 user_id，避免越权。
type MemberAddressService struct {
	addresses *repository.MemberAddressRepository
}

// NewMemberAddressService 构造会员地址服务。
func NewMemberAddressService() *MemberAddressService {
	return &MemberAddressService{addresses: repository.NewMemberAddressRepository()}
}

// List 查询当前会员的全部地址，默认地址排在最前。
func (s *MemberAddressService) List(ctx context.Context, memberID types.ID) ([]model.MemberAddressModel, error) {
	return s.addresses.ListByUser(ctx, memberID)
}

// Create 新增地址；首条地址或显式要求时设为默认。
func (s *MemberAddressService) Create(ctx context.Context, memberID types.ID, in MemberAddressInput) (types.ID, error) {
	if err := normalizeAddressInput(&in); err != nil {
		return 0, err
	}

	list, err := s.addresses.ListByUser(ctx, memberID)
	if err != nil {
		return 0, err
	}
	if len(list) >= MemberAddressMaxCount {
		return 0, baseErrs.New(memberErrs.MemberAddressLimitExceeded, MemberAddressMaxCount)
	}

	m := &model.MemberAddressModel{
		UserID:        memberID,
		Name:          in.Name,
		Mobile:        in.Mobile,
		AreaID:        in.AreaID,
		DetailAddress: in.DetailAddress,
		DefaultStatus: 0,
	}
	if err := s.addresses.Create(ctx, m); err != nil {
		return 0, err
	}

	// 首条地址自动默认，显式要求默认时也走一次互斥逻辑
	if in.DefaultStatus == 1 || len(list) == 0 {
		if err := s.SetDefault(ctx, memberID, m.ID); err != nil {
			return 0, err
		}
	}

	return m.ID, nil
}

// Update 编辑地址内容；默认地址状态只能通过 SetDefault 修改，避免绕过互斥逻辑。
func (s *MemberAddressService) Update(ctx context.Context, memberID, id types.ID, in MemberAddressInput) error {
	if err := normalizeAddressInput(&in); err != nil {
		return err
	}

	m, err := s.addresses.FindByID(ctx, id, memberID)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(memberErrs.MemberAddressNotExists, id)
	}

	m.Name = in.Name
	m.Mobile = in.Mobile
	m.AreaID = in.AreaID
	m.DetailAddress = in.DetailAddress

	return s.addresses.Update(ctx, m)
}

// Delete 删除自己名下的地址。
func (s *MemberAddressService) Delete(ctx context.Context, memberID, id types.ID) error {
	m, err := s.addresses.FindByID(ctx, id, memberID)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(memberErrs.MemberAddressNotExists, id)
	}

	return s.addresses.Delete(ctx, id, memberID)
}

// SetDefault 把指定地址设为默认：同一会员的其他地址先取消默认，再设置当前地址。
func (s *MemberAddressService) SetDefault(ctx context.Context, memberID, id types.ID) error {
	m, err := s.addresses.FindByID(ctx, id, memberID)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(memberErrs.MemberAddressNotExists, id)
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.addresses.ClearDefault(ctx, tx, memberID); err != nil {
			return err
		}
		return s.addresses.SetDefault(ctx, tx, id, memberID)
	})
}

// normalizeAddressInput 清洗并校验地址字段。
func normalizeAddressInput(in *MemberAddressInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Mobile = strings.TrimSpace(in.Mobile)
	in.DetailAddress = strings.TrimSpace(in.DetailAddress)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "收件人不能为空")
	}
	if len([]rune(in.Name)) > 30 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "收件人不能超过 30 个字符")
	}
	if !mobilePattern.MatchString(in.Mobile) {
		return baseErrs.New(memberErrs.MemberMobileInvalid)
	}
	if in.DetailAddress == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "详细地址不能为空")
	}
	if len([]rune(in.DetailAddress)) > 255 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "详细地址不能超过 255 个字符")
	}
	if in.AreaID < 0 {
		in.AreaID = 0
	}

	return nil
}
