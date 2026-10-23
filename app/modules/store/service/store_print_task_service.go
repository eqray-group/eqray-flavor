package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	support "eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
	"eqray-flavor/app/modules/store/repository"
)

const (
	// printTaskPrinterPageSize 单次取设备的条数上限，门店同一类型设备不会超过这个量级。
	printTaskPrinterPageSize = 100
	// printTaskDefaultCopies 设备未配置份数时的兜底值。
	printTaskDefaultCopies = 1
)

// PrintTaskItem 是打印正文里的一条菜品明细。
type PrintTaskItem struct {
	// Name 菜品名称
	Name string
	// SkuName 规格名称，可为空
	SkuName string
	// Quantity 数量
	Quantity int
	// Subtotal 小计（分）
	Subtotal int64
	// Remark 明细备注，如少辣
	Remark string
}

// PrintTaskOrderInput 是订单打印任务的入参。
// 所有字段在生成时快照进正文，后续订单改动不影响已入队的单据。
type PrintTaskOrderInput struct {
	// StoreID 门店 ID
	StoreID types.ID
	// OrderID 订单 ID
	OrderID types.ID
	// OrderNo 订单号
	OrderNo string
	// PickupCode 取餐码
	PickupCode string
	// TableName 桌台名称，外卖为空
	TableName string
	// TypeText 订单类型文案
	TypeText string
	// ChannelText 下单渠道文案
	ChannelText string
	// Remark 顾客备注
	Remark string
	// PayAmount 应付金额（分）
	PayAmount int64
	// CreatedAt 下单时间文案
	CreatedAt string
	// Items 菜品明细
	Items []PrintTaskItem
}

// PrintTaskListQuery 是打印任务列表的筛选条件。
type PrintTaskListQuery struct {
	PageNo int
	PageSize int
	// PrintType 打印类型，nil 表示不限
	PrintType *enums.PrintTypeEnum
	// Status 任务状态，nil 表示不限
	Status *enums.PrintTaskStatusEnum
	// PrinterID 设备编号，0 表示不限
	PrinterID types.ID
	// OrderNo 订单号关键字
	OrderNo string
}

// PrintTaskListRow 是打印任务列表输出，补上门店名称与枚举文案。
type PrintTaskListRow struct {
	model.StorePrintTaskModel
	StoreName    string `json:"storeName"`
	PrintTypeName string `json:"printTypeName"`
	StatusName    string `json:"statusName"`
}

// StorePrintTaskService 负责打印任务的生成、查询与重打。
// 任务在订单支付成功时生成，正文按当时的模板版本渲染并存快照，重打直接复用正文。
type StorePrintTaskService struct {
	tasks     *repository.StorePrintTaskRepository
	printers  *repository.StorePrinterRepository
	templates *repository.StorePrintTemplateRepository
	// stores 复用门店服务：可见范围判定与门店名查询都要走同一套口径
	stores *StoreService
}

// NewStorePrintTaskService 构造打印任务服务。
func NewStorePrintTaskService() *StorePrintTaskService {
	return &StorePrintTaskService{
		tasks:     repository.NewStorePrintTaskRepository(),
		printers:  repository.NewStorePrinterRepository(),
		templates: repository.NewStorePrintTemplateRepository(),
		stores:    NewStoreService(),
	}
}

// CreateForOrder 为支付成功的订单生成打印任务：前台结账单 + 后厨厨打单。
// 入参：订单快照数据；返回生成的任务条数，0 表示门店没配设备或模板。
// 说明：门店没配打印设备或模板属于正常情况（很多店只接外卖不打印），直接跳过不算错误。
func (s *StorePrintTaskService) CreateForOrder(ctx context.Context, in PrintTaskOrderInput) (int, error) {
	if in.StoreID == 0 || in.OrderID == 0 {
		return 0, nil
	}

	store, err := s.stores.FindByID(ctx, in.StoreID)
	if err != nil {
		return 0, err
	}

	tasks := make([]model.StorePrintTaskModel, 0, 2)
	for _, printType := range []enums.PrintTypeEnum{enums.PrintTypeCheckout, enums.PrintTypeKitchen} {
		generated, err := s.buildTasks(ctx, in, store, printType)
		if err != nil {
			return len(tasks), err
		}
		tasks = append(tasks, generated...)
	}
	if len(tasks) == 0 {
		return 0, nil
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.tasks.CreateBatch(ctx, tx, tasks)
	}); err != nil {
		return 0, err
	}

	return len(tasks), nil
}

