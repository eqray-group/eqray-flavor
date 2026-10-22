import request from '@/utils/http'

// 分页查询桌台区域：门店范围由请求头 X-Store-Id 或当前用户可见门店集合决定
export function fetchGetTableAreaList(params: Api.Store.TableAreaSearchParams) {
  return request.get<Api.Store.TableAreaList>({
    url: '/api/v1/store/table-areas',
    params
  })
}

// 查询区域下拉选项：桌台筛选器与新增 / 编辑弹窗共用
export function fetchGetTableAreaOptions() {
  return request.get<Api.Store.TableAreaOption[]>({
    url: '/api/v1/store/table-areas/options'
  })
}

// 新增桌台区域：同一门店内区域名称唯一
export function fetchCreateTableArea(data: Api.Store.TableAreaSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/store/table-areas',
    params: data
  })
}

// 编辑桌台区域：区域不能跨门店移动
export function fetchUpdateTableArea(id: string, data: Api.Store.TableAreaSaveParams) {
  return request.put<null>({
    url: `/api/v1/store/table-areas/${id}`,
    params: data
  })
}

// 删除桌台区域：区域下仍有桌台时后端拒绝
export function fetchDeleteTableArea(id: string) {
  return request.del<null>({
    url: `/api/v1/store/table-areas/${id}`
  })
}

// 分页查询桌台：支持桌号 / 桌码、区域与状态筛选
export function fetchGetTableList(params: Api.Store.TableSearchParams) {
  return request.get<Api.Store.TableList>({
    url: '/api/v1/store/tables',
    params
  })
}

// 新增桌台：桌码留空时后端按「门店编码-桌号」自动生成
export function fetchCreateTable(data: Api.Store.TableSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/store/tables',
    params: data
  })
}

// 编辑桌台：桌码创建后不可修改，后端会忽略该字段
export function fetchUpdateTable(id: string, data: Api.Store.TableSaveParams) {
  return request.put<null>({
    url: `/api/v1/store/tables/${id}`,
    params: data
  })
}

// 删除桌台：使用中的桌台必须清台后才能删除
export function fetchDeleteTable(id: string) {
  return request.del<null>({
    url: `/api/v1/store/tables/${id}`
  })
}

// 批量生成桌台：桌号按前缀补零，单次最多 100 张
export function fetchBatchCreateTables(data: Api.Store.TableBatchParams) {
  return request.post<Api.Store.TableBatchResult>({
    url: '/api/v1/store/tables/batch',
    params: data
  })
}

// 开台：仅空闲桌台可开，订单模块未接入时 currentOrderId 传 '0'
export function fetchOpenTable(id: string, currentOrderId = '0') {
  return request.post<null>({
    url: `/api/v1/store/tables/${id}/open`,
    params: { currentOrderId }
  })
}

// 清台：使用中或待清台的桌台恢复空闲
export function fetchClearTable(id: string) {
  return request.post<null>({
    url: `/api/v1/store/tables/${id}/clear`,
    params: {}
  })
}
