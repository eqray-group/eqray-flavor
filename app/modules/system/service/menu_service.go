package service

import (
	"context"
	"strings"

	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/system/enums"
	systemErrs "eqray-flavor/app/modules/system/errs"
	"eqray-flavor/app/modules/system/model"
	"eqray-flavor/app/modules/system/repository"
)

// MenuTreeNode 是菜单树的节点；children 为空时输出空数组而不是 null。
type MenuTreeNode struct {
	model.SystemMenuModel
	Children []*MenuTreeNode `json:"children"`
}

// MenuSaveInput 是新增 / 编辑菜单的入参，由 Controller 从请求体转换而来。
type MenuSaveInput struct {
	Name          string
	Permission    string
	Type          enums.MenuTypeEnum
	Sort          int
	ParentID      types.ID
	Path          string
	Icon          string
	Component     string
	ComponentName string
	Status        baseEnums.CommonStatusEnum
	Visible       int8
	KeepAlive     int8
	AlwaysShow    int8
}

// MenuService 负责 system_menu 的业务规则（docs 6.3）。
type MenuService struct {
	menus      *repository.MenuRepository
	permission *PermissionService
}

func NewMenuService() *MenuService {
	return &MenuService{
		menus:      repository.NewMenuRepository(),
		permission: NewPermissionService(),
	}
}

// Tree 返回完整菜单树（含按钮），菜单管理与角色授权共用。
func (s *MenuService) Tree(ctx context.Context) ([]*MenuTreeNode, error) {
	list, err := s.menus.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return buildMenuTree(list), nil
}

// RouteMeta 是动态路由的 meta 字段，字段名与前端 AppRouteRecord.meta 对齐。
type RouteMeta struct {
	Title     string          `json:"title"`
	Icon      string          `json:"icon,omitempty"`
	KeepAlive bool            `json:"keepAlive,omitempty"`
	IsHide    bool            `json:"isHide,omitempty"`
	AuthList  []RouteAuthItem `json:"authList,omitempty"`
}

// RouteAuthItem 描述页面内的按钮权限点。
type RouteAuthItem struct {
	Title    string `json:"title"`
	AuthMark string `json:"authMark"`
}

// RouteRecord 是返回给前端的动态路由节点（后端权限模式）。
type RouteRecord struct {
	Path      string         `json:"path"`
	Name      string         `json:"name"`
	Component string         `json:"component,omitempty"`
	Meta      RouteMeta      `json:"meta"`
	Children  []*RouteRecord `json:"children,omitempty"`
}

const (
	// layoutComponent 是一级目录使用的布局组件，与前端 RoutesAlias.Layout 一致。
	layoutComponent = "/index/index"
	// userCenterComponent 是个人中心页面，属于登录者自助页，不参与角色授权。
	userCenterComponent = "/system/user-center"
	// systemDirPath 是系统管理目录路径，个人中心固定挂在该目录下。
	systemDirPath = "/system"
)

// Routes 返回当前用户可见的动态路由（后端权限模式）。
// 按钮不生成路由，转为所属页面的 meta.authList；个人中心作为内置隐藏路由附加。
func (s *MenuService) Routes(ctx context.Context, userID types.ID) ([]*RouteRecord, error) {
	list, err := s.permission.Menus(ctx, userID)
	if err != nil {
		return nil, err
	}

	routes := buildRouteTree(list)
	if len(routes) == 0 {
		return routes, nil
	}

	return appendUserCenter(routes), nil
}

// buildRouteTree 按 parent_id 组装路由树，目录与菜单进路由，按钮进 authList。
func buildRouteTree(list []model.SystemMenuModel) []*RouteRecord {
	menus := make(map[types.ID]bool, len(list))
	for _, menu := range list {
		if menu.Type != enums.MenuTypeButton {
			menus[menu.ID] = true
		}
	}

	children := make(map[types.ID][]model.SystemMenuModel)
	buttons := make(map[types.ID][]model.SystemMenuModel)
	for _, menu := range list {
		if menu.Type == enums.MenuTypeButton {
			buttons[menu.ParentID] = append(buttons[menu.ParentID], menu)
			continue
		}
		children[menu.ParentID] = append(children[menu.ParentID], menu)
	}

	roots := make([]*RouteRecord, 0, len(menus))
	for _, menu := range list {
		if menu.Type == enums.MenuTypeButton || menus[menu.ParentID] {
			continue
		}
		roots = append(roots, buildRoute(menu, children, buttons))
	}

	return roots
}

