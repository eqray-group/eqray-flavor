import request from '@/utils/http'

// 分页查询当前门店的菜品配置：门店由请求头 X-Store-Id 决定，未配置的规格也会列出
export function fetchGetStoreSkuList(params: Api.Dish.StoreSkuSearchParams) {
  return request.get<Api.Dish.StoreSkuList>({
    url: '/api/v1/dish/store-skus',
    params
  })
}

// 保存单规格的门店配置与门店价：门店取请求头，门店价必须在品牌价允许浮动范围内
export function fetchSaveStoreSku(skuId: string, data: Api.Dish.StoreSkuSaveParams) {
  return request.put<null>({
    url: `/api/v1/dish/store-skus/${skuId}`,
    params: data
  })
}

// 批量调整门店可售开关：用于列表批量上架 / 下架 / 沽清
export function fetchBatchStoreSku(data: Api.Dish.StoreSkuBatchParams) {
  return request.post<null>({
    url: '/api/v1/dish/store-skus/batch',
    params: data
  })
}
