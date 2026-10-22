package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
	"time"

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
	// riderDefaultMaxOrders 是新建骑手的默认接单并发上限。
	riderDefaultMaxOrders = 3
	// riderMaxOrdersLimit 是并发上限的硬上限，避免一次占用过多配送单。
	riderMaxOrdersLimit = 20
	// riderDateLayout 与 store_rider.stat_date 的存储格式一致。
	riderDateLayout = "2006-01-02"
	// riderPasswordLength 是自动开通账号时返回的一次性初始密码长度。
	riderPasswordLength = 12
)

// riderPasswordAlphabet 去掉了 0/O、1/I/l 等易混淆字符，便于骑手口述录入。
const riderPasswordAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

// riderMobilePattern 是骑手登录手机号的格式，账号即手机号。
var riderMobilePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// RiderListQuery 是骑手档案列表的查询条件。
type RiderListQuery struct {
	Keyword  string
	Status   *enums.RiderStatusEnum
	PageNo   int
	PageSize int
}

// RiderCreateInput 是新增骑手的入参。
// UserID 为 0 时按 Mobile 自动开通后台账号，Name 作为账号昵称。
type RiderCreateInput struct {
	StoreID   types.ID
	UserID    types.ID
	Name      string
	Mobile    string
	MaxOrders int
	Status    enums.RiderStatusEnum
	Remark    string
}

// RiderUpdateInput 是编辑骑手档案的入参；门店、员工与账号不允许变更。
type RiderUpdateInput struct {
	MaxOrders int
	Status    enums.RiderStatusEnum
	Remark    string
}

// RiderCreateResult 是新增骑手的返回。
// InitialPassword 只在自动开通账号时返回一次，数据库不保存明文。
type RiderCreateResult struct {
	ID              types.ID `json:"id"`
	EmployeeID      types.ID `json:"employeeId"`
	UserID          types.ID `json:"userId"`
	Username        string   `json:"username"`
	InitialPassword string   `json:"initialPassword,omitempty"`
}

// RiderListRow 是骑手列表输出：档案记录 + 门店、账号与状态展示字段。
type RiderListRow struct {
	model.StoreRiderModel
	StoreName      string                     `json:"storeName"`
	UserID         types.ID                   `json:"userId"`
	Username       string                     `json:"username"`
	Nickname       string                     `json:"nickname"`
	Mobile         string                     `json:"mobile"`
	StatusName     string                     `json:"statusName"`
	EmployeeStatus baseEnums.CommonStatusEnum `json:"employeeStatus"`
}

