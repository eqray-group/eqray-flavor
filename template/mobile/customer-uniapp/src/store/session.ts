/**
 * 登录态：只在顾客端使用，令牌与后台管理端互不相通。
 */
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { fetchMe, smsLogin, type MemberProfile } from '@/api/member'

const TOKEN_KEY = 'customer_access_token'
const REFRESH_KEY = 'customer_refresh_token'

export const useSessionStore = defineStore('session', () => {
  const token = ref<string>((uni.getStorageSync(TOKEN_KEY) as string) || '')
  const member = ref<MemberProfile | null>(null)

  const isLogin = computed(() => !!token.value)
  /** 界面上叫号用：昵称优先，其次手机号打码 */
  const displayName = computed(() => {
    if (!member.value) return ''
    return member.value.nickname || member.value.mobile?.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2') || '会员'
  })

  const login = async (mobile: string, code: string): Promise<void> => {
    const res = await smsLogin(mobile, code)
    token.value = res.accessToken
    uni.setStorageSync(TOKEN_KEY, res.accessToken)
    uni.setStorageSync(REFRESH_KEY, res.refreshToken || '')
    member.value = res.member || null
    if (!member.value) {
      await loadMember()
    }
  }

  const loadMember = async (): Promise<void> => {
    if (!token.value) return
    try {
      member.value = await fetchMe()
    } catch {
      // 令牌失效由请求层统一处理，这里不额外弹提示
    }
  }

  const logout = (): void => {
    token.value = ''
    member.value = null
    uni.removeStorageSync(TOKEN_KEY)
    uni.removeStorageSync(REFRESH_KEY)
  }

  return { token, member, isLogin, displayName, login, loadMember, logout }
})
