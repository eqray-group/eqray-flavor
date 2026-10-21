import request from '@/utils/http'

/** 获取文件配置列表 */
export function fetchGetFileConfigList() {
  return request.get<{ list: Api.Infra.FileConfigItem[] }>({
    url: '/api/v1/infra/file-configs'
  })
}

/** 新增文件配置 */
export function fetchCreateFileConfig(data: Api.Infra.FileConfigParams) {
  return request.post<null>({
    url: '/api/v1/infra/file-configs',
    params: data
  })
}

/** 编辑文件配置 */
export function fetchUpdateFileConfig(id: string, data: Api.Infra.FileConfigParams) {
  return request.put<null>({
    url: `/api/v1/infra/file-configs/${id}`,
    params: data
  })
}

/** 设为主配置 */
export function fetchSetFileConfigMaster(id: string) {
  return request.put<null>({
    url: `/api/v1/infra/file-configs/${id}/master`
  })
}

/** 删除文件配置 */
export function fetchDeleteFileConfig(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/file-configs/${id}`
  })
}