// buildTasks 按打印类型找到门店的启用设备与模板，为每台设备生成一条任务。
// 入参：订单快照、门店与打印类型；返回待插入的任务列表，无模板时返回空。
func (s *StorePrintTaskService) buildTasks(
	ctx context.Context,
	in PrintTaskOrderInput,
	store *model.StoreModel,
	printType enums.PrintTypeEnum,
) ([]model.StorePrintTaskModel, error) {
	enabled := baseEnums.CommonStatusEnabled
	printerType := enums.PrinterTypeFor(printType)

	printers, _, err := s.printers.Paginate(ctx, repository.StorePrinterFilter{
		StoreID: in.StoreID,
		Type:    &printerType,
		Status:  &enabled,
	}, 1, printTaskPrinterPageSize)
	if err != nil {
		return nil, err
	}
	if len(printers) == 0 {
		return nil, nil
	}

	templates, _, err := s.templates.Paginate(ctx, repository.StorePrintTemplateFilter{
		StoreID:     in.StoreID,
		PrinterType: &printerType,
		PrintType:   &printType,
		Status:      &enabled,
	}, 1, 1)
	if err != nil {
		return nil, err
	}
	if len(templates) == 0 {
		return nil, nil
	}
	template := &templates[0]

	storeName := ""
	if store != nil {
		storeName = store.Name
	}
	content := renderPrintContent(template.Content, in, storeName)

	tasks := make([]model.StorePrintTaskModel, 0, len(printers))
	for i := range printers {
		printer := &printers[i]
		copies := printer.Copies
		if copies <= 0 {
			copies = printTaskDefaultCopies
		}
		tasks = append(tasks, model.StorePrintTaskModel{
			StoreID:         in.StoreID,
			PrinterID:       printer.ID,
			PrinterName:     printer.Name,
			TemplateID:      template.ID,
			TemplateVersion: template.Version,
			PrintType:       printType,
			OrderID:         in.OrderID,
			OrderNo:         in.OrderNo,
			Content:         content,
			Status:          enums.PrintTaskStatusPending,
			Copies:          copies,
		})
	}

	return tasks, nil
}

// Page 分页查询打印任务，按门店可见范围过滤。
// 入参：筛选条件与分页；返回补好文案的任务列表与总数。
func (s *StorePrintTaskService) Page(ctx context.Context, q PrintTaskListQuery) ([]PrintTaskListRow, int64, error) {
	filter, err := s.buildFilter(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	rows, total, err := s.tasks.Paginate(ctx, filter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]PrintTaskListRow, 0, len(rows))
	if len(rows) == 0 {
		return list, total, nil
	}

	// 门店名按去重后的门店集合一次性查出，避免逐行查询
	storeIDs := make([]types.ID, 0, len(rows))
	seen := make(map[types.ID]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.StoreID]; ok {
			continue
		}
		seen[row.StoreID] = struct{}{}
		storeIDs = append(storeIDs, row.StoreID)
	}
	storeNames := s.storeNames(ctx, storeIDs)

	for _, row := range rows {
		list = append(list, PrintTaskListRow{
			StorePrintTaskModel: row,
			StoreName:           storeNames[row.StoreID],
			PrintTypeName:       row.PrintType.String(),
			StatusName:          row.Status.String(),
		})
	}

	return list, total, nil
}

// Detail 查询打印任务详情，含渲染后的正文。
// 入参：任务 ID；返回任务与门店名，无权限或不存在时返回业务错误。
func (s *StorePrintTaskService) Detail(ctx context.Context, id types.ID) (*PrintTaskListRow, error) {
	task, err := s.tasks.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, baseErrs.New(errs.StorePrintTaskNotExists)
	}

	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !scope.All && !scope.Contains(task.StoreID) {
		return nil, baseErrs.New(errs.StorePrintTaskAccessDenied)
	}

	row := PrintTaskListRow{
		StorePrintTaskModel: *task,
		StoreName:           s.storeNames(ctx, []types.ID{task.StoreID})[task.StoreID],
		PrintTypeName:       task.PrintType.String(),
		StatusName:          task.Status.String(),
	}

	return &row, nil
}

