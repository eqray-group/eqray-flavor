// Package repository 定义 promotion 模块的数据访问层。
// 券状态流转与活动名额占用一律走「带原状态 / 带原计数」的条件更新，防并发重复使用与超额占用。
package repository

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/database/db"
	frameworkDB "github.com/goravel/framework/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/promotion/enums"
	"eqray-flavor/app/modules/promotion/model"
)

// toAny 把 ID 列表转为查询构造器需要的 []any。
func toAny(ids []types.ID) []any {
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}

	return values
}

// CouponTemplateRepository 封装优惠券模板的读写。
type CouponTemplateRepository struct{}

// NewCouponTemplateRepository 构造券模板仓储。
func NewCouponTemplateRepository() *CouponTemplateRepository {
	return &CouponTemplateRepository{}
}

// Create 写入一张券模板。
func (r *CouponTemplateRepository) Create(ctx context.Context, m *model.CouponTemplateModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(model.CouponTemplateTable).Insert(map[string]any{
		"id":                  m.ID,
		"name":                m.Name,
		"type":                int8(m.Type),
		"description":         m.Description,
		"threshold_amount":    m.ThresholdAmount,
		"discount_amount":     m.DiscountAmount,
		"discount_rate":       m.DiscountRate,
		"max_discount_amount": m.MaxDiscountAmount,
		"scope":               int8(m.Scope),
		"scope_ids":           m.ScopeIDs,
		"channel_limit":       int8(m.ChannelLimit),
		"issue_type":          int8(m.IssueType),
		"total_count":         m.TotalCount,
		"issued_count":        m.IssuedCount,
		"per_member_limit":    m.PerMemberLimit,
		"start_time":          m.StartTime,
		"end_time":            m.EndTime,
		"valid_days":          m.ValidDays,
		"status":              int8(m.Status),
		"creator":             m.Creator,
		"created_at":          m.CreatedAt,
		"updater":             m.Updater,
		"updated_at":          m.UpdatedAt,
	})

	return err
}

// Update 更新券模板的规则字段，不含发行统计。
func (r *CouponTemplateRepository) Update(ctx context.Context, m *model.CouponTemplateModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(model.CouponTemplateTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":                m.Name,
			"type":                int8(m.Type),
			"description":         m.Description,
			"threshold_amount":    m.ThresholdAmount,
			"discount_amount":     m.DiscountAmount,
			"discount_rate":       m.DiscountRate,
			"max_discount_amount": m.MaxDiscountAmount,
			"scope":               int8(m.Scope),
			"scope_ids":           m.ScopeIDs,
			"channel_limit":       int8(m.ChannelLimit),
			"issue_type":          int8(m.IssueType),
			"total_count":         m.TotalCount,
			"per_member_limit":    m.PerMemberLimit,
			"start_time":          m.StartTime,
			"end_time":            m.EndTime,
			"valid_days":          m.ValidDays,
			"status":              int8(m.Status),
			"updater":             m.Updater,
			"updated_at":          m.UpdatedAt,
		})

	return err
}

// Delete 物理删除券模板。
func (r *CouponTemplateRepository) Delete(ctx context.Context, id types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(model.CouponTemplateTable).
		Where("id", id).
		WhereNull("deleted_at").
		Delete()

	return err
}

