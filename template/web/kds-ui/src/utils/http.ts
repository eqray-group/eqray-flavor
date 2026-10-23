/**
 * 统一请求层：解析后端 { code, msg, data } 结构，自动带令牌与门店上下文。
 * 与门店端刻意使用不同的存储键：后厨机器可能和收银台共用一个浏览器。
 */
import axios, { type AxiosRequestConfig } from 'axios'

export const TOKEN_KEY = 'kds_access_token'
export const REFRESH_KEY = 'kds_refresh_token'
export const STORE_KEY = 'kds_store_id'

interface ApiResponse<T> {
  code: number
  msg: string
  data: T
}

/** 业务成功码，与后端 errs.Success 一致 */
const SUCCESS_CODE = 0

const instance = axios.create({
  timeout: 15000
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

/**
 * 轻量提示：KDS 不引 UI 库，避免后厨弱机加载几百 KB 样式。
 * @param message 提示文案
 * @param type 颜色类型
 */
export function notify(message: string, type: 'error' | 'success' = 'error'): void {
  const el = document.createElement('div')
  el.className = `kds-toast kds-toast-${type}`
  el.textContent = message
  document.body.appendChild(el)
  setTimeout(() => {
    el.classList.add('kds-toast-hide')
    setTimeout(() => el.remove(), 300)
  }, 2400)
}

/** 业务异常：携带后端错误码 */
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
  } catch {
    notify('网络异常，请检查后厨网络')

    throw new ApiError(-1, 'network error')
  }

  if (body.code !== SUCCESS_CODE) {
    if (body.code === 401) {
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(REFRESH_KEY)
      if (!location.hash.includes('/login')) {
        location.hash = '#/login'
      }
    }
    notify(body.msg || '请求失败')

    throw new ApiError(body.code, body.msg)
  }

  return body.data
}

export const http = {
  get: <T>(url: string, params?: Record<string, unknown>) => request<T>({ url, method: 'GET', params }),
  post: <T>(url: string, params?: Record<string, unknown>) => request<T>({ url, method: 'POST', params })
}
