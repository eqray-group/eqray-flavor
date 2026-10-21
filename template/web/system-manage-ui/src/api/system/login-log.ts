import request from '@/utils/http'

// 获取登录日志列表
export function fetchGetLoginLogList(params: Api.SystemManage.LoginLogSearchParams) {
  return request.get<Api.SystemManage.LoginLogList>({
    url: '/api/v1/system/logs/login',
    params
  })
}