// find 按查询取一条券模板，未命中返回 (nil, nil)。
func (r *CouponTemplateRepository) find(ctx context.Context, query db.Query) (*model.CouponTemplateModel, error) {
	var m model.CouponTemplateModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByID 按主键查询券模板。
func (r *CouponTemplateRepository) FindByID(ctx context.Context, id types.ID) (*model.CouponTemplateModel, error) {
	query := facades.DB().WithContext(ctx).Table(model.CouponTemplateTable).Where("id", id)

	return r.find(ctx, query)
}

// FindByIDs 批量查询券模板。
func (r *CouponTemplateRepository) FindByIDs(ctx context.Context, ids []types.ID) ([]model.CouponTemplateModel, error) {
	list := make([]model.CouponTemplateModel, 0, len(ids))
	if len(ids) == 0 {
		return list, nil
	}
	if err := facades.DB().WithContext(ctx).Table(model.CouponTemplateTable).
		WhereIn("id", toAny(ids)).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// TemplateListFilter 是券模板列表筛选条件。
type TemplateListFilter struct {
	// Name 券名称模糊匹配
	Name string
	// Type 券类型，0 表示不限
	Type enums.CouponType
	// Status 启用状态，nil 表示不限
	Status *int8
}

// Page 分页查询券模板，返回列表与总数。
func (r *CouponTemplateRepository) Page(ctx context.Context, filter TemplateListFilter, pageNo, pageSize int) ([]model.CouponTemplateModel, int64, error) {
	query := facades.DB().WithContext(ctx).Table(model.CouponTemplateTable).WhereNull("deleted_at")
	if filter.Name != "" {
		query = query.WhereLike("name", "%"+filter.Name+"%")
	}
	if filter.Type > 0 {
		query = query.Where("type", int8(filter.Type))
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	list := make([]model.CouponTemplateModel, 0)
	var total int64
	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// IncrementIssuedCount 原子累加已发放数量；限量模板按「已发放 + 本次 <= 发行总量」条件更新，防超发。
// 入参：事务、模板 ID、发放数量；返回是否成功（false 表示券已发完）。
func (r *CouponTemplateRepository) IncrementIssuedCount(ctx context.Context, tx db.Tx, id types.ID, count int) (bool, error) {
	// 先读发行总量：0 表示不限量，走无上限更新；限量模板才加条件
	row := struct {
		TotalCount int `db:"total_count"`
	}{}
	if err := tx.Table(model.CouponTemplateTable).
		Select("total_count").
		Where("id", id).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return false, err
	}

	query := tx.Table(model.CouponTemplateTable).
		Where("id", id).
		WhereNull("deleted_at")

	if row.TotalCount > 0 {
		query = query.Where("issued_count + ? <= total_count", count)
	}

	res, err := query.Update(map[string]any{
		"issued_count": frameworkDB.Raw("issued_count + ?", count),
		"updated_at":   types.Now(),
	})
	if err != nil {
		return false, err
	}

	return res.RowsAffected > 0, nil
}

// CouponRepository 封装会员券的读写。
type CouponRepository struct{}

// NewCouponRepository 构造会员券仓储。
func NewCouponRepository() *CouponRepository {
	return &CouponRepository{}
}

// Create 写入一张会员券。
func (r *CouponRepository) Create(ctx context.Context, tx db.Tx, m *model.CouponModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(model.CouponTable).Insert(map[string]any{
		"id":          m.ID,
		"template_id": m.TemplateID,
		"member_id":   m.MemberID,
		"status":      int8(m.Status),
		"order_id":    m.OrderID,
		"use_time":    m.UseTime,
		"expire_time": m.ExpireTime,
		"get_time":    m.GetTime,
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})

	return err
}

// find 按查询取一张券，未命中返回 (nil, nil)。
func (r *CouponRepository) find(ctx context.Context, query db.Query) (*model.CouponModel, error) {
	var m model.CouponModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByID 按主键查询会员券。
func (r *CouponRepository) FindByID(ctx context.Context, id types.ID) (*model.CouponModel, error) {
	query := facades.DB().WithContext(ctx).Table(model.CouponTable).Where("id", id)

	return r.find(ctx, query)
}

// FindUsable 查询会员可用的券：状态未使用且未过期，按过期时间升序，便于先用快过期的。
func (r *CouponRepository) FindUsable(ctx context.Context, memberID types.ID, now types.DateTime) ([]model.CouponModel, error) {
	list := make([]model.CouponModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.CouponTable).
		Where("member_id", memberID).
		Where("status", int8(enums.CouponStatusUnused)).
		Where("expire_time >= ?", now).
		WhereNull("deleted_at").
		OrderBy("expire_time").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// CountByTemplateAndMember 统计会员持有某模板的券数量，用于每人限领校验。
func (r *CouponRepository) CountByTemplateAndMember(ctx context.Context, templateID, memberID types.ID) (int, error) {
	row := struct {
		Total int `db:"total"`
	}{}
	if err := facades.DB().WithContext(ctx).Table(model.CouponTable).
		Select("COUNT(*) AS total").
		Where("template_id", templateID).
		Where("member_id", memberID).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, err
	}

	return row.Total, nil
}

// Transit 把券从原状态流转到新状态，带原状态条件更新防并发；返回是否流转成功。
// 入参：事务、券 ID、原状态、新状态与附加字段。
func (r *CouponRepository) Transit(ctx context.Context, tx db.Tx, id types.ID, from, to enums.CouponStatus, extra map[string]any) (bool, error) {
	updates := map[string]any{
		"status":     int8(to),
		"updated_at": types.Now(),
	}
	for key, value := range extra {
		updates[key] = value
	}

	res, err := tx.Table(model.CouponTable).
		Where("id", id).
		Where("status", int8(from)).
		WhereNull("deleted_at").
		Update(updates)
	if err != nil {
		return false, err
	}

	return res.RowsAffected > 0, nil
}

// TransitOutside 把券从原状态流转到新状态，不开事务。
// 用于单条状态变更（如后台作废），同样带原状态条件更新防并发。
func (r *CouponRepository) TransitOutside(ctx context.Context, id types.ID, from, to enums.CouponStatus, extra map[string]any) (bool, error) {
	updates := map[string]any{
		"status":     int8(to),
		"updated_at": types.Now(),
	}
	for key, value := range extra {
		updates[key] = value
	}

	res, err := facades.DB().WithContext(ctx).Table(model.CouponTable).
		Where("id", id).
		Where("status", int8(from)).
		WhereNull("deleted_at").
		Update(updates)
	if err != nil {
		return false, err
	}

	return res.RowsAffected > 0, nil
}

// CouponListFilter 是会员券列表筛选条件。
type CouponListFilter struct {
	// TemplateID 券模板，0 表示不限
	TemplateID types.ID
	// MemberID 会员，0 表示不限
	MemberID types.ID
	// Status 券状态，nil 表示不限
	Status *enums.CouponStatus
}

// Page 分页查询会员券，返回列表与总数。
func (r *CouponRepository) Page(ctx context.Context, filter CouponListFilter, pageNo, pageSize int) ([]model.CouponModel, int64, error) {
	query := facades.DB().WithContext(ctx).Table(model.CouponTable).WhereNull("deleted_at")
	if filter.TemplateID > 0 {
		query = query.Where("template_id", filter.TemplateID)
	}
	if filter.MemberID > 0 {
		query = query.Where("member_id", filter.MemberID)
	}
	if filter.Status != nil {
		query = query.Where("status", int8(*filter.Status))
	}

	list := make([]model.CouponModel, 0)
	var total int64
	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ListByOrderID 查询某订单占用的全部券，用于订单取消 / 退款时批量释放或返还。
func (r *CouponRepository) ListByOrderID(ctx context.Context, orderID types.ID) ([]model.CouponModel, error) {
	list := make([]model.CouponModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.CouponTable).
		Where("order_id", orderID).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ActivityRepository 封装营销活动与参与商品的读写。
type ActivityRepository struct{}

// NewActivityRepository 构造活动仓储。
func NewActivityRepository() *ActivityRepository {
	return &ActivityRepository{}
}

// Create 写入一条活动。
func (r *ActivityRepository) Create(ctx context.Context, m *model.ActivityModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(model.ActivityTable).Insert(map[string]any{
		"id":          m.ID,
		"name":        m.Name,
		"type":        int8(m.Type),
		"level":       int8(m.Level),
		"description": m.Description,
		"rule":        m.Rule,
		"scope":       int8(m.Scope),
		"scope_ids":   m.ScopeIDs,
		"priority":    m.Priority,
		"mutex_group": m.MutexGroup,
		"exclusive":   m.Exclusive,
		"stackable":   m.Stackable,
		"quota_limit": m.QuotaLimit,
		"quota_cycle": int8(m.QuotaCycle),
		"start_time":  m.StartTime,
		"end_time":    m.EndTime,
		"status":      int8(m.Status),
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})

	return err
}

// Update 更新活动字段。
func (r *ActivityRepository) Update(ctx context.Context, m *model.ActivityModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(model.ActivityTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":        m.Name,
			"type":        int8(m.Type),
			"level":       int8(m.Level),
			"description": m.Description,
			"rule":        m.Rule,
			"scope":       int8(m.Scope),
			"scope_ids":   m.ScopeIDs,
			"priority":    m.Priority,
			"mutex_group": m.MutexGroup,
			"exclusive":   m.Exclusive,
			"stackable":   m.Stackable,
			"quota_limit": m.QuotaLimit,
			"quota_cycle": int8(m.QuotaCycle),
			"start_time":  m.StartTime,
			"end_time":    m.EndTime,
			"status":      int8(m.Status),
			"updater":     m.Updater,
			"updated_at":  m.UpdatedAt,
		})

	return err
}

// Delete 物理删除活动。
func (r *ActivityRepository) Delete(ctx context.Context, id types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(model.ActivityTable).
		Where("id", id).
		WhereNull("deleted_at").
		Delete()

	return err
}

// find 按查询取一条活动，未命中返回 (nil, nil)。
func (r *ActivityRepository) find(ctx context.Context, query db.Query) (*model.ActivityModel, error) {
	var m model.ActivityModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByID 按主键查询活动。
func (r *ActivityRepository) FindByID(ctx context.Context, id types.ID) (*model.ActivityModel, error) {
	query := facades.DB().WithContext(ctx).Table(model.ActivityTable).Where("id", id)

	return r.find(ctx, query)
}

// ListEnabled 查询指定时间点处于有效期内的已启用活动，按优先级降序返回。
// 入参：门店 ID、当前时间；返回候选活动列表。
func (r *ActivityRepository) ListEnabled(ctx context.Context, storeID types.ID, now types.DateTime) ([]model.ActivityModel, error) {
	list := make([]model.ActivityModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.ActivityTable).
		Where("status", int8(enums.ActivityStatusEnabled)).
		Where("start_time <= ?", now).
		Where("end_time >= ?", now).
		WhereNull("deleted_at").
		OrderByDesc("priority").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	// 指定门店范围的活动要按门店过滤；全品牌活动直接保留
	filtered := make([]model.ActivityModel, 0, len(list))
	for i := range list {
		activity := &list[i]
		if activity.Scope == enums.ActivityScopeAll || containsScopeIDs(activity.ScopeIDs, storeID) {
			filtered = append(filtered, *activity)
		}
	}

	return filtered, nil
}

// ActivityListFilter 是活动列表筛选条件。
type ActivityListFilter struct {
	// Name 活动名称模糊匹配
	Name string
	// Type 活动玩法，0 表示不限
	Type enums.ActivityType
	// Status 启用状态，nil 表示不限
	Status *enums.ActivityStatus
}

// Page 分页查询活动，返回列表与总数。
func (r *ActivityRepository) Page(ctx context.Context, filter ActivityListFilter, pageNo, pageSize int) ([]model.ActivityModel, int64, error) {
	query := facades.DB().WithContext(ctx).Table(model.ActivityTable).WhereNull("deleted_at")
	if filter.Name != "" {
		query = query.WhereLike("name", "%"+filter.Name+"%")
	}
	if filter.Type > 0 {
		query = query.Where("type", int8(filter.Type))
	}
	if filter.Status != nil {
		query = query.Where("status", int8(*filter.Status))
	}

	list := make([]model.ActivityModel, 0)
	var total int64
	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ReplaceProducts 重设活动参与商品：先删旧关联再写新关联。
func (r *ActivityRepository) ReplaceProducts(ctx context.Context, activityID types.ID, spuIDs []types.ID) error {
	if _, err := facades.DB().WithContext(ctx).Table(model.ActivityProductTable).
		Where("activity_id", activityID).
		Delete(); err != nil {
		return err
	}
	if len(spuIDs) == 0 {
		return nil
	}

	for _, spuID := range spuIDs {
		item := &model.ActivityProductModel{ActivityID: activityID, SpuID: spuID}
		item.FillCreate(ctx)
		if _, err := facades.DB().WithContext(ctx).Table(model.ActivityProductTable).Insert(map[string]any{
			"id":          item.ID,
			"activity_id": item.ActivityID,
			"spu_id":      item.SpuID,
			"creator":     item.Creator,
			"created_at":  item.CreatedAt,
			"updater":     item.Updater,
			"updated_at":  item.UpdatedAt,
		}); err != nil {
			return err
		}
	}

	return nil
}

// ListProductSPUIDs 查询活动参与商品的 SPU ID 列表。
func (r *ActivityRepository) ListProductSPUIDs(ctx context.Context, activityID types.ID) ([]types.ID, error) {
	list := make([]model.ActivityProductModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.ActivityProductTable).
		Where("activity_id", activityID).
		WhereNull("deleted_at").
		Get(&list); err != nil {
		return nil, err
	}

	ids := make([]types.ID, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.SpuID)
	}

	return ids, nil
}

// ListProductSPUIDsByActivityIDs 批量查询多个活动的参与商品，按活动 ID 分组返回。
func (r *ActivityRepository) ListProductSPUIDsByActivityIDs(ctx context.Context, activityIDs []types.ID) (map[types.ID][]types.ID, error) {
	items := make([]model.ActivityProductModel, 0)
	if len(activityIDs) == 0 {
		return map[types.ID][]types.ID{}, nil
	}
	if err := facades.DB().WithContext(ctx).Table(model.ActivityProductTable).
		WhereIn("activity_id", toAny(activityIDs)).
		WhereNull("deleted_at").
		Get(&items); err != nil {
		return nil, err
	}

	grouped := make(map[types.ID][]types.ID, len(activityIDs))
	for _, item := range items {
		grouped[item.ActivityID] = append(grouped[item.ActivityID], item.SpuID)
	}

	return grouped, nil
}

// ActivityQuotaRepository 封装活动名额计数的读写。
type ActivityQuotaRepository struct{}

// NewActivityQuotaRepository 构造名额仓储。
func NewActivityQuotaRepository() *ActivityQuotaRepository {
	return &ActivityQuotaRepository{}
}

// Occupy 占用一个名额：带「已用数 < 上限」条件更新，保证并发下不会超发。
// 入参：事务、活动 ID、周期键、名额上限（必须大于 0）；返回是否占用成功。
// 说明：先走条件更新命中已有计数行；未命中说明计数行不存在，插入初始行，
// 唯一键冲突时由调用方按「名额已被抢占」处理。
func (r *ActivityQuotaRepository) Occupy(ctx context.Context, tx db.Tx, activityID types.ID, periodKey string, limit int) (bool, error) {
	if limit <= 0 {
		return false, nil
	}

	// 条件更新：已用数 < 上限才加一，并发下不会超发
	res, err := tx.Table(model.ActivityQuotaTable).
		Where("activity_id", activityID).
		Where("period_key", periodKey).
		Where("used_count < ?", limit).
		WhereNull("deleted_at").
		Update(map[string]any{
			"used_count": frameworkDB.Raw("used_count + 1"),
			"updated_at": types.Now(),
		})
	if err != nil {
		return false, err
	}
	if res.RowsAffected > 0 {
		return true, nil
	}

	// 未命中：可能是计数行不存在，也可能是名额已满。
	// 读一次当前计数区分两种情况：行不存在才插入，已满直接返回失败。
	// 读必须走事务：事务外的连接看不到本事务刚插入的计数行。
	used, countErr := r.GetUsedCountTx(ctx, tx, activityID, periodKey)
	if countErr != nil {
		return false, countErr
	}
	if used > 0 {
		// 计数行存在但名额已满
		return false, nil
	}

	// 计数行不存在：插入初始行。唯一键兜底并发插入。
	quota := &model.ActivityQuotaModel{ActivityID: activityID, PeriodKey: periodKey}
	quota.FillCreate(ctx)
	if _, err := tx.Table(model.ActivityQuotaTable).Insert(map[string]any{
		"id":          quota.ID,
		"activity_id": quota.ActivityID,
		"period_key":  quota.PeriodKey,
		"used_count":  1,
		"creator":     quota.Creator,
		"created_at":  quota.CreatedAt,
		"updater":     quota.Updater,
		"updated_at":  quota.UpdatedAt,
	}); err != nil {
		if !support.IsDuplicate(err) {
			return false, err
		}
		// 并发下别人先插入了计数行：回到条件更新重试一次，
		// 由「已用数 < 上限」决定谁真正占到这个名额。
		res, retryErr := tx.Table(model.ActivityQuotaTable).
			Where("activity_id", activityID).
			Where("period_key", periodKey).
			Where("used_count < ?", limit).
			WhereNull("deleted_at").
			Update(map[string]any{
				"used_count": frameworkDB.Raw("used_count + 1"),
				"updated_at": types.Now(),
			})
		if retryErr != nil {
			return false, retryErr
		}

		return res.RowsAffected > 0, nil
	}

	return true, nil
}

// Release 释放一个名额：取消 / 超时关单 / 整单退款时回退已用数。
// 带「已用数 > 0」条件更新，避免重复释放把计数扣成负数。
func (r *ActivityQuotaRepository) Release(ctx context.Context, tx db.Tx, activityID types.ID, periodKey string) error {
	_, err := tx.Table(model.ActivityQuotaTable).
		Where("activity_id", activityID).
		Where("period_key", periodKey).
		Where("used_count > 0").
		WhereNull("deleted_at").
		Update(map[string]any{
			"used_count": frameworkDB.Raw("used_count - 1"),
			"updated_at": types.Now(),
		})

	return err
}

// GetUsedCount 读取当前周期已占用名额数。
func (r *ActivityQuotaRepository) GetUsedCount(ctx context.Context, activityID types.ID, periodKey string) (int, error) {
	return r.countUsed(ctx, facades.DB().WithContext(ctx).Table(model.ActivityQuotaTable), activityID, periodKey)
}

// GetUsedCountTx 在事务内读取当前周期已占用名额数。
// 事务内的读必须走事务句柄：事务外的连接看不到本事务尚未提交的写入。
func (r *ActivityQuotaRepository) GetUsedCountTx(ctx context.Context, tx db.Tx, activityID types.ID, periodKey string) (int, error) {
	return r.countUsed(ctx, tx.Table(model.ActivityQuotaTable), activityID, periodKey)
}

// countUsed 按活动与周期键读取已占用名额数。
func (r *ActivityQuotaRepository) countUsed(ctx context.Context, query db.Query, activityID types.ID, periodKey string) (int, error) {
	row := struct {
		UsedCount int `db:"used_count"`
	}{}
	if err := query.
		Select("used_count").
		Where("activity_id", activityID).
		Where("period_key", periodKey).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, err
	}

	return row.UsedCount, nil
}

// OrderActivityRepository 封装订单占用活动名额关联的读写。
type OrderActivityRepository struct{}

// NewOrderActivityRepository 构造订单活动关联仓储。
func NewOrderActivityRepository() *OrderActivityRepository {
	return &OrderActivityRepository{}
}

// Create 写入一条订单占用活动名额的关联记录。
func (r *OrderActivityRepository) Create(ctx context.Context, tx db.Tx, m *model.OrderActivityModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(model.OrderActivityTable).Insert(map[string]any{
		"id":          m.ID,
		"order_id":    m.OrderID,
		"activity_id": m.ActivityID,
		"period_key":  m.PeriodKey,
		"status":      int8(m.Status),
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})

	return err
}

// ListByOrderID 查询某订单占用的全部活动名额关联记录（含已释放的）。
func (r *OrderActivityRepository) ListByOrderID(ctx context.Context, tx db.Tx, orderID types.ID) ([]model.OrderActivityModel, error) {
	list := make([]model.OrderActivityModel, 0)
	if err := tx.Table(model.OrderActivityTable).
		Where("order_id", orderID).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Release 把关联记录从「占用中」流转到「已释放」，带原状态条件更新防并发重复释放。
// 返回是否流转成功；false 表示已被其它请求释放过。
func (r *OrderActivityRepository) Release(ctx context.Context, tx db.Tx, id types.ID) (bool, error) {
	res, err := tx.Table(model.OrderActivityTable).
		Where("id", id).
		Where("status", int8(enums.OrderActivityStatusOccupied)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":     int8(enums.OrderActivityStatusReleased),
			"updated_at": types.Now(),
		})
	if err != nil {
		return false, err
	}

	return res.RowsAffected > 0, nil
}

// PointRuleRepository 封装积分抵扣规则的读写。
type PointRuleRepository struct{}

// NewPointRuleRepository 构造积分规则仓储。
func NewPointRuleRepository() *PointRuleRepository {
	return &PointRuleRepository{}
}

// FindEnabled 查询生效中的积分抵扣规则；未配置时返回 (nil, nil)。
func (r *PointRuleRepository) FindEnabled(ctx context.Context) (*model.PointRuleModel, error) {
	var m model.PointRuleModel
	if err := facades.DB().WithContext(ctx).Table(model.PointRuleTable).
		Where("status", 1).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// Save 保存积分抵扣规则：存在则更新，不存在则插入。
func (r *PointRuleRepository) Save(ctx context.Context, m *model.PointRuleModel) error {
	if m.ID > 0 {
		m.FillUpdate(ctx)
		_, err := facades.DB().WithContext(ctx).Table(model.PointRuleTable).
			Where("id", m.ID).
			WhereNull("deleted_at").
			Update(map[string]any{
				"exchange_rate":     m.ExchangeRate,
				"max_deduct_amount": m.MaxDeductAmount,
				"min_order_amount":  m.MinOrderAmount,
				"status":            int8(m.Status),
				"updater":           m.Updater,
				"updated_at":        m.UpdatedAt,
			})

		return err
	}

	m.FillCreate(ctx)
	_, err := facades.DB().WithContext(ctx).Table(model.PointRuleTable).Insert(map[string]any{
		"id":                m.ID,
		"exchange_rate":     m.ExchangeRate,
		"max_deduct_amount": m.MaxDeductAmount,
		"min_order_amount":  m.MinOrderAmount,
		"status":            int8(m.Status),
		"creator":           m.Creator,
		"created_at":        m.CreatedAt,
		"updater":           m.Updater,
		"updated_at":        m.UpdatedAt,
	})

	return err
}

// containsScopeIDs 判断逗号分隔的 ID 列表里是否含目标 ID。
func containsScopeIDs(scopeIDs string, target types.ID) bool {
	if scopeIDs == "" || target == 0 {
		return false
	}
	for _, part := range strings.Split(scopeIDs, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == target.String() {
			return true
		}
	}

	return false
}
