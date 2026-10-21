import request from '@/utils/http'

// 获取角色列表
export function fetchGetRoleList(params: Api.SystemManage.RoleSearchParams) {
  return request.get<Api.SystemManage.RoleList>({
    url: '/api/v1/system/roles',
    params
  })
}

// 获取全部启用角色（分配角色时使用）
export function fetchGetRoleOptions() {
  return request.get<Api.SystemManage.RoleItem[]>({
    url: '/api/v1/system/roles/options'
  })
}

// 获取角色详情
export function fetchGetRoleDetail(id: string) {
  return request.get<Api.SystemManage.RoleItem>({
    url: `/api/v1/system/roles/${id}`
  })
}

// 新增角色
export function fetchCreateRole(data: Api.SystemManage.RoleSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/system/roles',
    params: data
  })
}

// 编辑角色
export function fetchUpdateRole(id: string, data: Api.SystemManage.RoleSaveParams) {
  return request.put<null>({
    url: `/api/v1/system/roles/${id}`,
    params: data
  })
}

// 删除角色
export function fetchDeleteRole(id: string) {
  return request.del<null>({
    url: `/api/v1/system/roles/${id}`
  })
}

// 获取角色已授权的菜单 ID
export function fetchGetRoleMenus(id: string) {
  return request.get<string[]>({
    url: `/api/v1/system/roles/${id}/menus`
  })
}

// 角色授权（覆盖式）
export function fetchAssignRoleMenus(id: string, menuIds: string[]) {
  return request.put<null>({
    url: `/api/v1/system/roles/${id}/menus`,
    params: { menuIds }
  })
}
