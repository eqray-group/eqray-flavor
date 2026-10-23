package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-flavor/app/facades"
	baseErrs "eqray-flavor/app/modules/base/errs"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	payErrs "eqray-flavor/app/modules/pay/errs"
	"eqray-flavor/app/modules/pay/model"
	"eqray-flavor/app/modules/pay/repository"
)

const (
	// payRefundReasonMaxLength 退款原因长度上限，与 pay_refund.reason 对齐
	payRefundReasonMaxLength = 255
)

// PayRefundListQuery 是后台退款单列表的筛选条件。
type PayRefundListQuery struct {
	// Keyword 退款单号 / 支付单号 / 渠道退款单号关键字
	Keyword string
	// AppID 应用编号，0 表示不限
	AppID types.ID
	// OrderID 支付单编号，0 表示不限
	OrderID types.ID
	// ChannelCode 渠道编码
	ChannelCode string
	// Status 退款状态，nil 表示不限
	Status *payEnums.PayRefundStatusEnum
	// AuditStatus 审核状态，nil 表示不限
	AuditStatus *payEnums.PayRefundAuditStatusEnum
}

// PayRefundCreateInput 是发起退款的入参，金额单位为分。
type PayRefundCreateInput struct {
	// OrderID 支付单主键
	OrderID types.ID
	// MerchantRefundID 业务退款单号，用于退款幂等
	MerchantRefundID string
	// RefundPrice 本次退款金额（分）
	RefundPrice int64
	// Reason 退款原因
	Reason string
	// UserIP 操作人 IP
	UserIP string
	// NeedAudit 是否需要审核：true 时只落退款单待审核，不动钱也不提交渠道
	NeedAudit bool
}

// PayRefundRow 是退款单列表 / 详情输出，补上渠道与状态文案。
type PayRefundRow struct {
	model.PayRefundModel
	// ChannelName 渠道中文名，如「微信小程序」
	ChannelName string `json:"channelName"`
	// StatusName 状态文案，如「等待退款」
	StatusName string `json:"statusName"`
	// AuditStatusName 审核状态文案，如「待审核」
	AuditStatusName string `json:"auditStatusName"`
}

// PayRefundService 负责退款申请的受理、渠道调用与状态推进，对齐 yudao PayRefundService。
// 退款结果以渠道为准：渠道同步确认完成的立即置成功，否则保持等待并等退款回调补齐。
type PayRefundService struct {
	refunds  *repository.PayRefundRepository
	orders   *repository.PayOrderRepository
	channels *repository.PayChannelRepository
	// notifyTasks 退款成功后登记业务通知任务
	notifyTasks *PayNotifyTaskService
	// wallet 钱包支付单没有渠道记录，退款校验与回冲由钱包服务完成
	wallet *PayWalletService
}

// NewPayRefundService 构造退款单服务。
func NewPayRefundService() *PayRefundService {
	return &PayRefundService{
		refunds:     repository.NewPayRefundRepository(),
		orders:      repository.NewPayOrderRepository(),
		channels:    repository.NewPayChannelRepository(),
		notifyTasks: NewPayNotifyTaskService(),
		wallet:      NewPayWalletService(),
	}
}

// Page 分页查询退款单，支持关键字、应用、支付单、渠道与状态筛选。
// 入参：筛选条件与分页；返回补好文案的退款单列表与总数。
func (s *PayRefundService) Page(ctx context.Context, q PayRefundListQuery, pageNo, pageSize int) ([]PayRefundRow, int64, error) {
	filter := repository.PayRefundFilter{
		Keyword:     strings.TrimSpace(q.Keyword),
		AppID:       q.AppID,
		OrderID:     q.OrderID,
		ChannelCode: strings.TrimSpace(q.ChannelCode),
	}
	if q.Status != nil {
		status := int8(*q.Status)
		filter.Status = &status
	}
	if q.AuditStatus != nil {
		auditStatus := int8(*q.AuditStatus)
		filter.AuditStatus = &auditStatus
	}

	rows, total, err := s.refunds.Paginate(ctx, filter, pageNo, pageSize)
	if err != nil {
		return nil, 0, err
	}

	list := make([]PayRefundRow, 0, len(rows))
	for _, row := range rows {
		list = append(list, decoratePayRefund(row))
	}

	return list, total, nil
}

// Detail 查询退款单详情。
// 入参：退款单主键；返回详情，不存在时报「退款单不存在」。
func (s *PayRefundService) Detail(ctx context.Context, id types.ID) (*PayRefundRow, error) {
	m, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}

	row := decoratePayRefund(*m)

	return &row, nil
}