// buildRoute 把单个菜单转成路由节点；目录使用布局组件，菜单使用表里的 component。
func buildRoute(menu model.SystemMenuModel, children map[types.ID][]model.SystemMenuModel, buttons map[types.ID][]model.SystemMenuModel) *RouteRecord {
	route := &RouteRecord{
		Path:      menu.Path,
		Name:      routeName(menu),
		Component: layoutComponent,
		Meta: RouteMeta{
			Title:     menu.Name,
			Icon:      menu.Icon,
			KeepAlive: menu.KeepAlive == 1,
			IsHide:    menu.Visible == 0,
			AuthList:  routeAuthList(buttons[menu.ID]),
		},
	}

	if menu.Type == enums.MenuTypeMenu {
		route.Component = menu.Component
	}

	for _, child := range children[menu.ID] {
		route.Children = append(route.Children, buildRoute(child, children, buttons))
	}

	return route
}

// routeName 优先用 component_name，未填时按 path 转成 PascalCase。
func routeName(menu model.SystemMenuModel) string {
	if menu.ComponentName != "" {
		return menu.ComponentName
	}

	parts := strings.FieldsFunc(menu.Path, func(r rune) bool {
		return r == '/' || r == '-' || r == '_'
	})
	if len(parts) == 0 {
		return menu.Name
	}

	var name strings.Builder
	for _, part := range parts {
		runes := []rune(part)
		name.WriteString(strings.ToUpper(string(runes[0])))
		name.WriteString(string(runes[1:]))
	}

	return name.String()
}

// routeAuthList 把按钮权限点转成 meta.authList，authMark 取权限标识的最后一段。
func routeAuthList(buttons []model.SystemMenuModel) []RouteAuthItem {
	if len(buttons) == 0 {
		return nil
	}

	items := make([]RouteAuthItem, 0, len(buttons))
	for _, button := range buttons {
		mark := button.Permission
		if i := strings.LastIndex(mark, ":"); i >= 0 {
			mark = mark[i+1:]
		}
		if mark == "" {
			continue
		}
		items = append(items, RouteAuthItem{Title: button.Name, AuthMark: mark})
	}

	return items
}

// appendUserCenter 把个人中心挂到系统管理目录下，作为隐藏的内置路由。
// 优先找 /system，避免「控制台」排到第一位后个人中心被挂到 /dashboard。
func appendUserCenter(routes []*RouteRecord) []*RouteRecord {
	var target *RouteRecord
	for _, route := range routes {
		if route.Component != layoutComponent {
			continue
		}
		if target == nil {
			target = route
		}
		if route.Path == systemDirPath {
			target = route
			break
		}
	}
	if target == nil {
		return routes
	}

	target.Children = append(target.Children, &RouteRecord{
		Path:      "user-center",
		Name:      "UserCenter",
		Component: userCenterComponent,
		Meta: RouteMeta{
			Title:     "个人中心",
			KeepAlive: true,
			IsHide:    true,
		},
	})

	return routes
}

// Detail 查询菜单详情，不存在时返回业务错误码。
func (s *MenuService) Detail(ctx context.Context, id types.ID) (*model.SystemMenuModel, error) {
	menu, err := s.menus.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, baseErrs.New(systemErrs.SystemMenuNotExists, id)
	}

	return menu, nil
}

// Create 新增菜单：先校验类型、上级与权限标识，再落库。
func (s *MenuService) Create(ctx context.Context, in MenuSaveInput) (types.ID, error) {
	menu, err := s.prepare(ctx, in, 0)
	if err != nil {
		return 0, err
	}

	if err := s.menus.Create(ctx, menu); err != nil {
		return 0, err
	}

	s.permission.Invalidate()

	return menu.ID, nil
}

// Update 编辑菜单：上级不能指向自己或自己的子级，避免成环。
func (s *MenuService) Update(ctx context.Context, id types.ID, in MenuSaveInput) error {
	existing, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	menu, err := s.prepare(ctx, in, existing.ID)
	if err != nil {
		return err
	}
	menu.ID = existing.ID

	if err := s.menus.Update(ctx, menu); err != nil {
		return err
	}

	s.permission.Invalidate()

	return nil
}

