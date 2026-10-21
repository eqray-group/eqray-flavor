import request from '@/utils/http'
import type { AppRouteRecord } from '@/types/router'

// 获取当前登录者的动态路由菜单（后端权限模式）
export function fetchGetMenuList() {
  return request.get<AppRouteRecord[]>({
    url: '/api/v1/system/menus/routes'
  })
}

// 获取菜单树（菜单管理页）
export function fetchGetMenuTree() {
  return request.get<Api.SystemManage.MenuItem[]>({
    url: '/api/v1/system/menus'
  })
}

// 获取菜单树（角色授权、选择上级菜单复用）
export function fetchGetMenuOptions() {
  return request.get<Api.SystemManage.MenuItem[]>({
    url: '/api/v1/system/menus/simple'
  })
}

// 获取菜单详情
export function fetchGetMenuDetail(id: string) {
  return request.get<Api.SystemManage.MenuItem>({
    url: `/api/v1/system/menus/${id}`
  })
}

// 新增菜单
export function fetchCreateMenu(data: Api.SystemManage.MenuSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/system/menus',
    params: data
  })
}

// 编辑菜单
export function fetchUpdateMenu(id: string, data: Api.SystemManage.MenuSaveParams) {
  return request.put<null>({
    url: `/api/v1/system/menus/${id}`,
    params: data
  })
}

// 删除菜单
export function fetchDeleteMenu(id: string) {
  return request.del<null>({
    url: `/api/v1/system/menus/${id}`
  })
}
