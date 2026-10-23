import { http } from '@/utils/http'

/** 后厨看板订单 */
export interface KdsOrder {
  id: string
  no: string
  status: number
  statusText: string
  type: number
  typeText: string
  tableName: string
  pickupCode: string
  remark: string
  createdAt: string
  payTime?: string | null
  items: {
    spuName: string
    skuName: string
    quantity: number
    remark: string
    orderer: string
    addons: string[]
  }[]
}

// 一次取回待接单 / 制作中 / 待取餐的全部订单与明细
export function fetchKdsBoard() {
  return http.get<KdsOrder[]>('/api/v1/order/pos/kds')
}

// 出餐等状态流转：to 为目标状态（300 制作中 / 400 待取餐 / 500 已完成）
export function transitOrder(id: string, to: number, reason = '') {
  return http.post<null>(`/api/v1/order/orders/${id}/transit`, { to, reason })
}
