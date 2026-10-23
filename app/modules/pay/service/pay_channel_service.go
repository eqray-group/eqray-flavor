package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payErrs "eqray-flavor/app/modules/pay/errs"
	"eqray-flavor/app/modules/pay/model"
	"eqray-flavor/app/modules/pay/repository"
)

const (
	// payChannelConfigMaxLength 渠道配置 JSON 长度上限，与 pay_channel.config 列对齐
	payChannelConfigMaxLength = 8192
	// payChannelFeeRateMax 渠道费率上限（百分比），防止误填导致手续费算错
	payChannelFeeRateMax = 100
)

// paySensitiveConfigKeys 是渠道配置里需要脱敏返回的字段名，
// 同时覆盖微信（apiV3Key / privateKey）与支付宝（privateKey / publicKey）的常见写法。
var paySensitiveConfigKeys = map[string]bool{
	"apiv3key":        true,
	"api_v3_key":      true,
	"apikey":          true,
	"api_key":         true,
	"privatekey":      true,
	"private_key":     true,
	"publickey":       true,
	"public_key":      true,
	"appsecret":       true,
	"app_secret":      true,
	"key":             true,
	"password":        true,
	"cert":            true,
	"certcontent":     true,
	"cert_content":    true,
	"privatekeypem":   true,
	"private_key_pem": true,
}

// PayChannelListQuery 是后台渠道列表的筛选条件。
type PayChannelListQuery struct {
	// AppID 应用编号，0 表示不限
	AppID types.ID
	// Code 渠道编码
	Code string
	// Status 开启状态，nil 表示不限
	Status *int8
}

// PayChannelSaveInput 是新增 / 编辑渠道的入参。
// Config 为渠道参数 JSON，敏感字段提交掩码值表示沿用原值。
type PayChannelSaveInput struct {
	AppID   types.ID
	Code    string
	Status  payEnums.PayEnableStatusEnum
	Remark  string
	FeeRate float64
	Config  string
}

// PayChannelRow 是渠道列表 / 详情输出，config 已做敏感字段脱敏。
type PayChannelRow struct {
	model.PayChannelModel
	// ChannelName 渠道中文名，如「微信小程序」
	ChannelName string `json:"channelName"`
	// StatusName 状态文案，如「开启」
	StatusName string `json:"statusName"`
}

// PayChannelService 负责支付渠道的增删改查，对齐 yudao PayChannelService。
// 渠道配置以 JSON 存储，返回时对密钥类字段脱敏，编辑时可回传掩码表示沿用原值。
type PayChannelService struct {
	channels *repository.PayChannelRepository
	apps     *repository.PayAppRepository
}

// NewPayChannelService 构造支付渠道服务。
func NewPayChannelService() *PayChannelService {
	return &PayChannelService{
		channels: repository.NewPayChannelRepository(),
		apps:     repository.NewPayAppRepository(),
	}
}

// Page 分页查询渠道，支持应用、编码与状态筛选。
// 入参：筛选条件与分页；返回脱敏后的渠道列表与总数。
func (s *PayChannelService) Page(ctx context.Context, q PayChannelListQuery, pageNo, pageSize int) ([]PayChannelRow, int64, error) {
	rows, total, err := s.channels.Paginate(ctx, repository.PayChannelFilter{
		AppID:  q.AppID,
		Code:   strings.TrimSpace(q.Code),
		Status: q.Status,
	}, pageNo, pageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]PayChannelRow, 0, len(rows))
	for _, row := range rows {
		list = append(list, decoratePayChannel(row))
	}

	return list, total, nil
}

// ListByApp 查询指定应用下的全部渠道，供后台配置页展示。
// 入参：应用编号；返回脱敏后的渠道列表。
func (s *PayChannelService) ListByApp(ctx context.Context, appID types.ID) ([]PayChannelRow, error) {
	rows, err := s.channels.AllByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	list := make([]PayChannelRow, 0, len(rows))
	for _, row := range rows {
		list = append(list, decoratePayChannel(row))
	}

	return list, nil
}

// Detail 查询渠道详情，config 中的密钥只返回掩码。
// 入参：渠道主键；返回渠道详情，不存在时报「支付渠道不存在」。
func (s *PayChannelService) Detail(ctx context.Context, id types.ID) (*PayChannelRow, error) {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}

	row := decoratePayChannel(*m)

	return &row, nil
}