// Reprint 重打一张已打印 / 失败 / 取消的单据：把任务重置为待打印并累加重打次数。
// 入参：任务 ID；返回错误表示任务不存在、无门店权限或已在待打印队列中。
func (s *StorePrintTaskService) Reprint(ctx context.Context, id types.ID) error {
	task, err := s.tasks.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return baseErrs.New(errs.StorePrintTaskNotExists)
	}

	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return err
	}
	if !scope.All && !scope.Contains(task.StoreID) {
		return baseErrs.New(errs.StorePrintTaskAccessDenied)
	}

	rows, err := s.tasks.ResetForReprint(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return baseErrs.New(errs.StorePrintTaskStatusInvalid, task.Status.String())
	}

	return nil
}

// CancelByOrder 取消订单下仍待打印的任务，订单取消后调用，避免继续出纸。
// 入参：订单 ID；返回错误表示取消失败，调用方按非致命处理。
func (s *StorePrintTaskService) CancelByOrder(ctx context.Context, orderID types.ID) error {
	if orderID == 0 {
		return nil
	}

	_, err := s.tasks.CancelByOrder(ctx, orderID)

	return err
}

// buildFilter 按门店上下文与可见范围拼装筛选条件。
func (s *StorePrintTaskService) buildFilter(ctx context.Context, q PrintTaskListQuery) (repository.StorePrintTaskFilter, error) {
	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return repository.StorePrintTaskFilter{}, err
	}

	filter := repository.StorePrintTaskFilter{
		PrintType: q.PrintType,
		Status:    q.Status,
		PrinterID: q.PrinterID,
		OrderNo:   strings.TrimSpace(q.OrderNo),
	}
	if contextStoreID := support.StoreID(ctx); contextStoreID > 0 {
		if !scope.All && !scope.Contains(contextStoreID) {
			return filter, baseErrs.New(errs.StoreAccessDenied)
		}
		filter.StoreID = contextStoreID
	} else if !scope.All {
		filter.RestrictStores = true
		filter.StoreIDs = scope.IDs
	}

	return filter, nil
}

// storeNames 批量取门店名称，缺失时返回空串。
func (s *StorePrintTaskService) storeNames(ctx context.Context, ids []types.ID) map[types.ID]string {
	names := make(map[types.ID]string, len(ids))
	if len(ids) == 0 {
		return names
	}

	stores, err := s.stores.FindByIDs(ctx, ids)
	if err != nil {
		return names
	}
	for _, store := range stores {
		names[store.ID] = store.Name
	}

	return names
}

// renderPrintContent 用订单数据渲染模板正文。
// 支持占位符：{{storeName}} {{orderNo}} {{pickupCode}} {{tableName}} {{orderType}}
// {{channel}} {{createdAt}} {{remark}} {{payAmount}} {{items}}
// 说明：未知占位符原样保留，方便运营在模板里直接写文字。
func renderPrintContent(template string, in PrintTaskOrderInput, storeName string) string {
	replacer := strings.NewReplacer(
		"{{storeName}}", storeName,
		"{{orderNo}}", in.OrderNo,
		"{{pickupCode}}", in.PickupCode,
		"{{tableName}}", in.TableName,
		"{{orderType}}", in.TypeText,
		"{{channel}}", in.ChannelText,
		"{{createdAt}}", in.CreatedAt,
		"{{remark}}", in.Remark,
		"{{payAmount}}", formatPrintYuan(in.PayAmount),
		"{{items}}", renderPrintItems(in.Items),
	)

	return replacer.Replace(template)
}

// renderPrintItems 把菜品明细渲染成多行文本，一行一道菜。
func renderPrintItems(items []PrintTaskItem) string {
	if len(items) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, item := range items {
		name := item.Name
		if item.SkuName != "" {
			name += "(" + item.SkuName + ")"
		}
		builder.WriteString(fmt.Sprintf("%s x%d  %s\n", name, item.Quantity, formatPrintYuan(item.Subtotal)))
		if item.Remark != "" {
			builder.WriteString("  备注：" + item.Remark + "\n")
		}
	}

	return strings.TrimRight(builder.String(), "\n")
}

// formatPrintYuan 把金额（分）转成两位小数的元。
func formatPrintYuan(cent int64) string {
	return strconv.FormatFloat(float64(cent)/100, 'f', 2, 64)
}
