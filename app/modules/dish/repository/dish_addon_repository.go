package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	dishErrs "eqray-flavor/app/modules/dish/errs"
	"eqray-flavor/app/modules/dish/model"
)

const (
	dishAddonGroupTable = model.DishAddonGroupTable
	dishAddonItemTable  = model.DishAddonItemTable
	dishSpuAddonTable   = model.DishSpuAddonTable
)

// DishAddonRepository 是加料组、加料项与菜品加料关联三张表的仓储。
type DishAddonRepository struct{}

// NewDishAddonRepository 构造加料仓储。
func NewDishAddonRepository() *DishAddonRepository {
	return &DishAddonRepository{}
}

// PaginateGroups 分页查询加料组，支持名称关键字筛选。
func (r *DishAddonRepository) PaginateGroups(ctx context.Context, keyword string, pageNo, pageSize int) ([]model.DishAddonGroupModel, int64, error) {
	list := make([]model.DishAddonGroupModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(dishAddonGroupTable).WhereNull("deleted_at")
	if keyword != "" {
		query = query.WhereLike("name", "%"+keyword+"%")
	}

	if err := query.OrderBy("sort").OrderBy("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindGroupByID 按主键查询未删除的加料组；未命中返回 (nil, nil)。
func (r *DishAddonRepository) FindGroupByID(ctx context.Context, id types.ID) (*model.DishAddonGroupModel, error) {
	var m model.DishAddonGroupModel
	if err := facades.DB().WithContext(ctx).Table(dishAddonGroupTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindGroupsByIDs 批量查询未删除的加料组，供菜品保存时校验关联。
func (r *DishAddonRepository) FindGroupsByIDs(ctx context.Context, ids []types.ID) ([]model.DishAddonGroupModel, error) {
	list := make([]model.DishAddonGroupModel, 0)
	values := idValues(uniqueIDs(ids))
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishAddonGroupTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// GroupExistsByName 判断加料组名称是否已占用；excludeID > 0 时排除自身。
func (r *DishAddonRepository) GroupExistsByName(ctx context.Context, name string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(dishAddonGroupTable).
		Where("name", name).
		WhereNull("deleted_at")
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	total, err := query.Count()
	if err != nil {
		return false, err
	}

	return total > 0, nil
}

// CreateGroup 在事务内插入加料组。
func (r *DishAddonRepository) CreateGroup(ctx context.Context, tx db.Tx, m *model.DishAddonGroupModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(dishAddonGroupTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"required":   m.Required,
		"min_select": m.MinSelect,
		"max_select": m.MaxSelect,
		"sort":       m.Sort,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishAddonGroupExists, m.Name)
	}

	return err
}

// UpdateGroup 在事务内更新加料组。
func (r *DishAddonRepository) UpdateGroup(ctx context.Context, tx db.Tx, m *model.DishAddonGroupModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(dishAddonGroupTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"required":   m.Required,
			"min_select": m.MinSelect,
			"max_select": m.MaxSelect,
			"sort":       m.Sort,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishAddonGroupExists, m.Name)
	}

	return err
}

// DeleteGroup 物理删除加料组，加料项由服务层在同一事务内清理。
func (r *DishAddonRepository) DeleteGroup(ctx context.Context, tx db.Tx, id types.ID) error {
	res, err := tx.Table(dishAddonGroupTable).
		Where("id", id).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(dishErrs.DishAddonGroupNotExists)
	}

	return nil
}

// CountSpuUsage 统计加料组被多少菜品引用，删除前校验用。
func (r *DishAddonRepository) CountSpuUsage(ctx context.Context, groupID types.ID) (int64, error) {
	total, err := facades.DB().WithContext(ctx).Table(dishSpuAddonTable).
		Where("addon_group_id", groupID).
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return 0, err
	}

	return total, nil
}

// CountSpuUsageByGroupIDs 批量统计加料组被引用次数，返回加料组 ID 到次数的映射。
func (r *DishAddonRepository) CountSpuUsageByGroupIDs(ctx context.Context, groupIDs []types.ID) (map[types.ID]int64, error) {
	rows := make([]struct {
		GroupID types.ID `db:"addon_group_id"`
		Total   int64    `db:"total"`
	}, 0)

	values := idValues(uniqueIDs(groupIDs))
	if len(values) == 0 {
		return map[types.ID]int64{}, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishSpuAddonTable).
		Select("addon_group_id", "COUNT(*) AS total").
		WhereIn("addon_group_id", values).
		WhereNull("deleted_at").
		GroupBy("addon_group_id").
		Get(&rows); err != nil {
		return nil, err
	}

	counts := make(map[types.ID]int64, len(rows))
	for _, row := range rows {
		counts[row.GroupID] = row.Total
	}

	return counts, nil
}

// FindItemsByGroupIDs 批量查询加料组下的加料项，按组与排序升序。
func (r *DishAddonRepository) FindItemsByGroupIDs(ctx context.Context, groupIDs []types.ID) ([]model.DishAddonItemModel, error) {
	list := make([]model.DishAddonItemModel, 0)
	values := idValues(uniqueIDs(groupIDs))
	if len(values) == 0 {
		return list, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishAddonItemTable).
		WhereIn("group_id", values).
		WhereNull("deleted_at").
		OrderBy("group_id").
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// CountItemsByGroupIDs 批量统计各加料组的加料项数量，返回加料组 ID 到数量的映射。
func (r *DishAddonRepository) CountItemsByGroupIDs(ctx context.Context, groupIDs []types.ID) (map[types.ID]int64, error) {
	rows := make([]struct {
		GroupID types.ID `db:"group_id"`
		Total   int64    `db:"total"`
	}, 0)

	values := idValues(uniqueIDs(groupIDs))
	if len(values) == 0 {
		return map[types.ID]int64{}, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishAddonItemTable).
		Select("group_id", "COUNT(*) AS total").
		WhereIn("group_id", values).
		WhereNull("deleted_at").
		GroupBy("group_id").
		Get(&rows); err != nil {
		return nil, err
	}

	counts := make(map[types.ID]int64, len(rows))
	for _, row := range rows {
		counts[row.GroupID] = row.Total
	}

	return counts, nil
}

// FindItemsByIDs 批量查询加料项；供订单试算校验加料选择使用。
func (r *DishAddonRepository) FindItemsByIDs(ctx context.Context, ids []types.ID) ([]model.DishAddonItemModel, error) {
	list := make([]model.DishAddonItemModel, 0)
	if len(ids) == 0 {
		return list, nil
	}
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}
	if err := facades.DB().WithContext(ctx).Table(dishAddonItemTable).
		WhereIn("id", values).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// FindItemByID 按主键查询未删除的加料项；未命中返回 (nil, nil)。
func (r *DishAddonRepository) FindItemByID(ctx context.Context, id types.ID) (*model.DishAddonItemModel, error) {
	var m model.DishAddonItemModel
	if err := facades.DB().WithContext(ctx).Table(dishAddonItemTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ItemExistsByName 判断同一加料组内加料项名称是否已占用。
func (r *DishAddonRepository) ItemExistsByName(ctx context.Context, groupID types.ID, name string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(dishAddonItemTable).
		Where("group_id", groupID).
		Where("name", name).
		WhereNull("deleted_at")
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	total, err := query.Count()
	if err != nil {
		return false, err
	}

	return total > 0, nil
}

// CreateItem 在事务内插入加料项。
func (r *DishAddonRepository) CreateItem(ctx context.Context, tx db.Tx, m *model.DishAddonItemModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(dishAddonItemTable).Insert(map[string]any{
		"id":         m.ID,
		"group_id":   m.GroupID,
		"name":       m.Name,
		"price":      m.Price,
		"sort":       m.Sort,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishAddonItemExists, m.Name)
	}

	return err
}

// UpdateItem 在事务内更新加料项；所属加料组不允许修改。
func (r *DishAddonRepository) UpdateItem(ctx context.Context, tx db.Tx, m *model.DishAddonItemModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(dishAddonItemTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"price":      m.Price,
			"sort":       m.Sort,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(dishErrs.DishAddonItemExists, m.Name)
	}

	return err
}

// DeleteItem 物理删除加料项。
func (r *DishAddonRepository) DeleteItem(ctx context.Context, tx db.Tx, id types.ID) error {
	res, err := tx.Table(dishAddonItemTable).
		Where("id", id).
		WhereNull("deleted_at").
		Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(dishErrs.DishAddonItemNotExists)
	}

	return nil
}

// DeleteItemsByGroupID 在事务内删除某个加料组下的全部加料项。
func (r *DishAddonRepository) DeleteItemsByGroupID(ctx context.Context, tx db.Tx, groupID types.ID) error {
	_, err := tx.Table(dishAddonItemTable).
		Where("group_id", groupID).
		WhereNull("deleted_at").
		Delete()

	return err
}

// FindGroupIDsBySpuID 返回菜品挂载的加料组 ID，按菜品内排序升序。
func (r *DishAddonRepository) FindGroupIDsBySpuID(ctx context.Context, spuID types.ID) ([]types.ID, error) {
	rows := make([]model.DishSpuAddonModel, 0)
	if err := facades.DB().WithContext(ctx).Table(dishSpuAddonTable).
		Select("addon_group_id", "sort").
		Where("spu_id", spuID).
		WhereNull("deleted_at").
		OrderBy("sort").
		OrderBy("id").
		Get(&rows); err != nil {
		return nil, err
	}

	ids := make([]types.ID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AddonGroupID)
	}

	return ids, nil
}

// FindSpuAddonMap 批量返回菜品挂载的加料组，供菜单聚合一次取完。
// 入参：菜品 ID 列表；返回菜品 ID 到加料组 ID 列表（按菜品内排序）的映射。
func (r *DishAddonRepository) FindSpuAddonMap(ctx context.Context, spuIDs []types.ID) (map[types.ID][]types.ID, error) {
	rows := make([]model.DishSpuAddonModel, 0)
	values := idValues(uniqueIDs(spuIDs))
	if len(values) == 0 {
		return map[types.ID][]types.ID{}, nil
	}

	if err := facades.DB().WithContext(ctx).Table(dishSpuAddonTable).
		Select("spu_id", "addon_group_id", "sort").
		WhereIn("spu_id", values).
		WhereNull("deleted_at").
		OrderBy("spu_id").
		OrderBy("sort").
		OrderBy("id").
		Get(&rows); err != nil {
		return nil, err
	}

	result := make(map[types.ID][]types.ID, len(rows))
	for _, row := range rows {
		result[row.SpuID] = append(result[row.SpuID], row.AddonGroupID)
	}

	return result, nil
}

// SyncSpuGroups 在事务内把菜品的加料组对齐到 groupIDs：先清空再按顺序插入。
func (r *DishAddonRepository) SyncSpuGroups(ctx context.Context, tx db.Tx, spuID types.ID, groupIDs []types.ID) error {
	if _, err := tx.Table(dishSpuAddonTable).
		Where("spu_id", spuID).
		WhereNull("deleted_at").
		Delete(); err != nil {
		return err
	}

	for index, groupID := range uniqueIDs(groupIDs) {
		m := model.DishSpuAddonModel{
			SpuID:        spuID,
			AddonGroupID: groupID,
			Sort:         index,
		}
		m.FillCreate(ctx)

		if _, err := tx.Table(dishSpuAddonTable).Insert(map[string]any{
			"id":             m.ID,
			"spu_id":         m.SpuID,
			"addon_group_id": m.AddonGroupID,
			"sort":           m.Sort,
			"remark":         m.Remark,
			"creator":        m.Creator,
			"created_at":     m.CreatedAt,
			"updater":        m.Updater,
			"updated_at":     m.UpdatedAt,
		}); err != nil {
			return err
		}
	}

	return nil
}

// DeleteSpuAddons 在事务内删除菜品与加料组的全部关联，菜品删除时级联调用。
func (r *DishAddonRepository) DeleteSpuAddons(ctx context.Context, tx db.Tx, spuID types.ID) error {
	_, err := tx.Table(dishSpuAddonTable).
		Where("spu_id", spuID).
		WhereNull("deleted_at").
		Delete()

	return err
}
