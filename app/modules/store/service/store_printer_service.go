package service

import (
	"context"
	"encoding/json"
	"strings"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
	"eqray-flavor/app/modules/store/repository"
)

const (
	// printerSecretMaxLength 是设备密钥的明文字节上限，避免加密后的密文列超长。
	printerSecretMaxLength = 512
	// printerConfigMaxLength 是非敏感配置 JSON 的长度上限。
	printerConfigMaxLength = 2000
	// printerMaxCopies 是默认打印份数的硬上限，避免误配置造成纸张浪费。
	printerMaxCopies = 9
)

// PrinterListQuery 是打印设备列表的查询条件。
type PrinterListQuery struct {
	Keyword    string
	Type       *enums.PrinterTypeEnum
	AccessMode *enums.PrinterAccessModeEnum
	Status     *baseEnums.CommonStatusEnum
	PageNo     int
	PageSize   int
}

// PrinterSaveInput 是新增 / 编辑打印设备的入参。
// Secret 只在新增或修改密钥时传明文；编辑时留空表示沿用原密钥。
type PrinterSaveInput struct {
	StoreID    types.ID
	Name       string
	DeviceNo   string
	Type       enums.PrinterTypeEnum
	AccessMode enums.PrinterAccessModeEnum
	Config     string
	Secret     string
	Copies     int
	Status     baseEnums.CommonStatusEnum
	Remark     string
}

// PrinterListRow 是打印设备列表输出：设备记录 + 门店与展示字段。
// 模型中的 Secret 带 json:"-"，列表只返回 SecretMask 与 HasSecret。
type PrinterListRow struct {
	model.StorePrinterModel
	StoreName      string `json:"storeName"`
	TypeName       string `json:"typeName"`
	AccessModeName string `json:"accessModeName"`
	StatusName     string `json:"statusName"`
	HasSecret      bool   `json:"hasSecret"`
}

// StorePrinterService 负责打印设备的增删改查、门店隔离与密钥保护。
type StorePrinterService struct {
	printers *repository.StorePrinterRepository
	stores   *StoreService
}

// NewStorePrinterService 构造打印设备服务。
func NewStorePrinterService() *StorePrinterService {
	return &StorePrinterService{
		printers: repository.NewStorePrinterRepository(),
		stores:   NewStoreService(),
	}
}

// Page 分页查询当前用户可见范围内的打印设备。
// 请求带门店上下文时只查该门店；否则按可见门店集合过滤。
func (s *StorePrinterService) Page(ctx context.Context, q PrinterListQuery) ([]PrinterListRow, int64, error) {
	filter, err := s.buildFilter(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	rows, total, err := s.printers.Paginate(ctx, filter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list, err := s.decorateRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Create 新增打印设备：校验门店可见、设备号唯一并加密密钥后落库。
// 入参：未传门店时回填请求头门店；返回新设备 ID。明文密钥不落库、不返回。
func (s *StorePrinterService) Create(ctx context.Context, in PrinterSaveInput) (types.ID, error) {
	store, err := s.stores.ResolveWritableStore(ctx, in.StoreID)
	if err != nil {
		return 0, err
	}
	in.StoreID = store.ID
	if err := normalizePrinterInput(&in); err != nil {
		return 0, err
	}

	exists, err := s.printers.ExistsByDeviceNo(ctx, in.StoreID, in.DeviceNo, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(storeErrs.StorePrinterDeviceExists, in.DeviceNo)
	}

	secret, mask, err := s.protectSecret(in.Secret)
	if err != nil {
		return 0, err
	}

	m := &model.StorePrinterModel{
		StoreID:    in.StoreID,
		Name:       in.Name,
		DeviceNo:   in.DeviceNo,
		Type:       in.Type,
		AccessMode: in.AccessMode,
		Config:     in.Config,
		Secret:     secret,
		SecretMask: mask,
		Copies:     in.Copies,
		Status:     in.Status,
		Remark:     in.Remark,
	}
	if err := s.printers.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑打印设备，只允许调整设备配置，设备不能跨门店移动。
// 入参 secret 为空表示沿用原密钥；传新明文时重新加密并更新脱敏值。
func (s *StorePrinterService) Update(ctx context.Context, id types.ID, in PrinterSaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if in.StoreID > 0 && in.StoreID != m.StoreID {
		return baseErrs.New(storeErrs.StorePrinterStoreImmutable)
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}
	in.StoreID = m.StoreID
	if err := normalizePrinterInput(&in); err != nil {
		return err
	}

	exists, err := s.printers.ExistsByDeviceNo(ctx, m.StoreID, in.DeviceNo, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(storeErrs.StorePrinterDeviceExists, in.DeviceNo)
	}

	if in.Secret != "" {
		secret, mask, err := s.protectSecret(in.Secret)
		if err != nil {
			return err
		}
		m.Secret = secret
		m.SecretMask = mask
	}
	m.Name = in.Name
	m.DeviceNo = in.DeviceNo
	m.Type = in.Type
	m.AccessMode = in.AccessMode
	m.Config = in.Config
	m.Copies = in.Copies
	m.Status = in.Status
	m.Remark = in.Remark

	return s.printers.Update(ctx, m)
}

// Delete 物理删除打印设备；设备号释放后可重新绑定到其他设备记录。
func (s *StorePrinterService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}

	return s.printers.Delete(ctx, m)
}

// findByID 查询打印设备，不存在时返回固定错误码。
func (s *StorePrinterService) findByID(ctx context.Context, id types.ID) (*model.StorePrinterModel, error) {
	m, err := s.printers.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(storeErrs.StorePrinterNotExists)
	}

	return m, nil
}

// buildFilter 根据请求上下文与当前用户可见范围构造打印设备筛选条件。
func (s *StorePrinterService) buildFilter(ctx context.Context, q PrinterListQuery) (repository.StorePrinterFilter, error) {
	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return repository.StorePrinterFilter{}, err
	}

	filter := repository.StorePrinterFilter{
		Keyword:    strings.TrimSpace(q.Keyword),
		Type:       q.Type,
		AccessMode: q.AccessMode,
		Status:     q.Status,
	}
	if contextStoreID := support.StoreID(ctx); contextStoreID > 0 {
		if !scope.All && !scope.Contains(contextStoreID) {
			return filter, baseErrs.New(storeErrs.StoreAccessDenied)
		}
		filter.StoreID = contextStoreID
	} else if !scope.All {
		filter.RestrictStores = true
		filter.StoreIDs = scope.IDs
	}

	return filter, nil
}

// decorateRows 为打印设备列表补齐门店名称、枚举文案与密钥配置状态。
func (s *StorePrinterService) decorateRows(ctx context.Context, rows []model.StorePrinterModel) ([]PrinterListRow, error) {
	list := make([]PrinterListRow, 0, len(rows))
	if len(rows) == 0 {
		return list, nil
	}

	storeNames, err := s.storeNames(ctx, rows)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		list = append(list, PrinterListRow{
			StorePrinterModel: row,
			StoreName:         storeNames[row.StoreID],
			TypeName:          row.Type.String(),
			AccessModeName:    row.AccessMode.String(),
			StatusName:        row.Status.String(),
			HasSecret:         row.Secret != "",
		})
	}

	return list, nil
}

