package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payErrs "eqray-flavor/app/modules/pay/errs"
	"eqray-flavor/app/modules/pay/model"
	"eqray-flavor/app/modules/pay/repository"
)

const (
	// payWalletPackageNameMaxLength 充值套餐名长度上限
	payWalletPackageNameMaxLength = 64
	// payWalletPackageDescMaxLength 充值套餐描述长度上限
	payWalletPackageDescMaxLength = 255
	// payWalletRechargeSubjectPrefix 充值支付单的商品标题前缀
	payWalletRechargeSubjectPrefix = "钱包充值-"
)

// PayWalletChangeInput 是钱包余额变动入参，金额单位为分且恒为正数。
type PayWalletChangeInput struct {
	// WalletID 钱包编号
	WalletID types.ID
	// Price 变动金额（分），必须大于 0
	Price int64
	// Remark 备注
	Remark string
}

// PayWalletRechargeCreateInput 是发起钱包充值的入参。
type PayWalletRechargeCreateInput struct {
	// UserID 充值用户
	UserID types.ID
	// UserType 用户类型
	UserType int8
	// PackageID 充值套餐编号
	PackageID types.ID
	// AppID 支付应用编号
	AppID types.ID
	// ChannelCode 支付渠道编码
	ChannelCode string
	// OpenID 微信 JSAPI / 小程序支付必填
	OpenID string
	// AuthCode 支付宝条码支付必填
	AuthCode string
	// UserIP 用户 IP
	UserIP string
}

// PayWalletRechargeResult 是发起充值的返回结果，前端用 order 里的参数调起支付。
type PayWalletRechargeResult struct {
	// RechargeID 充值记录编号
	RechargeID types.ID `json:"rechargeId"`
	// TotalPrice 到账金额（分），含赠送
	TotalPrice int64 `json:"totalPrice"`
	// PayPrice 应付金额（分）
	PayPrice int64 `json:"payPrice"`
	// BonusPrice 赠送金额（分）
	BonusPrice int64 `json:"bonusPrice"`
	// Order 支付单调起参数
	Order *PayOrderSubmitResult `json:"order"`
}

// PayWalletRechargeRow 是充值记录输出，补上状态文案与到账金额。
type PayWalletRechargeRow struct {
	model.PayWalletRechargeModel
	// StatusName 状态文案，如「已支付」
	StatusName string `json:"statusName"`
}

// PayWalletService 负责会员钱包的余额变动、充值套餐与充值记录，对齐 yudao PayWalletService。
// 余额变动一律走条件更新 + 流水，避免并发把余额扣成负数。
type PayWalletService struct {
	wallets *repository.PayWalletRepository
	orders  *PayOrderService
	// ordersRepo 钱包余额支付需要直接落支付单
	ordersRepo *repository.PayOrderRepository
	// apps 支付应用配置，余额支付要写支付单的业务通知地址
	apps *repository.PayAppRepository
	// notifyTasks 余额支付成功后登记业务通知任务
	notifyTasks *PayNotifyTaskService
}

// NewPayWalletService 构造钱包服务。
func NewPayWalletService() *PayWalletService {
	return &PayWalletService{
		wallets:     repository.NewPayWalletRepository(),
		orders:      NewPayOrderService(),
		ordersRepo:  repository.NewPayOrderRepository(),
		apps:        repository.NewPayAppRepository(),
		notifyTasks: NewPayNotifyTaskService(),
	}
}

// Page 分页查询钱包，支持按用户与用户类型筛选。
// 入参：筛选条件与分页；返回钱包列表与总数。
func (s *PayWalletService) Page(ctx context.Context, filter repository.PayWalletFilter, pageNo, pageSize int) ([]model.PayWalletModel, int64, error) {
	return s.wallets.Paginate(ctx, filter, pageNo, pageSize)
}

// Detail 按主键查询钱包。
// 入参：钱包主键；返回钱包，不存在时报「钱包不存在」。
func (s *PayWalletService) Detail(ctx context.Context, id types.ID) (*model.PayWalletModel, error) {
	wallet, err := s.wallets.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, baseErrs.New(payErrs.PayWalletNotExists)
	}

	return wallet, nil
}

