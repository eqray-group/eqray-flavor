import request from '@/utils/http'

// 获取会员列表
export function fetchGetMemberList(params: Api.Member.MemberSearchParams) {
  return request.get<Api.Member.MemberList>({
    url: '/api/v1/member/users',
    params
  })
}

// 获取会员详情
export function fetchGetMemberDetail(id: string) {
  return request.get<Api.Member.MemberItem>({
    url: `/api/v1/member/users/${id}`
  })
}

// 新增会员
export function fetchCreateMember(data: Api.Member.MemberSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/member/users',
    params: data
  })
}

// 编辑会员
export function fetchUpdateMember(id: string, data: Api.Member.MemberSaveParams) {
  return request.put<null>({
    url: `/api/v1/member/users/${id}`,
    params: data
  })
}

// 启用 / 禁用会员
export function fetchUpdateMemberStatus(id: string, status: number) {
  return request.put<null>({
    url: `/api/v1/member/users/${id}/status`,
    params: { status }
  })
}

// 删除会员（软删除）
export function fetchDeleteMember(id: string) {
  return request.del<null>({
    url: `/api/v1/member/users/${id}`
  })
}

// 调整会员积分
export function fetchAdjustMemberPoint(id: string, data: Api.Member.PointAdjustParams) {
  return request.post<null>({
    url: `/api/v1/member/users/${id}/point`,
    params: data
  })
}

// 获取会员积分记录
export function fetchGetMemberPointRecords(
  id: string,
  params: { pageNo?: number; pageSize?: number }
) {
  return request.get<Api.Member.PointRecordList>({
    url: `/api/v1/member/users/${id}/points`,
    params
  })
}

// 调整会员等级
export function fetchAdjustMemberLevel(id: string, data: Api.Member.LevelAdjustParams) {
  return request.put<null>({
    url: `/api/v1/member/users/${id}/level`,
    params: data
  })
}

// 获取会员等级记录
export function fetchGetMemberLevelRecords(
  id: string,
  params: { pageNo?: number; pageSize?: number }
) {
  return request.get<Api.Member.LevelRecordList>({
    url: `/api/v1/member/users/${id}/level-records`,
    params
  })
}

// 获取会员签到记录
export function fetchGetMemberSignRecords(
  id: string,
  params: { pageNo?: number; pageSize?: number }
) {
  return request.get<Api.Member.SignRecordList>({
    url: `/api/v1/member/users/${id}/sign-records`,
    params
  })
}

// 获取会员收货地址
export function fetchGetMemberAddresses(id: string) {
  return request.get<Api.Member.AddressItem[]>({
    url: `/api/v1/member/users/${id}/addresses`
  })
}

// 获取会员第三方账号绑定（后台只读）
export function fetchGetMemberSocialBindings(id: string) {
  return request.get<Api.Member.SocialBindingItem[]>({
    url: `/api/v1/member/users/${id}/social-bindings`
  })
}
