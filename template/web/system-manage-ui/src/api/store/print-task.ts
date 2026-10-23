import request from '@/utils/http'

// 分页查询打印任务：门店范围由请求头 X-Store-Id 或当前用户可见门店集合决定
export function fetchGetPrintTaskList(params: Api.Store.PrintTaskSearchParams) {
  return request.get<Api.Store.PrintTaskList>({
    url: '/api/v1/store/print-tasks',
    params
  })
}

// 查询打印任务详情：返回渲染后的正文，便于核对打印内容
export function fetchGetPrintTaskDetail(id: string) {
  return request.get<Api.Store.PrintTaskItem>({
    url: `/api/v1/store/print-tasks/${id}`
  })
}

// 重打小票：任务重置为待打印并累加重打次数，正文沿用生成时的快照
export function fetchReprintPrintTask(id: string) {
  return request.post<null>({
    url: `/api/v1/store/print-tasks/${id}/reprint`
  })
}
