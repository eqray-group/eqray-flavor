<!-- 退款单详情弹窗：展示退款金额、状态、渠道退款单号与回调原文 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="退款单详情"
    width="820px"
    align-center
    :close-on-click-modal="false"
  >
    <div v-loading="loading">
      <ElDescriptions v-if="detail" :column="2" border size="small">
        <ElDescriptionsItem label="退款单号" :span="2">{{ detail.no }}</ElDescriptionsItem>
        <ElDescriptionsItem label="支付单号">{{ detail.orderNo }}</ElDescriptionsItem>
        <ElDescriptionsItem label="业务退款单号">{{ detail.merchantRefundId }}</ElDescriptionsItem>
        <ElDescriptionsItem label="业务订单号">{{ detail.merchantOrderId }}</ElDescriptionsItem>
        <ElDescriptionsItem label="支付渠道">{{ detail.channelName }}</ElDescriptionsItem>
        <ElDescriptionsItem label="状态">
          <ElTag :type="statusTagType(detail.status)" effect="plain">{{ detail.statusName }}</ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="原支付金额（元）">{{
          formatYuan(detail.payPrice)
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="退款金额（元）">{{
          formatYuan(detail.refundPrice)
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="渠道退款单号">{{
          detail.channelRefundNo || '-'
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="成功时间">{{ detail.successTime || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="退款原因" :span="2">{{
          detail.reason || '-'
        }}</ElDescriptionsItem>
        <ElDescriptionsItem v-if="detail.channelErrorMsg" label="渠道报错" :span="2">
          {{ detail.channelErrorCode }} - {{ detail.channelErrorMsg }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="更新时间">{{ detail.updatedAt || '-' }}</ElDescriptionsItem>
      </ElDescriptions>

      <template v-if="detail?.channelNotifyData">
        <div class="mt-5 mb-2 font-medium">渠道退款回调原文</div>
        <ElInput
          :model-value="detail.channelNotifyData"
          type="textarea"
          :rows="6"
          readonly
          resize="none"
        />
      </template>
    </div>

    <template #footer>
      <ElButton type="primary" @click="dialogVisible = false">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchGetPayRefundDetail } from '@/api/pay/refund'

  defineOptions({ name: 'PayRefundDetailDialog' })

  interface Props {
    visible: boolean
    refundId?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    refundId: ''
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const loading = ref(false)
  const detail = ref<Api.Pay.RefundItem | null>(null)

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  // 退款状态对应的标签类型
  const statusTagType = (status: number): 'warning' | 'success' | 'danger' => {
    const map: Record<number, 'warning' | 'success' | 'danger'> = {
      0: 'warning',
      10: 'success',
      20: 'danger'
    }

    return map[status] ?? 'warning'
  }

  // 打开弹窗时按主键加载退款单详情
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible || !props.refundId) return

      loading.value = true
      detail.value = null
      try {
        detail.value = await fetchGetPayRefundDetail(props.refundId)
      } catch {
        // 错误提示已由请求层统一弹出
      } finally {
        loading.value = false
      }
    },
    { immediate: true }
  )
</script>
