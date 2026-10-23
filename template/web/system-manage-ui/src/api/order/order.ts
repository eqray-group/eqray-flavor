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

// 修改订单备注：顾客备注与商家备注分列保存
export function fetchUpdateOrderRemark(id: string, data: Api.Order.OrderRemarkParams) {
  return request.put<null>({
    url: `/api/v1/order/orders/${id}/remark`,
    params: data
  })
}

// 订单改价：delta 为差额（分），正数加价、负数减免
export function fetchAdjustOrderPrice(id: string, data: Api.Order.OrderAdjustPriceParams) {
  return request.put<null>({
    url: `/api/v1/order/orders/${id}/price`,
    params: data
  })
}

// 查询可派单骑手：仅返回该订单门店下在线接单中的骑手
export function fetchGetOrderRiders(id: string) {
  return request.get<Api.Order.OrderRiderOption[]>({
    url: `/api/v1/order/orders/${id}/riders`
  })
}

// 订单派单：给外卖订单指派骑手
export function fetchAssignOrderRider(id: string, data: Api.Order.OrderAssignRiderParams) {
  return request.post<null>({
    url: `/api/v1/order/orders/${id}/rider`,
    params: data
  })
}

// 取消派单：清空骑手并归还接单额度
export function fetchCancelOrderRider(id: string) {
  return request.del<null>({
    url: `/api/v1/order/orders/${id}/rider`
  })
}

// 查询小票数据：用于前台补打 / 重打
export function fetchGetOrderTicket(id: string) {
  return request.get<Api.Order.OrderTicket>({
    url: `/api/v1/order/orders/${id}/ticket`
  })
}

// 导出订单 CSV：返回二进制流，由调用方触发浏览器下载
export function fetchExportOrders(params: Api.Order.OrderSearchParams) {
  return request.download({
    url: '/api/v1/order/orders/export',
    params
  })
}