// Create 新增渠道：校验应用存在、编码合法且同应用下不重复，config 必须是 JSON 对象。
// 入参：渠道字段；返回新渠道主键。
func (s *PayChannelService) Create(ctx context.Context, in PayChannelSaveInput) (types.ID, error) {
	if err := s.normalizeInput(ctx, &in); err != nil {
		return 0, err
	}

	exists, err := s.channels.ExistsByAppAndCode(ctx, in.AppID, in.Code, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(payErrs.PayChannelCodeDuplicate, payEnums.ChannelLabel(in.Code))
	}

	m := &model.PayChannelModel{
		AppID:   in.AppID,
		Code:    in.Code,
		Status:  in.Status,
		Remark:  in.Remark,
		FeeRate: in.FeeRate,
		Config:  in.Config,
	}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.channels.Create(ctx, tx, m)
	}); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑渠道：应用与编码不可变更，config 中掩码字段沿用原值。
// 入参：渠道主键与新字段；返回错误表示渠道不存在或校验失败。
func (s *PayChannelService) Update(ctx context.Context, id types.ID, in PayChannelSaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}

	// 应用与渠道编码决定对接哪个收款账户，创建后不允许变更
	in.AppID = m.AppID
	in.Code = m.Code
	if err := s.normalizeInput(ctx, &in); err != nil {
		return err
	}

	m.Status = in.Status
	m.Remark = in.Remark
	m.FeeRate = in.FeeRate
	m.Config = mergeSensitiveConfig(m.Config, in.Config)

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.channels.Update(ctx, tx, m)
	})
}

// Delete 软删除渠道，历史支付单仍保留渠道快照。
// 入参：渠道主键；返回错误表示渠道不存在或删除失败。
func (s *PayChannelService) Delete(ctx context.Context, id types.ID) error {
	if _, err := s.findByID(ctx, id); err != nil {
		return err
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.channels.Delete(ctx, tx, id)
	})
}

// ResolveEnabled 取下单用的渠道与应用，并校验两者均处于开启状态。
// 入参：应用编号、渠道编码；返回渠道与应用，任一未开启或不存在时返回业务错误。
// 说明：钱包余额是内置渠道，没有 pay_channel 配置行，直接返回内存态渠道对象。
func (s *PayChannelService) ResolveEnabled(ctx context.Context, appID types.ID, code string) (*model.PayChannelModel, *model.PayAppModel, error) {
	if !payEnums.IsValidChannel(code) {
		return nil, nil, baseErrs.New(payErrs.PayChannelCodeInvalid, code)
	}

	app, err := s.apps.FindByID(ctx, appID)
	if err != nil {
		return nil, nil, err
	}
	if app == nil {
		return nil, nil, baseErrs.New(payErrs.PayAppNotExists)
	}
	if app.Status != payEnums.PayEnableStatusEnable {
		return nil, nil, baseErrs.New(payErrs.PayAppIsDisable, app.Name)
	}

	// 钱包余额不走渠道配置与网关，由支付单服务直接转交钱包服务处理
	if code == payEnums.ChannelWallet {
		return &model.PayChannelModel{
			AppID:   app.ID,
			Code:    payEnums.ChannelWallet,
			Status:  payEnums.PayEnableStatusEnable,
			FeeRate: 0,
			Remark:  "系统内置渠道：钱包余额",
		}, app, nil
	}

	channel, err := s.channels.FindByAppAndCode(ctx, appID, code)
	if err != nil {
		return nil, nil, err
	}
	if channel == nil {
		return nil, nil, baseErrs.New(payErrs.PayChannelNotExists)
	}
	if channel.Status != payEnums.PayEnableStatusEnable {
		return nil, nil, baseErrs.New(payErrs.PayChannelIsDisable, payEnums.ChannelLabel(code))
	}

	return channel, app, nil
}

// ByID 按主键取渠道且不校验开关状态，供补单、关单、退款等需要对历史支付单继续操作的场景使用。
// 入参：渠道主键；返回渠道，不存在时返回业务错误。
func (s *PayChannelService) ByID(ctx context.Context, id types.ID) (*model.PayChannelModel, error) {
	return s.findByID(ctx, id)
}

// EnabledByID 按主键取渠道并校验其仍处于开启状态，供已建支付单重新调起时使用。
// 入参：渠道主键；返回渠道，不存在或已关闭时返回业务错误。
func (s *PayChannelService) EnabledByID(ctx context.Context, id types.ID) (*model.PayChannelModel, error) {
	channel, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if channel.Status != payEnums.PayEnableStatusEnable {
		return nil, baseErrs.New(payErrs.PayChannelIsDisable, payEnums.ChannelLabel(channel.Code))
	}

	return channel, nil
}

