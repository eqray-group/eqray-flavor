import { http, request } from '@/utils/http'

// 订单分页：门店由请求头 X-Store-Id 决定，这里不传 storeId
export function fetchOrderPage(params: Api.Order.PageParams) {
  return http.get<Api.Order.Page>('/api/v1/order/orders', params as Record<string, unknown>)
}

// 订单详情：含明细与状态时间轴
export function fetchOrderDetail(id: string) {
  return http.get<Record<string, unknown>>(`/api/v1/order/orders/${id}`)
}

// 状态流转：接单 / 出餐 / 完成
export function fetchTransitOrder(id: string, to: number, reason?: string) {
  return http.post<null>(`/api/v1/order/orders/${id}/transit`, { to, reason })
}

// 取消订单：仅待支付状态可用
export function fetchCancelOrder(id: string, reason?: string) {
  return http.post<null>(`/api/v1/order/orders/${id}/cancel`, { reason })
}

// 门店今日经营概览：营收、客单价、退款与各状态单量
export function fetchPosOverview() {
  return http.get<Api.Order.PosOverview>('/api/v1/order/pos/overview')
}

// 门店经营报表：趋势 + 收款构成 + 菜品排行
export function fetchPosReport(days: number, top = 10) {
  return http.get<Api.Order.PosReport>('/api/v1/order/pos/reports', { days, top })
}

// POS 代客试算：明细是嵌套结构，走 JSON 请求体
export function fetchPosEstimate(data: Record<string, unknown>) {
  return http.postBody<Api.Order.EstimateResult>('/api/v1/order/pos/estimates', data)
}

// POS 代客下单：散客不传 memberId
export function fetchPosCreateOrder(data: Record<string, unknown>) {
  return http.postBody<Api.Order.CreateResult>('/api/v1/order/pos/orders', data)
}

// 现金收款：不生成支付单，直接置订单已支付
export function fetchSettleCash(id: string) {
  return http.post<null>(`/api/v1/order/pos/orders/${id}/settle-cash`)
}

// 扫码收款：微信返回二维码，支付宝传付款码
export function fetchPosPay(id: string, payType: number, authCode?: string) {
  return http.post<Api.Order.PayResult>(`/api/v1/order/pos/orders/${id}/pay`, {
    payType,
    authCode: authCode || ''
  })
}

// 收款状态轮询：微信扫码用
export function fetchPosPayStatus(id: string) {
  return http.get<Api.Order.PayStatusResult>(`/api/v1/order/pos/orders/${id}/pay-status`)
}
