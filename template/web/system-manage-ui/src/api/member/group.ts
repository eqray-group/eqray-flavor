import request from '@/utils/http'

// 获取分组列表
export function fetchGetGroupList(params: Api.Member.GroupSearchParams) {
  return request.get<Api.Member.GroupList>({
    url: '/api/v1/member/groups',
    params
  })
}

// 获取启用中的分组（会员表单下拉）
export function fetchGetGroupOptions() {
  return request.get<Api.Member.GroupItem[]>({
    url: '/api/v1/member/groups/options'
  })
}

// 新增分组
export function fetchCreateGroup(data: Api.Member.GroupSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/member/groups',
    params: data
  })
}

// 编辑分组
export function fetchUpdateGroup(id: string, data: Api.Member.GroupSaveParams) {
  return request.put<null>({
    url: `/api/v1/member/groups/${id}`,
    params: data
  })
}

// 删除分组
export function fetchDeleteGroup(id: string) {
  return request.del<null>({
    url: `/api/v1/member/groups/${id}`
  })
}
