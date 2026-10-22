package service

import (
	"context"
	"strings"

	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	infraErrs "eqray-flavor/app/modules/infra/errs"
	"eqray-flavor/app/modules/infra/model"
	"eqray-flavor/app/modules/infra/repository"
)

// DistrictSaveInput 是新增 / 编辑行政区划的入参。
// Pid 与 AreaCode 仅新增时生效，编辑时由服务端沿用库中值。
type DistrictSaveInput struct {
	Pid       types.ID
	AreaName  string
	ShortName string
	AreaCode  string
	Sort      int
}

// DistrictListQuery 是行政区划列表的查询条件。
type DistrictListQuery struct {
	Keyword  string
	Level    int
	Pid      *types.ID
	PageNo   int
	PageSize int
}

// DistrictTreeNode 是区划树节点：区划自身字段 + 下级节点，仅用于输出。
type DistrictTreeNode struct {
	model.InfraDistrictModel
	Children []*DistrictTreeNode `json:"children"`
}

// InfraDistrictService 负责行政区划的层级维护与唯一性校验。
type InfraDistrictService struct {
	districts *repository.InfraDistrictRepository
}

// NewInfraDistrictService 构造行政区划服务。
func NewInfraDistrictService() *InfraDistrictService {
	return &InfraDistrictService{districts: repository.NewInfraDistrictRepository()}
}

// Page 分页查询区划，支持关键字、级别与上级筛选。
func (s *InfraDistrictService) Page(ctx context.Context, q DistrictListQuery) ([]model.InfraDistrictModel, int64, error) {
	return s.districts.Paginate(ctx, repository.DistrictFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Level:   q.Level,
		Pid:     q.Pid,
	}, q.PageNo, q.PageSize)
}

// Tree 返回完整区划树，以国家级为根；父级缺失的节点按根节点兜底展示。
func (s *InfraDistrictService) Tree(ctx context.Context) ([]*DistrictTreeNode, error) {
	list, err := s.districts.All(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make(map[types.ID]*DistrictTreeNode, len(list))
	ordered := make([]*DistrictTreeNode, 0, len(list))
	for i := range list {
		node := &DistrictTreeNode{InfraDistrictModel: list[i], Children: make([]*DistrictTreeNode, 0)}
		nodes[node.ID] = node
		ordered = append(ordered, node)
	}

	roots := make([]*DistrictTreeNode, 0, 1)
	for _, node := range ordered {
		parent, ok := nodes[node.Pid]
		if node.Pid == 0 || !ok {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	return roots, nil
}

// Children 返回指定上级的直接下级，pid 传 0 表示顶级；供级联选择懒加载使用。
func (s *InfraDistrictService) Children(ctx context.Context, pid types.ID) ([]model.InfraDistrictModel, error) {
	return s.districts.Children(ctx, pid)
}

// Detail 查询区划详情，不存在时返回业务错误码。
func (s *InfraDistrictService) Detail(ctx context.Context, id types.ID) (*model.InfraDistrictModel, error) {
	m, err := s.districts.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(infraErrs.InfraDistrictNotExists, id)
	}

	return m, nil
}

// Create 新增区划：按上级推导层级与父级编码，并校验行政编码唯一。
// 返回新区划 ID。
func (s *InfraDistrictService) Create(ctx context.Context, in DistrictSaveInput) (types.ID, error) {
	if err := normalizeDistrictInput(&in); err != nil {
		return 0, err
	}

	level, parentCode, err := s.resolveParent(ctx, in.Pid)
	if err != nil {
		return 0, err
	}

	exists, err := s.districts.ExistsByCode(ctx, in.AreaCode, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(infraErrs.InfraDistrictCodeExists, in.AreaCode)
	}

	m := &model.InfraDistrictModel{
		Pid:            in.Pid,
		AreaName:       in.AreaName,
		ShortName:      in.ShortName,
		AreaCode:       in.AreaCode,
		ParentAreaCode: parentCode,
		Level:          level,
		Sort:           in.Sort,
	}
	if err := s.districts.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑区划：只允许改名称、简称与排序。
// 上级与行政编码是层级标识，改动会让子级编码失真，因此编辑时忽略入参。
func (s *InfraDistrictService) Update(ctx context.Context, id types.ID, in DistrictSaveInput) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	if err := normalizeDistrictInput(&in); err != nil {
		return err
	}

	m.AreaName = in.AreaName
	m.ShortName = in.ShortName
	m.Sort = in.Sort

	return s.districts.Update(ctx, m)
}

// Delete 删除区划：仍有下级时拒绝，避免层级断裂。
func (s *InfraDistrictService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	count, err := s.districts.CountChildren(ctx, m.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(infraErrs.InfraDistrictHasChildren, m.AreaName)
	}

	return s.districts.Delete(ctx, m)
}

// resolveParent 按上级推导行政级别与父级编码。
// 入参：pid 上级 ID（0 表示顶级）；返回层级、父级编码，上级不存在时返回业务错误码。
func (s *InfraDistrictService) resolveParent(ctx context.Context, pid types.ID) (int, string, error) {
	if pid == 0 {
		return 1, "", nil
	}

	parent, err := s.districts.FindByID(ctx, pid)
	if err != nil {
		return 0, "", err
	}
	if parent == nil {
		return 0, "", baseErrs.New(infraErrs.InfraDistrictParentMissing, pid)
	}

	return parent.Level + 1, parent.AreaCode, nil
}

// normalizeDistrictInput 清洗字符串并校验必填项与排序下限。
func normalizeDistrictInput(in *DistrictSaveInput) error {
	in.AreaName = strings.TrimSpace(in.AreaName)
	in.ShortName = strings.TrimSpace(in.ShortName)
	in.AreaCode = strings.TrimSpace(in.AreaCode)

	if in.AreaName == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "区划名称不能为空")
	}
	if in.AreaCode == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "行政编码不能为空")
	}
	// 简称留空时用全称兜底，保证列表与下拉都有可读文案
	if in.ShortName == "" {
		in.ShortName = in.AreaName
	}
	if in.Sort < 0 {
		in.Sort = 0
	}

	return nil
}
