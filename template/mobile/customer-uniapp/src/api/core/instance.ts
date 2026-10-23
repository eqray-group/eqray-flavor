import AdapterUniapp from '@alova/adapter-uniapp'
import { createAlova } from 'alova'
import vueHook from 'alova/vue'
import { TOKEN_KEY, handleAlovaError, handleAlovaResponse } from './handlers'

export const alovaInstance = createAlova({
  // 开发期由 vite 代理到本地后端，生产环境同域部署
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  ...AdapterUniapp(),
  statesHook: vueHook,
  beforeRequest: (method) => {
    if (['POST', 'PUT', 'PATCH'].includes(method.type)) {
      method.config.headers['Content-Type'] = 'application/json'
    }

    // 顾客端全部接口都要会员令牌，统一在这里注入，页面不关心
    const token = uni.getStorageSync(TOKEN_KEY) as string
    if (token) {
      method.config.headers.Authorization = `Bearer ${token}`
    }

    // 小程序与 WebView 会缓存 GET，加时间戳避免取到旧数据
    if (method.type === 'GET' && CommonUtil.isObj(method.config.params)) {
      method.config.params._t = Date.now()
    }
  },

  responded: {
    onSuccess: handleAlovaResponse,
    onError: handleAlovaError,
  },

  timeout: 20000,
  // 菜单与订单都要实时，全局关闭缓存
  cacheFor: null,
})

export default alovaInstance
