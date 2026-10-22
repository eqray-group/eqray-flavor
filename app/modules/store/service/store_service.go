// Package service 承载门店域的业务规则：组织树层级约束、编码唯一、删除保护。
package service

import (
	"context"
	"strconv"
	"strings"
	"time"

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

// storeTimeLayout 是营业时间的统一格式，库中按 HH:mm 定长存储。
const storeTimeLayout = "15:04"

const (
	// visibleStoreCachePrefix 是用户可见门店集合的缓存键前缀。
	visibleStoreCachePrefix = "store:visible:"
	// visibleStoreCacheTTL 与权限缓存一致，员工变更时会主动失效。
	visibleStoreCacheTTL = 30 * time.Minute
)

// StoreSaveInput 是新增 / 编辑门店的入参。
// Code 仅在新增时生效，编辑时后端会忽略并沿用库中编码。
type StoreSaveInput struct {
	ParentID        types.ID
	Type            enums.StoreTypeEnum
	Code            string
	Name            string
	ShortName       string
	Status          enums.StoreStatusEnum
	ContactName     string
	ContactMobile   string
	DistrictID      types.ID
	Address         string
	Longitude       float64
	Latitude        float64
	OpenTime        string
	CloseTime       string
	DeliveryRule    string
	AutoAccept      int8
	OnlineEnabled   int8
	DeliveryEnabled int8
	Sort            int
	Remark          string
}

// StoreListQuery 是门店列表的查询条件。
// EntityID 表示「只看某节点下（含自身）的门店」，配合左侧组织树使用。
type StoreListQuery struct {
	Keyword  string
	Type     *enums.StoreTypeEnum
	Status   *enums.StoreStatusEnum
	EntityID *types.ID
	PageNo   int
	PageSize int
}

// StoreTreeNode 是组织树节点：门店自身字段 + 下级节点。
// 仅用于树接口输出，不落库。
type StoreTreeNode struct {
	model.StoreModel
	Children []*StoreTreeNode `json:"children"`
}

// StoreScope 描述当前用户可见的门店范围。
// All=true 表示超管或未关联任何门店的总部账号，可跨店访问；否则只能访问 IDs。
type StoreScope struct {
	All bool
	IDs []types.ID
}

// Contains 判断门店是否在可见范围内。
func (s StoreScope) Contains(id types.ID) bool {
	for _, visible := range s.IDs {
		if visible == id {
			return true
		}
	}

	return false
}

// VisibleStore 是门店切换器使用的最小门店信息。
type VisibleStore struct {
	ID     types.ID              `json:"id"`
	Code   string                `json:"code"`
	Name   string                `json:"name"`
	Status enums.StoreStatusEnum `json:"status"`
}

// StoreService 负责门店配置与组织树的业务规则。
type StoreService struct {
	stores    *repository.StoreRepository
	employees *repository.StoreEmployeeRepository
}

// NewStoreService 构造门店服务。
func NewStoreService() *StoreService {
	return &StoreService{
		stores:    repository.NewStoreRepository(),
		employees: repository.NewStoreEmployeeRepository(),
	}
}

// Page 分页查询门店，支持名称 / 编码关键字与类型、状态筛选。
// 传了 EntityID 时只返回该节点及其下级的门店（总部 / 大区本身不出现在结果里）。
func (s *StoreService) Page(ctx context.Context, q StoreListQuery) ([]model.StoreModel, int64, error) {
	filter := repository.StoreFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Type:    q.Type,
		Status:  q.Status,
	}

	if q.EntityID != nil && *q.EntityID > 0 {
		ids, err := s.shopIDsUnder(ctx, *q.EntityID)
		if err != nil {
			return nil, 0, err
		}
		// 节点下没有门店时直接返回空，避免 WhereIn 空集退化成全表
		if len(ids) == 0 {
			return []model.StoreModel{}, 0, nil
		}
		filter.IDs = ids
	}

	return s.stores.Paginate(ctx, filter, q.PageNo, q.PageSize)
}

