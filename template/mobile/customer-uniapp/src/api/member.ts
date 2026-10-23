/**
 * 会员登录与资料。令牌与后台管理端完全隔离，不可互用。
 */
import alovaInstance from './core/instance'

export interface MemberProfile {
  id: string
  mobile: string
  nickname: string
  avatar: string
  point: number
  levelId: string
}

export interface LoginResult {
  accessToken: string
  refreshToken: string
  expiresIn?: number
  member?: MemberProfile
}

// 发送登录验证码
export function sendSmsCode(mobile: string) {
  return alovaInstance.Post<null>('/api/v1/client/member/auth/sms-code', { mobile })
}

// 手机号 + 验证码登录；手机号未注册时后端自动建会员
export function smsLogin(mobile: string, code: string) {
  return alovaInstance.Post<LoginResult>('/api/v1/client/member/auth/sms-login', { mobile, code })
}

// 当前会员
export function fetchMe() {
  return alovaInstance.Get<MemberProfile>('/api/v1/client/member/auth/me')
}
