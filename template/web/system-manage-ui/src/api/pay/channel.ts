import request from '@/utils/http'

// 分页查询支付渠道：支持按应用、渠道编码与开启状态筛选
export function fetchGetPayChannelList(params: Api.Pay.ChannelSearchParams) {
  return request.get<Api.Pay.ChannelList>({
    url: '/api/v1/pay/channels',
    params
  })
}

// 查询后端支持的渠道编码清单，供新增 / 编辑弹窗的下拉使用
export function fetchGetPayChannelOptions() {
  return request.get<Api.Pay.ChannelOption[]>({
    url: '/api/v1/pay/channels/options'
  })
}

// 查询支付渠道详情：config 中的密钥字段只返回掩码
export function fetchGetPayChannelDetail(id: string) {
  return request.get<Api.Pay.ChannelItem>({
    url: `/api/v1/pay/channels/${id}`
  })
}

// 新增支付渠道：同一应用下渠道编码不可重复，config 必须是 JSON 对象
export function fetchCreatePayChannel(data: Api.Pay.ChannelSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/pay/channels',
    params: data
  })
}

// 编辑支付渠道：应用与渠道编码不可变更，掩码字段沿用原密钥
export function fetchUpdatePayChannel(id: string, data: Api.Pay.ChannelSaveParams) {
  return request.put<null>({
    url: `/api/v1/pay/channels/${id}`,
    params: data
  })
}

// 删除支付渠道：历史支付单仍保留渠道快照
export function fetchDeletePayChannel(id: string) {
  return request.del<null>({
    url: `/api/v1/pay/channels/${id}`
  })
}
