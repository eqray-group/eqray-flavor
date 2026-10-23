<!-- 支付单详情弹窗：展示支付单字段与每次渠道调用的拓展记录，便于排查下单报错 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="支付单详情"
    width="880px"
    align-center
    :close-on-click-modal="false"
  >
    <div v-loading="loading">
      <ElDescriptions v-if="detail" :column="2" border size="small">
        <ElDescriptionsItem label="支付单号" :span="2">{{ detail.no }}</ElDescriptionsItem>
        <ElDescriptionsItem label="业务订单号">{{ detail.merchantOrderId }}</ElDescriptionsItem>
        <ElDescriptionsItem label="支付渠道">{{ detail.channelName }}</ElDescriptionsItem>
        <ElDescriptionsItem label="状态">
          <ElTag :type="statusTagType(detail.status)" effect="plain">
            {{ detail.statusName }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="商品标题">{{ detail.subject }}</ElDescriptionsItem>
        <ElDescriptionsItem label="支付金额（元）">{{
          formatYuan(detail.price)
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="已退金额（元）">{{
          formatYuan(detail.refundPrice)
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="渠道手续费（元）">
          {{ formatYuan(detail.channelFeePrice) }}（费率 {{ detail.channelFeeRate }}%）
        </ElDescriptionsItem>
        <ElDescriptionsItem label="渠道订单号">{{
          detail.channelOrderNo || '-'
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="渠道用户标识">{{
          detail.channelUserId || '-'
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="用户 IP">{{ detail.userIp || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="失效时间">{{ detail.expireTime || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="成功时间">{{ detail.successTime || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="创建时间">{{ detail.createdAt || '-' }}</ElDescriptionsItem>
      </ElDescriptions>

      <div class="mt-5 mb-2 font-medium">渠道调用记录</div>
      <ElTable :data="detail?.extensions ?? []" size="small" border max-height="320">
        <ElTableColumn prop="createdAt" label="调用时间" width="170" />
        <ElTableColumn prop="channelCode" label="渠道编码" width="130" />
        <ElTableColumn label="状态" width="100">
          <template #default="{ row }">
            <ElTag :type="statusTagType(row.status)" effect="plain">{{ row.status }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="渠道额外参数" min-width="220">
          <template #default="{ row }">
            <span class="text-xs">{{ row.channelExtras || '-' }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="渠道报错" min-width="220">
          <template #default="{ row }">
            <span v-if="row.channelErrorMsg" class="text-xs">
              {{ row.channelErrorCode }} - {{ row.channelErrorMsg }}
            </span>
            <span v-else class="text-xs text-g-600">-</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="回调原文" min-width="200">
          <template #default="{ row }">
            <ElTooltip
              v-if="row.channelNotifyData"
              :content="row.channelNotifyData"
              placement="top"
            >
              <span class="cursor-pointer text-xs">已收到</span>
            </ElTooltip>
            <span v-else class="text-xs text-g-600">-</span>
          </template>
        </ElTableColumn>
      </ElTable>
    </div>

    <template #footer>
      <ElButton type="primary" @click="dialogVisible = false">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchGetPayOrderDetail } from '@/api/pay/order'

  defineOptions({ name: 'PayOrderDetailDialog' })

  interface Props {
    visible: boolean
    orderId?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    orderId: ''
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const loading = ref(false)
  const detail = ref<Api.Pay.OrderDetail | null>(null)

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  // 支付状态对应的标签类型
  const statusTagType = (status: number): 'warning' | 'success' | 'info' | 'danger' => {
    const map: Record<number, 'warning' | 'success' | 'info' | 'danger'> = {
      0: 'warning',
      10: 'success',
      20: 'info',
      30: 'danger'
    }

    return map[status] ?? 'info'
  }

  // 打开弹窗时按主键加载支付单详情
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible || !props.orderId) return

      loading.value = true
      detail.value = null
      try {
        detail.value = await fetchGetPayOrderDetail(props.orderId)
      } catch {
        // 错误提示已由请求层统一弹出
      } finally {
        loading.value = false
      }
    },
    { immediate: true }
  )
</script>
