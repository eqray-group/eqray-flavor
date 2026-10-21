package service

import (
	"context"
	"strings"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	infraModel "eqray-flavor/app/modules/infra/model"
	"eqray-flavor/app/modules/infra/repository"
)

// InfraSmsChannelSaveInput 是新增 / 编辑短信渠道的入参。
type InfraSmsChannelSaveInput struct {
	Name      string
	Code      string
	APIKey    string
	APISecret string
	Signature string
	Extra     string
	Status    int8
	Remark    string
}

// InfraSmsChannelListQuery 是渠道列表的查询条件。
type InfraSmsChannelListQuery struct {
	Keyword  string
	Code     string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// InfraSmsChannelService 负责短信渠道配置。
type InfraSmsChannelService struct {
	channels *repository.InfraSmsChannelRepository
}

func NewInfraSmsChannelService() *InfraSmsChannelService {
	return &InfraSmsChannelService{channels: repository.NewInfraSmsChannelRepository()}
}

// Page 分页查询渠道配置。
func (s *InfraSmsChannelService) Page(ctx context.Context, q InfraSmsChannelListQuery) ([]infraModel.InfraSmsChannelModel, int64, error) {
	return s.channels.Paginate(ctx, repository.InfraSmsChannelFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Code:    strings.TrimSpace(q.Code),
		Status:  q.Status,
	}, q.PageNo, q.PageSize)
}

// Options 返回启用中的渠道，供短信模板表单下拉。
func (s *InfraSmsChannelService) Options(ctx context.Context) ([]infraModel.InfraSmsChannelModel, error) {
	return s.channels.Enabled(ctx)
}

// Detail 查询渠道详情。
func (s *InfraSmsChannelService) Detail(ctx context.Context, id types.ID) (*infraModel.InfraSmsChannelModel, error) {
	m, err := s.channels.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(infraErrs.InfraSmsChannelNotExists, id)
	}

	return m, nil
}

// Create 新增渠道：代码唯一，且必须是受支持的渠道。
func (s *InfraSmsChannelService) Create(ctx context.Context, in InfraSmsChannelSaveInput) (types.ID, error) {
	if err := normalizeChannelInput(&in, true); err != nil {
		return 0, err
	}

	exists, err := s.channels.ExistsByCode(ctx, in.Code, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(infraErrs.InfraSmsChannelCodeExists, in.Code)
	}

	m := &infraModel.InfraSmsChannelModel{
		Name:      in.Name,
		Code:      in.Code,
		APIKey:    in.APIKey,
		APISecret: in.APISecret,
		Signature: in.Signature,
		Extra:     in.Extra,
		Status:    baseEnums.CommonStatusEnum(in.Status),
		Remark:    in.Remark,
	}
	if err := s.channels.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑渠道；api_secret 留空表示不修改。
func (s *InfraSmsChannelService) Update(ctx context.Context, id types.ID, in InfraSmsChannelSaveInput) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	if err := normalizeChannelInput(&in, false); err != nil {
		return err
	}

	m.Name = in.Name
	m.APIKey = in.APIKey
	m.Signature = in.Signature
	m.Extra = in.Extra
	m.Status = baseEnums.CommonStatusEnum(in.Status)
	m.Remark = in.Remark
	if in.APISecret != "" {
		m.APISecret = in.APISecret
	}

	return s.channels.Update(ctx, m)
}

// Delete 删除渠道；渠道下仍有模板时拒绝。
func (s *InfraSmsChannelService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	count, err := s.channels.CountTemplatesByChannelID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(infraErrs.InfraSmsChannelInUse, m.Name)
	}

	return s.channels.Delete(ctx, m)
}

// normalizeChannelInput 清洗入参并校验取值范围。
func normalizeChannelInput(in *InfraSmsChannelSaveInput, needSecret bool) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Code = strings.TrimSpace(in.Code)
	in.APIKey = strings.TrimSpace(in.APIKey)
	in.APISecret = strings.TrimSpace(in.APISecret)
	in.Signature = strings.TrimSpace(in.Signature)
	in.Extra = strings.TrimSpace(in.Extra)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "渠道名称不能为空")
	}
	if !IsSupportedSmsChannel(in.Code) {
		return baseErrs.WithMsg(baseErrs.BadRequest, "不支持的短信渠道")
	}
	if in.Code != infraModel.InfraSmsChannelDebugLog {
		if in.APIKey == "" {
			return baseErrs.WithMsg(baseErrs.BadRequest, "APIKey 不能为空")
		}
		if needSecret && in.APISecret == "" {
			return baseErrs.WithMsg(baseErrs.BadRequest, "APISecret 不能为空")
		}
		if in.Signature == "" {
			return baseErrs.WithMsg(baseErrs.BadRequest, "短信签名不能为空")
		}
	}
	if !baseEnums.CommonStatusEnum(in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "状态取值不合法")
	}

	return nil
}

// IsSupportedSmsChannel 判断渠道代码是否受支持。
func IsSupportedSmsChannel(code string) bool {
	switch code {
	case infraModel.InfraSmsChannelDebugLog, infraModel.InfraSmsChannelAliyun, infraModel.InfraSmsChannelTencent:
		return true
	default:
		return false
	}
}
