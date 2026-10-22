import request from '@/utils/http'

// 分页查询行政区划：支持名称 / 简称 / 编码关键字与级别、上级筛选
export function fetchGetDistrictList(params: Api.Infra.DistrictSearchParams) {
  return request.get<Api.Infra.DistrictList>({
    url: '/api/v1/infra/districts',
    params
  })
}

// 查询完整区划树：管理页上级选择与父级名称展示共用
export function fetchGetDistrictTree() {
  return request.get<Api.Infra.DistrictTreeNode[]>({
    url: '/api/v1/infra/districts/tree'
  })
}

// 查询指定上级的直接下级，pid 传 0 表示顶级；用于级联逐级加载
export function fetchGetDistrictOptions(pid: number) {
  return request.get<Api.Infra.DistrictItem[]>({
    url: '/api/v1/infra/districts/options',
    params: { pid }
  })
}

// 新增行政区划：层级与父级编码由后端按上级推导
export function fetchCreateDistrict(data: Api.Infra.DistrictSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/infra/districts',
    params: data
  })
}

// 编辑行政区划：pid 与 areaCode 创建后不可改，后端会忽略这两个字段
export function fetchUpdateDistrict(id: string, data: Api.Infra.DistrictSaveParams) {
  return request.put<null>({
    url: `/api/v1/infra/districts/${id}`,
    params: data
  })
}

// 删除行政区划：仍有下级时后端会拒绝
export function fetchDeleteDistrict(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/districts/${id}`
  })
}
