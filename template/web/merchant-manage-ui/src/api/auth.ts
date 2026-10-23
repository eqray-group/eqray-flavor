import { http, request } from '@/utils/http'

// 图片验证码：登录页刷新可换一张
export function fetchCaptcha() {
  return http.get<Api.Auth.Captcha>('/api/v1/system/auth/captcha')
}

// 登录：后端按查询串绑定参数，与管理后台保持一致
export function fetchLogin(params: Api.Auth.LoginParams) {
  return request<Api.Auth.LoginResponse>({
    url: '/api/v1/system/auth/login',
    method: 'POST',
    params
  })
}

// 当前登录者信息：含按钮权限，前端据此控制操作可见性
export function fetchUserInfo() {
  return http.get<Api.Auth.UserInfo>('/api/v1/system/auth/profile')
}

// 退出登录：令牌失效不影响本地清理
export function fetchLogout() {
  return request<null>({
    url: '/api/v1/system/auth/logout',
    method: 'POST'
  })
}
