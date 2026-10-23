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

// SpuListQuery 是菜品列表的查询条件。
type SpuListQuery struct {
	// Keyword 菜品名称或编码关键字
	Keyword string
	// CategoryID 分类 ID，包含其下级分类；0 表示全部分类
	CategoryID types.ID
	// Status 上下架状态，nil 表示不限
	Status   *int8
	PageNo   int
	PageSize int
}

// SpuListRow 是菜品列表输出：菜品本体 + 分类名称 + 规格数量。
type SpuListRow struct {
	model.DishSpuModel
	CategoryName string `json:"categoryName"`
	SkuCount     int64  `json:"skuCount"`
}

// SpuDetail 是菜品详情输出：菜品本体 + 规格列表 + 已挂载的加料组 ID。
type SpuDetail struct {
	model.DishSpuModel
	Skus          []model.DishSkuModel `json:"skus"`
	AddonGroupIDs []types.ID           `json:"addonGroupIds"`
}

// SpuSkuInput 是菜品保存时提交的一条规格；ID 为空表示新增。
type SpuSkuInput struct {
	ID            types.ID
	Name          string
	Code          string
	StandardPrice int64
	MemberPrice   int64
	CostPrice     int64
	Status        baseEnums.CommonStatusEnum
	Sort          int
	Remark        string
}

// SpuSaveInput 是新增 / 编辑菜品的入参；规格与加料组随菜品一起整表提交。
type SpuSaveInput struct {
	CategoryID    types.ID
	Name          string
	Code          string
	Unit          string
	Description   string
	Image         string
	Tags          string
	Allergens     string
	Status        baseEnums.CommonStatusEnum
	Sort          int
	Remark        string
	Skus          []SpuSkuInput
	AddonGroupIDs []types.ID
}

// SpuService 负责菜品（SPU）与其规格、加料组关联的维护。
type SpuService struct {
	spus       *repository.DishSpuRepository
	skus       *repository.DishSkuRepository
	addons     *repository.DishAddonRepository
	categories *repository.DishCategoryRepository
	category   *CategoryService
}

// NewSpuService 构造菜品服务。
func NewSpuService() *SpuService {
	return &SpuService{
		spus:       repository.NewDishSpuRepository(),
		skus:       repository.NewDishSkuRepository(),
		addons:     repository.NewDishAddonRepository(),
		categories: repository.NewDishCategoryRepository(),
		category:   NewCategoryService(),
	}
}

// Page 分页查询菜品，按分类（含下级）与关键字、上下架状态筛选，并补齐分类名称与规格数。
func (s *SpuService) Page(ctx context.Context, q SpuListQuery) ([]SpuListRow, int64, error) {
	filter := repository.DishSpuFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Status:  q.Status,
	}
	if q.CategoryID > 0 {
		subtree, err := s.category.DescendantIDs(ctx, q.CategoryID)
		if err != nil {
			return nil, 0, err
		}
		if len(subtree) == 0 {
			subtree = map[types.ID]bool{q.CategoryID: true}
		}
		ids := make([]types.ID, 0, len(subtree))
		for id := range subtree {
			ids = append(ids, id)
		}
		filter.CategoryIDs = ids
	}

	rows, total, err := s.spus.Paginate(ctx, filter, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	list, err := s.decorate(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Detail 返回菜品详情，包含规格列表与已挂载的加料组。
func (s *SpuService) Detail(ctx context.Context, id types.ID) (*SpuDetail, error) {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}

	skus, err := s.skus.FindBySpuIDs(ctx, []types.ID{m.ID})
	if err != nil {
		return nil, err
	}
	groupIDs, err := s.addons.FindGroupIDsBySpuID(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	return &SpuDetail{
		DishSpuModel:  *m,
		Skus:          skus,
		AddonGroupIDs: groupIDs,
	}, nil
}

// Create 新增菜品：校验分类、重名、编码与规格后，在一个事务内写入菜品、规格与加料组关联。
// 入参：菜品字段 + 规格列表（至少一条）+ 加料组 ID 列表；返回新菜品 ID。
func (s *SpuService) Create(ctx context.Context, in SpuSaveInput) (types.ID, error) {
	if err := normalizeSpuInput(&in); err != nil {
		return 0, err
	}
	if _, err := s.ensureCategory(ctx, in.CategoryID); err != nil {
		return 0, err
	}
	if err := s.ensureNameAndCode(ctx, in.CategoryID, in.Name, in.Code, 0); err != nil {
		return 0, err
	}
	skuModels := make([]*model.DishSkuModel, 0, len(in.Skus))
	for _, sku := range in.Skus {
		skuModels = append(skuModels, toSkuModel(0, sku))
	}
	if err := s.ensureAddonGroups(ctx, in.AddonGroupIDs); err != nil {
		return 0, err
	}

	m := &model.DishSpuModel{
		CategoryID:  in.CategoryID,
		Name:        in.Name,
		Code:        in.Code,
		Unit:        in.Unit,
		Description: in.Description,
		Image:       in.Image,
		Tags:        in.Tags,
		Allergens:   in.Allergens,
		Status:      in.Status,
		Sort:        in.Sort,
		Remark:      in.Remark,
	}

	err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.spus.Create(ctx, tx, m); err != nil {
			return err
		}
		for _, sku := range skuModels {
			sku.SpuID = m.ID
			if err := s.skus.Create(ctx, tx, []*model.DishSkuModel{sku}); err != nil {
				return err
			}
		}

		return s.addons.SyncSpuGroups(ctx, tx, m.ID, in.AddonGroupIDs)
	})
	if err != nil {
		return 0, err
	}

	InvalidateAllMenus(ctx)

	return m.ID, nil
}

