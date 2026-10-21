import request from '@/utils/http'

/**
 * 登录
 * @param params 登录参数
 * @returns 登录响应
 */
export function fetchLogin(params: Api.Auth.LoginParams) {
  return request.post<Api.Auth.LoginResponse>({
    url: '/api/v1/system/auth/login',
    params
    // showSuccessMessage: true // 显示成功消息
    // showErrorMessage: false // 不显示错误消息
  })
}

/**
 * 退出登录
 * @param refreshToken 刷新令牌
 */
export function fetchLogout(refreshToken?: string) {
  return request.post<null>({
    url: '/api/v1/system/auth/logout',
    params: { refreshToken },
    showErrorMessage: false
  })
}

/**
 * 获取图片验证码
 * @returns captchaId 与 base64 图片
 */
export function fetchCaptcha() {
  return request.get<Api.Auth.Captcha>({
    url: '/api/v1/system/auth/captcha'
  })
}

/**
 * 获取用户信息
 * @returns 用户信息
 */
export function fetchGetUserInfo() {
  return request.get<Api.Auth.UserInfo>({
    url: '/api/v1/system/profile'
  })
}

/**
 * 更新个人资料
 * @param data 昵称、邮箱、手机号、性别、头像
 */
export function fetchUpdateProfile(data: Api.Auth.ProfileUpdateParams) {
  return request.put<null>({
    url: '/api/v1/system/profile',
    params: data
  })
}

/**
 * 修改当前登录者密码
 * @param data 当前密码与新密码
 */
export function fetchUpdateProfilePassword(data: Api.Auth.ProfilePasswordParams) {
  return request.put<null>({
    url: '/api/v1/system/profile/password',
    params: data
  })
}
