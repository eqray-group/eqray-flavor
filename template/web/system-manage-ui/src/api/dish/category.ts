import request from '@/utils/http'

// 查询菜品分类树：一次返回整棵树，节点带直属菜品数，分类管理与菜品表单共用
export function fetchGetCategoryTree() {
  return request.get<Api.Dish.CategoryNode[]>({
    url: '/api/v1/dish/categories/tree'
  })
}

// 新增菜品分类：同一上级下名称唯一，parentId 传 '0' 表示根分类
export function fetchCreateCategory(data: Api.Dish.CategorySaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/dish/categories',
    params: data
  })
}

// 编辑菜品分类：上级不能是自身或自身下级，同级不能重名
export function fetchUpdateCategory(id: string, data: Api.Dish.CategorySaveParams) {
  return request.put<null>({
    url: `/api/v1/dish/categories/${id}`,
    params: data
  })
}

// 删除菜品分类：仍有子分类或菜品时后端拒绝
export function fetchDeleteCategory(id: string) {
  return request.del<null>({
    url: `/api/v1/dish/categories/${id}`
  })
}