// EnsureWallet 查询会员钱包，没有则创建；并发创建由唯一键兜底后重查一次。
// 入参：用户编号与用户类型；返回钱包。
func (s *PayWalletService) EnsureWallet(ctx context.Context, userID types.ID, userType int8) (*model.PayWalletModel, error) {
	wallet, err := s.wallets.FindByUser(ctx, userID, userType)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	created := &model.PayWalletModel{UserID: userID, UserType: userType}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.wallets.Create(ctx, tx, created)
	}); err != nil {
		// 并发下另一请求已创建，重查一次即可拿到同一钱包
		existing, findErr := s.wallets.FindByUser(ctx, userID, userType)
		if findErr == nil && existing != nil {
			return existing, nil
		}

		return nil, err
	}

	return created, nil
}

// Transactions 分页查询钱包流水。
// 入参：筛选条件与分页；返回流水列表与总数。
func (s *PayWalletService) Transactions(ctx context.Context, filter repository.PayWalletTransactionFilter, pageNo, pageSize int) ([]model.PayWalletTransactionModel, int64, error) {
	return s.wallets.PaginateTransactions(ctx, filter, pageNo, pageSize)
}

// Income 给钱包入账（后台加款、退款回充等），并写一条收入流水。
// 入参：钱包编号、金额（分）与备注；返回错误表示钱包不存在或金额非法。
func (s *PayWalletService) Income(ctx context.Context, in PayWalletChangeInput) error {
	return s.changeBalance(ctx, in, payEnums.WalletBizAdjust, 1, "后台加款")
}

// Expense 从钱包扣款（消费、后台扣款等），余额不足时拒绝，并写一条支出流水。
// 入参：钱包编号、金额（分）与备注；返回错误表示余额不足或钱包不存在。
func (s *PayWalletService) Expense(ctx context.Context, in PayWalletChangeInput) error {
	return s.changeBalance(ctx, in, payEnums.WalletBizAdjust, -1, "后台扣款")
}

// changeBalance 在事务内调整余额并写流水，扣减时依赖条件更新保证不透支。
// 入参：变动入参、业务类型、方向（1 收入 / -1 支出）与默认备注；返回业务错误表示金额非法或余额不足。
// 说明：加款计入累计充值，扣款计入累计支出，保证钱包首页的累计收支口径正确。
func (s *PayWalletService) changeBalance(ctx context.Context, in PayWalletChangeInput, bizType payEnums.PayWalletBizTypeEnum, direction int64, defaultRemark string) error {
	if in.Price <= 0 {
		return baseErrs.New(payErrs.PayWalletPriceInvalid)
	}

	wallet, err := s.Detail(ctx, in.WalletID)
	if err != nil {
		return err
	}

	remark := strings.TrimSpace(in.Remark)
	if remark == "" {
		remark = defaultRemark
	}

	deltaExpense := int64(0)
	deltaRecharge := int64(0)
	if direction < 0 {
		deltaExpense = in.Price
	} else {
		deltaRecharge = in.Price
	}

	return s.applyBalance(
		ctx,
		wallet,
		in.Price*direction,
		deltaExpense,
		deltaRecharge,
		bizType,
		strconv.FormatInt(int64(wallet.ID), 10),
		remark,
	)
}

// applyBalance 开事务执行余额变更 + 流水落库。
// 入参：钱包、余额增量（正入账 / 负出账）、累计支出与累计充值增量、业务类型、业务编号与备注；返回业务错误。
func (s *PayWalletService) applyBalance(ctx context.Context, wallet *model.PayWalletModel, delta, deltaExpense, deltaRecharge int64, bizType payEnums.PayWalletBizTypeEnum, bizID, remark string) error {
	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.applyBalanceInTx(ctx, tx, wallet, delta, deltaExpense, deltaRecharge, bizType, bizID, remark)
	})
}

