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

// CategoryNode 是分类树节点：分类本体 + 直属菜品数 + 子节点。
type CategoryNode struct {
	model.DishCategoryModel
	// SpuCount 该分类下的直属菜品数量
	SpuCount int64 `json:"spuCount"`
	// Children 子分类，无子节点时为空数组
	Children []*CategoryNode `json:"children"`
}

// CategorySaveInput 是新增 / 编辑菜品分类的入参。
type CategorySaveInput struct {
	ParentID types.ID
	Name     string
	Image    string
	Sort     int
	Status   baseEnums.CommonStatusEnum
	Remark   string
}

// CategoryService 负责菜品分类树的查询与维护。
type CategoryService struct {
	categories *repository.DishCategoryRepository
	spus       *repository.DishSpuRepository
}

// NewCategoryService 构造菜品分类服务。
func NewCategoryService() *CategoryService {
	return &CategoryService{
		categories: repository.NewDishCategoryRepository(),
		spus:       repository.NewDishSpuRepository(),
	}
}

// Tree 返回完整分类树；孤儿子节点（上级已删）作为根节点返回，避免数据丢失。
func (s *CategoryService) Tree(ctx context.Context) ([]*CategoryNode, error) {
	rows, err := s.categories.All(ctx)
	if err != nil {
		return nil, err
	}

	categoryIDs := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		categoryIDs = append(categoryIDs, row.ID)
	}
	counts, err := s.spus.CountByCategoryIDs(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}

	nodes := make(map[types.ID]*CategoryNode, len(rows))
	for _, row := range rows {
		nodes[row.ID] = &CategoryNode{
			DishCategoryModel: row,
			SpuCount:          counts[row.ID],
			Children:          make([]*CategoryNode, 0),
		}
	}

	roots := make([]*CategoryNode, 0)
	for _, row := range rows {
		node := nodes[row.ID]
		parent, ok := nodes[row.ParentID]
		if row.ParentID == 0 || !ok {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	return roots, nil
}

// Create 新增分类：校验上级存在、同级重名后落库。
// 入参：parentId、名称、图标、排序、状态、备注；返回新分类 ID。
func (s *CategoryService) Create(ctx context.Context, in CategorySaveInput) (types.ID, error) {
	if err := normalizeCategoryInput(&in); err != nil {
		return 0, err
	}
	if err := s.ensureParent(ctx, in.ParentID); err != nil {
		return 0, err
	}

	exists, err := s.categories.ExistsByName(ctx, in.ParentID, in.Name, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(dishErrs.DishCategoryExists, in.Name)
	}

	m := &model.DishCategoryModel{
		ParentID: in.ParentID,
		Name:     in.Name,
		Image:    in.Image,
		Sort:     in.Sort,
		Status:   in.Status,
		Remark:   in.Remark,
	}
	err = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.categories.Create(ctx, tx, m)
	})
	if err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑分类：可调整上级、名称、图标、排序、状态与备注。
// 约束：上级不能是自身或自身下级，同级不能重名。
func (s *CategoryService) Update(ctx context.Context, id types.ID, in CategorySaveInput) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if err := normalizeCategoryInput(&in); err != nil {
		return err
	}
	if err := s.ensureParent(ctx, in.ParentID); err != nil {
		return err
	}
	if in.ParentID > 0 {
		subtree, err := s.DescendantIDs(ctx, m.ID)
		if err != nil {
			return err
		}
		if subtree[in.ParentID] {
			return baseErrs.New(dishErrs.DishCategoryParentCycle)
		}
	}

	exists, err := s.categories.ExistsByName(ctx, in.ParentID, in.Name, m.ID)
	if err != nil {
		return err
	}
	if exists {
		return baseErrs.New(dishErrs.DishCategoryExists, in.Name)
	}

	m.ParentID = in.ParentID
	m.Name = in.Name
	m.Image = in.Image
	m.Sort = in.Sort
	m.Status = in.Status
	m.Remark = in.Remark

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.categories.Update(ctx, tx, m)
	}); err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// Delete 删除分类：仍有子分类或下级菜品时拒绝，避免菜品失去归属。
func (s *CategoryService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}

	hasChildren, err := s.categories.HasChildren(ctx, m.ID)
	if err != nil {
		return err
	}
	if hasChildren {
		return baseErrs.New(dishErrs.DishCategoryHasChildren, m.Name)
	}

	count, err := s.spus.CountByCategoryID(ctx, m.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(dishErrs.DishCategoryHasDish, m.Name, count)
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.categories.Delete(ctx, tx, m.ID)
	}); err != nil {
		return err
	}

	InvalidateAllMenus(ctx)

	return nil
}

// findByID 查询分类，不存在时返回固定错误码。
func (s *CategoryService) findByID(ctx context.Context, id types.ID) (*model.DishCategoryModel, error) {
	m, err := s.categories.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(dishErrs.DishCategoryNotExists)
	}

	return m, nil
}

// ensureParent 校验上级分类存在且未删除；parentID 为 0 表示根分类。
func (s *CategoryService) ensureParent(ctx context.Context, parentID types.ID) error {
	if parentID <= 0 {
		return nil
	}

	parent, err := s.categories.FindByID(ctx, parentID)
	if err != nil {
		return err
	}
	if parent == nil {
		return baseErrs.New(dishErrs.DishCategoryNotExists)
	}

	return nil
}

// DescendantIDs 返回以 id 为根的整棵子树 ID 集合（含自身），用于上级成环校验与分类范围筛选。
func (s *CategoryService) DescendantIDs(ctx context.Context, id types.ID) (map[types.ID]bool, error) {
	rows, err := s.categories.All(ctx)
	if err != nil {
		return nil, err
	}

	children := make(map[types.ID][]types.ID, len(rows))
	for _, row := range rows {
		children[row.ParentID] = append(children[row.ParentID], row.ID)
	}

	result := map[types.ID]bool{id: true}
	queue := []types.ID{id}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range children[current] {
			if result[childID] {
				continue
			}
			result[childID] = true
			queue = append(queue, childID)
		}
	}

	return result, nil
}

// normalizeCategoryInput 清洗分类入参并校验状态取值。
func normalizeCategoryInput(in *CategorySaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Image = strings.TrimSpace(in.Image)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.ParentID < 0 {
		in.ParentID = 0
	}
	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "分类名称不能为空")
	}
	if len([]rune(in.Name)) > 64 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "分类名称不能超过 64 个字符")
	}
	if in.Sort < 0 {
		in.Sort = 0
	}
	if !in.Status.IsValid() {
		return baseErrs.New(dishErrs.DishCategoryStatusInvalid)
	}

	return nil
}