// Create 受理退款申请并调用渠道退款。
// 入参：支付单主键、业务退款单号、退款金额（分）与原因；返回退款单。
// 约束：只允许对已支付的单退款；等待中与已成功的退款金额合计不得超过实付金额。
func (s *PayRefundService) Create(ctx context.Context, in PayRefundCreateInput) (*PayRefundRow, error) {
	in.MerchantRefundID = strings.TrimSpace(in.MerchantRefundID)
	in.Reason = strings.TrimSpace(in.Reason)
	in.UserIP = strings.TrimSpace(in.UserIP)

	if in.MerchantRefundID == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "商户退款单号不能为空")
	}
	if in.RefundPrice <= 0 {
		return nil, baseErrs.New(payErrs.PayRefundPriceInvalid, 0)
	}
	if len([]rune(in.Reason)) > payRefundReasonMaxLength {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "退款原因不能超过 255 个字符")
	}

	order, err := s.orders.FindByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, baseErrs.New(payErrs.PayOrderNotExists)
	}
	if !order.Status.CanRefund() {
		return nil, baseErrs.New(payErrs.PayRefundOrderStatusInvalid, order.No)
	}

	exists, err := s.refunds.FindByMerchantRefundID(ctx, order.AppID, in.MerchantRefundID)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, baseErrs.New(payErrs.PayRefundMerchantRefundIDDuplicate, in.MerchantRefundID)
	}

	// 防超退：把仍在等待与已成功的退款都算进去，避免并发申请把单退穿
	occupied, err := s.refunds.SumRefundPrice(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	refundable := order.Price - occupied
	if in.RefundPrice > refundable {
		return nil, baseErrs.New(payErrs.PayRefundPriceInvalid, refundable)
	}

	// 钱包支付单没有 pay_channel 记录，退款在本模块直接完成，不走渠道网关
	if order.ChannelCode == payEnums.ChannelWallet {
		return s.createWalletRefund(ctx, order, in)
	}

	// 退款不受渠道开关影响：渠道已停用也要能退历史单
	channel, err := s.channels.FindByID(ctx, order.ChannelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, baseErrs.New(payErrs.PayChannelNotExists)
	}

	// 待审核：只落一张退款单占住额度，既不调渠道也不动钱，
	// 由后台审核通过后再走 submitRefund
	if in.NeedAudit {
		refund := &model.PayRefundModel{
			No:               newPayRefundNo(),
			AppID:            order.AppID,
			ChannelID:        channel.ID,
			ChannelCode:      channel.Code,
			OrderID:          order.ID,
			OrderNo:          order.No,
			MerchantOrderID:  order.MerchantOrderID,
			MerchantRefundID: in.MerchantRefundID,
			Status:           payEnums.PayRefundStatusCreate,
			PayPrice:         order.Price,
			RefundPrice:      in.RefundPrice,
			Reason:           in.Reason,
			UserIP:           in.UserIP,
			ChannelOrderNo:   order.ChannelOrderNo,
			AuditStatus:      payEnums.PayRefundAuditPending,
		}
		if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
			return s.refunds.Create(ctx, tx, refund)
		}); err != nil {
			return nil, err
		}

		row := decoratePayRefund(*refund)

		return &row, nil
	}

	gateway, err := NewChannelGateway(channel)
	if err != nil {
		return nil, err
	}

	refund := &model.PayRefundModel{
		No:               newPayRefundNo(),
		AppID:            order.AppID,
		ChannelID:        channel.ID,
		ChannelCode:      channel.Code,
		OrderID:          order.ID,
		OrderNo:          order.No,
		MerchantOrderID:  order.MerchantOrderID,
		MerchantRefundID: in.MerchantRefundID,
		Status:           payEnums.PayRefundStatusCreate,
		PayPrice:         order.Price,
		RefundPrice:      in.RefundPrice,
		Reason:           in.Reason,
		UserIP:           in.UserIP,
		ChannelOrderNo:   order.ChannelOrderNo,
		AuditStatus:      payEnums.PayRefundAuditNone,
	}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		return s.refunds.Create(ctx, tx, refund)
	}); err != nil {
		return nil, err
	}

	return s.submitRefund(ctx, order, refund, gateway)
}