// shopIDsUnder 收集指定节点及其下级中的门店 ID。
// 入参：rootID 组织树节点 ID；返回该子树内 type=门店 的 ID 集合。
func (s *StoreService) shopIDsUnder(ctx context.Context, rootID types.ID) ([]types.ID, error) {
	list, err := s.stores.All(ctx)
	if err != nil {
		return nil, err
	}

	children := make(map[types.ID][]types.ID, len(list))
	typesByID := make(map[types.ID]enums.StoreTypeEnum, len(list))
	for _, item := range list {
		children[item.ParentID] = append(children[item.ParentID], item.ID)
		typesByID[item.ID] = item.Type
	}

	ids := make([]types.ID, 0, len(list))
	queue := []types.ID{rootID}
	visited := map[types.ID]bool{rootID: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if typesByID[current] == enums.StoreTypeShop {
			ids = append(ids, current)
		}
		for _, child := range children[current] {
			if visited[child] {
				continue
			}
			visited[child] = true
			queue = append(queue, child)
		}
	}

	return ids, nil
}

// Tree 返回完整组织树，根节点为无上级的总部；孤儿节点按根节点兜底展示。
func (s *StoreService) Tree(ctx context.Context) ([]*StoreTreeNode, error) {
	list, err := s.stores.All(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make(map[types.ID]*StoreTreeNode, len(list))
	ordered := make([]*StoreTreeNode, 0, len(list))
	for i := range list {
		node := &StoreTreeNode{StoreModel: list[i], Children: make([]*StoreTreeNode, 0)}
		nodes[node.ID] = node
		ordered = append(ordered, node)
	}

	roots := make([]*StoreTreeNode, 0, 1)
	for _, node := range ordered {
		parent, ok := nodes[node.ParentID]
		if node.ParentID == 0 || !ok {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	return roots, nil
}

// Detail 查询门店详情，不存在时返回业务错误码。
func (s *StoreService) Detail(ctx context.Context, id types.ID) (*model.StoreModel, error) {
	m, err := s.stores.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(storeErrs.StoreNotExists, id)
	}

	return m, nil
}

// FindByID 查询未删除门店，未命中返回 (nil, nil)。
// 门店上下文中间件用它区分「门店不存在」与「存在但无权访问」。
func (s *StoreService) FindByID(ctx context.Context, id types.ID) (*model.StoreModel, error) {
	return s.stores.FindByID(ctx, id)
}

// FindByIDs 批量查询门店，供员工列表补齐门店名称。
func (s *StoreService) FindByIDs(ctx context.Context, ids []types.ID) ([]model.StoreModel, error) {
	return s.stores.FindByIDs(ctx, ids)
}

// VisibleStores 返回当前用户可见的门店节点（仅 type=3 的门店），供切换器使用。
// 超管与未关联任何门店的总部账号返回全部门店；门店员工按启用记录返回。
func (s *StoreService) VisibleStores(ctx context.Context, userID types.ID) ([]VisibleStore, error) {
	scope, err := s.Scope(ctx, userID)
	if err != nil {
		return nil, err
	}

	var shops []model.StoreModel
	if scope.All {
		shops, err = s.stores.AllShops(ctx)
	} else {
		shops, err = s.stores.ActiveShopsByIDs(ctx, scope.IDs)
	}
	if err != nil {
		return nil, err
	}

	list := make([]VisibleStore, 0, len(shops))
	for _, shop := range shops {
		list = append(list, VisibleStore{
			ID:     shop.ID,
			Code:   shop.Code,
			Name:   shop.Name,
			Status: shop.Status,
		})
	}

	return list, nil
}

// Scope 计算当前用户可见的门店范围，并缓存门店员工的计算结果。
// 规则（design/23 §1）：超管全量；无任何员工关联视为总部账号、全量；
// 有员工关联则仅启用记录对应的有效门店，空集合表示无可见门店。
func (s *StoreService) Scope(ctx context.Context, userID types.ID) (StoreScope, error) {
	super, err := support.IsSuperAdmin(ctx, userID)
	if err != nil {
		return StoreScope{}, err
	}
	if super {
		return StoreScope{All: true}, nil
	}
	// 未登录或用户 ID 非法时按无可见门店处理，避免误放行
	if userID <= 0 {
		return StoreScope{}, nil
	}

	cacheKey := visibleStoreCacheKey(userID)
	if cached := facades.Cache().GetString(cacheKey); cached != "" {
		if scope, ok := decodeStoreScope(cached); ok {
			return scope, nil
		}
	}

	// 只要存在任何关联（含禁用）就不再视为总部账号，避免禁用员工后反而跨店可见
	hasAny, err := s.employees.HasAnyByUser(ctx, userID)
	if err != nil {
		return StoreScope{}, err
	}

	scope := StoreScope{All: !hasAny}
	if hasAny {
		enabled := baseEnums.CommonStatusEnabled
		storeIDs, err := s.employees.ListStoreIDsByUser(ctx, userID, &enabled)
		if err != nil {
			return StoreScope{}, err
		}
		shops, err := s.stores.ActiveShopsByIDs(ctx, storeIDs)
		if err != nil {
			return StoreScope{}, err
		}

		scope.IDs = make([]types.ID, 0, len(shops))
		for _, shop := range shops {
			scope.IDs = append(scope.IDs, shop.ID)
		}
	}

	_ = facades.Cache().Put(cacheKey, encodeStoreScope(scope), visibleStoreCacheTTL)

	return scope, nil
}

// InvalidateScope 删除指定用户的可见门店缓存，员工分配变更后调用。
func (s *StoreService) InvalidateScope(userID types.ID) {
	if userID <= 0 {
		return
	}

	facades.Cache().Forget(visibleStoreCacheKey(userID))
}

// EnsureStoreVisible 校验门店是否在用户可见范围内，越权返回固定错误码。
// 入参：ctx、userID、storeID；返回 nil 表示允许访问。
func (s *StoreService) EnsureStoreVisible(ctx context.Context, userID, storeID types.ID) error {
	if storeID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店 ID 不能为空")
	}

	scope, err := s.Scope(ctx, userID)
	if err != nil {
		return err
	}
	if scope.All || scope.Contains(storeID) {
		return nil
	}

	return baseErrs.New(storeErrs.StoreAccessDenied)
}

// ResolveWritableStore 解析门店级写操作的目标门店。
// 未传 storeID 时回填请求头门店，再校验上下文一致、门店存在、类型为门店且在当前用户可见范围内。
// 入参：ctx 请求上下文、storeID 请求体门店（可为 0）；返回可用于后续业务校验的门店。
func (s *StoreService) ResolveWritableStore(ctx context.Context, storeID types.ID) (*model.StoreModel, error) {
	if storeID == 0 {
		storeID = support.StoreID(ctx)
	}
	if storeID <= 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "门店不能为空")
	}
	if contextStoreID := support.StoreID(ctx); contextStoreID > 0 && contextStoreID != storeID {
		return nil, baseErrs.New(storeErrs.StoreContextMismatch)
	}

	store, err := s.FindByID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, baseErrs.New(storeErrs.StoreNotExists, storeID)
	}
	if store.Type != enums.StoreTypeShop {
		return nil, baseErrs.New(storeErrs.StoreNotShop)
	}
	if err := s.EnsureStoreVisible(ctx, types.ID(support.UserID(ctx)), storeID); err != nil {
		return nil, err
	}

	return store, nil
}

