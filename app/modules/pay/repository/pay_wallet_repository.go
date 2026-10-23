package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"
	frameworkDB "github.com/goravel/framework/database/db"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/model"
)

const (
	// payWalletTable 是会员钱包表名，对应 yudao pay_wallet
	payWalletTable = model.PayWalletTable
	// payWalletTransactionTable 是钱包流水表名，对应 yudao pay_wallet_transaction
	payWalletTransactionTable = model.PayWalletTransactionTable
	// payWalletRechargePackageTable 是充值套餐表名，对应 yudao pay_wallet_recharge_package
	payWalletRechargePackageTable = model.PayWalletRechargePackageTable
	// payWalletRechargeTable 是充值记录表名，对应 yudao pay_wallet_recharge
	payWalletRechargeTable = model.PayWalletRechargeTable
)

// PayWalletRepository 负责钱包、流水、充值套餐与充值记录的读写。
// 余额变动统一走 ChangeBalance 的条件更新，避免并发覆盖。
type PayWalletRepository struct{}

// NewPayWalletRepository 构造钱包仓储。
func NewPayWalletRepository() *PayWalletRepository {
	return &PayWalletRepository{}
}

// PayWalletFilter 是钱包列表的筛选条件。
type PayWalletFilter struct {
	// UserID 用户编号
	UserID types.ID
	// UserType 用户类型
	UserType *int8
}

