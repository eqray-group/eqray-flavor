import request from '@/utils/http'

// 获取字典类型列表
export function fetchGetDictTypeList(params: Api.SystemManage.DictTypeSearchParams) {
  return request.get<Api.SystemManage.DictTypeList>({
    url: '/api/v1/system/dict-types',
    params
  })
}

// 获取全部启用字典类型（字典数据弹窗选择所属类型）
export function fetchGetDictTypeOptions() {
  return request.get<Api.SystemManage.DictTypeItem[]>({
    url: '/api/v1/system/dict-types/options'
  })
}

// 获取字典类型详情
export function fetchGetDictTypeDetail(id: string) {
  return request.get<Api.SystemManage.DictTypeItem>({
    url: `/api/v1/system/dict-types/${id}`
  })
}

// 新增字典类型
export function fetchCreateDictType(data: Api.SystemManage.DictTypeSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/system/dict-types',
    params: data
  })
}

// 编辑字典类型
export function fetchUpdateDictType(id: string, data: Api.SystemManage.DictTypeSaveParams) {
  return request.put<null>({
    url: `/api/v1/system/dict-types/${id}`,
    params: data
  })
}

// 删除字典类型（连同其下字典数据）
export function fetchDeleteDictType(id: string) {
  return request.del<null>({
    url: `/api/v1/system/dict-types/${id}`
  })
}

// 获取字典数据列表
export function fetchGetDictDataList(params: Api.SystemManage.DictDataSearchParams) {
  return request.get<Api.SystemManage.DictDataList>({
    url: '/api/v1/system/dict-data',
    params
  })
}

// 获取字典数据详情
export function fetchGetDictDataDetail(id: string) {
  return request.get<Api.SystemManage.DictDataItem>({
    url: `/api/v1/system/dict-data/${id}`
  })
}

// 新增字典数据
export function fetchCreateDictData(data: Api.SystemManage.DictDataSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/system/dict-data',
    params: data
  })
}

// 编辑字典数据
export function fetchUpdateDictData(id: string, data: Api.SystemManage.DictDataSaveParams) {
  return request.put<null>({
    url: `/api/v1/system/dict-data/${id}`,
    params: data
  })
}

// 删除字典数据
export function fetchDeleteDictData(id: string) {
  return request.del<null>({
    url: `/api/v1/system/dict-data/${id}`
  })
}

// 获取某个字典类型的启用项（下拉 / 标签渲染）
export function fetchGetDictItems(type: string) {
  return request.get<Api.SystemManage.DictItem[]>({
    url: `/api/v1/system/dicts/${type}`
  })
}