// applyBalanceInTx 在调用方事务内执行余额变更 + 流水落库，供退款回冲、余额支付等复用。
// 入参：事务、钱包、余额增量（正入账 / 负出账）、累计支出与累计充值增量、业务类型、业务编号与备注；返回业务错误。
// 说明：累计充值与累计支出只统计真实发生的充值与消费，退款回冲不冲减历史累计值。
func (s *PayWalletService) applyBalanceInTx(ctx context.Context, tx db.Tx, wallet *model.PayWalletModel, delta, deltaExpense, deltaRecharge int64, bizType payEnums.PayWalletBizTypeEnum, bizID, remark string) error {
	if delta == 0 {
		return baseErrs.New(payErrs.PayWalletPriceInvalid)
	}

	direction := payEnums.PayWalletTransactionTypeIncome
	price := delta
	if delta < 0 {
		direction = payEnums.PayWalletTransactionTypeExpense
		price = -delta
	}

	rows, err := s.wallets.ChangeBalance(ctx, tx, wallet.ID, delta, deltaExpense, deltaRecharge)
	if err != nil {
		return err
	}
	if rows == 0 {
		// 扣减为 0 行只有两种情况：钱包不存在或余额不足
		if delta < 0 {
			return baseErrs.New(payErrs.PayWalletBalanceNotEnough, wallet.Balance)
		}

		return baseErrs.New(payErrs.PayWalletNotExists)
	}

	balance, err := s.wallets.BalanceInTx(ctx, tx, wallet.ID)
	if err != nil {
		return err
	}

	return s.wallets.CreateTransaction(ctx, tx, &model.PayWalletTransactionModel{
		WalletID: wallet.ID,
		BizType:  bizType,
		BizID:    bizID,
		Type:     direction,
		Price:    price,
		Balance:  balance,
		Remark:   remark,
	})
}

// Packages 分页查询充值套餐（含已关闭），供后台维护使用。
// 入参：分页；返回套餐列表与总数。
func (s *PayWalletService) Packages(ctx context.Context, pageNo, pageSize int) ([]model.PayWalletRechargePackageModel, int64, error) {
	return s.wallets.PaginatePackages(ctx, pageNo, pageSize)
}

// EnabledPackages 返回开启状态的充值套餐，供前端选档。
// 入参：无；返回按实付金额升序的套餐列表。
func (s *PayWalletService) EnabledPackages(ctx context.Context) ([]model.PayWalletRechargePackageModel, error) {
	return s.wallets.AllEnabledPackages(ctx)
}

// CreatePackage 新增充值套餐。
// 入参：套餐名、实付金额、赠送金额、状态与描述；返回新套餐主键。
func (s *PayWalletService) CreatePackage(ctx context.Context, in model.PayWalletRechargePackageModel) (types.ID, error) {
	if err := validatePackageInput(&in); err != nil {
		return 0, err
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.wallets.CreatePackage(ctx, tx, &in)
	}); err != nil {
		return 0, err
	}

	return in.ID, nil
}

// UpdatePackage 编辑充值套餐，套餐已产生的充值记录不受影响。
// 入参：套餐字段（ID 必填）；返回错误表示套餐不存在或参数非法。
func (s *PayWalletService) UpdatePackage(ctx context.Context, in model.PayWalletRechargePackageModel) error {
	if _, err := s.findPackage(ctx, in.ID); err != nil {
		return err
	}
	if err := validatePackageInput(&in); err != nil {
		return err
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.wallets.UpdatePackage(ctx, tx, &in)
	})
}

// DeletePackage 删除充值套餐（软删除），历史充值记录保留。
// 入参：套餐主键；返回错误表示套餐不存在。
func (s *PayWalletService) DeletePackage(ctx context.Context, id types.ID) error {
	if _, err := s.findPackage(ctx, id); err != nil {
		return err
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.wallets.DeletePackage(ctx, tx, id)
	})
}

// Recharges 分页查询充值记录。
// 入参：筛选条件与分页；返回补好状态文案的充值记录与总数。
func (s *PayWalletService) Recharges(ctx context.Context, filter repository.PayWalletRechargeFilter, pageNo, pageSize int) ([]PayWalletRechargeRow, int64, error) {
	list, total, err := s.wallets.PaginateRecharges(ctx, filter, pageNo, pageSize)
	if err != nil {
		return nil, 0, err
	}

	rows := make([]PayWalletRechargeRow, 0, len(list))
	for _, item := range list {
		rows = append(rows, PayWalletRechargeRow{PayWalletRechargeModel: item, StatusName: item.Status.String()})
	}

	return rows, total, nil
}

