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

// InfraSmsTemplateSaveInput 是新增 / 编辑短信模板的入参。
type InfraSmsTemplateSaveInput struct {
	ChannelID     types.ID
	Code          string
	Name          string
	Type          int8
	APITemplateID string
	Content       string
	Params        string
	Status        int8
	Remark        string
}

// InfraSmsTemplateListQuery 是模板列表的查询条件。
type InfraSmsTemplateListQuery struct {
	Keyword   string
	ChannelID types.ID
	Type      *int8
	Status    *baseEnums.CommonStatusEnum
	PageNo    int
	PageSize  int
}

// InfraSmsTemplateService 负责短信模板配置。
type InfraSmsTemplateService struct {
	templates *repository.InfraSmsTemplateRepository
	channels  *repository.InfraSmsChannelRepository
}

func NewInfraSmsTemplateService() *InfraSmsTemplateService {
	return &InfraSmsTemplateService{
		templates: repository.NewInfraSmsTemplateRepository(),
		channels:  repository.NewInfraSmsChannelRepository(),
	}
}

// Page 分页查询模板。
func (s *InfraSmsTemplateService) Page(ctx context.Context, q InfraSmsTemplateListQuery) ([]infraModel.InfraSmsTemplateModel, int64, error) {
	return s.templates.Paginate(ctx, repository.InfraSmsTemplateFilter{
		Keyword:   strings.TrimSpace(q.Keyword),
		ChannelID: q.ChannelID,
		Type:      q.Type,
		Status:    q.Status,
	}, q.PageNo, q.PageSize)
}

// Detail 查询模板详情。
func (s *InfraSmsTemplateService) Detail(ctx context.Context, id types.ID) (*infraModel.InfraSmsTemplateModel, error) {
	m, err := s.templates.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(infraErrs.InfraSmsTemplateNotExists, id)
	}

	return m, nil
}

// Create 新增模板：代码唯一，渠道必须存在。
func (s *InfraSmsTemplateService) Create(ctx context.Context, in InfraSmsTemplateSaveInput) (types.ID, error) {
	if err := normalizeTemplateInput(&in); err != nil {
		return 0, err
	}
	if err := s.ensureChannel(ctx, in.ChannelID); err != nil {
		return 0, err
	}

	exists, err := s.templates.ExistsByCode(ctx, in.Code, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(infraErrs.InfraSmsTemplateCodeExists, in.Code)
	}

	m := &infraModel.InfraSmsTemplateModel{
		ChannelID:     in.ChannelID,
		Code:          in.Code,
		Name:          in.Name,
		Type:          in.Type,
		APITemplateID: in.APITemplateID,
		Content:       in.Content,
		Params:        in.Params,
		Status:        baseEnums.CommonStatusEnum(in.Status),
		Remark:        in.Remark,
	}
	if err := s.templates.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑模板；渠道可重新绑定。
func (s *InfraSmsTemplateService) Update(ctx context.Context, id types.ID, in InfraSmsTemplateSaveInput) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	if err := normalizeTemplateInput(&in); err != nil {
		return err
	}
	if err := s.ensureChannel(ctx, in.ChannelID); err != nil {
		return err
	}

	m.ChannelID = in.ChannelID
	m.Name = in.Name
	m.Type = in.Type
	m.APITemplateID = in.APITemplateID
	m.Content = in.Content
	m.Params = in.Params
	m.Status = baseEnums.CommonStatusEnum(in.Status)
	m.Remark = in.Remark

	return s.templates.Update(ctx, m)
}

// Delete 删除模板。
func (s *InfraSmsTemplateService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	return s.templates.Delete(ctx, m)
}

// ensureChannel 校验模板绑定的渠道存在。
func (s *InfraSmsTemplateService) ensureChannel(ctx context.Context, channelID types.ID) error {
	if channelID == 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "请选择短信渠道")
	}

	channel, err := s.channels.FindByID(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return baseErrs.New(infraErrs.InfraSmsChannelNotExists, channelID)
	}

	return nil
}

// normalizeTemplateInput 清洗入参并校验取值范围。
func normalizeTemplateInput(in *InfraSmsTemplateSaveInput) error {
	in.Code = strings.TrimSpace(in.Code)
	in.Name = strings.TrimSpace(in.Name)
	in.APITemplateID = strings.TrimSpace(in.APITemplateID)
	in.Content = strings.TrimSpace(in.Content)
	in.Params = strings.TrimSpace(in.Params)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.Code == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "模板代码不能为空")
	}
	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "模板名称不能为空")
	}
	if in.Type != infraModel.InfraSmsTemplateTypeCode && in.Type != infraModel.InfraSmsTemplateTypeNotify {
		return baseErrs.WithMsg(baseErrs.BadRequest, "模板类型取值不合法")
	}
	if in.Content == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "模板内容不能为空")
	}
	if !baseEnums.CommonStatusEnum(in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "状态取值不合法")
	}

	return nil
}
