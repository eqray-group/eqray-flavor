import request from '@/utils/http'

// 分页查询支付应用：按应用名 / appKey 关键字与开启状态筛选
export function fetchGetPayAppList(params: Api.Pay.AppSearchParams) {
  return request.get<Api.Pay.AppList>({
    url: '/api/v1/pay/apps',
    params
  })
}

// 查询开启状态的支付应用，供渠道配置与代下单页面的应用下拉使用
export function fetchGetPayAppOptions() {
  return request.get<Api.Pay.AppOption[]>({
    url: '/api/v1/pay/apps/options'
  })
}

// 查询支付应用详情：密钥字段后端只返回脱敏值
export function fetchGetPayAppDetail(id: string) {
  return request.get<Api.Pay.AppItem>({
    url: `/api/v1/pay/apps/${id}`
  })
}

// 新增支付应用：appKey 与 appSecret 由后端生成，明文密钥只在响应里返回一次
export function fetchCreatePayApp(data: Api.Pay.AppSaveParams) {
  return request.post<Api.Pay.AppItem>({
    url: '/api/v1/pay/apps',
    params: data
  })
}

// 编辑支付应用：appKey 不可修改，appSecret 留空表示保持原值
export function fetchUpdatePayApp(id: string, data: Api.Pay.AppSaveParams) {
  return request.put<null>({
    url: `/api/v1/pay/apps/${id}`,
    params: data
  })
}

// 删除支付应用：应用下仍有渠道配置时后端会拒绝
export function fetchDeletePayApp(id: string) {
  return request.del<null>({
    url: `/api/v1/pay/apps/${id}`
  })
}
