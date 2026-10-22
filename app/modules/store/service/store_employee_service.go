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

// EmployeeListQuery 是门店员工列表的查询条件。
type EmployeeListQuery struct {
	Keyword  string
	Role     *enums.StoreEmployeeRoleEnum
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// EmployeeSaveInput 是新增 / 编辑门店员工的入参。
type EmployeeSaveInput struct {
	StoreID types.ID
	UserID  types.ID
	Role    enums.StoreEmployeeRoleEnum
	Status  baseEnums.CommonStatusEnum
	Remark  string
}

// EmployeeListRow 是员工列表输出：关联记录 + 门店与账号展示字段。
type EmployeeListRow struct {
	model.StoreEmployeeModel
	StoreName string `json:"storeName"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Mobile    string `json:"mobile"`
	RoleName  string `json:"roleName"`
}

// EmployeeUserOption 是员工分配弹窗使用的账号下拉项。
type EmployeeUserOption struct {
	ID       types.ID `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Mobile   string   `json:"mobile"`
	Status   int8     `json:"status"`
}

// StoreEmployeeService 负责门店员工分配与可见门店缓存失效。
type StoreEmployeeService struct {
	employees *repository.StoreEmployeeRepository
	riders    *repository.StoreRiderRepository
	stores    *StoreService
}

// NewStoreEmployeeService 构造门店员工服务。
func NewStoreEmployeeService() *StoreEmployeeService {
	return &StoreEmployeeService{
		employees: repository.NewStoreEmployeeRepository(),
		riders:    repository.NewStoreRiderRepository(),
		stores:    NewStoreService(),
	}
}

// Page 分页查询当前用户可见范围内的门店员工。
// 请求带门店上下文时只查该门店；否则按可见门店集合过滤，总部账号可跨店查询。
func (s *StoreEmployeeService) Page(ctx context.Context, q EmployeeListQuery) ([]EmployeeListRow, int64, error) {
	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	filter := repository.StoreEmployeeFilter{
		Role:   q.Role,
		Status: q.Status,
	}
	if contextStoreID := support.StoreID(ctx); contextStoreID > 0 {
		if !scope.All && !scope.Contains(contextStoreID) {
			return nil, 0, baseErrs.New(storeErrs.StoreAccessDenied)
		}
		filter.StoreID = contextStoreID
	} else if !scope.All {
		filter.RestrictStores = true
		filter.StoreIDs = scope.IDs
	}

	keyword := strings.TrimSpace(q.Keyword)
	if keyword != "" {
		users, err := support.SearchUserSummaries(ctx, keyword, 100)
		if err != nil {
			return nil, 0, err
		}
		if len(users) == 0 {
			return []EmployeeListRow{}, 0, nil
		}
		filter.UserIDs = make([]types.ID, 0, len(users))
		for _, user := range users {
			filter.UserIDs = append(filter.UserIDs, user.ID)
		}
	}

	rows, total, err := s.employees.Paginate(ctx, filter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list, err := s.decorateRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UserOptions 搜索可分配的后台账号，供员工弹窗远程选择。
// 入参：keyword 账号 / 昵称 / 手机号；返回最多 20 个启用账号。
func (s *StoreEmployeeService) UserOptions(ctx context.Context, keyword string) ([]EmployeeUserOption, error) {
	users, err := support.SearchUserSummaries(ctx, strings.TrimSpace(keyword), 20)
	if err != nil {
		return nil, err
	}

	list := make([]EmployeeUserOption, 0, len(users))
	for _, user := range users {
		list = append(list, EmployeeUserOption{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Mobile:   user.Mobile,
			Status:   user.Status,
		})
	}

	return list, nil
}

// Create 新增门店员工关联：校验门店层级、可见范围与账号存在性后落库。
// 返回新记录 ID；成功后立即失效该账号的可见门店缓存。
func (s *StoreEmployeeService) Create(ctx context.Context, in EmployeeSaveInput) (types.ID, error) {
	// 未传门店时按请求头门店上下文回填，前端可只提交账号与角色
	if in.StoreID == 0 {
		in.StoreID = support.StoreID(ctx)
	}
	if err := normalizeEmployeeInput(&in); err != nil {
		return 0, err
	}
	if err := s.ensureAssignment(ctx, in.StoreID); err != nil {
		return 0, err
	}
	if err := s.ensureUserExists(ctx, in.UserID); err != nil {
		return 0, err
	}

	exists, err := s.employees.ExistsByStoreUser(ctx, in.StoreID, in.UserID, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(storeErrs.StoreEmployeeExists)
	}

	m := &model.StoreEmployeeModel{
		StoreID: in.StoreID,
		UserID:  in.UserID,
		Role:    in.Role,
		Status:  in.Status,
		Remark:  in.Remark,
	}
	if err := s.employees.Create(ctx, m); err != nil {
		return 0, err
	}

	s.stores.InvalidateScope(in.UserID)

	return m.ID, nil
}

// Update 编辑门店员工：允许调整门店、账号、角色、状态与备注。
// 变更账号时同时失效新旧账号的可见门店缓存。
func (s *StoreEmployeeService) Update(ctx context.Context, id types.ID, in EmployeeSaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	// 未传门店时沿用原记录门店，避免编辑角色/状态时被误判为清空门店
	if in.StoreID == 0 {
		in.StoreID = m.StoreID
	}
	if err := normalizeEmployeeInput(&in); err != nil {
		return err
	}
	if err := s.ensureContextMatch(ctx, m.StoreID); err != nil {
		return err
	}
	if err := s.stores.EnsureStoreVisible(ctx, types.ID(support.UserID(ctx)), m.StoreID); err != nil {
		return err
	}
	if err := s.ensureContextMatch(ctx, in.StoreID); err != nil {
		return err
	}
	if err := s.ensureAssignment(ctx, in.StoreID); err != nil {
		return err
	}
	if err := s.ensureUserExists(ctx, in.UserID); err != nil {
		return err
	}

	// 骑手档案与员工一对一，门店 / 账号 / 角色 / 状态由骑手管理页统一维护，避免档案漂移
	rider, err := s.riders.FindByEmployeeID(ctx, m.ID)
	if err != nil {
		return err
	}
	if rider != nil && (in.StoreID != m.StoreID || in.UserID != m.UserID ||
		in.Role != m.Role || in.Status != m.Status) {
		return baseErrs.New(storeErrs.StoreRiderEmployeeManaged)
	}

	exists, err := s.employees.ExistsByStoreUser(ctx, in.StoreID, in.UserID, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(storeErrs.StoreEmployeeExists)
	}

	oldUserID := m.UserID
	m.StoreID = in.StoreID
	m.UserID = in.UserID
	m.Role = in.Role
	m.Status = in.Status
	m.Remark = in.Remark
	if err := s.employees.Update(ctx, m); err != nil {
		return err
	}

	s.stores.InvalidateScope(oldUserID)
	if oldUserID != in.UserID {
		s.stores.InvalidateScope(in.UserID)
	}

	return nil
}

// Delete 删除门店员工关联；关联表采用物理删除，成功后失效该账号缓存。
func (s *StoreEmployeeService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.ensureContextMatch(ctx, m.StoreID); err != nil {
		return err
	}
	if err := s.stores.EnsureStoreVisible(ctx, types.ID(support.UserID(ctx)), m.StoreID); err != nil {
		return err
	}

	// 骑手档案仍存在时禁止直接删除员工，避免产生孤立档案
	rider, err := s.riders.FindByEmployeeID(ctx, m.ID)
	if err != nil {
		return err
	}
	if rider != nil {
		return baseErrs.New(storeErrs.StoreRiderEmployeeManaged)
	}

	if err := s.employees.Delete(ctx, m); err != nil {
		return err
	}
	s.stores.InvalidateScope(m.UserID)

	return nil
}

// findByID 查询员工关联，不存在时返回固定错误码。
func (s *StoreEmployeeService) findByID(ctx context.Context, id types.ID) (*model.StoreEmployeeModel, error) {
	m, err := s.employees.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(storeErrs.StoreEmployeeNotExists)
	}

	return m, nil
}

// ensureAssignment 校验门店可分配：上下文一致、门店存在且为门店类型、在可见范围内。
func (s *StoreEmployeeService) ensureAssignment(ctx context.Context, storeID types.ID) error {
	if err := s.ensureContextMatch(ctx, storeID); err != nil {
		return err
	}

	store, err := s.stores.FindByID(ctx, storeID)
	if err != nil {
		return err
	}
	if store == nil {
		return baseErrs.New(storeErrs.StoreNotExists, storeID)
	}
	if store.Type != enums.StoreTypeShop {
		return baseErrs.New(storeErrs.StoreEmployeeStoreTypeInvalid)
	}

	return s.stores.EnsureStoreVisible(ctx, types.ID(support.UserID(ctx)), storeID)
}

// ensureContextMatch 校验请求体门店与请求头门店上下文是否一致。
func (s *StoreEmployeeService) ensureContextMatch(ctx context.Context, storeID types.ID) error {
	if contextStoreID := support.StoreID(ctx); contextStoreID > 0 && contextStoreID != storeID {
		return baseErrs.New(storeErrs.StoreContextMismatch)
	}

	return nil
}

// ensureUserExists 校验后台账号存在，避免产生孤立员工记录。
func (s *StoreEmployeeService) ensureUserExists(ctx context.Context, userID types.ID) error {
	users, err := support.FindUserSummaries(ctx, []types.ID{userID})
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return baseErrs.New(storeErrs.StoreEmployeeUserNotExists, userID)
	}

	return nil
}

// decorateRows 为员工列表补齐门店名称与账号展示字段。
// 入参：rows 员工关联记录；返回带展示字段的列表，门店或账号目录查询失败时返回错误。
func (s *StoreEmployeeService) decorateRows(ctx context.Context, rows []model.StoreEmployeeModel) ([]EmployeeListRow, error) {
	list := make([]EmployeeListRow, 0, len(rows))
	if len(rows) == 0 {
		return list, nil
	}

	storeIDs := make([]types.ID, 0, len(rows))
	userIDs := make([]types.ID, 0, len(rows))
	seenStore := make(map[types.ID]bool, len(rows))
	seenUser := make(map[types.ID]bool, len(rows))
	for _, row := range rows {
		if !seenStore[row.StoreID] {
			seenStore[row.StoreID] = true
			storeIDs = append(storeIDs, row.StoreID)
		}
		if !seenUser[row.UserID] {
			seenUser[row.UserID] = true
			userIDs = append(userIDs, row.UserID)
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

	userMap := make(map[types.ID]support.UserSummary, len(userIDs))
	users, err := support.FindUserSummaries(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		userMap[user.ID] = user
	}

	for _, row := range rows {
		user := userMap[row.UserID]
		list = append(list, EmployeeListRow{
			StoreEmployeeModel: row,
			StoreName:          storeNames[row.StoreID],
			Username:           user.Username,
			Nickname:           user.Nickname,
			Mobile:             user.Mobile,
			RoleName:           row.Role.String(),
		})
	}

	return list, nil
}

// normalizeEmployeeInput 清洗备注并校验门店、账号、角色与状态。
func normalizeEmployeeInput(in *EmployeeSaveInput) error {
	in.Remark = strings.TrimSpace(in.Remark)

	if in.StoreID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店不能为空")
	}
	if in.UserID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "后台账号不能为空")
	}
	if !in.Role.IsValid() {
		return baseErrs.New(storeErrs.StoreEmployeeRoleInvalid)
	}
	if !in.Status.IsValid() {
		return baseErrs.New(storeErrs.StoreEmployeeStatusInvalid)
	}

	return nil
}
