import request from '@/utils/http'

// 获取短信渠道列表
export function fetchGetSmsChannelList(params: Api.Infra.SmsChannelSearchParams) {
  return request.get<Api.Infra.SmsChannelList>({
    url: '/api/v1/infra/sms-channels',
    params
  })
}

// 获取启用中的短信渠道（模板表单下拉）
export function fetchGetSmsChannelOptions() {
  return request.get<Api.Infra.SmsChannelItem[]>({
    url: '/api/v1/infra/sms-channels/options'
  })
}

// 新增短信渠道
export function fetchCreateSmsChannel(data: Api.Infra.SmsChannelSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/infra/sms-channels',
    params: data
  })
}

// 编辑短信渠道
export function fetchUpdateSmsChannel(id: string, data: Api.Infra.SmsChannelSaveParams) {
  return request.put<null>({
    url: `/api/v1/infra/sms-channels/${id}`,
    params: data
  })
}

// 删除短信渠道
export function fetchDeleteSmsChannel(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/sms-channels/${id}`
  })
}