// Update 编辑菜品：规格按「带 ID 更新、无 ID 新增、未提交删除」的差量处理。
// 约束：至少保留一个规格，删除规格时同步清理其门店可售与门店价格配置。
func (s *SpuService) Update(ctx context.Context, id types.ID, in SpuSaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if err := normalizeSpuInput(&in); err != nil {
		return err
	}
	if _, err := s.ensureCategory(ctx, in.CategoryID); err != nil {
		return err
	}
	if err := s.ensureNameAndCode(ctx, in.CategoryID, in.Name, in.Code, m.ID); err != nil {
		return err
	}
	if err := s.ensureAddonGroups(ctx, in.AddonGroupIDs); err != nil {
		return err
	}

	existing, err := s.skus.FindBySpuIDs(ctx, []types.ID{m.ID})
	if err != nil {
		return err
	}
	existingMap := make(map[types.ID]*model.DishSkuModel, len(existing))
	for index := range existing {
		existingMap[existing[index].ID] = &existing[index]
	}

	creates, updates, keepIDs, err := s.diffSkus(ctx, m.ID, in.Skus, existingMap)
	if err != nil {
		return err
	}

	m.CategoryID = in.CategoryID
	m.Name = in.Name
	m.Code = in.Code
	m.Unit = in.Unit
	m.Description = in.Description
	m.Image = in.Image
	m.Tags = in.Tags
	m.Allergens = in.Allergens
	m.Status = in.Status
	m.Sort = in.Sort
	m.Remark = in.Remark

	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.spus.Update(ctx, tx, m); err != nil {
			return err
		}
		for _, sku := range creates {
			if err := s.skus.Create(ctx, tx, []*model.DishSkuModel{sku}); err != nil {
				return err
			}
		}
		for _, sku := range updates {
			if err := s.skus.Update(ctx, tx, sku); err != nil {
				return err
			}
		}

		removeIDs := make([]types.ID, 0)
		for skuID := range existingMap {
			if keepIDs[skuID] {
				continue
			}
			removeIDs = append(removeIDs, skuID)
		}
		if len(removeIDs) > 0 {
			if err := s.skus.DeleteStoreConfigs(ctx, tx, removeIDs); err != nil {
				return err
			}
			if err := s.skus.DeleteByIDs(ctx, tx, removeIDs); err != nil {
				return err
			}
		}

		return s.addons.SyncSpuGroups(ctx, tx, m.ID, in.AddonGroupIDs)
	})
	if err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// UpdateStatus 只切换菜品上下架状态，供列表开关直接调用。
func (s *SpuService) UpdateStatus(ctx context.Context, id types.ID, status baseEnums.CommonStatusEnum) error {
	if !status.IsValid() {
		return baseErrs.New(dishErrs.DishSpuStatusInvalid)
	}

	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	m.Status = status

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.spus.Update(ctx, tx, m)
	}); err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// Delete 删除菜品：级联清理规格、门店可售与价格配置、加料组关联。
