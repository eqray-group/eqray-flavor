import request from '@/utils/http'

// 获取标签列表
export function fetchGetTagList(params: Api.Member.TagSearchParams) {
  return request.get<Api.Member.TagList>({
    url: '/api/v1/member/tags',
    params
  })
}

// 获取启用中的标签（会员表单多选）
export function fetchGetTagOptions() {
  return request.get<Api.Member.TagItem[]>({
    url: '/api/v1/member/tags/options'
  })
}

// 新增标签
export function fetchCreateTag(data: Api.Member.TagSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/member/tags',
    params: data
  })
}

// 编辑标签
export function fetchUpdateTag(id: string, data: Api.Member.TagSaveParams) {
  return request.put<null>({
    url: `/api/v1/member/tags/${id}`,
    params: data
  })
}

// 删除标签
export function fetchDeleteTag(id: string) {
  return request.del<null>({
    url: `/api/v1/member/tags/${id}`
  })
}