// Create 新增门店：校验层级、编码唯一后落库。
// 返回新门店 ID。
func (s *StoreService) Create(ctx context.Context, in StoreSaveInput) (types.ID, error) {
	if err := normalizeStoreInput(&in); err != nil {
		return 0, err
	}
	if in.Type == enums.StoreTypeHeadquarter {
		if err := s.ensureSingleHeadquarter(ctx, 0); err != nil {
			return 0, err
		}
	}
	if err := s.checkParent(ctx, 0, in.Type, in.ParentID); err != nil {
		return 0, err
	}

	exists, err := s.stores.ExistsByCode(ctx, in.Code, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(storeErrs.StoreCodeExists, in.Code)
	}

	m := &model.StoreModel{
		ParentID:        in.ParentID,
		Type:            in.Type,
		Code:            in.Code,
		Name:            in.Name,
		ShortName:       in.ShortName,
		Status:          in.Status,
		ContactName:     in.ContactName,
		ContactMobile:   in.ContactMobile,
		DistrictID:      in.DistrictID,
		Address:         in.Address,
		Longitude:       in.Longitude,
		Latitude:        in.Latitude,
		OpenTime:        in.OpenTime,
		CloseTime:       in.CloseTime,
		DeliveryRule:    in.DeliveryRule,
		AutoAccept:      in.AutoAccept,
		OnlineEnabled:   in.OnlineEnabled,
		DeliveryEnabled: in.DeliveryEnabled,
		Sort:            in.Sort,
		Remark:          in.Remark,
	}
	if err := s.stores.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑门店：编码沿用库中值，层级变更时重新做环检测。
func (s *StoreService) Update(ctx context.Context, id types.ID, in StoreSaveInput) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	// 编码是外部对接的稳定标识，编辑时一律忽略入参
	in.Code = m.Code
	if err := normalizeStoreInput(&in); err != nil {
		return err
	}
	// 品牌总部是唯一根节点，不允许改为大区或门店，避免组织树失去根
	if m.Type == enums.StoreTypeHeadquarter && in.Type != enums.StoreTypeHeadquarter {
		return baseErrs.New(storeErrs.StoreHeadquarterProtected)
	}
	// 只有当前不是总部时才需要校验唯一性，避免编辑总部自身被误拦
	if in.Type == enums.StoreTypeHeadquarter && m.Type != enums.StoreTypeHeadquarter {
		if err := s.ensureSingleHeadquarter(ctx, m.ID); err != nil {
			return err
		}
	}
	if err := s.checkParent(ctx, m.ID, in.Type, in.ParentID); err != nil {
		return err
	}
	// 门店是最末级，已有下属节点时不允许降级为门店，否则子树悬空
	if in.Type == enums.StoreTypeShop && m.Type != enums.StoreTypeShop {
		count, err := s.stores.CountChildren(ctx, m.ID)
		if err != nil {
			return err
		}
		if count > 0 {
			return baseErrs.New(storeErrs.StoreTypeHasChildren, m.Name)
		}
	}

	m.ParentID = in.ParentID
	m.Type = in.Type
	m.Name = in.Name
	m.ShortName = in.ShortName
	m.Status = in.Status
	m.ContactName = in.ContactName
	m.ContactMobile = in.ContactMobile
	m.DistrictID = in.DistrictID
	m.Address = in.Address
	m.Longitude = in.Longitude
	m.Latitude = in.Latitude
	m.OpenTime = in.OpenTime
	m.CloseTime = in.CloseTime
	m.DeliveryRule = in.DeliveryRule
	m.AutoAccept = in.AutoAccept
	m.OnlineEnabled = in.OnlineEnabled
	m.DeliveryEnabled = in.DeliveryEnabled
	m.Sort = in.Sort
	m.Remark = in.Remark

	return s.stores.Update(ctx, m)
}