func (s *SpuService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}

	skus, err := s.skus.FindBySpuIDs(ctx, []types.ID{m.ID})
	if err != nil {
		return err
	}
	skuIDs := make([]types.ID, 0, len(skus))
	for _, sku := range skus {
		skuIDs = append(skuIDs, sku.ID)
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.skus.DeleteStoreConfigs(ctx, tx, skuIDs); err != nil {
			return err
		}
		if err := s.skus.DeleteBySpuID(ctx, tx, m.ID); err != nil {
			return err
		}
		if err := s.addons.DeleteSpuAddons(ctx, tx, m.ID); err != nil {
			return err
		}

		return s.spus.Delete(ctx, tx, m.ID)
	}); err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// findByID 查询菜品，不存在时返回固定错误码。
func (s *SpuService) findByID(ctx context.Context, id types.ID) (*model.DishSpuModel, error) {
	m, err := s.spus.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(dishErrs.DishSpuNotExists)
	}

	return m, nil
}

// ensureCategory 校验菜品所属分类存在且未删除。
func (s *SpuService) ensureCategory(ctx context.Context, categoryID types.ID) (*model.DishCategoryModel, error) {
	if categoryID <= 0 {
		return nil, baseErrs.New(dishErrs.DishSpuCategoryNotExists)
	}

	category, err := s.categories.FindByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, baseErrs.New(dishErrs.DishSpuCategoryNotExists)
	}

	return category, nil
}

// ensureNameAndCode 校验同一分类下名称唯一与菜品编码唯一。
func (s *SpuService) ensureNameAndCode(ctx context.Context, categoryID types.ID, name, code string, excludeID types.ID) error {
	exists, err := s.spus.ExistsByName(ctx, categoryID, name, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(dishErrs.DishSpuNameExists, name)
	}

	codeExists, err := s.spus.ExistsByCode(ctx, code, excludeID)
	if err != nil {
		return err
	}
	if codeExists {
		return baseErrs.New(dishErrs.DishSpuCodeExists, code)
	}

	return nil
}

// ensureAddonGroups 校验菜品挂载的加料组全部存在。
func (s *SpuService) ensureAddonGroups(ctx context.Context, groupIDs []types.ID) error {
	wanted := uniqueIDs(groupIDs)
	if len(wanted) == 0 {
		return nil
	}

	groups, err := s.addons.FindGroupsByIDs(ctx, wanted)
	if err != nil {
		return err
	}
	if len(groups) != len(wanted) {
		return baseErrs.New(dishErrs.DishAddonGroupNotExists)
	}

	return nil
}

// toSkuModel 把规格入参转换为模型；spuID 为 0 时由调用方在菜品落库后回填。
func toSkuModel(spuID types.ID, in SpuSkuInput) *model.DishSkuModel {
	return &model.DishSkuModel{
		SpuID:         spuID,
		Name:          in.Name,
		Code:          in.Code,
		StandardPrice: in.StandardPrice,
		MemberPrice:   in.MemberPrice,
		CostPrice:     in.CostPrice,
		Status:        in.Status,
		Sort:          in.Sort,
		Remark:        in.Remark,
	}
}

// diffSkus 比较提交的规格与库中规格，产出新增、更新与保留集合。
// 约束：提交后至少保留一个规格；带 ID 的规格必须属于当前菜品。
func (s *SpuService) diffSkus(ctx context.Context, spuID types.ID, inputs []SpuSkuInput, existing map[types.ID]*model.DishSkuModel) ([]*model.DishSkuModel, []*model.DishSkuModel, map[types.ID]bool, error) {
	if len(inputs) == 0 {
		return nil, nil, nil, baseErrs.New(dishErrs.DishSkuRequired)
	}

	creates := make([]*model.DishSkuModel, 0, len(inputs))
	updates := make([]*model.DishSkuModel, 0, len(inputs))
	keepIDs := make(map[types.ID]bool, len(inputs))
	names := make(map[string]bool, len(inputs))
	codes := make(map[string]bool, len(inputs))

	for _, in := range inputs {
		name := strings.TrimSpace(in.Name)
		code := strings.TrimSpace(in.Code)
		if names[name] {
			return nil, nil, nil, baseErrs.New(dishErrs.DishSkuNameExists, name)
		}
		names[name] = true
		if code != "" {
			if codes[code] {
				return nil, nil, nil, baseErrs.New(dishErrs.DishSkuCodeExists, code)
			}
			codes[code] = true
		}

		if in.ID > 0 {
			m, ok := existing[in.ID]
			if !ok {
				return nil, nil, nil, baseErrs.New(dishErrs.DishSkuNotExists)
			}
			m.Name = name
			m.Code = code
			m.StandardPrice = in.StandardPrice
			m.MemberPrice = in.MemberPrice
			m.CostPrice = in.CostPrice
			m.Status = in.Status
			m.Sort = in.Sort
			m.Remark = strings.TrimSpace(in.Remark)
			updates = append(updates, m)
			keepIDs[in.ID] = true
			continue
		}

		in.Name = name
		in.Code = code
		in.Remark = strings.TrimSpace(in.Remark)
		creates = append(creates, toSkuModel(spuID, in))
	}

	if len(keepIDs)+len(creates) == 0 {
		return nil, nil, nil, baseErrs.New(dishErrs.DishSkuRequired)
	}

	return creates, updates, keepIDs, nil
}

