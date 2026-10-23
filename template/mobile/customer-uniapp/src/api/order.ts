/**
 * 下单与订单接口。金额一律以服务端试算为准，前端不自行计算优惠。
 */
import alovaInstance from './core/instance'

export interface OrderLineInput {
  skuId: string
  quantity: number
  addons: { addonItemId: string; quantity: number }[]
  remark?: string
}

export interface EstimateLine {
  skuId: string
  spuName: string
  skuName: string
  quantity: number
  unitPrice: number
  subtotal: number
  payable: number
  addons: { addonItemId: string; name: string; price: number; quantity: number }[]
}

export interface EstimateResult {
  lines: EstimateLine[]
  itemAmount: number
  addonAmount: number
  lineDiscountAmount: number
  activityAmount: number
  couponAmount: number
  pointsAmount: number
  payAmount: number
  activityNames: string[]
  couponName: string
  notes: string[]
}

export interface CreateOrderPayload {
  type: number
  tableId?: string
  remark?: string
  requestNo?: string
  items: OrderLineInput[]
}

export interface CreateOrderResult {
  order: {
    id: string
    no: string
    status: number
    statusText: string
    payAmount: number
    pickupCode: string
    expireTime: string
  }
  idempotentHit: boolean
}

export interface OrderItem {
  spuName: string
  skuName: string
  quantity: number
  amount?: number
}

export interface OrderRow {
  id: string
  no: string
  status: number
  statusText: string
  type: number
  typeText: string
  payStatus: number
  payAmount: number
  pickupCode: string
  tableName?: string
  createdAt: string
  items?: OrderItem[]
}

export interface OrderPage {
  list: OrderRow[]
  total: number
}

// 试算：明细是嵌套结构，必须走请求体
export function fetchEstimate(data: { items: OrderLineInput[], usePoints?: boolean }) {
  return alovaInstance.Post<EstimateResult>('/api/v1/client/order/estimates', data)
}

// 下单：requestNo 由调用方生成，失败重试复用同一个号，保证幂等
export function createOrder(data: CreateOrderPayload) {
  return alovaInstance.Post<CreateOrderResult>('/api/v1/client/order/orders', data)
}

// 我的订单
export function fetchOrders(params: { pageNo: number, pageSize: number }) {
  return alovaInstance.Get<OrderPage>('/api/v1/client/order/orders', { params })
}

// 发起支付：返回调起支付所需参数
export function payOrder(id: string, data: { channelCode?: string }) {
  return alovaInstance.Post<Record<string, unknown>>(`/api/v1/client/order/orders/${id}/pay`, data)
}
