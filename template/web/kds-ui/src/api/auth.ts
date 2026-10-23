import { http } from '@/utils/http'

export interface LoginParams {
  username: string
  password: string
  captchaId: string
  captchaCode: string
}

export interface LoginResponse {
  accessToken: string
  refreshToken: string
}

export interface UserInfo {
  userId: string
  userName: string
  nickname?: string
  buttons: string[]
}

export interface VisibleStore {
  id: string
  name: string
}

// 图片验证码：登录页点击刷新
export function fetchCaptcha() {
  return http.get<{ captchaId: string; image: string }>('/api/v1/system/auth/captcha')
}

// 登录：后端按查询串绑定参数
export function fetchLogin(params: LoginParams) {
  return http.post<LoginResponse>('/api/v1/system/auth/login', params as unknown as Record<string, unknown>)
}

// 当前登录者：含按钮权限
export function fetchProfile() {
  return http.get<UserInfo>('/api/v1/system/auth/profile')
}

// 当前用户可见门店：后厨只能看一个门店，多店时登录后选一次
export function fetchVisibleStores() {
  return http.get<VisibleStore[]>('/api/v1/store/stores/visible')
}