// RechargeDetail 查询充值记录详情。
// 入参：充值记录主键；返回详情，不存在时报「充值记录不存在」。
func (s *PayWalletService) RechargeDetail(ctx context.Context, id types.ID) (*PayWalletRechargeRow, error) {
	m, err := s.wallets.FindRecharge(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayWalletRechargeNotExists)
	}

	return &PayWalletRechargeRow{PayWalletRechargeModel: *m, StatusName: m.Status.String()}, nil
}

// CreateRecharge 创建充值记录并发起支付，支付成功后由回调入账。
// 入参：用户、套餐、支付应用与渠道；返回充值记录编号与支付调起参数。
// 约束：套餐必须开启；支付单创建失败时充值记录保持未支付，可重试下单。
func (s *PayWalletService) CreateRecharge(ctx context.Context, in PayWalletRechargeCreateInput) (*PayWalletRechargeResult, error) {
	pkg, err := s.findPackage(ctx, in.PackageID)
	if err != nil {
		return nil, err
	}
	if pkg.Status != payEnums.PayEnableStatusEnable {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "充值套餐已关闭，请选择其它档位")
	}

	wallet, err := s.EnsureWallet(ctx, in.UserID, in.UserType)
	if err != nil {
		return nil, err
	}

	recharge := &model.PayWalletRechargeModel{
		WalletID:   wallet.ID,
		TotalPrice: pkg.PayPrice + pkg.BonusPrice,
		PayPrice:   pkg.PayPrice,
		BonusPrice: pkg.BonusPrice,
		PackageID:  pkg.ID,
		Status:     payEnums.PayWalletRechargeStatusUnpaid,
	}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.wallets.CreateRecharge(ctx, tx, recharge)
	}); err != nil {
		return nil, err
	}

	// 充值单用记录 ID 作为商户订单号，保证同一笔充值只有一张支付单
	order, err := s.orders.Create(ctx, PayOrderCreateInput{
		AppID:           in.AppID,
		ChannelCode:     in.ChannelCode,
		UserID:          in.UserID,
		UserType:        in.UserType,
		MerchantOrderID: "WR" + strconv.FormatInt(int64(recharge.ID), 10),
		Subject:         payWalletRechargeSubjectPrefix + pkg.Name,
		Price:           pkg.PayPrice,
		OpenID:          in.OpenID,
		AuthCode:        in.AuthCode,
		UserIP:          in.UserIP,
	})
	if err != nil {
		return nil, err
	}

	var orderID types.ID
	if order != nil {
		orderID = order.OrderID
	}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.wallets.BindRechargePayOrder(ctx, tx, recharge.ID, orderID)
	}); err != nil {
		return nil, err
	}

	return &PayWalletRechargeResult{
		RechargeID: recharge.ID,
		TotalPrice: recharge.TotalPrice,
		PayPrice:   recharge.PayPrice,
		BonusPrice: recharge.BonusPrice,
		Order:      order,
	}, nil
}

