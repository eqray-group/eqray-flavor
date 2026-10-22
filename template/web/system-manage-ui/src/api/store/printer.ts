import request from '@/utils/http'

// 分页查询打印设备：门店范围由请求头 X-Store-Id 或当前用户可见门店集合决定
export function fetchGetPrinterList(params: Api.Store.PrinterSearchParams) {
  return request.get<Api.Store.PrinterList>({
    url: '/api/v1/store/printers',
    params
  })
}

// 新增打印设备：设备号在同一门店内唯一，密钥由后端加密保存
export function fetchCreatePrinter(data: Api.Store.PrinterSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/store/printers',
    params: data
  })
}

// 编辑打印设备：secret 留空表示沿用原密钥，设备不能跨门店移动
export function fetchUpdatePrinter(id: string, data: Api.Store.PrinterSaveParams) {
  return request.put<null>({
    url: `/api/v1/store/printers/${id}`,
    params: data
  })
}

// 删除打印设备：删除后释放该门店的设备号占用
export function fetchDeletePrinter(id: string) {
  return request.del<null>({
    url: `/api/v1/store/printers/${id}`
  })
}
