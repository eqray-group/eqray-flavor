package service

import (
	"context"
	"strings"

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
	// printTemplateMinPaperWidth 是纸宽下限，单位毫米。
	printTemplateMinPaperWidth = 40
	// printTemplateMaxPaperWidth 是纸宽上限，单位毫米。
	printTemplateMaxPaperWidth = 200
	// printTemplateContentMaxLength 是模板正文长度上限，避免异常大报文。
	printTemplateContentMaxLength = 65535
)

// PrintTemplateListQuery 是打印模板列表的查询条件。
type PrintTemplateListQuery struct {
	Keyword     string
	PrinterType *enums.PrinterTypeEnum
	PrintType   *enums.PrintTypeEnum
	Status      *baseEnums.CommonStatusEnum
	PageNo      int
	PageSize    int
}

// PrintTemplateSaveInput 是新增 / 编辑打印模板的入参。
// Version 不在入参中，由服务层在正文变化时自动递增。
type PrintTemplateSaveInput struct {
	StoreID     types.ID
	Name        string
	PrinterType enums.PrinterTypeEnum
	PrintType   enums.PrintTypeEnum
	PaperWidth  int
	Content     string
	Status      baseEnums.CommonStatusEnum
	Remark      string
}

// PrintTemplateListRow 是打印模板列表输出：模板记录 + 门店与枚举展示字段。
type PrintTemplateListRow struct {
	model.StorePrintTemplateModel
	StoreName       string `json:"storeName"`
	PrinterTypeName string `json:"printerTypeName"`
	PrintTypeName   string `json:"printTypeName"`
	StatusName      string `json:"statusName"`
}

// StorePrintTemplateService 负责打印模板的增删改查、门店隔离与版本递增。
type StorePrintTemplateService struct {
	templates *repository.StorePrintTemplateRepository
	stores    *StoreService
}

// NewStorePrintTemplateService 构造打印模板服务。
func NewStorePrintTemplateService() *StorePrintTemplateService {
	return &StorePrintTemplateService{
		templates: repository.NewStorePrintTemplateRepository(),
		stores:    NewStoreService(),
	}
}

