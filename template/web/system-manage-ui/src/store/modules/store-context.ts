import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

/**
 * 门店上下文状态
 *
 * 保存顶部门店切换器选中的门店，HTTP 请求拦截器读取 currentStoreId
 * 并注入 X-Store-Id。门店列表在切换器挂载时刷新，不持久化，避免权限
 * 变更后仍展示旧门店。
 */
export const useStoreContextStore = defineStore(
  'storeContext',
  () => {
    /** 当前门店 ID，空串表示全部可见门店 */
    const currentStoreId = ref('')
    /** 当前门店名称，仅用于顶部展示 */
    const currentStoreName = ref('')
    /** 当前用户可见门店，切换器挂载后刷新 */
    const stores = ref<Api.Store.VisibleStore[]>([])

    /** 是否选中了具体门店 */
    const hasCurrentStore = computed(() => currentStoreId.value !== '')

    /** 设置可见门店列表，当前门店失效时回退到全部可见门店 */
    const setStores = (list: Api.Store.VisibleStore[]): void => {
      stores.value = list
      if (!currentStoreId.value) return

      const current = list.find((item) => item.id === currentStoreId.value)
      if (!current) {
        currentStoreId.value = ''
        currentStoreName.value = ''
        return
      }
      currentStoreName.value = current.name
    }

    /** 切换当前门店；id 为空串表示全部可见门店 */
    const setCurrentStore = (id: string, name = ''): void => {
      currentStoreId.value = id
      currentStoreName.value = id ? name : ''
    }

    /** 退出登录时清空门店上下文，避免下一个账号携带旧门店请求头 */
    const clear = (): void => {
      currentStoreId.value = ''
      currentStoreName.value = ''
      stores.value = []
    }

    return {
      currentStoreId,
      currentStoreName,
      stores,
      hasCurrentStore,
      setStores,
      setCurrentStore,
      clear
    }
  },
  {
    persist: {
      key: 'storeContext',
      storage: localStorage,
      pick: ['currentStoreId', 'currentStoreName']
    }
  }
)
