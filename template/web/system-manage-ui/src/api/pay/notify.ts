import request from '@/utils/http'

// 分页查询业务回调通知任务：支付 / 退款成功时由系统自动登记
export function fetchGetPayNotifyTaskList(params: Api.Pay.NotifyTaskSearchParams) {
  return request.get<Api.Pay.NotifyTaskList>({
    url: '/api/v1/pay/notify-tasks',
    params
  })
}

// 查询某条通知任务的派发日志：含业务方响应内容，用于排查回调未生效
export function fetchGetPayNotifyTaskLogs(
  id: string,
  params?: { pageNo?: number; pageSize?: number }
) {
  return request.get<Api.Pay.NotifyLogList>({
    url: `/api/v1/pay/notify-tasks/${id}/logs`,
    params
  })
}