// submitRefund 把已落库的退款单提交渠道并回写结果。
// 入参：支付单、退款单与渠道网关；返回更新后的退款单。
// 说明：审核通过时复用这段逻辑，保证「直接退款」与「审核后退款」走同一条路径。
func (s *PayRefundService) submitRefund(ctx context.Context, order *model.PayOrderModel, refund *model.PayRefundModel, gateway ChannelGateway) (*PayRefundRow, error) {
	channelRefundNo, finished, refundErr := gateway.Refund(ctx, RefundRequest{
		No:          order.No,
		RefundNo:    refund.No,
		PayPrice:    order.Price,
		RefundPrice: refund.RefundPrice,
		Reason:      refund.Reason,
		NotifyURL:   channelRefundNotifyURL(order.ChannelID),
	})
	if refundErr != nil {
		s.recordRefundError(ctx, refund, refundErr)

		return nil, refundErr
	}

	refund.ChannelRefundNo = channelRefundNo
	if finished {
		now := types.DateTime(time.Now())
		refund.Status = payEnums.PayRefundStatusSuccess
		refund.SuccessTime = &now
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if _, err := s.refunds.UpdateResult(ctx, tx, refund); err != nil {
			return err
		}
		// 渠道确认退款成功后才累计支付单已退金额，等待中的退款只占额度不改累计值
		if finished {
			rows, err := s.orders.MarkRefunded(ctx, tx, order.ID, refund.RefundPrice)
			if err != nil {
				return err
			}
			// 并发退款时两笔都可能通过前置校验，这里用条件更新兜住累计超退
			if rows == 0 {
				return baseErrs.New(payErrs.PayRefundExceedOrderPrice)
			}

			// 退款完成即登记业务通知任务，应用未配置回调地址时内部跳过
			return s.notifyTasks.CreateRefundTask(ctx, tx, refund)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	row := decoratePayRefund(*refund)

	return &row, nil
}

// createWalletRefund 受理钱包支付单的退款：校验充值单规则后直接落一张已成功的退款单。
// 入参：支付单与退款入参；返回已在事务内完成回冲的退款单。
// 约束：充值单只允许整单退且余额够扣回；退款单、累计退款金额与钱包回冲在同一事务，避免半成品。
func (s *PayRefundService) createWalletRefund(ctx context.Context, order *model.PayOrderModel, in PayRefundCreateInput) (*PayRefundRow, error) {
	if err := s.wallet.CheckRechargeRefund(ctx, order.ID, in.RefundPrice); err != nil {
		return nil, err
	}

	// 待审核：只落退款单占住额度，余额回冲等审核通过后再做
	if in.NeedAudit {
		refund := &model.PayRefundModel{
			No:               newPayRefundNo(),
			AppID:            order.AppID,
			ChannelCode:      order.ChannelCode,
			OrderID:          order.ID,
			OrderNo:          order.No,
			MerchantOrderID:  order.MerchantOrderID,
			MerchantRefundID: in.MerchantRefundID,
			Status:           payEnums.PayRefundStatusCreate,
			PayPrice:         order.Price,
			RefundPrice:      in.RefundPrice,
			Reason:           in.Reason,
			UserIP:           in.UserIP,
			ChannelOrderNo:   order.ChannelOrderNo,
			AuditStatus:      payEnums.PayRefundAuditPending,
		}
		if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
			return s.refunds.Create(ctx, tx, refund)
		}); err != nil {
			return nil, err
		}

		row := decoratePayRefund(*refund)

		return &row, nil
	}

	now := types.DateTime(time.Now())
	refund := &model.PayRefundModel{
		No:               newPayRefundNo(),
		AppID:            order.AppID,
		ChannelCode:      order.ChannelCode,
		OrderID:          order.ID,
		OrderNo:          order.No,
		MerchantOrderID:  order.MerchantOrderID,
		MerchantRefundID: in.MerchantRefundID,
		Status:           payEnums.PayRefundStatusSuccess,
		PayPrice:         order.Price,
		RefundPrice:      in.RefundPrice,
		Reason:           in.Reason,
		UserIP:           in.UserIP,
		ChannelOrderNo:   order.ChannelOrderNo,
		SuccessTime:      &now,
	}
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if err := s.refunds.Create(ctx, tx, refund); err != nil {
			return err
		}

		rows, err := s.orders.MarkRefunded(ctx, tx, order.ID, refund.RefundPrice)
		if err != nil {
			return err
		}
		if rows == 0 {
			return baseErrs.New(payErrs.PayRefundExceedOrderPrice)
		}

		// 钱包不经过渠道，退款成功即在同一事务里回冲余额（充值单扣回、普通支付单退回）
		if err := s.wallet.ApplyWalletRefundInTx(ctx, tx, order, refund.ID, refund.RefundPrice); err != nil {
			return err
		}

		return s.notifyTasks.CreateRefundTask(ctx, tx, refund)
	}); err != nil {
		return nil, err
	}

	row := decoratePayRefund(*refund)

	return &row, nil
}

