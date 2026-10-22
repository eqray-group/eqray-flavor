<!-- 顶部门店切换器：只展示当前用户可见门店，切换后刷新当前页面 -->
<template>
  <ElSelect
    v-if="stores.length > 0"
    :model-value="currentStoreId"
    class="store-switcher w-44 max-lg:!hidden"
    :loading="loading"
    filterable
    @change="handleChange"
  >
    <ElOption label="全部可见门店" value="" />
    <ElOption v-for="store in stores" :key="store.id" :label="store.name" :value="store.id" />
  </ElSelect>
</template>

<script setup lang="ts">
  import { storeToRefs } from 'pinia'
  import { useStoreContextStore } from '@/store/modules/store-context'
  import { fetchGetVisibleStores } from '@/api/store/store'
  import { useCommon } from '@/hooks/core/useCommon'

  defineOptions({ name: 'StoreSwitcher' })

  const storeContext = useStoreContextStore()
  const { currentStoreId, stores } = storeToRefs(storeContext)
  const { refresh } = useCommon()

  const loading = ref(false)

  // 拉取当前用户可见门店；当前门店失效时回退到全部可见门店
  const loadStores = async (): Promise<void> => {
    loading.value = true
    try {
      const list = await fetchGetVisibleStores()
      storeContext.setStores(list)
      // 只有一家门店时直接选中，门店级页面无需再手动切换
      if (!currentStoreId.value && list.length === 1) {
        storeContext.setCurrentStore(list[0].id, list[0].name)
      }
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  // 切换门店后写入上下文并刷新当前路由页面，列表与统计按新门店重新加载
  const handleChange = (value: string): void => {
    const id = value || ''
    if (id === currentStoreId.value) return

    const current = stores.value.find((item) => item.id === id)
    storeContext.setCurrentStore(id, current?.name ?? '')
    refresh()
  }

  onMounted(loadStores)
</script>