// HandleRechargePaid 在支付成功事务内给充值单入账：置为已支付、加余额与累计充值、写流水。
// 入参：事务与支付单；返回错误表示入账失败（会让支付回调整体重试）。
// 说明：非充值支付单直接返回 nil；重复回调由 MarkRechargePaid 的条件更新兜底。
func (s *PayWalletService) HandleRechargePaid(ctx context.Context, tx db.Tx, order *model.PayOrderModel) error {
	recharge, err := s.wallets.FindRechargeByPayOrderID(ctx, order.ID)
	if err != nil {
		return err
	}
	if recharge == nil || recharge.Status != payEnums.PayWalletRechargeStatusUnpaid {
		return nil
	}

	rows, err := s.wallets.MarkRechargePaid(ctx, tx, recharge.ID, order.ID, order.ChannelCode, types.DateTime(time.Now()))
	if err != nil || rows == 0 {
		return err
	}

	// 余额按「实付 + 赠送」入账，累计充值只统计实付
	if rows, err = s.wallets.ChangeBalance(ctx, tx, recharge.WalletID, recharge.TotalPrice, 0, recharge.PayPrice); err != nil {
		return err
	}
	if rows == 0 {
		return baseErrs.New(payErrs.PayWalletNotExists)
	}

	balance, err := s.wallets.BalanceInTx(ctx, tx, recharge.WalletID)
	if err != nil {
		return err
	}

	return s.wallets.CreateTransaction(ctx, tx, &model.PayWalletTransactionModel{
		WalletID: recharge.WalletID,
		BizType:  payEnums.WalletBizRecharge,
		BizID:    strconv.FormatInt(int64(recharge.ID), 10),
		Type:     payEnums.PayWalletTransactionTypeIncome,
		Price:    recharge.TotalPrice,
		Balance:  balance,
		Remark:   "钱包充值",
	})
}

// PayWalletPayInput 是钱包余额支付的入参，金额单位为分。
type PayWalletPayInput struct {
	// AppID 支付应用编号，业务通知按应用配置的地址推送
	AppID types.ID
	// UserID 付款用户
	UserID types.ID
	// UserType 用户类型
	UserType int8
	// MerchantOrderID 业务订单号，用于下单幂等
	MerchantOrderID string
	// Subject 商品标题
	Subject string
	// Body 商品描述
	Body string
	// Price 支付金额（分）
	Price int64
	// NotifyURL 业务回调地址，留空时取应用配置的支付回调地址
	NotifyURL string
	// UserIP 用户 IP
	UserIP string
}

// PayWalletPayResult 是余额支付的返回结果，前端据此展示支付结果与最新余额。
type PayWalletPayResult struct {
	// OrderID 支付单主键
	OrderID types.ID `json:"orderId"`
	// No 支付单号
	No string `json:"no"`
	// Price 支付金额（分）
	Price int64 `json:"price"`
	// Status 支付状态
	Status payEnums.PayOrderStatusEnum `json:"status"`
	// StatusName 状态文案，如「已支付」
	StatusName string `json:"statusName"`
	// Balance 支付后的钱包余额（分）
	Balance int64 `json:"balance"`
	// SuccessTime 支付成功时间
	SuccessTime *types.DateTime `json:"successTime"`
}