// decorate 为菜品列表补齐分类名称与规格数量。
func (s *SpuService) decorate(ctx context.Context, rows []model.DishSpuModel) ([]SpuListRow, error) {
	list := make([]SpuListRow, 0, len(rows))
	if len(rows) == 0 {
		return list, nil
	}

	categoryIDs := make([]types.ID, 0, len(rows))
	spuIDs := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		categoryIDs = append(categoryIDs, row.CategoryID)
		spuIDs = append(spuIDs, row.ID)
	}

	categories, err := s.categories.FindByIDs(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}
	categoryNames := make(map[types.ID]string, len(categories))
	for _, category := range categories {
		categoryNames[category.ID] = category.Name
	}

	skuCounts, err := s.skus.CountBySpuIDs(ctx, spuIDs)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		list = append(list, SpuListRow{
			DishSpuModel: row,
			CategoryName: categoryNames[row.CategoryID],
			SkuCount:     skuCounts[row.ID],
		})
	}

	return list, nil
}

// normalizeSpuInput 清洗菜品入参并校验状态与价格合法性。
func normalizeSpuInput(in *SpuSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Code = strings.TrimSpace(in.Code)
	in.Unit = strings.TrimSpace(in.Unit)
	in.Description = strings.TrimSpace(in.Description)
	in.Image = strings.TrimSpace(in.Image)
	in.Tags = strings.TrimSpace(in.Tags)
	in.Allergens = strings.TrimSpace(in.Allergens)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "菜品名称不能为空")
	}
	if len([]rune(in.Name)) > 64 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "菜品名称不能超过 64 个字符")
	}
	if in.Unit == "" {
		in.Unit = "份"
	}
	if in.Sort < 0 {
		in.Sort = 0
	}
	if !in.Status.IsValid() {
		return baseErrs.New(dishErrs.DishSpuStatusInvalid)
	}

	if len(in.Skus) == 0 {
		return baseErrs.New(dishErrs.DishSkuRequired)
	}
	for index := range in.Skus {
		sku := &in.Skus[index]
		sku.Name = strings.TrimSpace(sku.Name)
		sku.Code = strings.TrimSpace(sku.Code)
		sku.Remark = strings.TrimSpace(sku.Remark)

		if sku.Name == "" {
			return baseErrs.WithMsg(baseErrs.BadRequest, "规格名称不能为空")
		}
		if len([]rune(sku.Name)) > 64 {
			return baseErrs.WithMsg(baseErrs.BadRequest, "规格名称不能超过 64 个字符")
		}
		if sku.StandardPrice <= 0 || sku.MemberPrice < 0 || sku.CostPrice < 0 {
			return baseErrs.New(dishErrs.DishSkuPriceInvalid)
		}
		if sku.Sort < 0 {
			sku.Sort = 0
		}
		if !sku.Status.IsValid() {
			return baseErrs.New(dishErrs.DishSkuStatusInvalid)
		}
	}

	return nil
}

// uniqueIDs 按传入顺序去重并丢弃非法 ID。
func uniqueIDs(ids []types.ID) []types.ID {
	list := make([]types.ID, 0, len(ids))
	seen := make(map[types.ID]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		list = append(list, id)
	}

	return list
}
