import request from '@/utils/http'

// 获取短信模板列表
export function fetchGetSmsTemplateList(params: Api.Infra.SmsTemplateSearchParams) {
  return request.get<Api.Infra.SmsTemplateList>({
    url: '/api/v1/infra/sms-templates',
    params
  })
}

// 新增短信模板
export function fetchCreateSmsTemplate(data: Api.Infra.SmsTemplateSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/infra/sms-templates',
    params: data
  })
}

// 编辑短信模板
export function fetchUpdateSmsTemplate(id: string, data: Api.Infra.SmsTemplateSaveParams) {
  return request.put<null>({
    url: `/api/v1/infra/sms-templates/${id}`,
    params: data
  })
}

// 删除短信模板
export function fetchDeleteSmsTemplate(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/sms-templates/${id}`
  })
}