// Delete 删除门店：总部根节点不可删除，仍有下级节点时拒绝，避免组织树断裂。
func (s *StoreService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}
	if m.Type == enums.StoreTypeHeadquarter {
		return baseErrs.New(storeErrs.StoreHeadquarterProtected)
	}

	count, err := s.stores.CountChildren(ctx, m.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(storeErrs.StoreHasChildren, m.Name)
	}

	return s.stores.Delete(ctx, m)
}

// ensureSingleHeadquarter 校验品牌总部唯一，全库只允许一个总部根节点。
// 入参：excludeID 编辑时排除自身，新增传 0；返回 StoreHeadquarterExists 表示已存在总部。
func (s *StoreService) ensureSingleHeadquarter(ctx context.Context, excludeID types.ID) error {
	count, err := s.stores.CountByType(ctx, enums.StoreTypeHeadquarter, excludeID)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(storeErrs.StoreHeadquarterExists)
	}

	return nil
}

// checkParent 校验上级节点是否符合层级约束，并阻止指向自身或自身下级。
// 入参：selfID 编辑时为当前门店 ID（新增传 0）、typ 当前节点类型、parentID 上级 ID。
func (s *StoreService) checkParent(ctx context.Context, selfID types.ID, typ enums.StoreTypeEnum, parentID types.ID) error {
	if typ == enums.StoreTypeHeadquarter {
		if parentID != 0 {
			return baseErrs.WithMsg(baseErrs.BadRequest, "总部是根节点，不能指定上级")
		}
		return nil
	}
	if parentID == 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "大区与门店必须指定上级节点")
	}

	parent, err := s.stores.FindByID(ctx, parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return baseErrs.New(storeErrs.StoreParentNotExists, parentID)
	}

	// 大区只能挂总部；门店可挂总部直营，也可挂大区
	switch typ {
	case enums.StoreTypeRegion:
		if parent.Type != enums.StoreTypeHeadquarter {
			return baseErrs.New(storeErrs.StoreParentTypeInvalid, parent.Type, typ)
		}
	case enums.StoreTypeShop:
		if !parent.Type.IsBranch() {
			return baseErrs.New(storeErrs.StoreParentTypeInvalid, parent.Type, typ)
		}
	}

	if selfID == 0 {
		return nil
	}
	return s.checkCycle(ctx, selfID, parent)
}

