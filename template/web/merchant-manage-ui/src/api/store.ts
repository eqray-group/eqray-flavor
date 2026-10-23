import { http } from '@/utils/http'

// 当前用户可见门店：登录后用于选择 / 校验门店上下文
export function fetchVisibleStores() {
  return http.get<Api.Store.VisibleItem[]>('/api/v1/store/stores/visible')
}

// 打印任务分页：门店由请求头决定
export function fetchPrintTaskPage(params: Record<string, unknown>) {
  return http.get<Api.Store.PrintTaskPage>('/api/v1/store/print-tasks', params)
}

// 桌台列表：堂食点单需要选台
export function fetchTablePage(params: Record<string, unknown>) {
  return http.get<Api.Store.TablePage>('/api/v1/store/tables', params)
}

// 开台并绑定台位单：桌台图由下单动作触发，避免出现「开了台却没有单」的空台
export function fetchOpenTable(id: string, currentOrderId: string) {
  return http.post<null>(`/api/v1/store/tables/${id}/open`, { currentOrderId })
}

// 清台：使用中或待清台恢复空闲
export function fetchClearTable(id: string) {
  return http.post<null>(`/api/v1/store/tables/${id}/clear`)
}

// 重打小票：任务重置为待打印
export function fetchReprintPrintTask(id: string) {
  return http.post<null>(`/api/v1/store/print-tasks/${id}/reprint`)
}