// storeNames 批量查询设备所属门店名称，返回门店 ID 到名称的映射。
func (s *StorePrinterService) storeNames(ctx context.Context, rows []model.StorePrinterModel) (map[types.ID]string, error) {
	names := make(map[types.ID]string, len(rows))
	if len(rows) == 0 {
		return names, nil
	}

	ids := make([]types.ID, 0, len(rows))
	seen := make(map[types.ID]bool, len(rows))
	for _, row := range rows {
		if seen[row.StoreID] {
			continue
		}
		seen[row.StoreID] = true
		ids = append(ids, row.StoreID)
	}
	stores, err := s.stores.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, store := range stores {
		names[store.ID] = store.Name
	}

	return names, nil
}

// protectSecret 把明文密钥加密为 AES 密文，并生成只保留末四位的脱敏值。
// 入参：secret 明文，空字符串表示不配置密钥；返回密文、脱敏值与错误。
func (s *StorePrinterService) protectSecret(secret string) (string, string, error) {
	if secret == "" {
		return "", "", nil
	}

	ciphertext, err := facades.Crypt().EncryptString(secret)
	if err != nil {
		return "", "", baseErrs.WithMsg(baseErrs.InternalServerError, "打印设备密钥加密失败")
	}

	return ciphertext, maskPrinterSecret(secret), nil
}

// maskPrinterSecret 生成密钥脱敏值：空值返回空串，长度不超过 4 返回 ****，
// 否则隐藏前段只保留末四位，如 ****abcd。
func maskPrinterSecret(secret string) string {
	if secret == "" {
		return ""
	}
	runes := []rune(secret)
	if len(runes) <= 4 {
		return "****"
	}

	return "****" + string(runes[len(runes)-4:])
}

// normalizePrinterInput 清洗打印设备文本字段并校验枚举、配置与密钥长度。
// 入参：待落库的设备字段；返回业务错误表示校验不通过。
func normalizePrinterInput(in *PrinterSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.DeviceNo = strings.TrimSpace(in.DeviceNo)
	in.Config = strings.TrimSpace(in.Config)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.StoreID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店不能为空")
	}
	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "设备名称不能为空")
	}
	if len([]rune(in.Name)) > 64 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "设备名称不能超过 64 个字符")
	}
	if in.DeviceNo == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "设备号不能为空")
	}
	if len([]rune(in.DeviceNo)) > 64 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "设备号不能超过 64 个字符")
	}
	if !in.Type.IsValid() {
		return baseErrs.New(storeErrs.StorePrinterTypeInvalid)
	}
	if !in.AccessMode.IsValid() {
		return baseErrs.New(storeErrs.StorePrinterAccessModeInvalid)
	}
	if !in.Status.IsValid() {
		return baseErrs.New(storeErrs.StorePrinterStatusInvalid)
	}
	if in.Copies == 0 {
		in.Copies = 1
	}
	if in.Copies < 1 || in.Copies > printerMaxCopies {
		return baseErrs.New(storeErrs.StorePrinterCopiesInvalid)
	}
	if in.Config != "" {
		if len([]rune(in.Config)) > printerConfigMaxLength {
			return baseErrs.New(storeErrs.StorePrinterConfigInvalid)
		}
		if !json.Valid([]byte(in.Config)) {
			return baseErrs.New(storeErrs.StorePrinterConfigInvalid)
		}
	}
	if len(in.Secret) > printerSecretMaxLength {
		return baseErrs.New(storeErrs.StorePrinterSecretInvalid)
	}

	return nil
}
