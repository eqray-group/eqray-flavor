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
