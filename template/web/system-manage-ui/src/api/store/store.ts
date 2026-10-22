import request from '@/utils/http'

// 分页查询门店：支持名称 / 编码关键字与类型、状态筛选
export function fetchGetStoreList(params: Api.Store.StoreSearchParams) {
  return request.get<Api.Store.StoreList>({
    url: '/api/v1/store/stores',
    params
  })
}

// 查询门店组织树：上级门店选择与列表父级名称展示共用
export function fetchGetStoreTree() {
  return request.get<Api.Store.StoreTreeNode[]>({
    url: '/api/v1/store/stores/tree'
  })
}

// 新增门店：parentId 传字符串 '0' 表示总部根节点，大区与门店必须挂上级
// 雪花 ID 统一用 string，避免 JS Number 超过安全整数后丢精度
export function fetchCreateStore(data: Api.Store.StoreSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/store/stores',
    params: data
  })
}

// 编辑门店：code 创建后不可修改，后端会忽略该字段
export function fetchUpdateStore(id: string, data: Api.Store.StoreSaveParams) {
  return request.put<null>({
    url: `/api/v1/store/stores/${id}`,
    params: data
  })
}

// 删除门店：仍有下级节点时后端会拒绝
export function fetchDeleteStore(id: string) {
  return request.del<null>({
    url: `/api/v1/store/stores/${id}`
  })
}

// 查询当前登录用户可见的门店：顶部门店切换器与员工分配弹窗共用
export function fetchGetVisibleStores() {
  return request.get<Api.Store.VisibleStore[]>({
    url: '/api/v1/store/stores/visible'
  })
}