// PayByBalance 用钱包余额支付：直接落一张已支付的支付单并扣减余额，不经过渠道网关。
// 入参：应用、用户、业务订单号与金额；返回支付结果与支付后余额，余额不足时整体回滚并返回业务错误。
// 约束：同一应用的商户订单号不可重复；扣款与建单在同一事务内，避免扣了钱没订单。
// 说明：这是钱包渠道的支付实现，由 PayOrderService.Create 在渠道编码为 wallet 时调用。
func (s *PayWalletService) PayByBalance(ctx context.Context, in PayWalletPayInput) (*PayWalletPayResult, error) {
	in.MerchantOrderID = strings.TrimSpace(in.MerchantOrderID)
	in.Subject = strings.TrimSpace(in.Subject)
	in.Body = strings.TrimSpace(in.Body)
	in.NotifyURL = strings.TrimSpace(in.NotifyURL)
	in.UserIP = strings.TrimSpace(in.UserIP)

	if in.MerchantOrderID == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "商户订单号不能为空")
	}
	if in.Subject == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "商品标题不能为空")
	}
	if in.Price <= 0 {
		return nil, baseErrs.New(payErrs.PayOrderPriceInvalid)
	}

	existing, err := s.ordersRepo.FindByMerchantOrderID(ctx, in.AppID, in.MerchantOrderID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, baseErrs.New(payErrs.PayOrderMerchantOrderIDDuplicate, in.MerchantOrderID)
	}

	// 钱包支付同样要校验支付应用：应用缺失或关闭时没有业务通知地址可推送
	app, err := s.apps.FindByID(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, baseErrs.New(payErrs.PayAppNotExists)
	}
	if app.Status != payEnums.PayEnableStatusEnable {
		return nil, baseErrs.New(payErrs.PayAppIsDisable, app.Name)
	}

	wallet, err := s.EnsureWallet(ctx, in.UserID, in.UserType)
	if err != nil {
		return nil, err
	}

	// 余额支付不走渠道回调，但仍要写上业务通知地址，便于后台排查与业务侧对齐
	notifyURL := in.NotifyURL
	if notifyURL == "" {
		notifyURL = app.OrderNotifyURL
	}

	now := types.DateTime(time.Now())
	order := &model.PayOrderModel{
		AppID:           in.AppID,
		ChannelCode:     payEnums.ChannelWallet,
		UserID:          in.UserID,
		UserType:        in.UserType,
		MerchantOrderID: in.MerchantOrderID,
		Subject:         in.Subject,
		Body:            in.Body,
		NotifyURL:       notifyURL,
		Price:           in.Price,
		Status:          payEnums.PayOrderStatusSuccess,
		UserIP:          in.UserIP,
		ExpireTime:      now,
		SuccessTime:     &now,
		No:              newPayOrderNo(),
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.ordersRepo.Create(ctx, tx, order); err != nil {
			return err
		}

		// 余额扣减与支付单在同一事务：余额不足会连带回滚支付单；本次消费计入累计支出
		if err := s.applyBalanceInTx(
			ctx,
			tx,
			wallet,
			-in.Price,
			in.Price,
			0,
			payEnums.WalletBizPay,
			strconv.FormatInt(int64(order.ID), 10),
			"钱包余额支付："+in.Subject,
		); err != nil {
			return err
		}

		return s.notifyTasks.CreateOrderTask(ctx, tx, order)
	}); err != nil {
		return nil, err
	}

	// 事务提交后回读余额，前端直接用返回值刷新钱包展示
	balance := wallet.Balance - in.Price
	if latest, err := s.wallets.FindByID(ctx, wallet.ID); err == nil && latest != nil {
		balance = latest.Balance
	}

	return &PayWalletPayResult{
		OrderID:     order.ID,
		No:          order.No,
		Price:       order.Price,
		Status:      order.Status,
		StatusName:  order.Status.String(),
		Balance:     balance,
		SuccessTime: order.SuccessTime,
	}, nil
}

// CheckRechargeRefund 在发起退款前校验充值单退款规则与余额是否够扣回。
// 入参：支付单编号与退款金额（分）；返回业务错误，非充值单直接通过。
// 约束：充值单只支持整单退款（退款金额等于实付金额），且钱包余额需足够扣回到账金额。
func (s *PayWalletService) CheckRechargeRefund(ctx context.Context, orderID types.ID, refundPrice int64) error {
	recharge, err := s.wallets.FindRechargeByPayOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	if recharge == nil {
		return nil
	}
	if recharge.Status != payEnums.PayWalletRechargeStatusPaid {
		return baseErrs.New(payErrs.PayWalletRechargeStatusInvalid, recharge.Status.String())
	}
	if refundPrice != recharge.PayPrice {
		return baseErrs.WithMsg(baseErrs.BadRequest, "充值单仅支持整单退款，退款金额必须等于实付金额")
	}

	wallet, err := s.wallets.FindByID(ctx, recharge.WalletID)
	if err != nil {
		return err
	}
	if wallet == nil {
		return baseErrs.New(payErrs.PayWalletNotExists)
	}
	// 余额已被消费时不允许退款，避免把余额扣成负数
	if wallet.Balance < recharge.TotalPrice {
		return baseErrs.New(payErrs.PayWalletBalanceNotEnough, wallet.Balance)
	}

	return nil
}

