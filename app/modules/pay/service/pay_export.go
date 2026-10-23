// pay_export.go 提供支付域三张账单的 CSV 导出：支付单、退款单、钱包流水。
// 财务对账按「支付 - 退款 - 余额变动」三张表核对，导出列固定，不随页面列设置变化。
package service

import (
	"context"
	"strconv"
	"strings"

	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	payEnums "eqray-flavor/app/modules/pay/enums"
	"eqray-flavor/app/modules/pay/repository"
)

// Export 按筛选条件导出支付单 CSV。
// 入参：与列表一致的筛选条件；返回文件字节，单次最多导出 support.ExportRowLimit 条。
func (s *PayOrderService) Export(ctx context.Context, q PayOrderListQuery) ([]byte, error) {
	filter := repository.PayOrderFilter{
		Keyword:     strings.TrimSpace(q.Keyword),
		AppID:       q.AppID,
		ChannelCode: strings.TrimSpace(q.ChannelCode),
	}
	if q.Status != nil {
		status := int8(*q.Status)
		filter.Status = &status
	}

	rows, _, err := s.orders.Paginate(ctx, filter, 1, support.ExportRowLimit)
	if err != nil {
		return nil, err
	}

	header := []string{
		"支付单号", "商户订单号", "渠道", "支付状态", "商品标题",
		"支付金额(元)", "渠道手续费(元)", "已退金额(元)", "付款用户", "渠道订单号",
		"支付时间", "失效时间", "创建时间",
	}
	list := make([][]string, 0, len(rows))
	for i := range rows {
		m := &rows[i]
		list = append(list, []string{
			m.No,
			m.MerchantOrderID,
			payEnums.ChannelLabel(m.ChannelCode),
			m.Status.String(),
			m.Subject,
			support.FormatYuan(m.Price),
			support.FormatYuan(m.ChannelFeePrice),
			support.FormatYuan(m.RefundPrice),
			strconv.FormatInt(int64(m.UserID), 10),
			m.ChannelOrderNo,
			formatExportTime(m.SuccessTime),
			m.ExpireTime.String(),
			m.CreatedAt.String(),
		})
	}

	return support.BuildCSV(header, list)
}

// Export 按筛选条件导出售后退款单 CSV，含审核状态便于对账溯源。
// 入参：与列表一致的筛选条件；返回文件字节。
func (s *PayRefundService) Export(ctx context.Context, q PayRefundListQuery) ([]byte, error) {
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

	rows, _, err := s.refunds.Paginate(ctx, filter, 1, support.ExportRowLimit)
	if err != nil {
		return nil, err
	}

	header := []string{
		"退款单号", "支付单号", "商户订单号", "商户退款单号", "渠道", "退款状态", "审核状态",
		"退款金额(元)", "原支付金额(元)", "退款原因", "审核意见", "渠道退款单号",
		"退款成功时间", "创建时间",
	}
	list := make([][]string, 0, len(rows))
	for i := range rows {
		m := &rows[i]
		list = append(list, []string{
			m.No,
			m.OrderNo,
			m.MerchantOrderID,
			m.MerchantRefundID,
			payEnums.ChannelLabel(m.ChannelCode),
			m.Status.String(),
			m.AuditStatus.String(),
			support.FormatYuan(m.RefundPrice),
			support.FormatYuan(m.PayPrice),
			m.Reason,
			m.AuditRemark,
			m.ChannelRefundNo,
			formatExportTime(m.SuccessTime),
			m.CreatedAt.String(),
		})
	}

	return support.BuildCSV(header, list)
}

// ExportTransactions 导出某个钱包的流水 CSV，用于会员余额对账。
// 入参：钱包流水筛选条件（必须带 WalletID）；返回文件字节。
func (s *PayWalletService) ExportTransactions(ctx context.Context, filter repository.PayWalletTransactionFilter) ([]byte, error) {
	rows, _, err := s.wallets.PaginateTransactions(ctx, filter, 1, support.ExportRowLimit)
	if err != nil {
		return nil, err
	}

	header := []string{
		"流水ID", "钱包ID", "业务类型", "收支方向", "变动金额(元)", "变动后余额(元)", "业务编号", "备注", "时间",
	}
	list := make([][]string, 0, len(rows))
	for i := range rows {
		m := &rows[i]
		list = append(list, []string{
			strconv.FormatInt(int64(m.ID), 10),
			strconv.FormatInt(int64(m.WalletID), 10),
			m.BizType.String(),
			m.Type.String(),
			support.FormatYuan(m.Price),
			support.FormatYuan(m.Balance),
			m.BizID,
			m.Remark,
			m.CreatedAt.String(),
		})
	}

	return support.BuildCSV(header, list)
}

// formatExportTime 格式化可空时间，空值导出空串而不是 null。
func formatExportTime(t *types.DateTime) string {
	if t == nil || t.IsZero() {
		return ""
	}

	return t.String()
}
