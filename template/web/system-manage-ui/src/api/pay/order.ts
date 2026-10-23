import request from '@/utils/http'

// 分页查询支付单：支持单号关键字、应用、渠道与支付状态筛选
export function fetchGetPayOrderList(params: Api.Pay.OrderSearchParams) {
  return request.get<Api.Pay.OrderList>({
    url: '/api/v1/pay/orders',
    params
  })
}

// 查询支付单详情：带出每次渠道调用的拓展记录，便于排查下单报错
export function fetchGetPayOrderDetail(id: string) {
  return request.get<Api.Pay.OrderDetail>({
    url: `/api/v1/pay/orders/${id}`
  })
}

// 创建支付单并立即调起渠道：金额单位为分，只传业务侧试算结果
export function fetchCreatePayOrder(data: Api.Pay.OrderCreateParams) {
  return request.post<Api.Pay.OrderSubmitResult>({
    url: '/api/v1/pay/orders',
    params: data
  })
}

// 重新调起仍待支付的支付单：每次调用都会新增一条渠道调用记录
export function fetchSubmitPayOrder(id: string, data: Api.Pay.OrderSubmitParams) {
  return request.post<Api.Pay.OrderSubmitResult>({
    url: `/api/v1/pay/orders/${id}/submit`,
    params: data
  })
}

// 主动查询渠道订单并补齐本地状态：回调丢失时用于人工补单
export function fetchSyncPayOrder(id: string) {
  return request.post<Api.Pay.OrderItem>({
    url: `/api/v1/pay/orders/${id}/sync`
  })
}

// 关闭待支付的支付单：已支付或已退款的单后端会拒绝
export function fetchClosePayOrder(id: string) {
  return request.post<null>({
    url: `/api/v1/pay/orders/${id}/close`
  })
}
