import { http } from '@/utils/http'

// 会员检索：收银台按手机号或昵称找人，mobile / nickname 都是模糊匹配
export function fetchMemberPage(params: Record<string, unknown>) {
  return http.get<Api.Member.Page>('/api/v1/member/users', params)
}
