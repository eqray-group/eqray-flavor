/**
 * 统一请求层：解析后端 { code, msg, data } 结构，自动带令牌与门店上下文。
 * 门店端所有接口都要求 X-Store-Id，这里从 localStorage 注入，避免每个页面手写。
 */
import axios, { type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

/** 令牌与门店的本地存储键，登录与切换门店时写入 */
export const TOKEN_KEY = 'merchant_access_token'
export const REFRESH_KEY = 'merchant_refresh_token'
export const STORE_KEY = 'merchant_store_id'

/** 后端统一响应体 */
interface ApiResponse<T> {
  code: number
  msg: string
  data: T
}

/** 业务成功码，与后端 errs.Success 一致 */
const SUCCESS_CODE = 0

const instance = axios.create({
  timeout: 20000
})

instance.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  const storeID = localStorage.getItem(STORE_KEY)
  if (storeID) {
    config.headers['X-Store-Id'] = storeID
  }

  return config
})

/** 业务异常：携带后端错误码，便于页面区分「未登录」「无权限」 */
export class ApiError extends Error {
  code: number

  constructor(code: number, message: string) {
    super(message)
    this.code = code
  }
}

/**
 * 发起请求并解包 data。
 * @param config axios 配置
 */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  let body: ApiResponse<T>
  try {
    const res = await instance.request<ApiResponse<T>>(config)
    body = res.data
  } catch (error) {
    ElMessage.error('网络异常，请检查连接后重试')
    throw error
  }

  if (body.code !== SUCCESS_CODE) {
    // 未登录统一回登录页，避免每个页面都写一遍跳转
    if (body.code === 401) {
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(REFRESH_KEY)
      if (!location.hash.includes('/login')) {
        location.hash = '#/login'
      }
    }
    ElMessage.error(body.msg || '请求失败')
    throw new ApiError(body.code, body.msg)
  }

  return body.data
}

// 标量参数走查询串（与管理后台一致）；嵌套结构（如订单明细）走 JSON 请求体
export const http = {
  get: <T>(url: string, params?: Record<string, unknown>) => request<T>({ url, method: 'GET', params }),
  post: <T>(url: string, params?: Record<string, unknown>) => request<T>({ url, method: 'POST', params }),
  postBody: <T>(url: string, data?: unknown) => request<T>({ url, method: 'POST', data }),
  put: <T>(url: string, params?: Record<string, unknown>) => request<T>({ url, method: 'PUT', params }),
  del: <T>(url: string) => request<T>({ url, method: 'DELETE' })
}