// ApproveRefund 审核通过退款申请：提交渠道，钱包单直接回冲余额。
// 入参：退款单 ID、审核人与审核意见；返回更新后的退款单。
// 约束：只有待审核的单能通过；审核结果先落库再调渠道，重复点击由条件更新挡住。
func (s *PayRefundService) ApproveRefund(ctx context.Context, id, auditorID types.ID, remark string) (*PayRefundRow, error) {
	refund, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if refund.AuditStatus != payEnums.PayRefundAuditPending {
		return nil, baseErrs.New(payErrs.PayRefundAuditStatusInvalid, refund.AuditStatus.String())
	}

	order, err := s.orders.FindByID(ctx, refund.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, baseErrs.New(payErrs.PayOrderNotExists)
	}

	now := types.DateTime(time.Now())
	refund.AuditStatus = payEnums.PayRefundAuditApproved
	refund.AuditTime = &now
	refund.AuditBy = auditorID
	refund.AuditRemark = strings.TrimSpace(remark)

	// 钱包支付单没有渠道，审核通过即在事务内回冲余额并置为成功
	if refund.ChannelCode == payEnums.ChannelWallet {
		if err := s.approveWalletRefund(ctx, order, refund); err != nil {
			return nil, err
		}

		row := decoratePayRefund(*refund)

		return &row, nil
	}

	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		rows, err := s.refunds.UpdateAudit(ctx, tx, refund)
		if err != nil {
			return err
		}
		if rows == 0 {
			return baseErrs.New(payErrs.PayRefundAuditStatusInvalid, refund.AuditStatus.String())
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// 退款不受渠道开关影响：渠道已停用也要能退历史单
	channel, err := s.channels.FindByID(ctx, order.ChannelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, baseErrs.New(payErrs.PayChannelNotExists)
	}

	gateway, err := NewChannelGateway(channel)
	if err != nil {
		return nil, err
	}

	return s.submitRefund(ctx, order, refund, gateway)
}

// approveWalletRefund 审核通过钱包退款：累计已退金额、回冲余额并登记通知任务。
// 入参：支付单与已带审核结果的退款单；返回错误表示充值单校验或金额条件不通过。
// 约束：审核结果、支付单累计退款与钱包回冲在同一事务，避免出现「审核过了钱没退」。
func (s *PayRefundService) approveWalletRefund(ctx context.Context, order *model.PayOrderModel, refund *model.PayRefundModel) error {
	if err := s.wallet.CheckRechargeRefund(ctx, order.ID, refund.RefundPrice); err != nil {
		return err
	}

	now := types.DateTime(time.Now())
	refund.Status = payEnums.PayRefundStatusSuccess
	refund.SuccessTime = &now

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		rows, err := s.refunds.UpdateAudit(ctx, tx, refund)
		if err != nil {
			return err
		}
		if rows == 0 {
			return baseErrs.New(payErrs.PayRefundAuditStatusInvalid, refund.AuditStatus.String())
		}

		refunded, err := s.orders.MarkRefunded(ctx, tx, order.ID, refund.RefundPrice)
		if err != nil {
			return err
		}
		if refunded == 0 {
			return baseErrs.New(payErrs.PayRefundExceedOrderPrice)
		}

		if err := s.wallet.ApplyWalletRefundInTx(ctx, tx, order, refund.ID, refund.RefundPrice); err != nil {
			return err
		}

		return s.notifyTasks.CreateRefundTask(ctx, tx, refund)
	})
}

// RejectRefund 审核驳回退款申请：退款单直接关闭，不动钱。
// 入参：退款单 ID、审核人与驳回原因；返回错误表示状态不允许或原因为空。
func (s *PayRefundService) RejectRefund(ctx context.Context, id, auditorID types.ID, remark string) error {
	remark = strings.TrimSpace(remark)
	if remark == "" {
		return baseErrs.New(payErrs.PayRefundAuditRemarkRequired)
	}

	refund, err := s.findByID(ctx, id)
	if err != nil {
		return err
	}
	if refund.AuditStatus != payEnums.PayRefundAuditPending {
		return baseErrs.New(payErrs.PayRefundAuditStatusInvalid, refund.AuditStatus.String())
	}

	now := types.DateTime(time.Now())
	refund.AuditStatus = payEnums.PayRefundAuditRejected
	refund.AuditTime = &now
	refund.AuditBy = auditorID
	refund.AuditRemark = remark
	// 关闭退款单，避免一条永远停在「退款订单生成」的记录继续占用可退额度
	refund.Status = payEnums.PayRefundStatusClose

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		rows, err := s.refunds.UpdateAudit(ctx, tx, refund)
		if err != nil {
			return err
		}
		if rows == 0 {
			return baseErrs.New(payErrs.PayRefundAuditStatusInvalid, refund.AuditStatus.String())
		}

		return nil
	})
}

