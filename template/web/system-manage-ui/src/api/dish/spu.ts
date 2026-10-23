import request from '@/utils/http'

// 分页查询菜品：支持名称 / 编码关键字、分类（含下级）与上下架状态筛选
export function fetchGetSpuList(params: Api.Dish.SpuSearchParams) {
  return request.get<Api.Dish.SpuList>({
    url: '/api/v1/dish/spus',
    params
  })
}

// 查询菜品详情：返回菜品本体、规格列表与已挂载的加料组 ID
export function fetchGetSpuDetail(id: string) {
  return request.get<Api.Dish.SpuDetail>({
    url: `/api/v1/dish/spus/${id}`
  })
}

// 新增菜品：规格与加料组随菜品一起提交，至少需要一条规格
export function fetchCreateSpu(data: Api.Dish.SpuSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/dish/spus',
    params: data
  })
}

// 编辑菜品：规格整表覆盖，未提交的存量规格会被删除
export function fetchUpdateSpu(id: string, data: Api.Dish.SpuSaveParams) {
  return request.put<null>({
    url: `/api/v1/dish/spus/${id}`,
    params: data
  })
}

// 切换菜品上下架状态：列表开关直接调用
export function fetchUpdateSpuStatus(id: string, status: number) {
  return request.put<null>({
    url: `/api/v1/dish/spus/${id}/status`,
    params: { status }
  })
}

// 删除菜品：规格、门店可售与价格配置、加料组关联一并清理
export function fetchDeleteSpu(id: string) {
  return request.del<null>({
    url: `/api/v1/dish/spus/${id}`
  })
}
