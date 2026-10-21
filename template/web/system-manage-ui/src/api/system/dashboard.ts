import request from '@/utils/http'

/** 控制台聚合统计 */
export function fetchGetDashboardOverview() {
  return request.get<Api.Dashboard.Overview>({
    url: '/api/v1/system/dashboard/overview'
  })
}
