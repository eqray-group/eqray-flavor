package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	orderRepository "eqray-flavor/app/modules/order/repository"
	"eqray-flavor/app/modules/store/enums"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
	"eqray-flavor/app/modules/store/repository"
)

// TableListQuery 是桌台列表的查询条件。
type TableListQuery struct {
	Keyword  string
	AreaID   types.ID
	Status   *enums.TableStatusEnum
	PageNo   int
	PageSize int
}

// TableSaveInput 是新增 / 编辑桌台的入参。
// Code 仅新增时生效，编辑时沿用库中桌码，保证二维码长期有效。
type TableSaveInput struct {
	StoreID types.ID
	AreaID  types.ID
	Name    string
	Code    string
	Seats   int
	Sort    int
	Remark  string
}

// TableBatchInput 是批量生成桌台的入参。
type TableBatchInput struct {
	StoreID types.ID
	AreaID  types.ID
	Prefix  string
	Start   int
	Count   int
	Seats   int
}

// TableListRow 是桌台列表输出：桌台记录 + 门店、区域与状态展示字段。
type TableListRow struct {
	model.StoreTableModel
	StoreName  string `json:"storeName"`
	AreaName   string `json:"areaName"`
	StatusName string `json:"statusName"`
	// CurrentOrderNo 台位单号，没有台位单时为空
	CurrentOrderNo string `json:"currentOrderNo"`
	// CurrentOrderAmount 台位单应收金额（分）
	CurrentOrderAmount int64 `json:"currentOrderAmount"`
	// OpenedMinutes 已开台分钟数，未开台为 0
	OpenedMinutes int `json:"openedMinutes"`
}

// StoreTableService 负责桌台的增删改查、批量生成与开台 / 清台。
type StoreTableService struct {
	tables *repository.StoreTableRepository
	areas  *repository.StoreTableAreaRepository
	stores *StoreService
	// orders 换台后同步订单上的桌台快照；只依赖订单仓储，不会与订单服务形成循环引用
	orders *orderRepository.CaterOrderRepository
}

// NewStoreTableService 构造桌台服务。
func NewStoreTableService() *StoreTableService {
	return &StoreTableService{
		tables: repository.NewStoreTableRepository(),
		areas:  repository.NewStoreTableAreaRepository(),
		stores: NewStoreService(),
		orders: orderRepository.NewCaterOrderRepository(),
	}
}

