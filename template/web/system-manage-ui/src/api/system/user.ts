import request from '@/utils/http'

// 获取用户列表
export function fetchGetUserList(params: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: '/api/v1/system/users',
    params
  })
}

// 获取用户详情
export function fetchGetUserDetail(id: string) {
  return request.get<Api.SystemManage.UserListItem>({
    url: `/api/v1/system/users/${id}`
  })
}

// 新增用户
export function fetchCreateUser(data: Api.SystemManage.UserSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/system/users',
    params: data
  })
}

// 编辑用户
export function fetchUpdateUser(id: string, data: Api.SystemManage.UserSaveParams) {
  return request.put<null>({
    url: `/api/v1/system/users/${id}`,
    params: data
  })
}

// 删除用户
export function fetchDeleteUser(id: string) {
  return request.del<null>({
    url: `/api/v1/system/users/${id}`
  })
}

// 重置密码
export function fetchResetUserPassword(id: string, password: string) {
  return request.put<null>({
    url: `/api/v1/system/users/${id}/password`,
    params: { password }
  })
}

// 获取用户已绑定的角色 ID
export function fetchGetUserRoles(id: string) {
  return request.get<string[]>({
    url: `/api/v1/system/users/${id}/roles`
  })
}

// 给用户分配角色（覆盖式）
export function fetchAssignUserRoles(id: string, roleIds: string[]) {
  return request.put<null>({
    url: `/api/v1/system/users/${id}/roles`,
    params: { roleIds }
  })
}
