import request from '@/utils/http'

// 分页查询打印模板：门店范围由请求头 X-Store-Id 或当前用户可见门店集合决定
export function fetchGetPrintTemplateList(params: Api.Store.PrintTemplateSearchParams) {
  return request.get<Api.Store.PrintTemplateList>({
    url: '/api/v1/store/print-templates',
    params
  })
}

// 新增打印模板：同一门店下设备类型与打印类型唯一，初始版本由后端设为 1
export function fetchCreatePrintTemplate(data: Api.Store.PrintTemplateSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/store/print-templates',
    params: data
  })
}

// 编辑打印模板：正文变化时后端自动递增版本号，客户端不提交版本
export function fetchUpdatePrintTemplate(id: string, data: Api.Store.PrintTemplateSaveParams) {
  return request.put<null>({
    url: `/api/v1/store/print-templates/${id}`,
    params: data
  })
}

// 删除打印模板：历史打印任务的模板快照不依赖本表
export function fetchDeletePrintTemplate(id: string) {
  return request.del<null>({
    url: `/api/v1/store/print-templates/${id}`
  })
}