// Page 分页查询当前用户可见范围内的桌台。
// 请求带门店上下文时只查该门店；否则按可见门店集合过滤。
func (s *StoreTableService) Page(ctx context.Context, q TableListQuery) ([]TableListRow, int64, error) {
	filter, err := s.buildFilter(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	rows, total, err := s.tables.Paginate(ctx, filter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list, err := s.decorateRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Create 新增桌台：校验门店可见、区域归属与桌号 / 桌码唯一后落库。
// 桌码未传时按「门店编码-桌号」自动生成；返回新桌台 ID。
func (s *StoreTableService) Create(ctx context.Context, in TableSaveInput) (types.ID, error) {
	store, err := s.stores.ResolveWritableStore(ctx, in.StoreID)
	if err != nil {
		return 0, err
	}
	in.StoreID = store.ID
	if err := normalizeTableInput(&in); err != nil {
		return 0, err
	}
	if err := s.ensureArea(ctx, store.ID, in.AreaID); err != nil {
		return 0, err
	}
	if in.Code == "" {
		in.Code = tableCode(store.Code, in.Name)
	}
	if err := validateTableCode(in.Code); err != nil {
		return 0, err
	}

	exists, err := s.tables.ExistsByName(ctx, in.StoreID, in.Name, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(storeErrs.StoreTableNameExists, in.Name)
	}
	exists, err = s.tables.ExistsByCode(ctx, in.Code, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(storeErrs.StoreTableCodeExists, in.Code)
	}

	m := &model.StoreTableModel{
		StoreID:        in.StoreID,
		AreaID:         in.AreaID,
		Name:           in.Name,
		Code:           in.Code,
		Seats:          in.Seats,
		Sort:           in.Sort,
		Status:         enums.TableStatusIdle,
		CurrentOrderID: 0,
		Remark:         in.Remark,
	}
	if err := s.tables.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑桌台：只允许调整桌号、区域、座位数、排序与备注。
// 桌码与桌台状态不允许通过编辑接口修改，避免二维码失效或绕过开台 / 清台流程。
func (s *StoreTableService) Update(ctx context.Context, id types.ID, in TableSaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if in.StoreID > 0 && in.StoreID != m.StoreID {
		return baseErrs.New(storeErrs.StoreTableStoreImmutable)
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}
	in.StoreID = m.StoreID
	in.Code = m.Code
	if err := normalizeTableInput(&in); err != nil {
		return err
	}
	if err := s.ensureArea(ctx, m.StoreID, in.AreaID); err != nil {
		return err
	}

	exists, err := s.tables.ExistsByName(ctx, m.StoreID, in.Name, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(storeErrs.StoreTableNameExists, in.Name)
	}

	m.AreaID = in.AreaID
	m.Name = in.Name
	m.Seats = in.Seats
	m.Sort = in.Sort
	m.Remark = in.Remark
	return s.tables.Update(ctx, m)
}

// Delete 删除桌台；使用中或仍有台位单的桌台拒绝删除。
func (s *StoreTableService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}
	if m.Status != enums.TableStatusIdle || m.CurrentOrderID > 0 {
		return baseErrs.New(storeErrs.StoreTableBusy)
	}

	return s.tables.Delete(ctx, m)
}

// BatchCreate 批量生成桌台，桌号按「前缀 + 补零序号」生成，桌码按「门店编码-桌号」生成。
// 入参：区间、座位数与门店 / 区域；返回实际生成数量，任意冲突时整批回滚。
func (s *StoreTableService) BatchCreate(ctx context.Context, in TableBatchInput) (int, error) {
	store, err := s.stores.ResolveWritableStore(ctx, in.StoreID)
	if err != nil {
		return 0, err
	}
	in.StoreID = store.ID
	if err := normalizeTableBatchInput(&in); err != nil {
		return 0, err
	}
	if err := s.ensureArea(ctx, store.ID, in.AreaID); err != nil {
		return 0, err
	}

	list, err := buildTableBatch(store, in)
	if err != nil {
		return 0, err
	}
	names := make([]string, 0, len(list))
	codes := make([]string, 0, len(list))
	for _, m := range list {
		names = append(names, m.Name)
		codes = append(codes, m.Code)
	}

	existingNames, err := s.tables.ExistingNames(ctx, store.ID, names)
	if err != nil {
		return 0, err
	}
	existingCodes, err := s.tables.ExistingCodes(ctx, codes)
	if err != nil {
		return 0, err
	}
	for _, name := range names {
		if existingNames[name] || existingCodes[tableCode(store.Code, name)] {
			return 0, baseErrs.New(storeErrs.StoreTableBatchExists, name)
		}
	}

	if err := s.tables.BatchCreate(ctx, list); err != nil {
		return 0, err
	}

	return len(list), nil
}

// Open 开台：仅空闲桌台可开，状态改为使用中并记录当前台位单与开台时间。
// orderID 为 0 表示订单模块尚未建单，仅占台；后续订单服务可回填台位单。
func (s *StoreTableService) Open(ctx context.Context, id types.ID, orderID types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}
	if orderID < 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "台位单 ID 不能为负数")
	}

	return s.tables.Open(ctx, m, orderID)
}

// Clear 清台：使用中或待清台的桌台恢复空闲，并清空当前台位单与开台时间。
func (s *StoreTableService) Clear(ctx context.Context, id types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}

	return s.tables.Clear(ctx, m)
}

