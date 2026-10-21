import request from '@/utils/http'

// 获取会员等级列表
export function fetchGetLevelList(params: Api.Member.LevelSearchParams) {
  return request.get<Api.Member.LevelList>({
    url: '/api/v1/member/levels',
    params
  })
}

// 获取启用中的等级（会员表单下拉）
export function fetchGetLevelOptions() {
  return request.get<Api.Member.LevelItem[]>({
    url: '/api/v1/member/levels/options'
  })
}

// 新增会员等级
export function fetchCreateLevel(data: Api.Member.LevelSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/member/levels',
    params: data
  })
}

// 编辑会员等级
export function fetchUpdateLevel(id: string, data: Api.Member.LevelSaveParams) {
  return request.put<null>({
    url: `/api/v1/member/levels/${id}`,
    params: data
  })
}

// 删除会员等级
export function fetchDeleteLevel(id: string) {
  return request.del<null>({
    url: `/api/v1/member/levels/${id}`
  })
}
