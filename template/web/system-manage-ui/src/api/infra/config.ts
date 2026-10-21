import request from '@/utils/http'

/**
 * 获取参数配置列表
 * @param params 分组、名称关键字与分页参数
 */
export function fetchGetConfigList(params: Api.Infra.ConfigSearchParams) {
  return request.get<Api.Infra.ConfigList>({
    url: '/api/v1/infra/configs',
    params
  })
}

/**
 * 新增参数配置
 * @param data 参数分组、名称、键、值等信息
 */
export function fetchCreateConfig(data: Api.Infra.ConfigCreateParams) {
  return request.post<null>({
    url: '/api/v1/infra/configs',
    params: data
  })
}

/**
 * 编辑参数配置
 * @param id 参数 ID
 * @param data 可变字段，参数键与类型不可修改
 */
export function fetchUpdateConfig(id: string, data: Api.Infra.ConfigUpdateParams) {
  return request.put<null>({
    url: `/api/v1/infra/configs/${id}`,
    params: data
  })
}

/**
 * 删除参数配置
 * @param id 参数 ID
 */
export function fetchDeleteConfig(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/configs/${id}`
  })
}