// Delete 软删除；存在子节点时拒绝删除，避免出现孤儿菜单。
func (s *MenuService) Delete(ctx context.Context, id types.ID) error {
	menu, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	count, err := s.menus.CountChildren(ctx, menu.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return baseErrs.New(systemErrs.SystemMenuHasChildren)
	}

	if err := s.menus.Delete(ctx, menu); err != nil {
		return err
	}

	s.permission.Invalidate()

	return nil
}

// prepare 校验入参并组装 Model；selfID > 0 表示编辑场景，需要做环检测。
func (s *MenuService) prepare(ctx context.Context, in MenuSaveInput, selfID types.ID) (*model.SystemMenuModel, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "菜单名称不能为空")
	}
	if !in.Type.IsValid() {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "菜单类型不正确")
	}
	if !in.Status.IsValid() {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "菜单状态不正确")
	}
	if err := validateSwitch("是否显示", in.Visible); err != nil {
		return nil, err
	}
	if err := validateSwitch("是否缓存", in.KeepAlive); err != nil {
		return nil, err
	}
	if err := validateSwitch("是否始终显示", in.AlwaysShow); err != nil {
		return nil, err
	}

	// 只有按钮类节点填写权限标识，目录与菜单留空（docs 6.3）
	permission := strings.TrimSpace(in.Permission)
	if in.Type == enums.MenuTypeButton {
		if permission == "" {
			return nil, baseErrs.WithMsg(baseErrs.BadRequest, "按钮必须填写权限标识")
		}
	} else {
		permission = ""
	}

	if in.ParentID > 0 {
		parent, err := s.menus.FindByID(ctx, in.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil || parent.Type == enums.MenuTypeButton {
			return nil, baseErrs.New(systemErrs.SystemMenuParentInvalid)
		}
		if selfID > 0 {
			if in.ParentID == selfID {
				return nil, baseErrs.New(systemErrs.SystemMenuParentSelf)
			}
			list, err := s.menus.ListAll(ctx)
			if err != nil {
				return nil, err
			}
			if menuIsDescendant(list, selfID, in.ParentID) {
				return nil, baseErrs.New(systemErrs.SystemMenuParentSelf)
			}
		}
	}

	return &model.SystemMenuModel{
		Name:          in.Name,
		Permission:    permission,
		Type:          in.Type,
		Sort:          in.Sort,
		ParentID:      in.ParentID,
		Path:          strings.TrimSpace(in.Path),
		Icon:          strings.TrimSpace(in.Icon),
		Component:     strings.TrimSpace(in.Component),
		ComponentName: strings.TrimSpace(in.ComponentName),
		Status:        in.Status,
		Visible:       in.Visible,
		KeepAlive:     in.KeepAlive,
		AlwaysShow:    in.AlwaysShow,
	}, nil
}

// buildMenuTree 把扁平列表按 parent_id 组装成树；父节点缺失时按顶级处理，避免丢数据。
func buildMenuTree(list []model.SystemMenuModel) []*MenuTreeNode {
	nodes := make(map[types.ID]*MenuTreeNode, len(list))
	for i := range list {
		nodes[list[i].ID] = &MenuTreeNode{
			SystemMenuModel: list[i],
			Children:        make([]*MenuTreeNode, 0),
		}
	}

	roots := make([]*MenuTreeNode, 0, len(list))
	for i := range list {
		node := nodes[list[i].ID]
		if parent, ok := nodes[list[i].ParentID]; ok && list[i].ParentID != 0 {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}

	return roots
}

// menuIsDescendant 判断 candidateID 是否位于 ancestorID 的子树中。
func menuIsDescendant(list []model.SystemMenuModel, ancestorID, candidateID types.ID) bool {
	parentOf := make(map[types.ID]types.ID, len(list))
	for i := range list {
		parentOf[list[i].ID] = list[i].ParentID
	}

	for id := candidateID; id != 0; {
		if id == ancestorID {
			return true
		}
		next, ok := parentOf[id]
		if !ok {
			return false
		}
		id = next
	}

	return false
}

// validateSwitch 校验 0 / 1 开关字段。
func validateSwitch(name string, value int8) error {
	if value != 0 && value != 1 {
		return baseErrs.WithMsg(baseErrs.BadRequest, name+"只能为 0 或 1")
	}

	return nil
}