// ApplyWalletRefundInTx 在退款成功的事务内回冲钱包，按支付单性质分两种处理：
// 充值单回冲充值时到账的金额并把充值单置为已退款；普通钱包支付单把退款金额退回余额。
// 入参：事务、支付单、退款单编号与退款金额（分）；返回业务错误，非钱包支付单直接通过。
func (s *PayWalletService) ApplyWalletRefundInTx(ctx context.Context, tx db.Tx, order *model.PayOrderModel, refundOrderID types.ID, refundPrice int64) error {
	if order == nil || order.ChannelCode != payEnums.ChannelWallet {
		return nil
	}

	recharge, err := s.wallets.FindRechargeByPayOrderID(ctx, order.ID)
	if err != nil {
		return err
	}
	if recharge != nil {
		return s.refundRechargeInTx(ctx, tx, recharge, refundOrderID, refundPrice)
	}

	return s.refundPaymentInTx(ctx, tx, order, refundOrderID, refundPrice)
}

// refundRechargeInTx 回冲充值单：置为已退款、扣回到账金额并写退款流水。
// 入参：事务、充值记录、退款单编号与退款金额（分）；返回业务错误，已退款时幂等跳过。
func (s *PayWalletService) refundRechargeInTx(ctx context.Context, tx db.Tx, recharge *model.PayWalletRechargeModel, refundOrderID types.ID, refundPrice int64) error {
	if recharge.Status != payEnums.PayWalletRechargeStatusPaid {
		return nil
	}

	rows, err := s.wallets.MarkRechargeRefund(ctx, tx, recharge.ID, refundOrderID, refundPrice, types.DateTime(time.Now()))
	if err != nil || rows == 0 {
		return err
	}

	wallet, err := s.wallets.FindByID(ctx, recharge.WalletID)
	if err != nil {
		return err
	}
	if wallet == nil {
		return baseErrs.New(payErrs.PayWalletNotExists)
	}

	// 退款回冲按到账金额扣（实付 + 赠送），与入账金额对称；余额不足时由条件更新兜底报错
	return s.applyBalanceInTx(
		ctx,
		tx,
		wallet,
		-recharge.TotalPrice,
		0,
		0,
		payEnums.WalletBizRechargeRefund,
		strconv.FormatInt(int64(recharge.ID), 10),
		"充值退款回冲",
	)
}

// refundPaymentInTx 把钱包支付的订单金额退回余额，用于普通业务订单退款，避免退款后钱不进钱包。
// 入参：事务、支付单、退款单编号与退款金额（分）；返回业务错误表示钱包不存在。
func (s *PayWalletService) refundPaymentInTx(ctx context.Context, tx db.Tx, order *model.PayOrderModel, refundOrderID types.ID, refundPrice int64) error {
	wallet, err := s.wallets.FindByUser(ctx, order.UserID, order.UserType)
	if err != nil {
		return err
	}
	if wallet == nil {
		return baseErrs.New(payErrs.PayWalletNotExists)
	}

	return s.applyBalanceInTx(
		ctx,
		tx,
		wallet,
		refundPrice,
		0,
		0,
		payEnums.WalletBizPayRefund,
		strconv.FormatInt(int64(refundOrderID), 10),
		"钱包支付退款："+order.Subject,
	)
}

// findPackage 按主键取套餐，未命中转换成固定业务错误。
func (s *PayWalletService) findPackage(ctx context.Context, id types.ID) (*model.PayWalletRechargePackageModel, error) {
	pkg, err := s.wallets.FindPackage(ctx, id)
	if err != nil {
		return nil, err
	}
	if pkg == nil {
		return nil, baseErrs.New(payErrs.PayWalletRechargePackageNotExists)
	}

	return pkg, nil
}

// validatePackageInput 校验充值套餐入参：名称、实付金额与描述长度。
// 入参：待落库的套餐模型指针；返回业务错误表示校验不通过。
func validatePackageInput(in *model.PayWalletRechargePackageModel) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "套餐名不能为空")
	}
	if len([]rune(in.Name)) > payWalletPackageNameMaxLength {
		return baseErrs.WithMsg(baseErrs.BadRequest, "套餐名不能超过 64 个字符")
	}
	if in.PayPrice <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "套餐支付金额必须大于 0")
	}
	if in.BonusPrice < 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "赠送金额不能为负数")
	}
	if len([]rune(in.Description)) > payWalletPackageDescMaxLength {
		return baseErrs.WithMsg(baseErrs.BadRequest, "套餐描述不能超过 255 个字符")
	}

	return nil
}
