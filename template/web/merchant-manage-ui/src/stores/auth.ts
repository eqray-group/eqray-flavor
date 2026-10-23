import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { fetchLogin, fetchLogout, fetchUserInfo } from '@/api/auth'
import { REFRESH_KEY, TOKEN_KEY } from '@/utils/http'

/**
 * 登录态与按钮权限。
 * 权限用 buttons 数组判断，页面里写 hasAuth('order:pos:settle') 即可。
 */
export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem(TOKEN_KEY) || '')
  const user = ref<Api.Auth.UserInfo | null>(null)

  const isLogin = computed(() => !!token.value)
  const buttons = computed(() => user.value?.buttons ?? [])
  const displayName = computed(() => user.value?.nickname || user.value?.userName || '未登录')

  /** 判断是否拥有某个按钮权限 */
  const hasAuth = (code: string): boolean => buttons.value.includes(code)

  /** 登录并落库令牌 */
  const login = async (params: Api.Auth.LoginParams): Promise<void> => {
    const res = await fetchLogin(params)
    token.value = res.accessToken
    localStorage.setItem(TOKEN_KEY, res.accessToken)
    localStorage.setItem(REFRESH_KEY, res.refreshToken)
  }

  /** 拉取当前用户信息与权限 */
  const loadUser = async (): Promise<void> => {
    user.value = await fetchUserInfo()
  }

  /** 退出登录：先清本地，服务端失败也不阻塞 */
  const logout = async (): Promise<void> => {
    try {
      await fetchLogout()
    } catch {
      // 令牌可能已过期，忽略服务端错误
    }
    token.value = ''
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
  }

  return { token, user, isLogin, buttons, displayName, hasAuth, login, loadUser, logout }
})
