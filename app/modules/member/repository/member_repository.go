package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"
	frameworkDB "github.com/goravel/framework/database/db"

	"eqray-flavor/app/facades"
	baseEnums "eqray-flavor/app/modules/base/enums"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	memberErrs "eqray-flavor/app/modules/member/errs"
	"eqray-flavor/app/modules/member/model"
)

const memberTable = model.MemberUserTable

// MemberFilter 是会员列表的筛选条件，零值表示不限制。
type MemberFilter struct {
	Nickname string
	Mobile   string
	LevelID  types.ID
	GroupID  types.ID
	Status   *baseEnums.CommonStatusEnum
}

// MemberRepository 是 member_user 表的仓储。
type MemberRepository struct{}

func NewMemberRepository() *MemberRepository {
	return &MemberRepository{}
}

// Create 插入会员；手机号冲突转换为统一的业务错误码。
func (r *MemberRepository) Create(ctx context.Context, m *model.MemberUserModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(memberTable).Insert(map[string]any{
		"id":          m.ID,
		"mobile":      m.Mobile,
		"password":    m.Password,
		"nickname":    m.Nickname,
		"avatar":      m.Avatar,
		"name":        m.Name,
		"sex":         m.Sex,
		"birthday":    m.Birthday,
		"area_id":     m.AreaID,
		"level_id":    m.LevelID,
		"point":       m.Point,
		"experience":  m.Experience,
		"tag_ids":     m.TagIDs,
		"mark":        m.Mark,
		"register_ip": m.RegisterIP,
		"login_ip":    m.LoginIP,
		"login_date":  m.LoginDate,
		"group_id":    m.GroupID,
		"status":      m.Status,
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(memberErrs.MemberMobileExists, m.Mobile)
	}

	return err
}

// Update 更新会员可变字段；手机号冲突同样转换为业务错误码。
func (r *MemberRepository) Update(ctx context.Context, m *model.MemberUserModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(memberTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"mobile":     m.Mobile,
			"nickname":   m.Nickname,
			"avatar":     m.Avatar,
			"name":       m.Name,
			"sex":        m.Sex,
			"birthday":   m.Birthday,
			"area_id":    m.AreaID,
			"level_id":   m.LevelID,
			"point":      m.Point,
			"experience": m.Experience,
			"tag_ids":    m.TagIDs,
			"mark":       m.Mark,
			"group_id":   m.GroupID,
			"status":     m.Status,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(memberErrs.MemberMobileExists, m.Mobile)
	}

	return err
}

// UpdateProfile 只更新会员自助可改的字段。
// 手机号、积分、等级、分组、标签、状态属于后台或独立流程，不走这里，避免越权覆盖。
func (r *MemberRepository) UpdateProfile(ctx context.Context, m *model.MemberUserModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(memberTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"nickname":   m.Nickname,
			"avatar":     m.Avatar,
			"name":       m.Name,
			"sex":        m.Sex,
			"birthday":   m.Birthday,
			"area_id":    m.AreaID,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// UpdatePoint 在事务内更新积分余额，与积分记录写入保持原子性。
func (r *MemberRepository) UpdatePoint(ctx context.Context, tx db.Tx, m *model.MemberUserModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(memberTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"point":      m.Point,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// UpdateLevel 在事务内更新等级与经验值，与等级记录写入保持原子性。
func (r *MemberRepository) UpdateLevel(ctx context.Context, tx db.Tx, m *model.MemberUserModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(memberTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"level_id":   m.LevelID,
			"experience": m.Experience,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// DeductPoint 原子扣减积分：带「余额 >= 扣减额」条件更新，保证并发下不会扣成负数。
// 入参：事务、会员 ID、扣减积分数（正数）、更新者；返回扣减后的余额与是否成功。
// 影响 0 行表示余额不足或会员不存在，调用方据此回滚事务。
func (r *MemberRepository) DeductPoint(ctx context.Context, tx db.Tx, id types.ID, point int, updater string) (int, bool, error) {
	if point <= 0 {
		return 0, false, nil
	}

	res, err := tx.Table(memberTable).
		Where("id", id).
		Where("point >= ?", point).
		WhereNull("deleted_at").
		Update(map[string]any{
			"point":      frameworkDB.Raw("point - ?", point),
			"updater":    updater,
			"updated_at": types.Now(),
		})
	if err != nil {
		return 0, false, err
	}
	if res.RowsAffected == 0 {
		return 0, false, nil
	}

	// 回读扣减后的余额，用于写积分流水的 total_point
	row := struct {
		Point int `db:"point"`
	}{}
	if err := tx.Table(memberTable).
		Select("point").
		Where("id", id).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, false, err
	}

	return row.Point, true, nil
}

// AddPoint 原子增加积分：整单退款返还积分等场景使用。
// 入参：事务、会员 ID、增加积分数（正数）、更新者；返回增加后的余额与是否成功。
func (r *MemberRepository) AddPoint(ctx context.Context, tx db.Tx, id types.ID, point int, updater string) (int, bool, error) {
	if point <= 0 {
		return 0, false, nil
	}

	res, err := tx.Table(memberTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"point":      frameworkDB.Raw("point + ?", point),
			"updater":    updater,
			"updated_at": types.Now(),
		})
	if err != nil {
		return 0, false, err
	}
	if res.RowsAffected == 0 {
		return 0, false, nil
	}

	row := struct {
		Point int `db:"point"`
	}{}
	if err := tx.Table(memberTable).
		Select("point").
		Where("id", id).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, false, err
	}

	return row.Point, true, nil
}

// AddExperience 原子累加成长值：等级只升不降，成长值按消费金额累计。
// 入参：事务、会员 ID、累加成长值、更新者；返回累加后的成长值与是否成功。
func (r *MemberRepository) AddExperience(ctx context.Context, tx db.Tx, id types.ID, experience int, updater string) (int, bool, error) {
	if experience <= 0 {
		return 0, false, nil
	}

	res, err := tx.Table(memberTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"experience": frameworkDB.Raw("experience + ?", experience),
			"updater":    updater,
			"updated_at": types.Now(),
		})
	if err != nil {
		return 0, false, err
	}
	if res.RowsAffected == 0 {
		return 0, false, nil
	}

	row := struct {
		Experience int `db:"experience"`
	}{}
	if err := tx.Table(memberTable).
		Select("experience").
		Where("id", id).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, false, err
	}

	return row.Experience, true, nil
}

// DeductExperience 原子回退成长值：整单退款扣回已赠送的成长值时使用。
// 入参：事务、会员 ID、回退成长值（正数）、更新者；返回回退后的成长值与是否成功。
// 说明：回退量按「当前成长值与待回退量的较小值」扣减，保证不会出现负数，
// 也不因成长值被后台调整过而阻断退款流程。
func (r *MemberRepository) DeductExperience(ctx context.Context, tx db.Tx, id types.ID, experience int, updater string) (int, bool, error) {
	if experience <= 0 {
		return 0, false, nil
	}

	res, err := tx.Table(memberTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"experience": frameworkDB.Raw("experience - LEAST(experience, ?)", experience),
			"updater":    updater,
			"updated_at": types.Now(),
		})
	if err != nil {
		return 0, false, err
	}
	if res.RowsAffected == 0 {
		return 0, false, nil
	}

	// 回读回退后的成长值
	row := struct {
		Experience int `db:"experience"`
	}{}
	if err := tx.Table(memberTable).
		Select("experience").
		Where("id", id).
		WhereNull("deleted_at").
		First(&row); err != nil {
		return 0, false, err
	}

	return row.Experience, true, nil
}

// UpdateStatus 只更新启用状态，列表切换开关时避免整行覆盖。
func (r *MemberRepository) UpdateStatus(ctx context.Context, m *model.MemberUserModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(memberTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":     m.Status,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// UpdateLoginInfo 记录会员最近一次登录，只更新登录相关字段。
func (r *MemberRepository) UpdateLoginInfo(ctx context.Context, id types.ID, ip string, at types.DateTime) error {
	_, err := facades.DB().WithContext(ctx).Table(memberTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"login_ip":   ip,
			"login_date": at,
			"updated_at": types.Now(),
		})

	return err
}

// Delete 软删除会员。
func (r *MemberRepository) Delete(ctx context.Context, m *model.MemberUserModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(memberTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(memberErrs.MemberNotExists, m.ID)
	}

	return nil
}

// FindByID 按主键查询未软删的会员；未命中返回 (nil, nil)。
func (r *MemberRepository) FindByID(ctx context.Context, id types.ID) (*model.MemberUserModel, error) {
	var m model.MemberUserModel

	if err := facades.DB().WithContext(ctx).Table(memberTable).
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

// FindByMobile 按手机号查询未软删的会员；未命中返回 (nil, nil)。
func (r *MemberRepository) FindByMobile(ctx context.Context, mobile string) (*model.MemberUserModel, error) {
	var m model.MemberUserModel

	if err := facades.DB().WithContext(ctx).Table(memberTable).
		Where("mobile", mobile).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ExistsByMobile 判断手机号是否已被占用，excludeID 用于编辑时排除自身。
func (r *MemberRepository) ExistsByMobile(ctx context.Context, mobile string, excludeID types.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(memberTable).
		Where("mobile", mobile).
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

// Paginate 按条件分页查询会员。
func (r *MemberRepository) Paginate(ctx context.Context, filter MemberFilter, pageNo, pageSize int) ([]model.MemberUserModel, int64, error) {
	list := make([]model.MemberUserModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(memberTable).WhereNull("deleted_at")
	if filter.Nickname != "" {
		query = query.WhereLike("nickname", "%"+filter.Nickname+"%")
	}
	if filter.Mobile != "" {
		query = query.WhereLike("mobile", "%"+filter.Mobile+"%")
	}
	if filter.LevelID > 0 {
		query = query.Where("level_id", filter.LevelID)
	}
	if filter.GroupID > 0 {
		query = query.Where("group_id", filter.GroupID)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CreateLevelRecord 在事务内写入一条等级变更记录。
func (r *MemberRepository) CreateLevelRecord(ctx context.Context, tx db.Tx, m *model.MemberLevelRecordModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(model.MemberLevelRecordTable).Insert(map[string]any{
		"id":          m.ID,
		"user_id":     m.UserID,
		"level_id":    m.LevelID,
		"experience":  m.Experience,
		"title":       m.Title,
		"description": m.Description,
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})

	return err
}

// LevelRecords 分页查询会员的等级变更记录。
func (r *MemberRepository) LevelRecords(ctx context.Context, userID types.ID, pageNo, pageSize int) ([]model.MemberLevelRecordModel, int64, error) {
	list := make([]model.MemberLevelRecordModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(model.MemberLevelRecordTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		OrderByDesc("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// SignRecords 分页查询会员的签到记录。
func (r *MemberRepository) SignRecords(ctx context.Context, userID types.ID, pageNo, pageSize int) ([]model.MemberSignRecordModel, int64, error) {
	list := make([]model.MemberSignRecordModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(model.MemberSignRecordTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		OrderByDesc("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Addresses 查询会员收货地址，数量有限不分页。
func (r *MemberRepository) Addresses(ctx context.Context, userID types.ID) ([]model.MemberAddressModel, error) {
	list := make([]model.MemberAddressModel, 0)

	if err := facades.DB().WithContext(ctx).Table(model.MemberAddressTable).
		Where("user_id", userID).
		WhereNull("deleted_at").
		OrderByDesc("default_status").
		OrderByDesc("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}