// checkCycle 从上级节点向上回溯，命中自身即说明会形成环。
// 入参：selfID 当前门店 ID、parent 新的上级节点；返回 StoreParentCycle 表示非法。
func (s *StoreService) checkCycle(ctx context.Context, selfID types.ID, parent *model.StoreModel) error {
	current := parent
	for current != nil {
		if current.ID == selfID {
			return baseErrs.New(storeErrs.StoreParentCycle)
		}
		if current.ParentID == 0 {
			return nil
		}

		next, err := s.stores.FindByID(ctx, current.ParentID)
		if err != nil {
			return err
		}
		current = next
	}

	return nil
}

// normalizeStoreInput 清洗字符串字段并校验枚举、金额与 JSON 合法性。
// 入参为待落库的门店字段，返回业务错误表示校验不通过。
func normalizeStoreInput(in *StoreSaveInput) error {
	in.Code = strings.TrimSpace(in.Code)
	in.Name = strings.TrimSpace(in.Name)
	in.ShortName = strings.TrimSpace(in.ShortName)
	in.ContactName = strings.TrimSpace(in.ContactName)
	in.ContactMobile = strings.TrimSpace(in.ContactMobile)
	in.Address = strings.TrimSpace(in.Address)
	in.OpenTime = strings.TrimSpace(in.OpenTime)
	in.CloseTime = strings.TrimSpace(in.CloseTime)
	in.Remark = strings.TrimSpace(in.Remark)

	if err := normalizeStoreTimes(in); err != nil {
		return err
	}
	rule, err := normalizeDeliveryRule(in.DeliveryRule)
	if err != nil {
		return err
	}
	in.DeliveryRule = rule

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店名称不能为空")
	}
	if in.Code == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店编码不能为空")
	}
	if !in.Type.IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店类型取值不合法")
	}
	if !in.Status.IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店状态取值不合法")
	}
	if in.DistrictID < 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "区县 ID 不能为负数")
	}
	if in.Longitude < -180 || in.Longitude > 180 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "经度需在 -180 ~ 180 之间")
	}
	if in.Latitude < -90 || in.Latitude > 90 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "纬度需在 -90 ~ 90 之间")
	}
	if in.Sort < 0 {
		in.Sort = 0
	}

	return nil
}

// visibleStoreCacheKey 返回用户可见门店集合的缓存键。
func visibleStoreCacheKey(userID types.ID) string {
	return visibleStoreCachePrefix + userID.String()
}

// encodeStoreScope 把可见范围编码进缓存：
// 「a」表示全量，「r:1,2」表示仅这些门店，「r:」表示有限制但无可见门店。
func encodeStoreScope(scope StoreScope) string {
	if scope.All {
		return "a"
	}

	parts := make([]string, 0, len(scope.IDs))
	for _, id := range scope.IDs {
		if id > 0 {
			parts = append(parts, id.String())
		}
	}

	return "r:" + strings.Join(parts, ",")
}

// decodeStoreScope 解析缓存值，格式非法时 ok=false，调用方回源数据库。
func decodeStoreScope(raw string) (StoreScope, bool) {
	if raw == "a" {
		return StoreScope{All: true}, true
	}
	if !strings.HasPrefix(raw, "r:") {
		return StoreScope{}, false
	}

	raw = strings.TrimPrefix(raw, "r:")
	ids := make([]types.ID, 0)
	if raw == "" {
		return StoreScope{IDs: ids}, true
	}
	for _, item := range strings.Split(raw, ",") {
		value, err := strconv.ParseInt(strings.TrimSpace(item), 10, 64)
		if err != nil || value <= 0 {
			return StoreScope{}, false
		}
		ids = append(ids, types.ID(value))
	}

	return StoreScope{IDs: ids}, true
}

// normalizeStoreTimes 校验营业起止时间，统一成 HH:mm 两位格式。
// 入参：待校验的门店入参；返回业务错误表示格式非法。
// 规则：结束时间小于开始时间表示跨夜营业，相等表示全天营业，二者都合法。
func normalizeStoreTimes(in *StoreSaveInput) error {
	open, err := time.Parse(storeTimeLayout, in.OpenTime)
	if err != nil {
		return baseErrs.WithMsg(baseErrs.BadRequest, "营业开始时间格式应为 HH:mm，如 09:00")
	}
	closeAt, err := time.Parse(storeTimeLayout, in.CloseTime)
	if err != nil {
		return baseErrs.WithMsg(baseErrs.BadRequest, "营业结束时间格式应为 HH:mm，如 23:00")
	}

	in.OpenTime = open.Format(storeTimeLayout)
	in.CloseTime = closeAt.Format(storeTimeLayout)

	return nil
}
