/**
 * 响应与异常处理。
 * 后端契约：HTTP 恒为 200，业务结果放在 { code, msg, data }，code === 0 才算成功。
 * 因此判定依据必须是 body.code，而不是 HTTP 状态码。
 */
import type { Method } from 'alova'
import router from '@/router'

export class ApiError extends Error {
  code: number
  data?: unknown

  constructor(message: string, code: number, data?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.data = data
  }
}

/** 后端统一响应体 */
interface ApiEnvelope<T> {
  code: number
  msg?: string
  data?: T
}

/** 业务成功码，与后端 errs.Success 一致 */
const SUCCESS_CODE = 0
/** 未登录码，与后端鉴权中间件一致 */
const UNAUTHORIZED = 401
/** 令牌存储键，与 store/session 保持一致 */
const TOKEN_KEY = 'customer_access_token'
const REFRESH_KEY = 'customer_refresh_token'

/** 清掉登录态并回到登录页：多个页面共用，集中一处避免各写一遍 */
function toLogin(): void {
  uni.removeStorageSync(TOKEN_KEY)
  uni.removeStorageSync(REFRESH_KEY)
  const pages = getCurrentPages()
  const current = pages[pages.length - 1]?.route || ''
  if (!current.includes('login')) {
    router.replaceAll({ name: 'login' })
  }
}

export async function handleAlovaResponse(
  response: UniApp.RequestSuccessCallbackResult | UniApp.UploadFileSuccessCallbackResult | UniApp.DownloadSuccessData,
) {
  const globalToast = useGlobalToast()
  const { statusCode, data } = response as UniNamespace.RequestSuccessCallbackResult

  // 网关层直接拒绝（很少出现，出现了也不是业务错误）
  if (statusCode >= 400 && statusCode !== 401) {
    globalToast.error({ msg: `服务异常（${statusCode}）`, duration: 1500 })
    throw new ApiError(`服务异常（${statusCode}）`, statusCode, data)
  }

  const body = data as ApiEnvelope<unknown>

  // 非标准响应（例如直接返回文件流）原样交给调用方
  if (!body || typeof body.code !== 'number') {
    return data
  }

  if (body.code === SUCCESS_CODE) {
    // 关键：解包 data，页面拿到的就是业务对象，不用每次 body.data
    return body.data
  }

  if (body.code === UNAUTHORIZED) {
    globalToast.error({ msg: '登录已过期，请重新登录', duration: 1200 })
    toLogin()
  } else {
    globalToast.error({ msg: body.msg || '请求失败', duration: 1800 })
  }

  throw new ApiError(body.msg || '请求失败', body.code, body.data)
}

export function handleAlovaError(error: unknown, method: Method) {
  const globalToast = useGlobalToast()

  if (import.meta.env.MODE === 'development') {
    console.error('[Alova Error]', error, method)
  }

  if (error instanceof ApiError) {
    // 业务错误已在 handleAlovaResponse 里提示过，这里只向上抛
    throw error
  }

  const name = (error as { name?: string })?.name
  if (name === 'NetworkError' || name === 'TimeoutError') {
    globalToast.error({ msg: '网络不稳定，请稍后重试', duration: 1800 })
  } else {
    globalToast.error({ msg: '请求失败，请稍后重试', duration: 1800 })
  }

  throw error
}

export { TOKEN_KEY, REFRESH_KEY }