// Transfer 换台：把进行中的台位单从一张桌台换到另一张空闲桌台。
// 入参：源桌台、目标桌台；返回错误表示源台未开台、目标台不空闲或不属于同一门店。
// 说明：只迁移台位绑定关系，订单本身不动——订单已经记录了桌台 ID 的，
// 需要在订单侧同步改；一期只做桌台状态与台位单迁移。
func (s *StoreTableService) Transfer(ctx context.Context, id, toID types.ID) error {
	if id == toID {
		return baseErrs.WithMsg(baseErrs.BadRequest, "目标桌台不能与当前桌台相同")
	}

	from, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	to, err := s.findByID(ctx, toID)
	if err != nil {
		return err
	}
	if from.StoreID != to.StoreID {
		return baseErrs.WithMsg(baseErrs.BadRequest, "目标桌台不属于当前门店")
	}
	if _, err := s.stores.ResolveWritableStore(ctx, from.StoreID); err != nil {
		return err
	}

	if from.Status != enums.TableStatusDining {
		return baseErrs.New(storeErrs.StoreTableStatusInvalid)
	}
	if to.Status != enums.TableStatusIdle {
		return baseErrs.New(storeErrs.StoreTableNotIdle)
	}

	if err := s.tables.Transfer(ctx, from, to); err != nil {
		return err
	}

	// 台位上有单时同步订单桌号，否则小票与后厨单还会打旧台。
	// 这一步失败不回滚换台（桌台状态已变更），报错让店员去订单页核对，避免静默不一致。
	if from.CurrentOrderID > 0 {
		if _, err := s.orders.UpdateTableBinding(ctx, from.CurrentOrderID, to.ID, to.Name); err != nil {
			return baseErrs.WithMsg(baseErrs.InternalServerError, "桌台已换，但订单桌号更新失败，请到订单页核对")
		}
	}

	return nil
}

// findByID 查询桌台，不存在时返回固定错误码。
func (s *StoreTableService) findByID(ctx context.Context, id types.ID) (*model.StoreTableModel, error) {
	m, err := s.tables.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(storeErrs.StoreTableNotExists)
	}

	return m, nil
}

// ensureArea 校验区域存在且属于目标门店。
func (s *StoreTableService) ensureArea(ctx context.Context, storeID, areaID types.ID) error {
	area, err := s.areas.FindByID(ctx, areaID)
	if err != nil {
		return err
	}
	if area == nil {
		return baseErrs.New(storeErrs.StoreTableAreaNotExists)
	}
	if area.StoreID != storeID {
		return baseErrs.New(storeErrs.StoreTableAreaMismatch)
	}

	return nil
}

// buildFilter 根据请求上下文与当前用户可见范围构造桌台筛选条件。
func (s *StoreTableService) buildFilter(ctx context.Context, q TableListQuery) (repository.StoreTableFilter, error) {
	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return repository.StoreTableFilter{}, err
	}

	filter := repository.StoreTableFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		AreaID:  q.AreaID,
		Status:  q.Status,
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

// decorateRows 为桌台列表补齐门店、区域与状态名称。
func (s *StoreTableService) decorateRows(ctx context.Context, rows []model.StoreTableModel) ([]TableListRow, error) {
	list := make([]TableListRow, 0, len(rows))
	if len(rows) == 0 {
		return list, nil
	}

	storeIDs := make([]types.ID, 0, len(rows))
	areaIDs := make([]types.ID, 0, len(rows))
	seenStore := make(map[types.ID]bool, len(rows))
	seenArea := make(map[types.ID]bool, len(rows))
	for _, row := range rows {
		if !seenStore[row.StoreID] {
			seenStore[row.StoreID] = true
			storeIDs = append(storeIDs, row.StoreID)
		}
		if !seenArea[row.AreaID] {
			seenArea[row.AreaID] = true
			areaIDs = append(areaIDs, row.AreaID)
		}
	}

	storeNames := make(map[types.ID]string, len(storeIDs))
	stores, err := s.stores.FindByIDs(ctx, storeIDs)
	if err != nil {
		return nil, err
	}
	for _, store := range stores {
		storeNames[store.ID] = store.Name
	}

	areaNames := make(map[types.ID]string, len(areaIDs))
	areas, err := s.areas.FindByIDs(ctx, areaIDs)
	if err != nil {
		return nil, err
	}
	for _, area := range areas {
		areaNames[area.ID] = area.Name
	}

	// 有台位单的桌台顺带把单号与金额带出来，收银台一眼能看到哪桌消费了多少
	orderIDs := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		if row.CurrentOrderID > 0 {
			orderIDs = append(orderIDs, row.CurrentOrderID)
		}
	}
	orders, err := s.orders.FindByIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, row := range rows {
		item := TableListRow{
			StoreTableModel: row,
			StoreName:       storeNames[row.StoreID],
			AreaName:        areaNames[row.AreaID],
			StatusName:      row.Status.String(),
		}
		if order, ok := orders[row.CurrentOrderID]; ok && order != nil {
			item.CurrentOrderNo = order.No
			item.CurrentOrderAmount = order.PayAmount
		}
		if row.OpenedAt != nil && !row.OpenedAt.IsZero() {
			item.OpenedMinutes = int(now.Sub(row.OpenedAt.Time()).Minutes())
		}
		list = append(list, item)
	}

	return list, nil
}

