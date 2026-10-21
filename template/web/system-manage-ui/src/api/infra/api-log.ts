import request from '@/utils/http'

/** 获取 API 访问日志 */
export function fetchGetApiAccessLog(params: Api.Infra.ApiLogSearchParams) {
  return request.get<Api.Infra.ApiAccessLogList>({
    url: '/api/v1/infra/api-logs/access',
    params
  })
}

/** 获取 API 错误日志 */
export function fetchGetApiErrorLog(params: Api.Infra.ApiLogSearchParams) {
  return request.get<Api.Infra.ApiErrorLogList>({
    url: '/api/v1/infra/api-logs/error',
    params
  })
}

/** 标记错误日志为已处理 */
export function fetchProcessApiErrorLog(id: string) {
  return request.put<null>({
    url: `/api/v1/infra/api-logs/error/${id}/process`
  })
}
