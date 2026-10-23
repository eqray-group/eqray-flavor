package service

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	dishErrs "eqray-flavor/app/modules/dish/errs"
	"eqray-flavor/app/modules/dish/model"
	"eqray-flavor/app/modules/dish/repository"
)

// AddonGroupListQuery 是加料组列表的查询条件。
type AddonGroupListQuery struct {
	Keyword  string
	PageNo   int
	PageSize int
}

// AddonGroupRow 是加料组列表输出：加料组本体 + 加料项数量 + 被菜品引用次数。
type AddonGroupRow struct {
	model.DishAddonGroupModel
	ItemCount int64 `json:"itemCount"`
	SpuCount  int64 `json:"spuCount"`
}

// AddonGroupDetail 是加料组详情输出：加料组本体 + 加料项列表。
type AddonGroupDetail struct {
	model.DishAddonGroupModel
	Items []model.DishAddonItemModel `json:"items"`
}

// AddonGroupSaveInput 是新增 / 编辑加料组的入参。
type AddonGroupSaveInput struct {
	Name      string
	Required  int8
	MinSelect int
	MaxSelect int
	Sort      int
	Status    baseEnums.CommonStatusEnum
	Remark    string
}

// AddonItemSaveInput 是新增 / 编辑加料项的入参。
type AddonItemSaveInput struct {
	Name   string
	Price  int64
	Sort   int
	Status baseEnums.CommonStatusEnum
	Remark string
}

// AddonService 负责加料组与加料项的维护。
type AddonService struct {
	addons *repository.DishAddonRepository
}

// NewAddonService 构造加料服务。
func NewAddonService() *AddonService {
	return &AddonService{addons: repository.NewDishAddonRepository()}
}

// Page 分页查询加料组，并补齐加料项数量与被菜品引用次数。
func (s *AddonService) Page(ctx context.Context, q AddonGroupListQuery) ([]AddonGroupRow, int64, error) {
	rows, total, err := s.addons.PaginateGroups(ctx, strings.TrimSpace(q.Keyword), q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]AddonGroupRow, 0, len(rows))
	if len(rows) == 0 {
		return list, total, nil
	}

	groupIDs := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		groupIDs = append(groupIDs, row.ID)
	}

	itemCounts, err := s.addons.CountItemsByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, 0, err
	}
	spuCounts, err := s.addons.CountSpuUsageByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, 0, err
	}

	for _, row := range rows {
		list = append(list, AddonGroupRow{
			DishAddonGroupModel: row,
			ItemCount:           itemCounts[row.ID],
			SpuCount:            spuCounts[row.ID],
		})
	}

	return list, total, nil
}

// Detail 返回加料组详情，包含其全部加料项。
func (s *AddonService) Detail(ctx context.Context, id types.ID) (*AddonGroupDetail, error) {
	m, err := s.findGroup(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.addons.FindItemsByGroupIDs(ctx, []types.ID{m.ID})
	if err != nil {
		return nil, err
	}

	return &AddonGroupDetail{DishAddonGroupModel: *m, Items: items}, nil
}

// Create 新增加料组：校验名称唯一与选择数量范围后落库。
func (s *AddonService) Create(ctx context.Context, in AddonGroupSaveInput) (types.ID, error) {
	if err := normalizeAddonGroupInput(&in); err != nil {
		return 0, err
	}

	exists, err := s.addons.GroupExistsByName(ctx, in.Name, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(dishErrs.DishAddonGroupExists, in.Name)
	}

	m := &model.DishAddonGroupModel{
		Name:      in.Name,
		Required:  in.Required,
		MinSelect: in.MinSelect,
		MaxSelect: in.MaxSelect,
		Sort:      in.Sort,
		Status:    in.Status,
		Remark:    in.Remark,
	}

	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.addons.CreateGroup(ctx, tx, m)
	})
	if err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑加料组：名称全局唯一，选择数量范围需合法。
func (s *AddonService) Update(ctx context.Context, id types.ID, in AddonGroupSaveInput) error {
	m, err := s.findGroup(ctx, id)
	if err != nil {
		return err
	}
	if err := normalizeAddonGroupInput(&in); err != nil {
		return err
	}

	exists, err := s.addons.GroupExistsByName(ctx, in.Name, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(dishErrs.DishAddonGroupExists, in.Name)
	}

	m.Name = in.Name
	m.Required = in.Required
	m.MinSelect = in.MinSelect
	m.MaxSelect = in.MaxSelect
	m.Sort = in.Sort
	m.Status = in.Status
	m.Remark = in.Remark

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.addons.UpdateGroup(ctx, tx, m)
	}); err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// Delete 删除加料组：已被菜品引用时拒绝，避免菜品详情出现空加料组；加料项随组一起清理。