// ConfirmRefund 由退款回调调用：把等待中的退款单置为成功并累计支付单已退金额。
// 入参：退款单、渠道退款单号与回调原文；返回影响行数，0 表示已被其它回调处理。
func (s *PayRefundService) ConfirmRefund(ctx context.Context, refund *model.PayRefundModel, channelRefundNo, notifyData string) (int64, error) {
	now := types.DateTime(time.Now())
	refund.Status = payEnums.PayRefundStatusSuccess
	refund.SuccessTime = &now
	if channelRefundNo != "" {
		refund.ChannelRefundNo = channelRefundNo
	}
	refund.ChannelNotifyData = notifyData

	var affected int64
	if err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		rows, err := s.refunds.UpdateResult(ctx, tx, refund)
		if err != nil {
			return err
		}
		affected = rows
		if rows == 0 {
			return nil
		}

		refundedRows, err := s.orders.MarkRefunded(ctx, tx, refund.OrderID, refund.RefundPrice)
		if err != nil {
			return err
		}
		// 回调阶段才发现超退说明数据异常：回滚退款单状态并报错，便于人工介入
		if refundedRows == 0 {
			return baseErrs.New(payErrs.PayRefundExceedOrderPrice)
		}

		// 退款成功即登记业务通知任务，应用未配置回调地址时内部跳过
		return s.notifyTasks.CreateRefundTask(ctx, tx, refund)
	}); err != nil {
		return 0, err
	}

	return affected, nil
}

// recordRefundError 把渠道退款失败的错误码与文案写进退款单，便于后台排查。
// 入参：退款单与渠道错误；写库失败只忽略，不影响向上返回原始错误。
func (s *PayRefundService) recordRefundError(ctx context.Context, refund *model.PayRefundModel, cause error) {
	refund.Status = payEnums.PayRefundStatusFail
	refund.ChannelErrorCode = "0"
	refund.ChannelErrorMsg = trimMessage(cause.Error())

	var businessErr *baseErrs.BusinessError
	if errorsAs(cause, &businessErr) {
		refund.ChannelErrorCode = strconv.Itoa(businessErr.Code())
		refund.ChannelErrorMsg = businessErr.Msg()
	}

	_ = facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		_, err := s.refunds.UpdateResult(ctx, tx, refund)

		return err
	})
}

// findByID 按主键取退款单，未命中转换成固定业务错误。
func (s *PayRefundService) findByID(ctx context.Context, id types.ID) (*model.PayRefundModel, error) {
	m, err := s.refunds.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayRefundNotExists)
	}

	return m, nil
}

// decoratePayRefund 为退款单补齐渠道中文名与状态文案。
func decoratePayRefund(m model.PayRefundModel) PayRefundRow {
	return PayRefundRow{
		PayRefundModel:  m,
		ChannelName:     payEnums.ChannelLabel(m.ChannelCode),
		StatusName:      m.Status.String(),
		AuditStatusName: m.AuditStatus.String(),
	}
}

// channelRefundNotifyURL 生成渠道退款通知地址：{pay.notify_base_url}/api/v1/pay/callbacks/refund/{渠道主键}。
// 入参：渠道主键；返回完整地址，未配置基础地址时返回空串（微信会回落到商户平台配置）。
func channelRefundNotifyURL(channelID types.ID) string {
	base := strings.TrimRight(facades.Config().GetString("pay.notify_base_url", ""), "/")
	if base == "" {
		return ""
	}

	return base + "/api/v1/pay/callbacks/refund/" + channelID.String()
}

// newPayRefundNo 生成退款单号：R + 秒级时间 + 雪花 ID 末 6 位，与支付单号同规则。
// 入参：无；返回形如 R20261022153000123456 的退款单号。
func newPayRefundNo() string {
	id := support.NewID().String()

	return "R" + time.Now().Format("20060102150405") + id[len(id)-6:]
}