// findByID 按主键取渠道，未命中转换成固定业务错误。
func (s *PayChannelService) findByID(ctx context.Context, id types.ID) (*model.PayChannelModel, error) {
	m, err := s.channels.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayChannelNotExists)
	}

	return m, nil
}

// normalizeInput 清洗渠道字段并校验应用、编码、费率与配置 JSON。
// 入参：待落库的渠道字段指针，会被原地清洗；返回业务错误表示校验不通过。
func (s *PayChannelService) normalizeInput(ctx context.Context, in *PayChannelSaveInput) error {
	in.Code = strings.TrimSpace(in.Code)
	in.Remark = strings.TrimSpace(in.Remark)
	in.Config = strings.TrimSpace(in.Config)

	if !payEnums.IsValidChannel(in.Code) {
		return baseErrs.New(payErrs.PayChannelCodeInvalid, in.Code)
	}
	// 钱包余额由代码内置处理，不允许在 pay_channel 里配置
	if in.Code == payEnums.ChannelWallet {
		return baseErrs.WithMsg(baseErrs.BadRequest, "钱包余额是系统内置渠道，无需在渠道配置里新增")
	}
	if !in.Status.IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "渠道状态只能是 0（开启）或 1（关闭）")
	}
	if in.FeeRate < 0 || in.FeeRate > payChannelFeeRateMax {
		return baseErrs.WithMsg(baseErrs.BadRequest, "渠道费率必须在 0 ~ 100 之间")
	}

	app, err := s.apps.FindByID(ctx, in.AppID)
	if err != nil {
		return err
	}
	if app == nil {
		return baseErrs.New(payErrs.PayAppNotExists)
	}
	if app.Status != payEnums.PayEnableStatusEnable {
		return baseErrs.New(payErrs.PayAppIsDisable, app.Name)
	}

	if in.Config == "" {
		return baseErrs.New(payErrs.PayChannelConfigMissing, payEnums.ChannelLabel(in.Code), "配置不能为空")
	}
	if len(in.Config) > payChannelConfigMaxLength {
		return baseErrs.WithMsg(baseErrs.BadRequest, "渠道配置不能超过 8192 个字符")
	}
	if !isJSONObject(in.Config) {
		return baseErrs.New(payErrs.PayChannelConfigInvalid)
	}

	return nil
}

// decoratePayChannel 为渠道补齐中文名与状态文案，并把 config 里的密钥脱敏。
func decoratePayChannel(m model.PayChannelModel) PayChannelRow {
	m.Config = maskConfigJSON(m.Config)

	return PayChannelRow{
		PayChannelModel: m,
		ChannelName:     payEnums.ChannelLabel(m.Code),
		StatusName:      m.Status.String(),
	}
}

// isJSONObject 判断字符串是否为 JSON 对象，渠道配置必须是键值对。
// 入参：待校验文本；返回是否为合法 JSON 对象。
func isJSONObject(raw string) bool {
	data := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return false
	}

	return true
}

// maskConfigJSON 把渠道配置里的密钥字段脱敏为 ****+末四位，其余字段原样返回。
// 入参：原始配置 JSON；返回脱敏后的 JSON，解析失败时原样返回。
func maskConfigJSON(raw string) string {
	if raw == "" {
		return raw
	}

	data := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return raw
	}

	changed := false
	for key, value := range data {
		if !paySensitiveConfigKeys[strings.ToLower(key)] {
			continue
		}
		text, ok := value.(string)
		if !ok || text == "" {
			continue
		}
		data[key] = maskSecret(text)
		changed = true
	}
	if !changed {
		return raw
	}

	masked, err := json.Marshal(data)
	if err != nil {
		return raw
	}

	return string(masked)
}

// mergeSensitiveConfig 在编辑渠道时合并配置：新配置中的掩码值沿用旧配置里的真实密钥。
// 入参：旧配置 JSON、新配置 JSON；返回合并后的 JSON。
func mergeSensitiveConfig(oldConfig, newConfig string) string {
	if newConfig == "" {
		return oldConfig
	}

	newData := make(map[string]any)
	if err := json.Unmarshal([]byte(newConfig), &newData); err != nil {
		return newConfig
	}

	oldData := make(map[string]any)
	if err := json.Unmarshal([]byte(oldConfig), &oldData); err != nil {
		return newConfig
	}

	for key, value := range newData {
		text, ok := value.(string)
		if !ok || !isMaskedValue(text) {
			continue
		}
		if oldValue, exists := oldData[key]; exists {
			newData[key] = oldValue
		}
	}

	merged, err := json.Marshal(newData)
	if err != nil {
		return newConfig
	}

	return string(merged)
}
