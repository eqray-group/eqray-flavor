import request from '@/utils/http'

// 分页查询订单：支持订单号关键字、门店、订单状态、支付状态与类型筛选
export function fetchGetOrderList(params: Api.Order.OrderSearchParams) {
  return request.get<Api.Order.OrderList>({
    url: '/api/v1/order/orders',
    params
  })
}

// 查询订单详情：带出明细、加料快照与状态流转日志
export function fetchGetOrderDetail(id: string) {
  return request.get<Api.Order.OrderDetail>({
    url: `/api/v1/order/orders/${id}`
  })
}

// 流转订单状态：接单、出餐、完成、拒单等，合法流转由后端状态机校验
export function fetchTransitOrder(id: string, data: Api.Order.OrderTransitParams) {
  return request.post<null>({
    url: `/api/v1/order/orders/${id}/transit`,
    params: data
  })
}

// 取消订单：仅待支付状态可由店员取消
export function fetchCancelOrder(id: string, reason?: string) {
  return request.post<null>({
    url: `/api/v1/order/orders/${id}/cancel`,
    params: { reason }
  })
}
