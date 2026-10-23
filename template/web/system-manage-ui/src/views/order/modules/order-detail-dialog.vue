<!-- 订单详情弹窗：展示订单金额拆分、明细（含加料快照）与状态流转日志 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="订单详情"
    width="920px"
    align-center
    :close-on-click-modal="false"
  >
    <div v-loading="loading">
      <ElDescriptions v-if="detail" :column="3" border size="small">
        <ElDescriptionsItem label="订单号" :span="2">{{ detail.no }}</ElDescriptionsItem>
        <ElDescriptionsItem label="取餐码">{{ detail.pickupCode || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="订单类型">{{ detail.typeText }}</ElDescriptionsItem>
        <ElDescriptionsItem label="下单渠道">{{ detail.channelText }}</ElDescriptionsItem>
        <ElDescriptionsItem label="桌台">{{ detail.tableName || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="订单状态">
          <ElTag :type="statusTagType(detail.status)" effect="plain">
            {{ detail.statusText }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="支付状态">
          <ElTag :type="payStatusTagType(detail.payStatus)" effect="plain">
            {{ detail.payStatusText }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="收货人">
          {{ detail.receiverName || '-' }} {{ detail.receiverPhone || '' }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="收货地址" :span="3">
          {{ detail.receiverAddress || '-' }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="商品小计（元）">
          {{ formatYuan(detail.itemAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="加料费（元）">
          {{ formatYuan(detail.addonAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="单品优惠（元）">
          {{ formatYuan(detail.lineDiscountAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="活动优惠（元）">
          {{ formatYuan(detail.activityAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="优惠券（元）">
          {{ formatYuan(detail.couponAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="积分抵扣（元）">
          {{ formatYuan(detail.pointsAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="应付金额（元）">
          {{ formatYuan(detail.payAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="已退金额（元）">
          {{ formatYuan(detail.refundAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="订单备注">{{ detail.remark || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="取消原因">{{ detail.cancelReason || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="支付时间">{{ detail.payTime || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="接单时间">{{ detail.acceptTime || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="完成时间">{{ detail.completeTime || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="下单时间">{{ detail.createdAt }}</ElDescriptionsItem>
      </ElDescriptions>

      <div class="mt-5 mb-2 font-medium">订单明细</div>
      <ElTable :data="detail?.items ?? []" size="small" border max-height="320">
        <ElTableColumn prop="spuName" label="菜品" min-width="140" />
        <ElTableColumn prop="skuName" label="规格" min-width="110" />
        <ElTableColumn label="单价（元）" width="100">
          <template #default="{ row }">{{ formatYuan(row.unitPrice) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="quantity" label="数量" width="70" />
        <ElTableColumn label="加料费（元）" width="110">
          <template #default="{ row }">{{ formatYuan(row.addonAmount) }}</template>
        </ElTableColumn>
        <ElTableColumn label="分摊优惠（元）" width="120">
          <template #default="{ row }">
            {{ formatYuan(row.itemDiscountAmount) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="小计（元）" width="110">
          <template #default="{ row }">{{ formatYuan(row.subtotal) }}</template>
        </ElTableColumn>
        <ElTableColumn label="加料" min-width="160">
          <template #default="{ row }">
            <span v-if="row.addons?.length">
              {{
                row.addons
                  .map((addon: Api.Order.OrderAddonItem) => `${addon.addonName}×${addon.quantity}`)
                  .join('、')
              }}
            </span>
            <span v-else class="text-g-600">-</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="orderer" label="点菜人" width="90" />
        <ElTableColumn prop="remark" label="备注" width="120" />
      </ElTable>

      <div class="mt-5 mb-2 font-medium">状态流转日志</div>
      <ElTable :data="detail?.logs ?? []" size="small" border max-height="260">
        <ElTableColumn prop="createdAt" label="时间" width="170" />
        <ElTableColumn label="原状态" width="100">
          <template #default="{ row }">
            {{ row.fromStatus === 0 ? '创建' : statusText(row.fromStatus) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="新状态" width="100">
          <template #default="{ row }">{{ statusText(row.toStatus) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="operatorName" label="操作方" width="110" />
        <ElTableColumn prop="reason" label="原因" min-width="200" />
      </ElTable>
    </div>

    <template #footer>
      <ElButton type="primary" @click="dialogVisible = false">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchGetOrderDetail } from '@/api/order/order'

  defineOptions({ name: 'OrderDetailDialog' })

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
  const detail = ref<Api.Order.OrderDetail | null>(null)

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  // 订单状态文案，与后端 order/enums 一致
  const STATUS_TEXT: Record<number, string> = {
    100: '待支付',
    200: '待接单',
    300: '制作中',
    400: '待取餐',
    500: '已完成',
    600: '已取消',
    700: '退款中',
    800: '已退款'
  }

  const statusText = (status: number): string => STATUS_TEXT[status] ?? '未知'

  // 订单状态对应的标签类型
  const statusTagType = (status: number): 'warning' | 'success' | 'info' | 'danger' | 'primary' => {
    const map: Record<number, 'warning' | 'success' | 'info' | 'danger' | 'primary'> = {
      100: 'warning',
      200: 'warning',
      300: 'primary',
      400: 'primary',
      500: 'success',
      600: 'info',
      700: 'danger',
      800: 'info'
    }

    return map[status] ?? 'info'
  }

  // 支付状态对应的标签类型
  const payStatusTagType = (status: number): 'warning' | 'success' | 'info' => {
    const map: Record<number, 'warning' | 'success' | 'info'> = {
      0: 'warning',
      10: 'success',
      20: 'info'
    }

    return map[status] ?? 'info'
  }

  // 打开弹窗时按主键加载订单详情
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible || !props.orderId) return

      loading.value = true
      detail.value = null
      try {
        detail.value = await fetchGetOrderDetail(props.orderId)
      } catch {
        // 错误提示已由请求层统一弹出
      } finally {
        loading.value = false
      }
    }
  )
</script>
