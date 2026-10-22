import request from '@/utils/http'

// 分页查询骑手档案：门店范围由请求头 X-Store-Id 或当前用户可见门店集合决定
export function fetchGetRiderList(params: Api.Store.RiderSearchParams) {
  return request.get<Api.Store.RiderList>({
    url: '/api/v1/store/riders',
    params
  })
}

// 搜索可绑定为骑手的启用后台账号：新增弹窗选择已有账号时使用
export function fetchSearchRiderUsers(keyword: string) {
  return request.get<Api.Store.RiderUserOption[]>({
    url: '/api/v1/store/riders/user-options',
    params: { keyword }
  })
}

// 新增骑手档案：未传 userId 时后端自动开通账号并返回一次性初始密码
export function fetchCreateRider(data: Api.Store.RiderCreateParams) {
  return request.post<Api.Store.RiderCreateResult>({
    url: '/api/v1/store/riders',
    params: data
  })
}

// 编辑骑手档案：只允许调整并发上限、状态与备注
export function fetchUpdateRider(id: string, data: Api.Store.RiderUpdateParams) {
  return request.put<null>({
    url: `/api/v1/store/riders/${id}`,
    params: data
  })
}

// 删除骑手档案：仍有进行中配送单时后端会拒绝
export function fetchDeleteRider(id: string) {
  return request.del<null>({
    url: `/api/v1/store/riders/${id}`
  })
}