// normalizeTableInput 清洗桌号、桌码与备注，并校验座位数与排序。
// 入参：待落库的桌台字段；返回业务错误表示校验不通过。
func normalizeTableInput(in *TableSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Code = strings.TrimSpace(in.Code)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.StoreID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店不能为空")
	}
	if in.AreaID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "所属区域不能为空")
	}
	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "桌号不能为空")
	}
	if len([]rune(in.Name)) > 32 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "桌号不能超过 32 个字符")
	}
	if in.Seats == 0 {
		in.Seats = 4
	}
	if in.Seats < 1 || in.Seats > 99 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "座位数需在 1 ~ 99 之间")
	}
	if in.Sort < 0 {
		in.Sort = 0
	}

	return nil
}

// validateTableCode 校验桌码长度与空白字符，避免生成无法写入二维码的标识。
func validateTableCode(code string) error {
	if code == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "桌码不能为空")
	}
	if len([]rune(code)) > 64 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "桌码不能超过 64 个字符")
	}
	if strings.ContainsAny(code, " \t\r\n") {
		return baseErrs.WithMsg(baseErrs.BadRequest, "桌码不能包含空白字符")
	}

	return nil
}

// normalizeTableBatchInput 清洗批量生成前缀并校验区间、数量与座位数。
func normalizeTableBatchInput(in *TableBatchInput) error {
	in.Prefix = strings.TrimSpace(in.Prefix)

	if in.StoreID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店不能为空")
	}
	if in.AreaID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "所属区域不能为空")
	}
	if in.Prefix == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "桌号前缀不能为空")
	}
	if len([]rune(in.Prefix)) > 16 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "桌号前缀不能超过 16 个字符")
	}
	if in.Start <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "起始序号必须大于 0")
	}
	if in.Start > 9999 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "起始序号不能超过 9999")
	}
	if in.Count <= 0 || in.Count > 100 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "单次生成数量需在 1 ~ 100 之间")
	}
	if in.Seats == 0 {
		in.Seats = 4
	}
	if in.Seats < 1 || in.Seats > 99 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "座位数需在 1 ~ 99 之间")
	}

	return nil
}

// buildTableBatch 按前缀与区间生成桌台模型，名称序号至少补足两位。
// 入参：store 目标门店、in 已校验的批量参数；返回待插入的桌台列表。
func buildTableBatch(store *model.StoreModel, in TableBatchInput) ([]*model.StoreTableModel, error) {
	maxNumber := in.Start + in.Count - 1
	width := len(strconv.Itoa(maxNumber))
	if width < 2 {
		width = 2
	}

	list := make([]*model.StoreTableModel, 0, in.Count)
	for i := 0; i < in.Count; i++ {
		number := in.Start + i
		name := fmt.Sprintf("%s%0*d", in.Prefix, width, number)
		code := tableCode(store.Code, name)
		if err := validateTableCode(code); err != nil {
			return nil, err
		}

		list = append(list, &model.StoreTableModel{
			StoreID:        store.ID,
			AreaID:         in.AreaID,
			Name:           name,
			Code:           code,
			Seats:          in.Seats,
			Sort:           number,
			Status:         enums.TableStatusIdle,
			CurrentOrderID: 0,
		})
	}

	return list, nil
}

// tableCode 按「门店编码-桌号」生成全局唯一桌码。
func tableCode(storeCode, name string) string {
	return storeCode + "-" + name
}
