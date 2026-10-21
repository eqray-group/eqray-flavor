import request from '@/utils/http'

// 获取第三方登录渠道配置列表
export function fetchGetSocialClientList(params: Api.Member.SocialClientSearchParams) {
  return request.get<Api.Member.SocialClientList>({
    url: '/api/v1/member/social-clients',
    params
  })
}

// 新增第三方登录渠道配置
export function fetchCreateSocialClient(data: Api.Member.SocialClientSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/member/social-clients',
    params: data
  })
}

// 编辑第三方登录渠道配置
export function fetchUpdateSocialClient(id: string, data: Api.Member.SocialClientSaveParams) {
  return request.put<null>({
    url: `/api/v1/member/social-clients/${id}`,
    params: data
  })
}

// 删除第三方登录渠道配置
export function fetchDeleteSocialClient(id: string) {
  return request.del<null>({
    url: `/api/v1/member/social-clients/${id}`
  })
}
