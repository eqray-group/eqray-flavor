import request from '@/utils/http'

// 分页查询退款单：支持单号关键字、应用、支付单、渠道与退款状态筛选
export function fetchGetPayRefundList(params: Api.Pay.RefundSearchParams) {
  return request.get<Api.Pay.RefundList>({
    url: '/api/v1/pay/refunds',
    params
  })
}

// 查询退款单详情：含渠道退款单号、渠道报错与回调原文
export function fetchGetPayRefundDetail(id: string) {
  return request.get<Api.Pay.RefundItem>({
    url: `/api/v1/pay/refunds/${id}`
  })
}

// 发起退款：后端校验可退金额后调用渠道，渠道同步确认完成的直接置为成功
export function fetchCreatePayRefund(data: Api.Pay.RefundCreateParams) {
  return request.post<Api.Pay.RefundItem>({
    url: '/api/v1/pay/refunds',
    params: data
  })
}

// 审核通过退款：钱包单回冲余额，渠道单提交退款
export function fetchApprovePayRefund(id: string, data: Api.Pay.RefundAuditParams) {
  return request.put<Api.Pay.RefundItem>({
    url: `/api/v1/pay/refunds/${id}/approve`,
    params: data
  })
}

// 审核驳回退款：退款单关闭，不动钱，原因必填
export function fetchRejectPayRefund(id: string, data: Api.Pay.RefundAuditParams) {
  return request.put<null>({
    url: `/api/v1/pay/refunds/${id}/reject`,
    params: data
  })
}

// 导出售后退款单 CSV：含审核状态，返回二进制流
export function fetchExportPayRefunds(params: Api.Pay.RefundSearchParams) {
  return request.download({
    url: '/api/v1/pay/refunds/export',
    params
  })
}
