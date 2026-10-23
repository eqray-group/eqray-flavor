<!-- 充值记录详情弹窗：展示实付/赠送/到账、支付单与退款信息 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="充值记录详情"
    width="680px"
    align-center
    :close-on-click-modal="false"
  >
    <div v-loading="loading">
      <ElDescriptions v-if="detail" :column="2" border size="small">
        <ElDescriptionsItem label="钱包 ID" :span="2">{{ detail.walletId }}</ElDescriptionsItem>
        <ElDescriptionsItem label="实付（元）">{{
          formatYuan(detail.payPrice)
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="赠送（元）">{{
          formatYuan(detail.bonusPrice)
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="到账（元）">{{
          formatYuan(detail.totalPrice)
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="状态">
          <ElTag :type="detail.status === 10 ? 'success' : 'info'" effect="plain">
            {{ detail.statusName }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="支付单 ID">{{ detail.payOrderId || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="支付渠道">{{ detail.payChannelCode || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="支付时间">{{ detail.payTime || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="退款单 ID">{{ detail.refundOrderId || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="退款金额（元）">{{
          formatYuan(detail.refundPrice)
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="退款时间">{{ detail.refundTime || '-' }}</ElDescriptionsItem>
      </ElDescriptions>
    </div>

    <template #footer>
      <ElButton type="primary" @click="dialogVisible = false">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchGetPayWalletRechargeDetail } from '@/api/pay/wallet'

  defineOptions({ name: 'PayWalletRechargeDetailDialog' })

  interface Props {
    visible: boolean
    rechargeId?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    rechargeId: ''
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const loading = ref(false)
  const detail = ref<Api.Pay.WalletRechargeItem | null>(null)

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  // 打开弹窗时按主键加载充值记录
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible || !props.rechargeId) return

      loading.value = true
      detail.value = null
      try {
        detail.value = await fetchGetPayWalletRechargeDetail(props.rechargeId)
      } catch {
        // 错误提示已由请求层统一弹出
      } finally {
        loading.value = false
      }
    },
    { immediate: true }
  )
</script>
