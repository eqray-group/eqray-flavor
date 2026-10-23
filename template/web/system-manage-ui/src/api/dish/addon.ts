import request from '@/utils/http'

// 分页查询加料组：支持名称关键字筛选，返回加料项数量与菜品引用次数
export function fetchGetAddonGroupList(params: Api.Dish.AddonSearchParams) {
  return request.get<Api.Dish.AddonGroupList>({
    url: '/api/v1/dish/addon-groups',
    params
  })
}

// 查询加料组详情：包含该组下全部加料项，供加料项管理弹窗使用
export function fetchGetAddonGroupDetail(id: string) {
  return request.get<Api.Dish.AddonGroupDetail>({
    url: `/api/v1/dish/addon-groups/${id}`
  })
}

// 新增加料组：名称为品牌级唯一，必选时最少选择数至少为 1
export function fetchCreateAddonGroup(data: Api.Dish.AddonGroupSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/dish/addon-groups',
    params: data
  })
}

// 编辑加料组：已被菜品引用时仍可修改，但组名不能与其它组重复
export function fetchUpdateAddonGroup(id: string, data: Api.Dish.AddonGroupSaveParams) {
  return request.put<null>({
    url: `/api/v1/dish/addon-groups/${id}`,
    params: data
  })
}

// 删除加料组：仍被菜品引用时后端拒绝
export function fetchDeleteAddonGroup(id: string) {
  return request.del<null>({
    url: `/api/v1/dish/addon-groups/${id}`
  })
}

// 在指定加料组下新增加料项：组内名称唯一，加价为 0 表示免费
export function fetchCreateAddonItem(groupId: string, data: Api.Dish.AddonItemSaveParams) {
  return request.post<{ id: string }>({
    url: `/api/v1/dish/addon-groups/${groupId}/items`,
    params: data
  })
}

// 编辑加料项：只能修改名称、加价、排序与状态，不能跨组移动
export function fetchUpdateAddonItem(id: string, data: Api.Dish.AddonItemSaveParams) {
  return request.put<null>({
    url: `/api/v1/dish/addon-items/${id}`,
    params: data
  })
}

// 删除加料项：加料项随加料组一起被菜品引用，删除后从菜品选项中移除
export function fetchDeleteAddonItem(id: string) {
  return request.del<null>({
    url: `/api/v1/dish/addon-items/${id}`
  })
}
