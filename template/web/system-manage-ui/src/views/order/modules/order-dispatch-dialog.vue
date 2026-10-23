<!-- 订单派单弹窗：只列该订单门店下在线接单中的骑手，含当前负载 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="订单派单"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <div class="mb-3 text-xs text-g-600">
      订单号：{{ props.order?.no || '-' }} · 门店配送单将占用骑手一个接单额度
    </div>

    <div v-loading="loading">
      <ElEmpty v-if="!loading && riders.length === 0" description="该门店暂无在线接单的骑手" />

      <ElRadioGroup v-else v-model="form.riderId" class="flex flex-col gap-2">
        <ElRadio
          v-for="rider in riders"
          :key="rider.riderId"
          :value="rider.riderId"
          :disabled="!rider.available"
          class="mr-0! w-full"
        >
          <div class="flex w-full items-center justify-between">
            <span>{{ rider.name || `骑手 ${rider.riderId}` }}</span>
            <span class="text-xs text-g-600">
              进行中 {{ rider.currentOrders }} / {{ rider.maxOrders }}
              <ElTag v-if="!rider.available" type="warning" effect="plain" class="ml-2">已满</ElTag>
            </span>
          </div>
        </ElRadio>
      </ElRadioGroup>
    </div>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton
        type="primary"
        :loading="submitting"
        :disabled="!form.riderId"
        @click="handleSubmit"
      >
        确认派单
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { fetchAssignOrderRider, fetchGetOrderRiders } from '@/api/order/order'

  defineOptions({ name: 'OrderDispatchDialog' })

  interface Props {
    visible: boolean
    order?: Api.Order.OrderItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    order: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const loading = ref(false)
  const submitting = ref(false)
  const riders = ref<Api.Order.OrderRiderOption[]>([])

  const form = reactive({
    riderId: ''
  })

  // 拉取该订单门店下可接单的骑手
  const loadRiders = async (): Promise<void> => {
    if (!props.order) return

    loading.value = true
    form.riderId = ''
    try {
      riders.value = (await fetchGetOrderRiders(props.order.id)) ?? []
    } catch {
      riders.value = []
    } finally {
      loading.value = false
    }
  }

  // 每次打开都重新拉取，避免骑手负载变化后选到已满的骑手
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      void loadRiders()
    },
    { immediate: true }
  )

  // 提交派单
  const handleSubmit = async (): Promise<void> => {
    if (!props.order || !form.riderId) return

    submitting.value = true
    try {
      await fetchAssignOrderRider(props.order.id, { riderId: form.riderId })
      ElMessage.success('派单成功')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
