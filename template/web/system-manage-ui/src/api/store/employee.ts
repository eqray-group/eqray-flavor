import request from '@/utils/http'

// 分页查询门店员工：门店范围由请求头 X-Store-Id 或当前用户可见门店集合决定
export function fetchGetEmployeeList(params: Api.Store.EmployeeSearchParams) {
  return request.get<Api.Store.EmployeeList>({
    url: '/api/v1/store/employees',
    params
  })
}

// 搜索可分配的后台账号：员工弹窗远程下拉使用，空关键字返回前 20 个启用账号
export function fetchSearchEmployeeUsers(keyword: string) {
  return request.get<Api.Store.EmployeeUserOption[]>({
    url: '/api/v1/store/employees/user-options',
    params: { keyword }
  })
}

// 新增门店员工：storeId、userId 必须是字符串雪花 ID
export function fetchCreateEmployee(data: Api.Store.EmployeeSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/store/employees',
    params: data
  })
}

// 编辑门店员工：允许调整门店、账号、角色、状态与备注
export function fetchUpdateEmployee(id: string, data: Api.Store.EmployeeSaveParams) {
  return request.put<null>({
    url: `/api/v1/store/employees/${id}`,
    params: data
  })
}

// 删除门店员工：关联表物理删除，后端会同步失效该账号的可见门店缓存
export function fetchDeleteEmployee(id: string) {
  return request.del<null>({
    url: `/api/v1/store/employees/${id}`
  })
}
