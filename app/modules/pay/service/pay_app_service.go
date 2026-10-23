package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	// payAppNameMaxLength 应用名长度上限，与 pay_app.name 对齐
	payAppNameMaxLength = 64
	// payAppKeyLength 应用标识长度，保证全局唯一且不便猜测
	payAppKeyLength = 16
	// payAppSecretLength 应用密钥长度（字节），十六进制编码后为 32 位
	payAppSecretLength = 16
)

// PayAppListQuery 是后台应用列表的筛选条件。
type PayAppListQuery struct {
	// Keyword 应用名或 appKey 关键字
	Keyword string
	// Status 开启状态（0 开启 / 1 关闭），nil 表示不限
	Status *int8
}

// PayAppSaveInput 是新增 / 编辑支付应用的入参。
// AppSecret 留空表示由系统生成（新增）或保持原值（编辑）。
type PayAppSaveInput struct {
	Name              string
	Status            payEnums.PayEnableStatusEnum
	Remark            string
	OrderNotifyURL    string
	RefundNotifyURL   string
	TransferNotifyURL string
	AppSecret         string
}

// PayAppRow 是应用列表 / 详情输出，补上状态文案与渠道数量。
// 模型里的 AppSecret 带 json:"-"，明文只在创建响应里通过 AppSecretPlain 返回一次。
type PayAppRow struct {
	model.PayAppModel
	// StatusName 状态文案，如「开启」
	StatusName string `json:"statusName"`
	// ChannelCount 已配置渠道数量
	ChannelCount int `json:"channelCount"`
	// AppSecretPlain 应用密钥明文，仅新增接口返回，其余场景为空
	AppSecretPlain string `json:"appSecret,omitempty"`
}

// PayAppOption 是应用下拉选项。
type PayAppOption struct {
	// ID 应用编号
	ID types.ID `json:"id"`
	// Name 应用名
	Name string `json:"name"`
}

// PayAppService 负责支付应用的增删改查，对齐 yudao PayAppService。
// 应用标识与密钥由系统生成，appKey 创建后不可修改。
type PayAppService struct {
	apps     *repository.PayAppRepository
	channels *repository.PayChannelRepository
}

// NewPayAppService 构造支付应用服务。
func NewPayAppService() *PayAppService {
	return &PayAppService{
		apps:     repository.NewPayAppRepository(),
		channels: repository.NewPayChannelRepository(),
	}
}

// Page 分页查询支付应用，支持关键字与开启状态筛选。
// 入参：筛选条件与分页；返回应用列表（含渠道数量）与总数。
func (s *PayAppService) Page(ctx context.Context, q PayAppListQuery, pageNo, pageSize int) ([]PayAppRow, int64, error) {
	rows, total, err := s.apps.Paginate(ctx, repository.PayAppFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Status:  q.Status,
	}, pageNo, pageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]PayAppRow, 0, len(rows))
	for _, row := range rows {
		list = append(list, s.decorate(ctx, row))
	}

	return list, total, nil
}

// Options 返回全部开启状态的应用，供渠道配置与后台下拉使用。
// 入参：无；返回 id / name 列表。
func (s *PayAppService) Options(ctx context.Context) ([]PayAppOption, error) {
	rows, err := s.apps.AllEnabled(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]PayAppOption, 0, len(rows))
	for _, row := range rows {
		list = append(list, PayAppOption{ID: row.ID, Name: row.Name})
	}

	return list, nil
}

// Detail 查询应用详情。
// 入参：应用主键；返回应用详情，不存在时报「支付应用不存在」。
func (s *PayAppService) Detail(ctx context.Context, id types.ID) (*PayAppRow, error) {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}

	row := s.decorate(ctx, *m)

	return &row, nil
}

