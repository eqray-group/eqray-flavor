import request from '@/utils/http'

/**
 * 获取操作日志列表
 * @param params 模块、动作关键字、成功状态与分页参数
 */
export function fetchGetOperateLogList(params: Api.SystemManage.OperateLogSearchParams) {
  return request.get<Api.SystemManage.OperateLogList>({
    url: '/api/v1/system/logs/operate',
    params
  })
}