// Paginate 分页查询钱包，按主键倒序。
// 入参：筛选条件与分页；返回钱包列表与总数。
func (r *PayWalletRepository) Paginate(ctx context.Context, filter PayWalletFilter, pageNo, pageSize int) ([]model.PayWalletModel, int64, error) {
	list := make([]model.PayWalletModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payWalletTable).WhereNull("deleted_at")
	if filter.UserID > 0 {
		query = query.Where("user_id", filter.UserID)
	}
	if filter.UserType != nil {
		query = query.Where("user_type", *filter.UserType)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByID 按主键查询未删除的钱包；未命中返回 (nil, nil)。
func (r *PayWalletRepository) FindByID(ctx context.Context, id types.ID) (*model.PayWalletModel, error) {
	return r.findWallet(ctx, facades.DB().WithContext(ctx).Table(payWalletTable).Where("id", id))
}

// FindByUser 按「用户 + 用户类型」查询钱包；未命中返回 (nil, nil)。
func (r *PayWalletRepository) FindByUser(ctx context.Context, userID types.ID, userType int8) (*model.PayWalletModel, error) {
	return r.findWallet(ctx, facades.DB().WithContext(ctx).Table(payWalletTable).
		Where("user_id", userID).
		Where("user_type", userType))
}

// Create 在事务内插入钱包；同一用户并发创建依赖 uk_user 唯一键兜底。
func (r *PayWalletRepository) Create(ctx context.Context, tx db.Tx, m *model.PayWalletModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payWalletTable).Insert(map[string]any{
		"id":             m.ID,
		"user_id":        m.UserID,
		"user_type":      m.UserType,
		"balance":        m.Balance,
		"freeze_price":   m.FreezePrice,
		"total_expense":  m.TotalExpense,
		"total_recharge": m.TotalRecharge,
		"creator":        m.Creator,
		"created_at":     m.CreatedAt,
		"updater":        m.Updater,
		"updated_at":     m.UpdatedAt,
	})

	return err
}

// ChangeBalance 以条件更新调整钱包余额与累计值，扣减时要求余额充足。
// 入参：事务、钱包 ID、余额增量（正入账 / 负出账）、累计支出增量、累计充值增量；
// 返回影响行数，0 表示钱包不存在或余额不足。
func (r *PayWalletRepository) ChangeBalance(ctx context.Context, tx db.Tx, id types.ID, deltaBalance, deltaExpense, deltaRecharge int64) (int64, error) {
	updates := map[string]any{"updated_at": types.Now()}
	if deltaBalance != 0 {
		updates["balance"] = frameworkDB.Raw("balance + ?", deltaBalance)
	}
	if deltaExpense != 0 {
		updates["total_expense"] = frameworkDB.Raw("total_expense + ?", deltaExpense)
	}
	if deltaRecharge != 0 {
		updates["total_recharge"] = frameworkDB.Raw("total_recharge + ?", deltaRecharge)
	}

	query := tx.Table(payWalletTable).
		Where("id", id).
		WhereNull("deleted_at")
	if deltaBalance < 0 {
		query = query.WhereRaw("balance + ? >= 0", []any{deltaBalance})
	}

	res, err := query.Update(updates)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// PayWalletTransactionFilter 是钱包流水的筛选条件。
type PayWalletTransactionFilter struct {
	// WalletID 钱包编号
	WalletID types.ID
	// BizType 业务类型，nil 表示不限
	BizType *payEnums.PayWalletBizTypeEnum
	// Type 收支方向，nil 表示不限
	Type *int8
}

// PaginateTransactions 分页查询钱包流水，按主键倒序。
// 入参：筛选条件与分页；返回流水列表与总数。
func (r *PayWalletRepository) PaginateTransactions(ctx context.Context, filter PayWalletTransactionFilter, pageNo, pageSize int) ([]model.PayWalletTransactionModel, int64, error) {
	list := make([]model.PayWalletTransactionModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payWalletTransactionTable).WhereNull("deleted_at")
	if filter.WalletID > 0 {
		query = query.Where("wallet_id", filter.WalletID)
	}
	if filter.BizType != nil {
		query = query.Where("biz_type", int8(*filter.BizType))
	}
	if filter.Type != nil {
		query = query.Where("type", *filter.Type)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CreateTransaction 在事务内写入一条钱包流水。
// 入参：事务与流水模型；返回错误表示落库失败。
func (r *PayWalletRepository) CreateTransaction(ctx context.Context, tx db.Tx, m *model.PayWalletTransactionModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payWalletTransactionTable).Insert(map[string]any{
		"id":         m.ID,
		"wallet_id":  m.WalletID,
		"biz_type":   int8(m.BizType),
		"biz_id":     m.BizID,
		"type":       int8(m.Type),
		"price":      m.Price,
		"balance":    m.Balance,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// PaginatePackages 分页查询充值套餐（含关闭），按主键倒序。
// 入参：分页；返回套餐列表与总数。
func (r *PayWalletRepository) PaginatePackages(ctx context.Context, pageNo, pageSize int) ([]model.PayWalletRechargePackageModel, int64, error) {
	list := make([]model.PayWalletRechargePackageModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(payWalletRechargePackageTable).
		WhereNull("deleted_at").
		OrderByDesc("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// AllEnabledPackages 返回全部开启状态的充值套餐，供 C 端选档。
// 入参：无；返回 status=0 的套餐列表，按支付金额升序。
func (r *PayWalletRepository) AllEnabledPackages(ctx context.Context) ([]model.PayWalletRechargePackageModel, error) {
	list := make([]model.PayWalletRechargePackageModel, 0)
	if err := facades.DB().WithContext(ctx).Table(payWalletRechargePackageTable).
		Where("status", int8(payEnums.PayEnableStatusEnable)).
		WhereNull("deleted_at").
		OrderBy("pay_price").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// FindPackage 按主键查询未删除的充值套餐；未命中返回 (nil, nil)。
func (r *PayWalletRepository) FindPackage(ctx context.Context, id types.ID) (*model.PayWalletRechargePackageModel, error) {
	var m model.PayWalletRechargePackageModel
	if err := facades.DB().WithContext(ctx).Table(payWalletRechargePackageTable).
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

// CreatePackage 在事务内新增充值套餐。
func (r *PayWalletRepository) CreatePackage(ctx context.Context, tx db.Tx, m *model.PayWalletRechargePackageModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payWalletRechargePackageTable).Insert(map[string]any{
		"id":          m.ID,
		"name":        m.Name,
		"pay_price":   m.PayPrice,
		"bonus_price": m.BonusPrice,
		"status":      int8(m.Status),
		"description": m.Description,
		"creator":     m.Creator,
		"created_at":  m.CreatedAt,
		"updater":     m.Updater,
		"updated_at":  m.UpdatedAt,
	})

	return err
}

// UpdatePackage 在事务内更新充值套餐。
func (r *PayWalletRepository) UpdatePackage(ctx context.Context, tx db.Tx, m *model.PayWalletRechargePackageModel) error {
	m.FillUpdate(ctx)

	_, err := tx.Table(payWalletRechargePackageTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":        m.Name,
			"pay_price":   m.PayPrice,
			"bonus_price": m.BonusPrice,
			"status":      int8(m.Status),
			"description": m.Description,
			"updater":     m.Updater,
			"updated_at":  m.UpdatedAt,
		})

	return err
}

// DeletePackage 软删除充值套餐，已产生的充值记录保留。
func (r *PayWalletRepository) DeletePackage(ctx context.Context, tx db.Tx, id types.ID) error {
	_, err := tx.Table(payWalletRechargePackageTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updated_at": types.Now(),
		})

	return err
}

// PayWalletRechargeFilter 是充值记录的筛选条件。
type PayWalletRechargeFilter struct {
	// WalletID 钱包编号
	WalletID types.ID
	// Status 充值状态，nil 表示不限
	Status *int8
}

// PaginateRecharges 分页查询充值记录，按主键倒序。
func (r *PayWalletRepository) PaginateRecharges(ctx context.Context, filter PayWalletRechargeFilter, pageNo, pageSize int) ([]model.PayWalletRechargeModel, int64, error) {
	list := make([]model.PayWalletRechargeModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payWalletRechargeTable).WhereNull("deleted_at")
	if filter.WalletID > 0 {
		query = query.Where("wallet_id", filter.WalletID)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// BalanceInTx 在事务内读取钱包最新余额，用于流水记录变动后余额。
// 入参：事务与钱包 ID；返回余额（分），钱包不存在时返回 0。
func (r *PayWalletRepository) BalanceInTx(ctx context.Context, tx db.Tx, id types.ID) (int64, error) {
	row := struct {
		Balance int64 `db:"balance"`
	}{}
	if err := tx.Table(payWalletTable).Select("balance").Where("id", id).First(&row); err != nil {
		return 0, err
	}

	return row.Balance, nil
}

// FindRechargeByPayOrderID 按支付单编号查充值记录，用于支付成功回调时判断是否为充值单。
// 入参：支付单 ID；返回充值记录，未命中返回 (nil, nil)。
func (r *PayWalletRepository) FindRechargeByPayOrderID(ctx context.Context, payOrderID types.ID) (*model.PayWalletRechargeModel, error) {
	var m model.PayWalletRechargeModel
	if err := facades.DB().WithContext(ctx).Table(payWalletRechargeTable).
		Where("pay_order_id", payOrderID).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// BindRechargePayOrder 回写充值记录关联的支付单，充值单创建后调用。
// 入参：事务、充值记录 ID 与支付单 ID；返回错误表示更新失败。
func (r *PayWalletRepository) BindRechargePayOrder(ctx context.Context, tx db.Tx, id, payOrderID types.ID) error {
	_, err := tx.Table(payWalletRechargeTable).
		Where("id", id).
		WhereNull("deleted_at").
		Update(map[string]any{
			"pay_order_id": payOrderID,
			"updated_at":   types.Now(),
		})

	return err
}

// FindRecharge 按主键查询未删除的充值记录；未命中返回 (nil, nil)。
func (r *PayWalletRepository) FindRecharge(ctx context.Context, id types.ID) (*model.PayWalletRechargeModel, error) {
	var m model.PayWalletRechargeModel
	if err := facades.DB().WithContext(ctx).Table(payWalletRechargeTable).
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

// CreateRecharge 在事务内插入充值记录（状态未支付）。
func (r *PayWalletRepository) CreateRecharge(ctx context.Context, tx db.Tx, m *model.PayWalletRechargeModel) error {
	m.FillCreate(ctx)

	_, err := tx.Table(payWalletRechargeTable).Insert(map[string]any{
		"id":               m.ID,
		"wallet_id":        m.WalletID,
		"total_price":      m.TotalPrice,
		"pay_price":        m.PayPrice,
		"bonus_price":      m.BonusPrice,
		"package_id":       m.PackageID,
		"status":           int8(m.Status),
		"pay_order_id":     m.PayOrderID,
		"pay_channel_code": m.PayChannelCode,
		"pay_time":         m.PayTime,
		"refund_order_id":  m.RefundOrderID,
		"refund_price":     m.RefundPrice,
		"refund_time":      m.RefundTime,
		"creator":          m.Creator,
		"created_at":       m.CreatedAt,
		"updater":          m.Updater,
		"updated_at":       m.UpdatedAt,
	})

	return err
}

// MarkRechargePaid 以条件更新把充值记录置为已支付，只命中未支付的记录。
// 入参：事务、充值记录 ID、支付单 ID、渠道编码、支付时间；返回影响行数。
func (r *PayWalletRepository) MarkRechargePaid(ctx context.Context, tx db.Tx, id, payOrderID types.ID, channelCode string, payTime types.DateTime) (int64, error) {
	res, err := tx.Table(payWalletRechargeTable).
		Where("id", id).
		Where("status", int8(payEnums.PayWalletRechargeStatusUnpaid)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":           int8(payEnums.PayWalletRechargeStatusPaid),
			"pay_order_id":     payOrderID,
			"pay_channel_code": channelCode,
			"pay_time":         payTime,
			"updated_at":       types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// MarkRechargeRefund 以条件更新回写充值退款信息，只命中已支付的记录。
// 入参：事务、充值记录 ID、退款单 ID、退款金额、退款时间；返回影响行数。
func (r *PayWalletRepository) MarkRechargeRefund(ctx context.Context, tx db.Tx, id, refundOrderID types.ID, refundPrice int64, refundTime types.DateTime) (int64, error) {
	res, err := tx.Table(payWalletRechargeTable).
		Where("id", id).
		Where("status", int8(payEnums.PayWalletRechargeStatusPaid)).
		WhereNull("deleted_at").
		Update(map[string]any{
			"status":          int8(payEnums.PayWalletRechargeStatusRefund),
			"refund_order_id": refundOrderID,
			"refund_price":    refundPrice,
			"refund_time":     refundTime,
			"updated_at":      types.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// findWallet 执行单条查询，未命中时返回 (nil, nil)。
func (r *PayWalletRepository) findWallet(ctx context.Context, query db.Query) (*model.PayWalletModel, error) {
	var m model.PayWalletModel
	if err := query.WhereNull("deleted_at").First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}
