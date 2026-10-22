package service

import (
	"context"
	"strings"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	storeErrs "eqray-flavor/app/modules/store/errs"
	"eqray-flavor/app/modules/store/model"
	"eqray-flavor/app/modules/store/repository"
)

// TableAreaListQuery 是桌台区域列表的查询条件。
type TableAreaListQuery struct {
	Keyword  string
	PageNo   int
	PageSize int
}

// TableAreaSaveInput 是新增 / 编辑桌台区域的入参。
type TableAreaSaveInput struct {
	StoreID types.ID
	Name    string
	Sort    int
	Remark  string
}

// TableAreaListRow 是区域列表输出：区域记录 + 门店名称与桌台数量。
type TableAreaListRow struct {
	model.StoreTableAreaModel
	StoreName  string `json:"storeName"`
	TableCount int64  `json:"tableCount"`
}

// TableAreaOption 是桌台弹窗与筛选器使用的区域下拉项。
type TableAreaOption struct {
	ID        types.ID `json:"id"`
	StoreID   types.ID `json:"storeId"`
	StoreName string   `json:"storeName"`
	Name      string   `json:"name"`
	Sort      int      `json:"sort"`
}

// StoreTableAreaService 负责桌台区域的增删改查与门店隔离。
type StoreTableAreaService struct {
	areas  *repository.StoreTableAreaRepository
	tables *repository.StoreTableRepository
	stores *StoreService
}

// NewStoreTableAreaService 构造桌台区域服务。
func NewStoreTableAreaService() *StoreTableAreaService {
	return &StoreTableAreaService{
		areas:  repository.NewStoreTableAreaRepository(),
		tables: repository.NewStoreTableRepository(),
		stores: NewStoreService(),
	}
}

// Page 分页查询当前用户可见范围内的桌台区域。
// 请求带门店上下文时只查该门店；否则按可见门店集合过滤。
func (s *StoreTableAreaService) Page(ctx context.Context, q TableAreaListQuery) ([]TableAreaListRow, int64, error) {
	filter, err := s.buildFilter(ctx, q.Keyword)
	if err != nil {
		return nil, 0, err
	}

	rows, total, err := s.areas.Paginate(ctx, filter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list, err := s.decorateRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Options 返回当前门店上下文或可见范围内的全部区域，供桌台筛选与弹窗选择。
func (s *StoreTableAreaService) Options(ctx context.Context) ([]TableAreaOption, error) {
	filter, err := s.buildFilter(ctx, "")
	if err != nil {
		return nil, err
	}

	rows, err := s.areas.All(ctx, filter)
	if err != nil {
		return nil, err
	}

	storeNames, err := s.storeNames(ctx, rows)
	if err != nil {
		return nil, err
	}

	list := make([]TableAreaOption, 0, len(rows))
	for _, row := range rows {
		list = append(list, TableAreaOption{
			ID:        row.ID,
			StoreID:   row.StoreID,
			StoreName: storeNames[row.StoreID],
			Name:      row.Name,
			Sort:      row.Sort,
		})
	}

	return list, nil
}

// Create 新增桌台区域：校验门店可见、名称唯一后落库。
// 入参：StoreID 未传时回填请求头门店；返回新区域 ID。
func (s *StoreTableAreaService) Create(ctx context.Context, in TableAreaSaveInput) (types.ID, error) {
	store, err := s.stores.ResolveWritableStore(ctx, in.StoreID)
	if err != nil {
		return 0, err
	}
	in.StoreID = store.ID
	if err := normalizeTableAreaInput(&in); err != nil {
		return 0, err
	}

	exists, err := s.areas.ExistsByName(ctx, in.StoreID, in.Name, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(storeErrs.StoreTableAreaExists, in.Name)
	}

	m := &model.StoreTableAreaModel{
		StoreID: in.StoreID,
		Name:    in.Name,
		Sort:    in.Sort,
		Remark:  in.Remark,
	}
	if err := s.areas.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑桌台区域，只允许调整名称、排序与备注，区域不能跨门店移动。
func (s *StoreTableAreaService) Update(ctx context.Context, id types.ID, in TableAreaSaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if in.StoreID > 0 && in.StoreID != m.StoreID {
		return baseErrs.New(storeErrs.StoreTableAreaStoreImmutable)
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}
	in.StoreID = m.StoreID
	if err := normalizeTableAreaInput(&in); err != nil {
		return err
	}

	exists, err := s.areas.ExistsByName(ctx, m.StoreID, in.Name, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(storeErrs.StoreTableAreaExists, in.Name)
	}

	m.Name = in.Name
	m.Sort = in.Sort
	m.Remark = in.Remark
	return s.areas.Update(ctx, m)
}

// Delete 删除桌台区域；区域下仍有桌台时拒绝，避免桌台失去归属。
func (s *StoreTableAreaService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.stores.ResolveWritableStore(ctx, m.StoreID); err != nil {
		return err
	}

	count, err := s.tables.CountByAreaID(ctx, m.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(storeErrs.StoreTableAreaInUse, count)
	}

	return s.areas.Delete(ctx, m)
}

// findByID 查询区域，不存在时返回固定错误码。
func (s *StoreTableAreaService) findByID(ctx context.Context, id types.ID) (*model.StoreTableAreaModel, error) {
	m, err := s.areas.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(storeErrs.StoreTableAreaNotExists)
	}

	return m, nil
}

// buildFilter 根据请求上下文与当前用户可见范围构造区域筛选条件。
func (s *StoreTableAreaService) buildFilter(ctx context.Context, keyword string) (repository.StoreTableAreaFilter, error) {
	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return repository.StoreTableAreaFilter{}, err
	}

	filter := repository.StoreTableAreaFilter{Keyword: strings.TrimSpace(keyword)}
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

// decorateRows 为区域列表补齐门店名称与桌台数量。
func (s *StoreTableAreaService) decorateRows(ctx context.Context, rows []model.StoreTableAreaModel) ([]TableAreaListRow, error) {
	list := make([]TableAreaListRow, 0, len(rows))
	if len(rows) == 0 {
		return list, nil
	}

	storeNames, err := s.storeNames(ctx, rows)
	if err != nil {
		return nil, err
	}

	areaIDs := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		areaIDs = append(areaIDs, row.ID)
	}
	counts, err := s.tables.CountByAreaIDs(ctx, areaIDs)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		list = append(list, TableAreaListRow{
			StoreTableAreaModel: row,
			StoreName:           storeNames[row.StoreID],
			TableCount:          counts[row.ID],
		})
	}

	return list, nil
}

// storeNames 批量查询区域所属门店名称，返回门店 ID 到名称的映射。
func (s *StoreTableAreaService) storeNames(ctx context.Context, rows []model.StoreTableAreaModel) (map[types.ID]string, error) {
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

// normalizeTableAreaInput 清洗区域名称与备注，并校验排序。
// 入参：待落库的区域字段；返回业务错误表示校验不通过。
func normalizeTableAreaInput(in *TableAreaSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.StoreID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店不能为空")
	}
	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "区域名称不能为空")
	}
	if len([]rune(in.Name)) > 32 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "区域名称不能超过 32 个字符")
	}
	if in.Sort < 0 {
		in.Sort = 0
	}

	return nil
}