func (s *AddonService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findGroup(ctx, id)
	if err != nil {
		return err
	}

	usage, err := s.addons.CountSpuUsage(ctx, m.ID)
	if err != nil {
		return err
	}
	if usage > 0 {
		return baseErrs.New(dishErrs.DishAddonGroupInUse, m.Name, usage)
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.addons.DeleteItemsByGroupID(ctx, tx, m.ID); err != nil {
			return err
		}

		return s.addons.DeleteGroup(ctx, tx, m.ID)
	}); err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// CreateItem 在指定加料组下新增加料项，同组内名称唯一。
func (s *AddonService) CreateItem(ctx context.Context, groupID types.ID, in AddonItemSaveInput) (types.ID, error) {
	group, err := s.findGroup(ctx, groupID)
	if err != nil {
		return 0, err
	}
	if err := normalizeAddonItemInput(&in); err != nil {
		return 0, err
	}

	exists, err := s.addons.ItemExistsByName(ctx, group.ID, in.Name, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(dishErrs.DishAddonItemExists, in.Name)
	}

	m := &model.DishAddonItemModel{
		GroupID: group.ID,
		Name:    in.Name,
		Price:   in.Price,
		Sort:    in.Sort,
		Status:  in.Status,
		Remark:  in.Remark,
	}

	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.addons.CreateItem(ctx, tx, m)
	})
	if err != nil {
		return 0, err
	}

	InvalidateAllMenus(ctx)

	return m.ID, nil
}

// UpdateItem 编辑加料项，加价不能为负，同组内名称唯一。
func (s *AddonService) UpdateItem(ctx context.Context, id types.ID, in AddonItemSaveInput) error {
	m, err := s.addons.FindItemByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(dishErrs.DishAddonItemNotExists)
	}
	if err := normalizeAddonItemInput(&in); err != nil {
		return err
	}

	exists, err := s.addons.ItemExistsByName(ctx, m.GroupID, in.Name, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(dishErrs.DishAddonItemExists, in.Name)
	}

	m.Name = in.Name
	m.Price = in.Price
	m.Sort = in.Sort
	m.Status = in.Status
	m.Remark = in.Remark

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.addons.UpdateItem(ctx, tx, m)
	}); err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// DeleteItem 删除加料项，不影响已下单的历史快照。
func (s *AddonService) DeleteItem(ctx context.Context, id types.ID) error {
	m, err := s.addons.FindItemByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(dishErrs.DishAddonItemNotExists)
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.addons.DeleteItem(ctx, tx, m.ID)
	}); err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// findGroup 查询加料组，不存在时返回固定错误码。
func (s *AddonService) findGroup(ctx context.Context, id types.ID) (*model.DishAddonGroupModel, error) {
	m, err := s.addons.FindGroupByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(dishErrs.DishAddonGroupNotExists)
	}

	return m, nil
}

// normalizeAddonGroupInput 清洗加料组入参并校验必选与数量范围。
func normalizeAddonGroupInput(in *AddonGroupSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "加料组名称不能为空")
	}
	if len([]rune(in.Name)) > 64 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "加料组名称不能超过 64 个字符")
	}
	if in.Required != 0 && in.Required != 1 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "必选标记只能是 0 或 1")
	}
	if in.Sort < 0 {
		in.Sort = 0
	}
	if in.MinSelect < 0 {
		in.MinSelect = 0
	}
	if in.MaxSelect < 0 {
		in.MaxSelect = 0
	}
	if in.Required == 1 && in.MinSelect < 1 {
		return baseErrs.New(dishErrs.DishAddonSelectRangeInvalid)
	}
	if in.MaxSelect > 0 && in.MaxSelect < in.MinSelect {
		return baseErrs.New(dishErrs.DishAddonSelectRangeInvalid)
	}
	if !in.Status.IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "加料组状态取值不合法")
	}

	return nil
}

// normalizeAddonItemInput 清洗加料项入参并校验加价不为负。
func normalizeAddonItemInput(in *AddonItemSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "加料项名称不能为空")
	}
	if len([]rune(in.Name)) > 64 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "加料项名称不能超过 64 个字符")
	}
	if in.Price < 0 {
		return baseErrs.New(dishErrs.DishAddonItemPriceInvalid)
	}
	if in.Sort < 0 {
		in.Sort = 0
	}
	if !in.Status.IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "加料项状态取值不合法")
	}

	return nil
}
