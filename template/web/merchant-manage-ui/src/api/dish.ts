import { http } from '@/utils/http'

// 分类树：沽清页按分类筛选用，一次取整棵
export function fetchCategoryTree() {
  return http.get<Api.Dish.CategoryNode[]>('/api/v1/dish/categories/tree')
}

// 当前门店的菜品配置：未配置的规格也会列出，门店由请求头决定
export function fetchStoreSkuPage(params: Record<string, unknown>) {
  return http.get<Api.Dish.StoreSkuPage>('/api/v1/dish/store-skus', params)
}

// POS 点单菜单：按门店上下文返回可售分类/菜品/规格/加料，传 memberId 按会员价
export function fetchPosMenu(memberId?: string) {
  return http.get<Api.Dish.MenuSnapshot>('/api/v1/dish/pos/menus', memberId ? { memberId } : undefined)
}

// 保存单规格门店配置：沽清改动只改 soldOut，其余字段原样回传避免覆盖门店价
export function fetchSaveStoreSku(skuId: string, data: Record<string, unknown>) {
  return http.put<null>(`/api/v1/dish/store-skus/${skuId}`, data)
}