// RiderUserOption 是骑手弹窗选择已有后台账号时的下拉项。
type RiderUserOption struct {
	ID       types.ID `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Mobile   string   `json:"mobile"`
	Status   int8     `json:"status"`
}

// StoreRiderService 负责骑手档案、门店员工关联与账号开通。
// 接单与释放名额通过仓储层条件更新完成，不在服务层先查再改。
type StoreRiderService struct {
	riders    *repository.StoreRiderRepository
	employees *repository.StoreEmployeeRepository
	stores    *StoreService
}

// NewStoreRiderService 构造骑手档案服务。
func NewStoreRiderService() *StoreRiderService {
	return &StoreRiderService{
		riders:    repository.NewStoreRiderRepository(),
		employees: repository.NewStoreEmployeeRepository(),
		stores:    NewStoreService(),
	}
}

// Page 分页查询当前用户可见范围内的骑手档案。
// 请求带门店上下文时只查该门店；关键字命中后台账号后按员工关联过滤。
func (s *StoreRiderService) Page(ctx context.Context, q RiderListQuery) ([]RiderListRow, int64, error) {
	filter, err := s.buildFilter(ctx, q.Status)
	if err != nil {
		return nil, 0, err
	}

	keyword := strings.TrimSpace(q.Keyword)
	if keyword != "" {
		users, err := support.SearchUserSummaries(ctx, keyword, 100)
		if err != nil {
			return nil, 0, err
		}
		if len(users) == 0 {
			return []RiderListRow{}, 0, nil
		}

		userIDs := make([]types.ID, 0, len(users))
		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}
		employees, err := s.employees.ListByUserIDs(ctx, filter.StoreID, userIDs)
		if err != nil {
			return nil, 0, err
		}
		employees = s.filterEmployeesByScope(employees, filter)
		if len(employees) == 0 {
			return []RiderListRow{}, 0, nil
		}
		filter.EmployeeIDs = make([]types.ID, 0, len(employees))
		for _, employee := range employees {
			filter.EmployeeIDs = append(filter.EmployeeIDs, employee.ID)
		}
	}

	rows, total, err := s.riders.Paginate(ctx, filter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list, err := s.decorateRows(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UserOptions 搜索可绑定为骑手的启用后台账号，供新增弹窗选择已有账号。
// 入参：keyword 账号 / 昵称 / 手机号；返回最多 20 个启用账号。
func (s *StoreRiderService) UserOptions(ctx context.Context, keyword string) ([]RiderUserOption, error) {
	users, err := support.SearchUserSummaries(ctx, strings.TrimSpace(keyword), 20)
	if err != nil {
		return nil, err
	}

	list := make([]RiderUserOption, 0, len(users))
	for _, user := range users {
		list = append(list, RiderUserOption{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Mobile:   user.Mobile,
			Status:   user.Status,
		})
	}

	return list, nil
}

// Create 新增骑手档案。
// 未指定 UserID 时以手机号为账号自动开通后台账号，初始密码只在本次响应返回。
// 指定已有账号时，若该账号在当前门店已有非骑手员工记录则拒绝，避免覆盖其它岗位。
func (s *StoreRiderService) Create(ctx context.Context, in RiderCreateInput) (*RiderCreateResult, error) {
	store, err := s.stores.ResolveWritableStore(ctx, in.StoreID)
	if err != nil {
		return nil, err
	}
	in.StoreID = store.ID
	if err := normalizeRiderCreateInput(&in); err != nil {
		return nil, err
	}

	userID := in.UserID
	username := ""
	initialPassword := ""
	if userID == 0 {
		password, err := generateRiderPassword()
		if err != nil {
			return nil, err
		}
		userID, err = support.CreateAccount(ctx, support.AccountProvisionInput{
			Username: in.Mobile,
			Password: password,
			Nickname: in.Name,
			Mobile:   in.Mobile,
			Remark:   "骑手账号自动开通",
			Status:   int8(baseEnums.CommonStatusEnabled),
		})
		if err != nil {
			return nil, err
		}
		username = in.Mobile
		initialPassword = password
	} else {
		summary, err := s.findEnabledUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		username = summary.Username
	}

	employee, err := s.employees.FindByStoreUser(ctx, in.StoreID, userID)
	if err != nil {
		return nil, err
	}
	if employee != nil {
		if employee.Role != enums.StoreEmployeeRoleRider {
			return nil, baseErrs.New(storeErrs.StoreRiderEmployeeRoleInvalid)
		}

		existing, err := s.riders.FindByEmployeeID(ctx, employee.ID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, baseErrs.New(storeErrs.StoreRiderExists)
		}

		deleted, err := s.riders.FindByEmployeeIDAny(ctx, employee.ID)
		if err != nil {
			return nil, err
		}
		if deleted != nil {
			// 软删除期间员工可能被移到其它门店，恢复时以员工当前门店为准
			deleted.StoreID = employee.StoreID
			deleted.Status = in.Status
			deleted.MaxOrders = in.MaxOrders
			deleted.CurrentOrders = 0
			deleted.TodayOrders = 0
			statDate := types.Now()
			deleted.StatDate = &statDate
			deleted.Remark = in.Remark
			if err := s.riders.Restore(ctx, deleted); err != nil {
				return nil, err
			}
			if err := s.syncEmployeeStatus(ctx, employee, in.Status); err != nil {
				return nil, err
			}

			return &RiderCreateResult{
				ID:              deleted.ID,
				EmployeeID:      deleted.EmployeeID,
				UserID:          employee.UserID,
				Username:        username,
				InitialPassword: initialPassword,
			}, nil
		}
	} else {
		employee = &model.StoreEmployeeModel{
			StoreID: in.StoreID,
			UserID:  userID,
			Role:    enums.StoreEmployeeRoleRider,
			Status:  employeeStatusForRider(in.Status),
			Remark:  "",
		}
		if err := s.employees.Create(ctx, employee); err != nil {
			return nil, err
		}
		// 直接走仓储新增员工，需要手动失效可见门店缓存
		s.stores.InvalidateScope(userID)
	}

	statDate := types.Now()
	rider := &model.StoreRiderModel{
		StoreID:       in.StoreID,
		EmployeeID:    employee.ID,
		Status:        in.Status,
		MaxOrders:     in.MaxOrders,
		CurrentOrders: 0,
		TodayOrders:   0,
		StatDate:      &statDate,
		Remark:        in.Remark,
	}
	if err := s.riders.Create(ctx, rider); err != nil {
		return nil, err
	}
	if err := s.syncEmployeeStatus(ctx, employee, in.Status); err != nil {
		return nil, err
	}

	return &RiderCreateResult{
		ID:              rider.ID,
		EmployeeID:      rider.EmployeeID,
		UserID:          userID,
		Username:        username,
		InitialPassword: initialPassword,
	}, nil
}

// Update 编辑骑手档案：只允许调整并发上限、状态与备注。
// 休息 / 停用前必须没有进行中的配送单；并发上限不能低于当前负载。
// 落库用带 current_orders 条件的更新，避免校验后被并发接单插入。
func (s *StoreRiderService) Update(ctx context.Context, id types.ID, in RiderUpdateInput) error {
	rider, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.stores.ResolveWritableStore(ctx, rider.StoreID); err != nil {
		return err
	}
	if err := normalizeRiderUpdateInput(&in); err != nil {
		return err
	}
	if in.Status.BlocksOrders() && rider.CurrentOrders > 0 {
		return baseErrs.New(storeErrs.StoreRiderHasActiveOrders, rider.CurrentOrders)
	}
	if in.MaxOrders < rider.CurrentOrders {
		return baseErrs.New(storeErrs.StoreRiderMaxOrdersInvalid)
	}

	employee, err := s.employees.FindByID(ctx, rider.EmployeeID)
	if err != nil {
		return err
	}
	if employee == nil {
		return baseErrs.New(storeErrs.StoreRiderEmployeeNotExists)
	}
	if employee.StoreID != rider.StoreID {
		return baseErrs.New(storeErrs.StoreRiderStoreMismatch)
	}

	rider.Status = in.Status
	rider.MaxOrders = in.MaxOrders
	rider.Remark = in.Remark
	var affected int64
	for attempt := 0; attempt < 2; attempt++ {
		affected, err = s.riders.UpdateConfig(ctx, rider)
		if err != nil {
			return err
		}
		if affected > 0 {
			break
		}

		// 条件更新失败时区分「字段未变化」与「并发接单后不再满足约束」
		latest, err := s.riders.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if latest == nil {
			return baseErrs.New(storeErrs.StoreRiderNotExists)
		}
		if latest.CurrentOrders > 0 && in.Status.BlocksOrders() {
			return baseErrs.New(storeErrs.StoreRiderHasActiveOrders, latest.CurrentOrders)
		}
		if latest.CurrentOrders > in.MaxOrders {
			return baseErrs.New(storeErrs.StoreRiderMaxOrdersInvalid)
		}
		// 约束已满足但影响行数为 0，通常是新旧值相同；再试一次后按成功处理
	}

	return s.syncEmployeeStatus(ctx, employee, in.Status)
}

// Delete 删除骑手档案：仍有进行中配送单时拒绝；删除后同步禁用门店员工关联。
// 软删除带 current_orders=0 条件，避免校验后被并发接单插入。
func (s *StoreRiderService) Delete(ctx context.Context, id types.ID) error {
	rider, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.stores.ResolveWritableStore(ctx, rider.StoreID); err != nil {
		return err
	}
	if rider.CurrentOrders > 0 {
		return baseErrs.New(storeErrs.StoreRiderHasActiveOrders, rider.CurrentOrders)
	}

	employee, err := s.employees.FindByID(ctx, rider.EmployeeID)
	if err != nil {
		return err
	}
	if employee != nil && employee.StoreID != rider.StoreID {
		return baseErrs.New(storeErrs.StoreRiderStoreMismatch)
	}

	var affected int64
	for attempt := 0; attempt < 2; attempt++ {
		affected, err = s.riders.Delete(ctx, rider)
		if err != nil {
			return err
		}
		if affected > 0 {
			break
		}

		// 条件删除失败时区分「已被并发删除」与「校验后刚抢到新单」
		latest, err := s.riders.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if latest == nil {
			return baseErrs.New(storeErrs.StoreRiderNotExists)
		}
		if latest.CurrentOrders > 0 {
			return baseErrs.New(storeErrs.StoreRiderHasActiveOrders, latest.CurrentOrders)
		}
	}
	if affected == 0 {
		return baseErrs.New(storeErrs.StoreRiderOperationConflict)
	}
	if employee != nil {
		return s.syncEmployeeStatus(ctx, employee, enums.RiderStatusDisabled)
	}

	return nil
}

// TryAccept 以条件更新抢占一个接单名额。
// 只有「接单中」且当前负载小于并发上限才会成功；并发失败后会再查一次档案给出具体原因。
// 入参：id 骑手档案 ID；返回 nil 表示抢占成功。
func (s *StoreRiderService) TryAccept(ctx context.Context, id types.ID) error {
	affected, err := s.riders.TryAccept(ctx, id, time.Now())
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	rider, err := s.riders.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if rider == nil {
		return baseErrs.New(storeErrs.StoreRiderNotExists)
	}
	if !rider.Status.CanAccept() {
		return baseErrs.New(storeErrs.StoreRiderStatusNotAccepting)
	}
	if rider.CurrentOrders >= rider.MaxOrders {
		return baseErrs.New(storeErrs.StoreRiderLimitExceeded)
	}

	// 条件更新失败但当前数据仍满足条件，通常是并发请求刚释放或状态变化，按上限处理最安全
	return baseErrs.New(storeErrs.StoreRiderLimitExceeded)
}

// Release 释放一个接单名额；重复释放保持幂等，供配送单送达、取消、退款时调用。
func (s *StoreRiderService) Release(ctx context.Context, id types.ID) error {
	_, err := s.riders.Release(ctx, id)

	return err
}

// findByID 查询骑手档案，不存在时返回固定错误码。
func (s *StoreRiderService) findByID(ctx context.Context, id types.ID) (*model.StoreRiderModel, error) {
	rider, err := s.riders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rider == nil {
		return nil, baseErrs.New(storeErrs.StoreRiderNotExists)
	}

	return rider, nil
}

// findEnabledUser 校验指定后台账号存在且启用，返回账号摘要。
func (s *StoreRiderService) findEnabledUser(ctx context.Context, userID types.ID) (support.UserSummary, error) {
	users, err := support.FindUserSummaries(ctx, []types.ID{userID})
	if err != nil {
		return support.UserSummary{}, err
	}
	if len(users) == 0 {
		return support.UserSummary{}, baseErrs.New(storeErrs.StoreRiderUserNotExists, userID)
	}
	if users[0].Status == int8(baseEnums.CommonStatusDisabled) {
		return support.UserSummary{}, baseErrs.New(storeErrs.StoreRiderUserDisabled, users[0].Username)
	}

	return users[0], nil
}

// buildFilter 根据请求上下文与当前用户可见范围构造骑手筛选条件。
func (s *StoreRiderService) buildFilter(ctx context.Context, status *enums.RiderStatusEnum) (repository.StoreRiderFilter, error) {
	userID := types.ID(support.UserID(ctx))
	scope, err := s.stores.Scope(ctx, userID)
	if err != nil {
		return repository.StoreRiderFilter{}, err
	}

	filter := repository.StoreRiderFilter{Status: status}
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

// filterEmployeesByScope 在关键字命中账号后，再按当前用户可见门店过滤员工关联。
func (s *StoreRiderService) filterEmployeesByScope(employees []model.StoreEmployeeModel, filter repository.StoreRiderFilter) []model.StoreEmployeeModel {
	if filter.StoreID == 0 && !filter.RestrictStores {
		return employees
	}

	allowed := make(map[types.ID]bool, len(filter.StoreIDs)+1)
	if filter.StoreID > 0 {
		allowed[filter.StoreID] = true
	}
	for _, storeID := range filter.StoreIDs {
		allowed[storeID] = true
	}

	list := make([]model.StoreEmployeeModel, 0, len(employees))
	for _, employee := range employees {
		if allowed[employee.StoreID] {
			list = append(list, employee)
		}
	}

	return list
}

// syncEmployeeStatus 让门店员工的启停状态跟随骑手状态：停用骑手同时停用员工关联。
// 状态变化后失效该账号的可见门店缓存。
func (s *StoreRiderService) syncEmployeeStatus(ctx context.Context, employee *model.StoreEmployeeModel, status enums.RiderStatusEnum) error {
	target := employeeStatusForRider(status)
	if employee.Status == target {
		return nil
	}

	employee.Status = target
	if err := s.employees.Update(ctx, employee); err != nil {
		return err
	}
	s.stores.InvalidateScope(employee.UserID)

	return nil
}

// decorateRows 为骑手列表补齐门店、账号与状态展示字段。
func (s *StoreRiderService) decorateRows(ctx context.Context, rows []model.StoreRiderModel) ([]RiderListRow, error) {
	list := make([]RiderListRow, 0, len(rows))
	if len(rows) == 0 {
		return list, nil
	}

	employeeIDs := make([]types.ID, 0, len(rows))
	storeIDs := make([]types.ID, 0, len(rows))
	seenStore := make(map[types.ID]bool, len(rows))
	for _, row := range rows {
		employeeIDs = append(employeeIDs, row.EmployeeID)
		if !seenStore[row.StoreID] {
			seenStore[row.StoreID] = true
			storeIDs = append(storeIDs, row.StoreID)
		}
	}

	employees, err := s.employees.FindByIDs(ctx, employeeIDs)
	if err != nil {
		return nil, err
	}
	employeeMap := make(map[types.ID]model.StoreEmployeeModel, len(employees))
	userIDs := make([]types.ID, 0, len(employees))
	for _, employee := range employees {
		employeeMap[employee.ID] = employee
		userIDs = append(userIDs, employee.UserID)
	}

	users, err := support.FindUserSummaries(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	userMap := make(map[types.ID]support.UserSummary, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	stores, err := s.stores.FindByIDs(ctx, storeIDs)
	if err != nil {
		return nil, err
	}
	storeNames := make(map[types.ID]string, len(stores))
	for _, store := range stores {
		storeNames[store.ID] = store.Name
	}

	today := time.Now().Format(riderDateLayout)
	for _, row := range rows {
		employee := employeeMap[row.EmployeeID]
		user := userMap[employee.UserID]
		todayOrders := row.TodayOrders
		if row.StatDate == nil || row.StatDate.Time().Format(riderDateLayout) != today {
			todayOrders = 0
		}
		row.TodayOrders = todayOrders

		list = append(list, RiderListRow{
			StoreRiderModel: row,
			StoreName:       storeNames[row.StoreID],
			UserID:          employee.UserID,
			Username:        user.Username,
			Nickname:        user.Nickname,
			Mobile:          user.Mobile,
			StatusName:      row.Status.String(),
			EmployeeStatus:  employee.Status,
		})
	}

	return list, nil
}

// normalizeRiderCreateInput 清洗骑手新增入参并校验账号、状态与并发上限。
func normalizeRiderCreateInput(in *RiderCreateInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Mobile = strings.TrimSpace(in.Mobile)
	in.Remark = strings.TrimSpace(in.Remark)
	if in.Status == 0 {
		in.Status = enums.RiderStatusIdle
	}
	if in.MaxOrders == 0 {
		in.MaxOrders = riderDefaultMaxOrders
	}

	if in.StoreID <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "门店不能为空")
	}
	if !in.Status.IsValid() {
		return baseErrs.New(storeErrs.StoreRiderStatusInvalid)
	}
	if in.MaxOrders < 1 || in.MaxOrders > riderMaxOrdersLimit {
		return baseErrs.New(storeErrs.StoreRiderMaxOrdersInvalid)
	}
	if len([]rune(in.Remark)) > 255 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "备注不能超过 255 个字符")
	}
	if in.UserID == 0 {
		if in.Name == "" {
			return baseErrs.WithMsg(baseErrs.BadRequest, "骑手姓名不能为空")
		}
		if len([]rune(in.Name)) > 64 {
			return baseErrs.WithMsg(baseErrs.BadRequest, "骑手姓名不能超过 64 个字符")
		}
		if !riderMobilePattern.MatchString(in.Mobile) {
			return baseErrs.WithMsg(baseErrs.BadRequest, "骑手手机号格式不正确")
		}
	} else if in.UserID < 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "后台账号 ID 不正确")
	}

	return nil
}

// normalizeRiderUpdateInput 清洗骑手编辑入参并校验状态与并发上限。
func normalizeRiderUpdateInput(in *RiderUpdateInput) error {
	in.Remark = strings.TrimSpace(in.Remark)
	if !in.Status.IsValid() {
		return baseErrs.New(storeErrs.StoreRiderStatusInvalid)
	}
	if in.MaxOrders < 1 || in.MaxOrders > riderMaxOrdersLimit {
		return baseErrs.New(storeErrs.StoreRiderMaxOrdersInvalid)
	}
	if len([]rune(in.Remark)) > 255 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "备注不能超过 255 个字符")
	}

	return nil
}

// employeeStatusForRider 把骑手状态映射为门店员工启停状态：只有停用骑手会禁用员工。
func employeeStatusForRider(status enums.RiderStatusEnum) baseEnums.CommonStatusEnum {
	if status == enums.RiderStatusDisabled {
		return baseEnums.CommonStatusDisabled
	}

	return baseEnums.CommonStatusEnabled
}

// generateRiderPassword 生成一次性的随机初始密码，使用 crypto/rand 避免可预测。
func generateRiderPassword() (string, error) {
	max := big.NewInt(int64(len(riderPasswordAlphabet)))
	buf := make([]byte, riderPasswordLength)
	for i := range buf {
		index, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = riderPasswordAlphabet[index.Int64()]
	}

	return string(buf), nil
}
