import request from '@/utils/http'

/**
 * 上传文件
 * @param file 浏览器 File 对象，后端表单字段名为 file
 */
export function fetchUploadFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  return request.post<Api.Infra.UploadedFile>({
    url: '/api/v1/infra/files',
    params: formData
  })
}

/**
 * 获取文件列表
 * @param params 文件名关键字与分页参数
 */
export function fetchGetFileList(params: Api.Infra.FileSearchParams) {
  return request.get<Api.Infra.FileList>({
    url: '/api/v1/infra/files',
    params
  })
}

/**
 * 删除文件
 * @param id 文件 ID
 */
export function fetchDeleteFile(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/files/${id}`
  })
}
