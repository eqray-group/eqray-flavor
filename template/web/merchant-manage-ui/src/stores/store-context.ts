import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { fetchVisibleStores } from '@/api/store'

/** 门店上下文使用的本地键，与请求层保持一致 */
const STORE_KEY = 'merchant_store_id'

/**
 * 门店上下文：门店端所有接口都带 X-Store-Id。
 * 单店员工自动进店；多店员工需要先选店，选完写 localStorage 由请求层注入。
 */
export const useStoreContext = defineStore('storeContext', () => {
  const stores = ref<Api.Store.VisibleItem[]>([])
  const currentID = ref(localStorage.getItem(STORE_KEY) || '')
  const loading = ref(false)

  const current = computed(() => stores.value.find((item) => item.id === currentID.value) || null)
  const hasStore = computed(() => !!currentID.value)

  /** 加载可见门店；只有一个时自动选中 */
  const load = async (): Promise<void> => {
    loading.value = true
    try {
      stores.value = await fetchVisibleStores()
      const exists = stores.value.some((item) => item.id === currentID.value)
      if (!exists && stores.value.length > 0) {
        // 原来选的门店不可见（或首次登录）时，回落到第一个可见门店
        select(stores.value[0].id)
      }
    } finally {
      loading.value = false
    }
  }

  /** 切换门店：写本地并刷新页面数据由调用方负责 */
  const select = (id: string): void => {
    currentID.value = id
    localStorage.setItem(STORE_KEY, id)
  }

  return { stores, currentID, current, loading, hasStore, load, select }
})
