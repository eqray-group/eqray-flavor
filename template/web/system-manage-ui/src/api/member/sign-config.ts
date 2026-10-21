import request from '@/utils/http'

// 获取签到奖励档位列表
export function fetchGetSignConfigList(params: Api.Member.SignConfigSearchParams) {
  return request.get<Api.Member.SignConfigList>({
    url: '/api/v1/member/sign-configs',
    params
  })
}

// 新增签到档位
export function fetchCreateSignConfig(data: Api.Member.SignConfigSaveParams) {
  return request.post<null>({
    url: '/api/v1/member/sign-configs',
    params: data
  })
}

// 编辑签到档位
export function fetchUpdateSignConfig(id: string, data: Api.Member.SignConfigSaveParams) {
  return request.put<null>({
    url: `/api/v1/member/sign-configs/${id}`,
    params: data
  })
}

// 删除签到档位
export function fetchDeleteSignConfig(id: string) {
  return request.del<null>({
    url: `/api/v1/member/sign-configs/${id}`
  })
}