// Page 分页查询当前用户可见范围内的打印模板。
// 请求带门店上下文时只查该门店；否则按可见门店集合过滤。
func (s *StorePrintTemplateService) Page(ctx context.Context, q PrintTemplateListQuery) ([]PrintTemplateListRow, int64, error) {
	filter, err := s.buildFilter(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	rows, total, err := s.templates.Paginate(ctx, filter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list, err := s.decorateRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Create 新增打印模板：校验门店可见与「设备类型 + 打印类型」唯一，初始版本为 1。
// 入参：未传门店时回填请求头门店；返回新模板 ID。
func (s *StorePrintTemplateService) Create(ctx context.Context, in PrintTemplateSaveInput) (types.ID, error) {
	store, err := s.stores.ResolveWritableStore(ctx, in.StoreID)
	if err != nil {
		return 0, err
	}
	in.StoreID = store.ID
	if err := normalizePrintTemplateInput(&in); err != nil {
		return 0, err
	}

	exists, err := s.templates.ExistsByKind(ctx, in.StoreID, in.PrinterType, in.PrintType, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(storeErrs.StorePrintTemplateKindExists, in.PrinterType.String(), in.PrintType.String())
	}

	m := &model.StorePrintTemplateModel{
		StoreID:     in.StoreID,
		Name:        in.Name,
		PrinterType: in.PrinterType,
		PrintType:   in.PrintType,
		PaperWidth:  in.PaperWidth,
		Content:     in.Content,
		Version:     1,
		Status:      in.Status,
		Remark:      in.Remark,
	}
	if err := s.templates.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑打印模板；模板不能跨门店移动，正文变化时版本号自动加 1。
func (s *StorePrintTemplateService) Update(ctx context.Context, id types.ID, in PrintTemplateSaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if in.StoreID > 0 && in.StoreID != m.StoreID {
		return baseErrs.New(storeErrs.StorePrintTemplateStoreImmutable)
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}
	in.StoreID = m.StoreID
	if err := normalizePrintTemplateInput(&in); err != nil {
		return err
	}

	exists, err := s.templates.ExistsByKind(ctx, m.StoreID, in.PrinterType, in.PrintType, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(storeErrs.StorePrintTemplateKindExists, in.PrinterType.String(), in.PrintType.String())
	}

	m.Name = in.Name
	m.PrinterType = in.PrinterType
	m.PrintType = in.PrintType
	m.PaperWidth = in.PaperWidth
	m.Status = in.Status
	m.Remark = in.Remark
	m.Version = nextPrintTemplateVersion(m.Version, m.Content, in.Content)
	m.Content = in.Content

	return s.templates.Update(ctx, m)
}

// Delete 物理删除打印模板；历史打印任务应在生成时保存模板版本与渲染快照。
func (s *StorePrintTemplateService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}

	return s.templates.Delete(ctx, m)
}

// findByID 查询打印模板，不存在时返回固定错误码。
func (s *StorePrintTemplateService) findByID(ctx context.Context, id types.ID) (*model.StorePrintTemplateModel, error) {
	m, err := s.templates.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(storeErrs.StorePrintTemplateNotExists)
	}

	return m, nil
}

// buildFilter 根据请求上下文与当前用户可见范围构造打印模板筛选条件。
func (s *StorePrintTemplateService) buildFilter(ctx context.Context, q PrintTemplateListQuery) (repository.StorePrintTemplateFilter, error) {
	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return repository.StorePrintTemplateFilter{}, err
	}

	filter := repository.StorePrintTemplateFilter{
		Keyword:     strings.TrimSpace(q.Keyword),
		PrinterType: q.PrinterType,
		PrintType:   q.PrintType,
		Status:      q.Status,
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

// decorateRows 为打印模板列表补齐门店名称、枚举文案与状态展示字段。
func (s *StorePrintTemplateService) decorateRows(ctx context.Context, rows []model.StorePrintTemplateModel) ([]PrintTemplateListRow, error) {
	list := make([]PrintTemplateListRow, 0, len(rows))
	if len(rows) == 0 {
		return list, nil
	}

	storeNames, err := s.storeNames(ctx, rows)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		list = append(list, PrintTemplateListRow{
			StorePrintTemplateModel: row,
			StoreName:               storeNames[row.StoreID],
			PrinterTypeName:         row.PrinterType.String(),
			PrintTypeName:           row.PrintType.String(),
			StatusName:              row.Status.String(),
		})
	}

	return list, nil
}

// storeNames 批量查询模板所属门店名称，返回门店 ID 到名称的映射。
func (s *StorePrintTemplateService) storeNames(ctx context.Context, rows []model.StorePrintTemplateModel) (map[types.ID]string, error) {
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

// nextPrintTemplateVersion 计算模板新版本号：正文未变化保持原版本，变化则加 1。
// 入参：当前版本、旧正文、新正文；返回更新后的版本号，异常旧版本按 1 兜底。
func nextPrintTemplateVersion(current int, oldContent, newContent string) int {
	if current < 1 {
		current = 1
	}
	if oldContent == newContent {
		return current
	}

	return current + 1
}

// normalizePrintTemplateInput 清洗打印模板文本字段并校验类型、纸宽与正文长度。
// 入参：待落库的模板字段；返回业务错误表示校验不通过。正文保留原始空白，避免破坏排版。
func normalizePrintTemplateInput(in *PrintTemplateSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.StoreID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店不能为空")
	}
	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "模板名称不能为空")
	}
	if len([]rune(in.Name)) > 64 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "模板名称不能超过 64 个字符")
	}
	if !in.PrinterType.IsValid() {
		return baseErrs.New(storeErrs.StorePrinterTypeInvalid)
	}
	if !in.PrintType.IsValid() {
		return baseErrs.New(storeErrs.StorePrintTemplatePrintTypeInvalid)
	}
	if in.PaperWidth < printTemplateMinPaperWidth || in.PaperWidth > printTemplateMaxPaperWidth {
		return baseErrs.New(storeErrs.StorePrintTemplatePaperWidthInvalid)
	}
	if strings.TrimSpace(in.Content) == "" {
		return baseErrs.New(storeErrs.StorePrintTemplateContentRequired)
	}
	if len([]rune(in.Content)) > printTemplateContentMaxLength {
		return baseErrs.WithMsg(baseErrs.BadRequest, "模板内容不能超过 65535 个字符")
	}
	if !in.Status.IsValid() {
		return baseErrs.New(storeErrs.StorePrintTemplateStatusInvalid)
	}

	return nil
}