// Create 新增支付应用：校验名称与回调地址，自动生成 appKey 与 appSecret。
// 入参：应用字段；返回新应用主键，appSecret 明文只在创建响应里出现一次。
func (s *PayAppService) Create(ctx context.Context, in PayAppSaveInput) (*PayAppRow, error) {
	if err := normalizePayAppInput(&in); err != nil {
		return nil, err
	}

	appKey, err := randomToken(payAppKeyLength)
	if err != nil {
		return nil, baseErrs.WithMsg(baseErrs.InternalServerError, "应用标识生成失败")
	}
	appSecret := strings.TrimSpace(in.AppSecret)
	if appSecret == "" {
		appSecret, err = randomToken(payAppSecretLength)
		if err != nil {
			return nil, baseErrs.WithMsg(baseErrs.InternalServerError, "应用密钥生成失败")
		}
	}

	m := &model.PayAppModel{
		Name:              in.Name,
		Status:            in.Status,
		Remark:            in.Remark,
		OrderNotifyURL:    in.OrderNotifyURL,
		RefundNotifyURL:   in.RefundNotifyURL,
		TransferNotifyURL: in.TransferNotifyURL,
		AppKey:            appKey,
		AppSecret:         appSecret,
	}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.apps.Create(ctx, tx, m)
	}); err != nil {
		return nil, err
	}

	// 创建后把明文密钥回给调用方，后续接口只返回脱敏值
	row := s.decorate(ctx, *m)
	row.AppSecretPlain = appSecret

	return &row, nil
}

// Update 编辑支付应用；appKey 不可修改，appSecret 留空保持原值。
// 入参：应用主键与新字段；返回错误表示应用不存在或校验失败。
func (s *PayAppService) Update(ctx context.Context, id types.ID, in PayAppSaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if err := normalizePayAppInput(&in); err != nil {
		return err
	}

	if secret := strings.TrimSpace(in.AppSecret); secret != "" {
		m.AppSecret = secret
	}
	m.Name = in.Name
	m.Status = in.Status
	m.Remark = in.Remark
	m.OrderNotifyURL = in.OrderNotifyURL
	m.RefundNotifyURL = in.RefundNotifyURL
	m.TransferNotifyURL = in.TransferNotifyURL

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.apps.Update(ctx, tx, m)
	})
}

// Delete 软删除支付应用；应用下仍有渠道时禁止删除，避免渠道成为孤儿配置。
// 入参：应用主键；返回错误表示存在渠道或删除失败。
func (s *PayAppService) Delete(ctx context.Context, id types.ID) error {
	if _, err := s.findByID(ctx, id); err != nil {
		return err
	}

	channels, err := s.channels.AllByApp(ctx, id)
	if err != nil {
		return err
	}
	if len(channels) > 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "应用下还有渠道配置，请先删除渠道")
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.apps.Delete(ctx, tx, id)
	})
}

// findByID 按主键取应用，未命中转换成固定业务错误。
func (s *PayAppService) findByID(ctx context.Context, id types.ID) (*model.PayAppModel, error) {
	m, err := s.apps.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayAppNotExists)
	}

	return m, nil
}

// decorate 为应用补齐状态文案与渠道数量，并隐藏密钥明文。
func (s *PayAppService) decorate(ctx context.Context, m model.PayAppModel) PayAppRow {
	count := 0
	if channels, err := s.channels.AllByApp(ctx, m.ID); err == nil {
		count = len(channels)
	}

	m.AppSecret = ""

	return PayAppRow{PayAppModel: m, StatusName: m.Status.String(), ChannelCount: count}
}

// normalizePayAppInput 清洗应用字段并校验名称、状态与回调地址。
// 入参：待落库的应用字段指针，会被原地清洗；返回业务错误表示校验不通过。
func normalizePayAppInput(in *PayAppSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)
	in.OrderNotifyURL = strings.TrimSpace(in.OrderNotifyURL)
	in.RefundNotifyURL = strings.TrimSpace(in.RefundNotifyURL)
	in.TransferNotifyURL = strings.TrimSpace(in.TransferNotifyURL)
	in.AppSecret = strings.TrimSpace(in.AppSecret)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "应用名不能为空")
	}
	if len([]rune(in.Name)) > payAppNameMaxLength {
		return baseErrs.WithMsg(baseErrs.BadRequest, "应用名不能超过 64 个字符")
	}
	if !in.Status.IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "应用状态只能是 0（开启）或 1（关闭）")
	}
	// 支付与退款回调是业务系统感知支付结果的唯一途径，必须配置
	if in.OrderNotifyURL == "" {
		return baseErrs.New(payErrs.PayAppNotifyURLRequired, in.Name)
	}
	if in.RefundNotifyURL == "" {
		return baseErrs.New(payErrs.PayAppNotifyURLRequired, in.Name)
	}

	return nil
}

// randomToken 生成指定字节长度的随机十六进制串，用于应用标识与密钥。
// 入参：字节长度；返回十六进制字符串（长度为字节数的两倍）。
func randomToken(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
